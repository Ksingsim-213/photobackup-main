package main

import (
	"errors"
	"strings"
)

type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count" gorm:"column:cnt"`
}

// GetTagCloud คืนแท็กพร้อมจำนวนภาพ (เฉพาะปลายทางที่ระบุ) สำหรับทำ Tag Cloud
func (a *App) GetTagCloud(dest string) ([]TagCount, error) {
	dest = normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	out := []TagCount{}
	err := a.db.Table("tags").
		Select("tags.name AS name, COUNT(image_tags.image_id) AS cnt").
		Joins("JOIN image_tags ON image_tags.tag_id = tags.id").
		Joins("JOIN images ON images.id = image_tags.image_id").
		Where("images.dest_dir = ?", dest).
		Group("tags.id").Order("cnt DESC, tags.name").Limit(120).
		Scan(&out).Error
	return out, err
}

func likeEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// SearchImages ค้นหาจากคำอธิบาย (LIKE) และ/หรือแท็ก
// matchAll=true: ต้องมีครบทุกแท็กที่เลือก, false: มีอย่างน้อยหนึ่งแท็ก
func (a *App) SearchImages(dest, caption string, tags []string, matchAll bool) ([]Image, error) {
	dest = normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	q := a.db.Model(&Image{}).Where("images.dest_dir = ?", dest)
	if caption = strings.TrimSpace(caption); caption != "" {
		q = q.Where("images.caption LIKE ? ESCAPE '\\'", "%"+likeEscape(caption)+"%")
	}
	uniq := map[string]bool{}
	var names []string
	for _, t := range tags {
		if t = strings.TrimSpace(t); t != "" && !uniq[t] {
			uniq[t] = true
			names = append(names, t)
		}
	}
	if len(names) > 0 {
		sub := a.db.Table("image_tags").Select("image_tags.image_id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id").
			Where("tags.name IN ?", names).Group("image_tags.image_id")
		if matchAll {
			sub = sub.Having("COUNT(DISTINCT tags.id) = ?", len(names))
		}
		q = q.Where("images.id IN (?)", sub)
	}
	var imgs []Image
	err := q.Preload("Tags").Order("images.backed_up_at DESC, images.file_name").Limit(300).Find(&imgs).Error
	return imgs, err
}
