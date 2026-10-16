package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type CheckResult struct {
	Dest      string  `json:"dest"`
	Total     int     `json:"total"`
	Present   int     `json:"present"`
	Missing   []Image `json:"missing"`
	ElapsedMs int64   `json:"elapsedMs"`
}

// SetDestination ตั้งโฟลเดอร์ปลายทาง แล้วตรวจความถูกต้อง (DB vs ไฟล์จริง) ทันที
func (a *App) SetDestination(dest string) (*CheckResult, error) {
	dest = normDir(dest)
	if err := mustBeDir(dest); err != nil {
		return nil, fmt.Errorf("โฟลเดอร์ปลายทาง: %w", err)
	}
	_ = a.setSetting("lastDest", dest)
	return a.CheckIntegrity(dest)
}

// CheckIntegrity อ่านรายการไฟล์ในโฟลเดอร์ครั้งเดียว แล้วเทียบกับประวัติใน DB ของปลายทางนั้น
func (a *App) CheckIntegrity(dest string) (*CheckResult, error) {
	dest = normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	start := time.Now()
	entries, err := os.ReadDir(dest)
	if err != nil {
		return nil, err
	}
	onDisk := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			onDisk[strings.ToLower(e.Name())] = struct{}{}
		}
	}
	var imgs []Image
	if err := a.db.Where("dest_dir = ?", dest).Find(&imgs).Error; err != nil {
		return nil, err
	}
	res := &CheckResult{Dest: dest, Total: len(imgs), Missing: []Image{}}
	for _, im := range imgs {
		if _, ok := onDisk[strings.ToLower(im.FileName)]; ok {
			res.Present++
		} else {
			res.Missing = append(res.Missing, im)
		}
	}
	res.ElapsedMs = time.Since(start).Milliseconds()
	return res, nil
}

// ListDB = /list db <path>: แสดงเฉพาะประวัติของโฟลเดอร์ปลายทางที่ระบุ
func (a *App) ListDB(dest string) ([]Image, error) {
	dest = normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	var imgs []Image
	err := a.db.Preload("Tags").Where("dest_dir = ?", dest).Order("backed_up_at DESC, file_name").Find(&imgs).Error
	return imgs, err
}
