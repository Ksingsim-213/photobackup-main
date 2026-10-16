# Photo Backup (Wails + Nuxt + GORM/SQLite + Claude Vision)

## เตรียมเครื่อง
- Go 1.22+, Node 20+, Wails v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- ตรวจด้วย `wails doctor`
- ใช้ `github.com/glebarez/sqlite` (SQLite แบบ pure Go) จึงไม่ต้องติดตั้ง gcc/cgo บน Windows

## รัน
```
go mod tidy
wails dev          # โหมดพัฒนา
wails build        # สร้างไฟล์ที่ build/bin  (build บน Windows ได้ .exe, บน macOS ได้ .app)
```

## ตั้งค่า Claude API key
ใส่ที่แถบซ้าย "Anthropic API key" (เก็บในฐานข้อมูลของแอป) หรือตั้ง env `ANTHROPIC_API_KEY`
เปลี่ยนโมเดลได้ที่ค่าคงที่ `claudeModel` ใน `analyze.go`

## ข้อมูลของแอป
`<UserConfigDir>/PhotoBackup/backup.db` และโฟลเดอร์ `thumbs/` (cache รูปย่อ)

## ความสามารถเทียบกับโจทย์
| ข้อกำหนด | ไฟล์ / หน้าจอ |
|---|---|
| Wails + Nuxt | `main.go`, `frontend/` (Nuxt SPA, `ssr: false`) |
| SQLite + GORM | `app.go` (เปิด DB + AutoMigrate), `models.go` |
| ย้ายไฟล์ / กันซ้ำ (โฟลเดอร์ปลายทาง + DB) | `move.go` → `MoveImages`, `knownNames` |
| Move แล้ว fallback เป็น Copy + Delete | `move.go` → `transfer` |
| จับเวลาทั้งหมด | `MoveResult.Elapsed` แสดงที่หน้า "สำรองรูปภาพ" |
| ตรวจความถูกต้อง DB vs ไฟล์จริง (Missing Files) | `check.go` → `CheckIntegrity`, หน้า "ประวัติและการตรวจสอบ" |
| ประวัติแยกตามปลายทาง (`/list db`) | `check.go` → `ListDB` |
| วิเคราะห์ภาพ (คำอธิบาย + Tag) | `analyze.go` (Claude Vision) |
| ค้นหาด้วยคำอธิบาย / Tag + Tag Cloud | `search.go`, หน้า "ค้นหาภาพ" |

## ข้อจำกัดที่ควรรู้
- ไฟล์ HEIC/HEIF (รูปจาก iPhone) สำรองได้ แต่ย่อภาพ/วิเคราะห์ไม่ได้ เพราะ Go ไม่มี decoder ในตัว
- การกันซ้ำเทียบจาก "ชื่อไฟล์" (ไม่สนตัวพิมพ์เล็ก/ใหญ่) ตามโจทย์ ไม่ได้เทียบเนื้อไฟล์
