package main

import "time"

// Image คือประวัติไฟล์ที่ถูกสำรองไว้ในโฟลเดอร์ปลายทางหนึ่ง ๆ
// (DestDir + FileName ต้องไม่ซ้ำกัน เพื่อป้องกันการบันทึกซ้ำ)
type Image struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DestDir    string    `gorm:"uniqueIndex:idx_dest_name;not null" json:"destDir"`
	FileName   string    `gorm:"uniqueIndex:idx_dest_name;not null" json:"fileName"`
	Path       string    `json:"path"`
	SourcePath string    `json:"sourcePath"`
	Size       int64     `json:"size"`
	BackedUpAt time.Time `json:"backedUpAt"`
	Caption    string    `json:"caption"`
	Analyzed   bool      `gorm:"index" json:"analyzed"`
	Tags       []Tag     `gorm:"many2many:image_tags;" json:"tags"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}
