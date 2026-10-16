package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"service-workbench/internal/model"
)

// Store 服务配置存储
type Store struct {
	mu    sync.RWMutex
	path  string
	items []model.ServiceConfig
}

// NewStore 创建配置存储，配置文件存在 exe 同目录
func NewStore(exeDir string) *Store {
	p := filepath.Join(exeDir, "services.json")
	s := &Store{path: p}
	s.load()
	return s
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		s.items = []model.ServiceConfig{}
		return
	}
	_ = json.Unmarshal(data, &s.items)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// List 返回所有配置
func (s *Store) List() []model.ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.ServiceConfig, len(s.items))
	copy(out, s.items)
	return out
}

// Upsert 新增或更新
func (s *Store) Upsert(cfg model.ServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.items {
		if c.ID == cfg.ID {
			s.items[i] = cfg
			return s.save()
		}
	}
	s.items = append(s.items, cfg)
	return s.save()
}

// Delete 按 ID 删除
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.items {
		if c.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return s.save()
		}
	}
	return nil
}

// Get 按 ID 查
func (s *Store) Get(id string) (model.ServiceConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.items {
		if c.ID == id {
			return c, true
		}
	}
	return model.ServiceConfig{}, false
}
