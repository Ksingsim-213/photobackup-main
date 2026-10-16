package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var backupExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".heic": true, ".heif": true}
var decodableExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

type MoveResult struct {
	Found     int      `json:"found"`
	Moved     int      `json:"moved"`
	Skipped   int      `json:"skipped"`
	Failed    int      `json:"failed"`
	Fallbacks int      `json:"fallbacks"` // จำนวนไฟล์ที่ Move ข้ามไดรฟ์ไม่ได้ จึงใช้ Copy + Delete
	Warnings  []string `json:"warnings"`
	Errors    []string `json:"errors"`
	ElapsedMs int64    `json:"elapsedMs"`
	Elapsed   string   `json:"elapsed"`
}

type moveJob struct{ src, dst, name string }
type moveOut struct {
	job      moveJob
	size     int64
	fallback bool
	warn     string
	err      error
}

func fmtElapsed(d time.Duration) string {
	return fmt.Sprintf("%d ms (%.2f วินาที)", d.Milliseconds(), d.Seconds())
}

func mustBeDir(p string) error {
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("%s ไม่ใช่โฟลเดอร์", p)
	}
	return nil
}

// MoveImages ย้าย (หรือคัดลอกถ้า deleteSource=false) รูปจาก source ไป dest
// - อ่านรายชื่อไฟล์ปลายทางและฐานข้อมูลครั้งเดียว แล้วตรวจชื่อซ้ำใน memory
// - คัดลอกแบบขนาน 4 workers, บันทึกฐานข้อมูลแบบ batch ครั้งเดียวตอนท้าย
func (a *App) MoveImages(source, dest string, deleteSource bool) (*MoveResult, error) {
	start := time.Now()
	source, dest = normDir(source), normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	if err := mustBeDir(source); err != nil {
		return nil, fmt.Errorf("โฟลเดอร์ต้นทาง: %w", err)
	}
	if err := mustBeDir(dest); err != nil {
		return nil, fmt.Errorf("โฟลเดอร์ปลายทาง: %w", err)
	}
	sa, _ := filepath.Abs(source)
	da, _ := filepath.Abs(dest)
	if sa == da {
		return nil, errors.New("ต้นทางและปลายทางต้องเป็นคนละโฟลเดอร์")
	}

	res := &MoveResult{Warnings: []string{}, Errors: []string{}}
	known, err := a.knownNames(dest)
	if err != nil {
		return nil, err
	}
	files, err := scanImages(source, dest)
	if err != nil {
		return nil, err
	}
	res.Found = len(files)

	jobs := make([]moveJob, 0, len(files))
	for _, f := range files {
		name := filepath.Base(f)
		key := strings.ToLower(name)
		if where, dup := known[key]; dup {
			res.Skipped++
			res.Warnings = append(res.Warnings, fmt.Sprintf("ข้าม %s: พบชื่อนี้%sแล้ว", name, where))
			continue
		}
		known[key] = "ซ้ำกันในชุดที่กำลังย้าย"
		jobs = append(jobs, moveJob{src: f, dst: filepath.Join(dest, name), name: name})
	}

	jobCh := make(chan moveJob)
	outCh := make(chan moveOut, 16)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				size, fb, warn, err := transfer(j.src, j.dst, deleteSource)
				outCh <- moveOut{job: j, size: size, fallback: fb, warn: warn, err: err}
			}
		}()
	}
	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
		wg.Wait()
		close(outCh)
	}()

	now := time.Now()
	records := make([]Image, 0, len(jobs))
	done := 0
	for o := range outCh {
		done++
		if o.err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", o.job.name, o.err))
		} else {
			res.Moved++
			if o.fallback {
				res.Fallbacks++
			}
			if o.warn != "" {
				res.Warnings = append(res.Warnings, o.job.name+": "+o.warn)
			}
			records = append(records, Image{DestDir: dest, FileName: o.job.name, Path: o.job.dst,
				SourcePath: o.job.src, Size: o.size, BackedUpAt: now})
		}
		if done%5 == 0 || done == len(jobs) {
			runtime.EventsEmit(a.ctx, "move:progress", map[string]any{"done": done, "total": len(jobs), "file": o.job.name})
		}
	}

	if len(records) > 0 {
		if err := a.db.CreateInBatches(records, 200).Error; err != nil {
			d := time.Since(start)
			res.ElapsedMs, res.Elapsed = d.Milliseconds(), fmtElapsed(d)
			return res, fmt.Errorf("ย้ายไฟล์แล้วแต่บันทึกฐานข้อมูลไม่สำเร็จ (%d ไฟล์): %w", len(records), err)
		}
	}
	d := time.Since(start)
	res.ElapsedMs, res.Elapsed = d.Milliseconds(), fmtElapsed(d)
	return res, nil
}

// knownNames รวมชื่อไฟล์ที่มีอยู่แล้วทั้งในโฟลเดอร์ปลายทางและในฐานข้อมูลของปลายทางนั้น
func (a *App) knownNames(dest string) (map[string]string, error) {
	known := map[string]string{}
	entries, err := os.ReadDir(dest)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			known[strings.ToLower(e.Name())] = "ในโฟลเดอร์ปลายทาง"
		}
	}
	var names []string
	if err := a.db.Model(&Image{}).Where("dest_dir = ?", dest).Pluck("file_name", &names).Error; err != nil {
		return nil, err
	}
	for _, n := range names {
		k := strings.ToLower(n)
		if _, ok := known[k]; ok {
			known[k] = "ทั้งในโฟลเดอร์ปลายทางและในฐานข้อมูล"
		} else {
			known[k] = "ในฐานข้อมูล"
		}
	}
	return known, nil
}

func scanImages(root, skip string) ([]string, error) {
	var out []string
	skipAbs, _ := filepath.Abs(skip)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ข้ามส่วนที่อ่านไม่ได้
		}
		name := d.Name()
		if d.IsDir() {
			if p != root {
				abs, _ := filepath.Abs(p)
				if abs == skipAbs || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "$") || name == "System Volume Information" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		if backupExts[strings.ToLower(filepath.Ext(name))] {
			out = append(out, p)
		}
		return nil
	})
	return out, err
}

// transfer: ลอง os.Rename ก่อน; ถ้าข้ามไดรฟ์ไม่ได้ (*os.LinkError) ให้ Copy ให้สมบูรณ์ แล้วค่อยลบต้นทาง
func transfer(src, dst string, deleteSource bool) (size int64, fallback bool, warn string, err error) {
	if deleteSource {
		st, e := os.Stat(src)
		if e != nil {
			return 0, false, "", e
		}
		e = os.Rename(src, dst)
		if e == nil {
			return st.Size(), false, "", nil
		}
		var le *os.LinkError
		if !errors.As(e, &le) {
			return 0, false, "", e
		}
		// ข้ามไดรฟ์ → ไป fallback
	}
	size, err = copyFile(src, dst)
	if err != nil {
		os.Remove(dst)
		return 0, false, "", err
	}
	if deleteSource {
		if e := os.Remove(src); e != nil {
			return size, true, "คัดลอกสำเร็จแต่ลบไฟล์ต้นทางไม่ได้: " + e.Error(), nil
		}
	}
	return size, deleteSource, "", nil
}

func copyFile(src, dst string) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil {
		return 0, err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(out, in)
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return 0, err
	}
	if n != st.Size() {
		return 0, fmt.Errorf("ขนาดไฟล์ไม่ตรงกัน (%d != %d)", n, st.Size())
	}
	os.Chtimes(dst, time.Now(), st.ModTime())
	return n, nil
}
