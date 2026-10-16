package manager

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"service-workbench/internal/model"
)

const maxLogBuffer = 2000 // 每个服务内存保留最近日志行数

type runningService struct {
	cfg      model.ServiceConfig
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	status   model.ServiceStatus
	pid      int
	startAt  int64
	exitCode int
	logBuf   []string
	mu       sync.Mutex
}

// Manager 服务管理器
type Manager struct {
	ctx      context.Context
	mu       sync.RWMutex
	services map[string]*runningService
}

func New() *Manager {
	return &Manager{services: map[string]*runningService{}}
}

func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// Start 启动一个服务
func (m *Manager) Start(cfg model.ServiceConfig) error {
	m.mu.Lock()
	if _, ok := m.services[cfg.ID]; ok {
		m.mu.Unlock()
		return nil // 已存在，可能在跑
	}
	rs := &runningService{cfg: cfg, status: model.StatusStarting, logBuf: []string{}}
	m.services[cfg.ID] = rs
	m.mu.Unlock()

	go m.run(rs)
	return nil
}

func (m *Manager) run(rs *runningService) {
	// 构建命令
	_, cancel := context.WithCancel(context.Background())
	rs.cancel = cancel

	// 根据文件类型选择解释器
	name := rs.cfg.ExePath
	args := rs.cfg.Args
	lower := strings.ToLower(name)

	var cmd *exec.Cmd
	switch {
	case strings.HasSuffix(lower, ".py"):
		// Python 脚本：用 python.exe 启动
		pythonExe, err := exec.LookPath("python")
		if err != nil {
			pythonExe = "python"
		}
		allArgs := append([]string{name}, args...)
		cmd = exec.Command(pythonExe, allArgs...)
	case strings.HasSuffix(lower, ".bat") || strings.HasSuffix(lower, ".cmd"):
		// 批处理：用 cmd.exe /c 启动
		allArgs := append([]string{"/c", name}, args...)
		cmd = exec.Command("cmd.exe", allArgs...)
	default:
		// 普通 exe 直接启动
		cmd = exec.Command(name, args...)
	}
	rs.cmd = cmd

	// 工作目录
	if rs.cfg.WorkDir != "" {
		cmd.Dir = rs.cfg.WorkDir
	} else {
		cmd.Dir = filepath.Dir(rs.cfg.ExePath)
	}

	// 环境变量：继承父进程 + 自定义
	env := os.Environ()
	for k, v := range rs.cfg.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	// Windows 下新建进程组、不弹黑窗
	// CREATE_NEW_PROCESS_GROUP = 0x00000200
	// CREATE_NO_WINDOW = 0x08000000
	cmd.SysProcAttr = sysProcAttrWindows(0x00000200 | 0x08000000)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.emitStatus(rs, model.StatusError, -1)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.emitStatus(rs, model.StatusError, -1)
		return
	}

	if err := cmd.Start(); err != nil {
		m.pushLog(rs, "启动失败: "+err.Error(), true)
		m.emitStatus(rs, model.StatusError, -1)
		return
	}

	rs.pid = cmd.Process.Pid
	rs.startAt = time.Now().Unix()
	m.emitStatus(rs, model.StatusRunning, 0)

	// 读日志
	stdoutReader := wrapReaderByEncoding(stdout, rs.cfg.LogEncoding)
	stderrReader := wrapReaderByEncoding(stderr, rs.cfg.LogEncoding)
	go m.readLog(rs, stdoutReader, false)
	go m.readLog(rs, stderrReader, true)

	// 等待结束
	err = cmd.Wait()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = -1
		}
	}
	rs.exitCode = code
	m.pushLog(rs, "进程已退出，退出码: "+itoa(code), code != 0)

	// 清理
	m.mu.Lock()
	delete(m.services, rs.cfg.ID)
	m.mu.Unlock()

	m.emitStatus(rs, model.StatusStopped, code)

	// TODO: 自动重启逻辑后续加
}

func (m *Manager) readLog(rs *runningService, r io.Reader, isErr bool) {
	scanner := bufio.NewScanner(r)
	// 增大 buffer，防止超长行截断
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		m.pushLog(rs, line, isErr)
	}
}

// wrapReaderByEncoding 根据编码配置包装 reader，GBK 自动转 UTF-8
func wrapReaderByEncoding(r io.Reader, enc string) io.Reader {
	if strings.ToLower(enc) == "gbk" || strings.ToLower(enc) == "cp936" {
		return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder())
	}
	return r
}

func (m *Manager) pushLog(rs *runningService, line string, isErr bool) {
	rs.mu.Lock()
	rs.logBuf = append(rs.logBuf, line)
	if len(rs.logBuf) > maxLogBuffer {
		rs.logBuf = rs.logBuf[len(rs.logBuf)-maxLogBuffer:]
	}
	rs.mu.Unlock()

	entry := model.LogEntry{
		ServiceID: rs.cfg.ID,
		Line:      line,
		IsErr:     isErr,
		Timestamp: time.Now().UnixMilli(),
	}
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "log", entry)
	}
}

// Stop 停止服务
func (m *Manager) Stop(id string) error {
	m.mu.RLock()
	rs, ok := m.services[id]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	m.emitStatus(rs, model.StatusStopping, 0)
	// 第一版：直接强杀
	if rs.cmd != nil && rs.cmd.Process != nil {
		_ = rs.cmd.Process.Kill()
	}
	if rs.cancel != nil {
		rs.cancel()
	}
	return nil
}

// StopAll 停止所有
func (m *Manager) StopAll() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.services))
	for id := range m.services {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	for _, id := range ids {
		_ = m.Stop(id)
	}
}

// GetStatus 返回所有服务当前状态
func (m *Manager) GetStatus(configs []model.ServiceConfig) []model.ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.ServiceInfo, 0, len(configs))
	for _, cfg := range configs {
		info := model.ServiceInfo{
			Config: cfg,
			Status: model.StatusStopped,
		}
		if rs, ok := m.services[cfg.ID]; ok {
			info.Status = rs.status
			info.PID = rs.pid
			info.StartTime = rs.startAt
			info.ExitCode = rs.exitCode
			rs.mu.Lock()
			if n := len(rs.logBuf); n > 0 {
				info.LastLog = rs.logBuf[n-1]
			}
			rs.mu.Unlock()
		}
		out = append(out, info)
	}
	return out
}

// GetLogs 返回某服务历史日志
func (m *Manager) GetLogs(id string) []string {
	m.mu.RLock()
	rs, ok := m.services[id]
	m.mu.RUnlock()
	if !ok {
		return []string{}
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	out := make([]string, len(rs.logBuf))
	copy(out, rs.logBuf)
	return out
}

func (m *Manager) emitStatus(rs *runningService, st model.ServiceStatus, code int) {
	rs.status = st
	info := model.ServiceInfo{
		Config:   rs.cfg,
		Status:   st,
		PID:      rs.pid,
		StartTime: rs.startAt,
		ExitCode: code,
	}
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "status", info)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
