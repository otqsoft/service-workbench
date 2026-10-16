package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"service-workbench/internal/config"
	"service-workbench/internal/manager"
	"service-workbench/internal/model"
)

// App struct
type App struct {
	ctx     context.Context
	store   *config.Store
	mgr     *manager.Manager
	exeDir  string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.mgr.SetContext(ctx)

	// 启动时自动恢复之前运行的服务？第一版先不自动启动，手动来
}

// ------- 配置 CRUD -------

// ListServices 返回所有服务配置 + 运行状态
func (a *App) ListServices() []model.ServiceInfo {
	cfgs := a.store.List()
	return a.mgr.GetStatus(cfgs)
}

// AddService 新增/更新服务配置
func (a *App) AddService(cfg model.ServiceConfig) error {
	if cfg.ID == "" {
		cfg.ID = uuid.NewString()
	}
	return a.store.Upsert(cfg)
}

// DeleteService 删除服务
func (a *App) DeleteService(id string) error {
	// 先停掉
	_ = a.mgr.Stop(id)
	return a.store.Delete(id)
}

// ------- 运行控制 -------

// StartService 启动单个服务
func (a *App) StartService(id string) error {
	cfg, ok := a.store.Get(id)
	if !ok {
		return nil
	}
	return a.mgr.Start(cfg)
}

// StopService 停止单个服务
func (a *App) StopService(id string) error {
	return a.mgr.Stop(id)
}

// StartAll 按顺序启动所有
func (a *App) StartAll() error {
	cfgs := a.store.List()
	// 按 StartOrder 排序
	for i := 0; i < len(cfgs); i++ {
		for j := i + 1; j < len(cfgs); j++ {
			if cfgs[j].StartOrder < cfgs[i].StartOrder {
				cfgs[i], cfgs[j] = cfgs[j], cfgs[i]
			}
		}
	}
	for _, cfg := range cfgs {
		_ = a.mgr.Start(cfg)
	}
	return nil
}

// StopAll 停止所有
func (a *App) StopAll() {
	a.mgr.StopAll()
}

// GetLogs 获取某服务历史日志
func (a *App) GetLogs(id string) []string {
	return a.mgr.GetLogs(id)
}

// OpenFileDialog 选 exe 文件
func (a *App) OpenFileDialog(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
		Filters: []runtime.FileFilter{
			{DisplayName: "程序 / 脚本", Pattern: "*.exe;*.py;*.bat;*.cmd;*.ps1"},
			{DisplayName: "所有文件", Pattern: "*.*"},
		},
	})
}

// OpenDirectoryDialog 选目录
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

// exeDir 返回 exe 所在目录（配置文件存这里）
func (a *App) ExeDir() string {
	return a.exeDir
}

func initExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
