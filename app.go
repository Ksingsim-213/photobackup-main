package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type App struct {
	ctx     context.Context
	db      *gorm.DB
	dataDir string
	thumbSm chan struct{} // จำกัดจำนวนงานย่อรูปพร้อมกัน
}

func NewApp() *App { return &App{thumbSm: make(chan struct{}, 4)} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, err := os.UserConfigDir()
	if err != nil {
		runtime.LogFatalf(ctx, "config dir: %v", err)
	}
	a.dataDir = filepath.Join(cfg, "PhotoBackup")
	if err := os.MkdirAll(filepath.Join(a.dataDir, "thumbs"), 0o755); err != nil {
		runtime.LogFatalf(ctx, "mkdir: %v", err)
	}
	dsn := filepath.Join(a.dataDir, "backup.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	a.db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		runtime.LogFatalf(ctx, "open db: %v", err)
	}
	if err := a.db.AutoMigrate(&Image{}, &Tag{}, &Setting{}); err != nil {
		runtime.LogFatalf(ctx, "migrate: %v", err)
	}
}

// normDir ทำให้ path ปลายทางอยู่ในรูปเดียวกันเสมอ (กันกรณี / ท้ายชื่อ หรือ path ซ้อน)
// เพื่อให้ DestDir ในฐานข้อมูลตรงกันทุกครั้งที่เลือกโฟลเดอร์เดิม
func normDir(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	return filepath.Clean(p)
}

// PickFolder เปิด dialog เลือกโฟลเดอร์
func (a *App) PickFolder(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

type Settings struct {
	HasAPIKey bool   `json:"hasApiKey"`
	LastDest  string `json:"lastDest"`
}

func (a *App) GetSettings() Settings {
	return Settings{HasAPIKey: a.apiKey() != "", LastDest: a.getSetting("lastDest")}
}

func (a *App) SaveAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("API key ว่างเปล่า")
	}
	return a.setSetting("geminiApiKey", key)
}

func (a *App) apiKey() string {
	if k := a.getSetting("geminiApiKey"); k != "" {
		return k
	}
	return os.Getenv("GEMINI_API_KEY")
}

func (a *App) getSetting(k string) string {
	var s Setting
	if err := a.db.First(&s, "key = ?", k).Error; err != nil {
		return ""
	}
	return s.Value
}

func (a *App) setSetting(k, v string) error {
	return a.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&Setting{Key: k, Value: v}).Error
}
