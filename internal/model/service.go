package model

// ServiceStatus 服务运行状态
type ServiceStatus string

const (
	StatusStopped  ServiceStatus = "stopped"
	StatusStarting ServiceStatus = "starting"
	StatusRunning  ServiceStatus = "running"
	StatusStopping ServiceStatus = "stopping"
	StatusError    ServiceStatus = "error"
)

// ServiceConfig 单个服务的配置信息
type ServiceConfig struct {
	ID          string            `json:"id"`          // 唯一标识，uuid
	Name        string            `json:"name"`        // 显示名称
	ExePath     string            `json:"exePath"`     // 可执行文件路径（exe / py / bat 均可）
	WorkDir     string            `json:"workDir"`     // 工作目录
	Args        []string          `json:"args"`        // 命令行参数
	Env         map[string]string `json:"env"`         // 额外环境变量
	LogFilePath string            `json:"logFilePath"` // 磁盘日志文件路径（可选，用于 tail）
	LogEncoding string            `json:"logEncoding"` // 日志编码：utf-8 / gbk
	AutoRestart bool              `json:"autoRestart"` // 崩溃后自动重启
	StartOrder  int               `json:"startOrder"`  // 启动顺序，数字小的先起
	StartDelay  int               `json:"startDelay"`  // 启动后等待秒数再判定成功
}

// ServiceInfo 服务运行时状态信息（推给前端）
type ServiceInfo struct {
	Config    ServiceConfig `json:"config"`
	Status    ServiceStatus `json:"status"`
	PID       int           `json:"pid"`
	StartTime int64         `json:"startTime"` // unix 秒
	ExitCode  int           `json:"exitCode"`
	LastLog   string        `json:"lastLog"`
}

// LogEntry 一条日志
type LogEntry struct {
	ServiceID string `json:"serviceId"`
	Line      string `json:"line"`
	IsErr     bool   `json:"isErr"` // stderr
	Timestamp int64  `json:"timestamp"`
}
