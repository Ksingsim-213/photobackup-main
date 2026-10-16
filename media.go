package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// loadResized โหลดรูป (หมุนตาม EXIF) ย่อให้ไม่เกิน maxDim แล้วเข้ารหัสเป็น JPEG
func loadResized(path string, maxDim int) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if !decodableExts[ext] {
		return nil, errors.New("ไม่รองรับรูปแบบไฟล์ " + ext)
	}
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	if b := img.Bounds(); b.Dx() > maxDim || b.Dy() > maxDim {
		img = imaging.Fit(img, maxDim, maxDim, imaging.Lanczos)
	}
	if ext != ".jpg" && ext != ".jpeg" { // เติมพื้นขาวให้รูปโปร่งใส
		bg := imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)
		img = imaging.OverlayCenter(bg, img, 1)
	}
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(82)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (a *App) imageByID(id uint) (*Image, error) {
	var im Image
	if err := a.db.First(&im, id).Error; err != nil {
		return nil, err
	}
	return &im, nil
}

// GetThumbnail คืน data URL ของรูปย่อ (cache ไว้ในโฟลเดอร์ข้อมูลของแอป)
func (a *App) GetThumbnail(id uint) (string, error) {
	cache := filepath.Join(a.dataDir, "thumbs", strconv.Itoa(int(id))+".jpg")
	if data, err := os.ReadFile(cache); err == nil {
		return dataURL(data), nil
	}
	im, err := a.imageByID(id)
	if err != nil {
		return "", err
	}
	a.thumbSm <- struct{}{}
	defer func() { <-a.thumbSm }()
	data, err := loadResized(im.Path, 360)
	if err != nil {
		return "", err
	}
	_ = os.WriteFile(cache, data, 0o644)
	return dataURL(data), nil
}

// GetPreview คืนรูปขนาดใหญ่ขึ้น (ไม่เกิน 1600px) สำหรับดูในหน้าต่างขยาย
func (a *App) GetPreview(id uint) (string, error) {
	im, err := a.imageByID(id)
	if err != nil {
		return "", err
	}
	data, err := loadResized(im.Path, 1600)
	if err != nil {
		return "", err
	}
	return dataURL(data), nil
}

func dataURL(b []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(b)
}
