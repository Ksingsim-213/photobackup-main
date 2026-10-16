package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

const geminiModel = "gemini-2.5-flash"

const analyzePrompt = `วิเคราะห์ภาพนี้ แล้วตอบกลับเป็น JSON เท่านั้น ห้ามมีข้อความอื่นหรือ markdown
รูปแบบ: {"caption": "...", "tags": ["...", "..."]}
- caption: คำอธิบายภาพเป็นภาษาไทย 1-2 ประโยค ระบุสิ่งสำคัญ สถานที่ กิจกรรม บรรยากาศ
- tags: 5-8 คำ ภาษาไทย คำสั้น ๆ เป็นคำนาม (วัตถุ ฉาก กิจกรรม สี ประเภทภาพ) ไม่ซ้ำกัน`

var errNoKey = errors.New("ยังไม่ได้ตั้งค่า Gemini API key")

type analysis struct {
	Caption string   `json:"caption"`
	Tags    []string `json:"tags"`
}

func (a *App) callGemini(ctx context.Context, jpg []byte) (*analysis, error) {
	key := a.apiKey()
	if key == "" {
		return nil, errNoKey
	}
	payload, err := json.Marshal(map[string]any{
		"contents": []any{map[string]any{
			"parts": []any{
				map[string]any{"inline_data": map[string]any{
					"mime_type": "image/jpeg",
					"data":      base64.StdEncoding.EncodeToString(jpg),
				}},
				map[string]any{"text": analyzePrompt},
			},
		}},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"maxOutputTokens":  400,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("สร้างคำขอ Gemini ไม่สำเร็จ: %w", err)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + geminiModel + ":generateContent?key=" + url.QueryEscape(key)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("สร้างคำขอ Gemini ไม่สำเร็จ: %w", err)
		}
		req.Header.Set("content-type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("อ่านผลตอบกลับจาก Gemini ไม่สำเร็จ: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("ปิดผลตอบกลับจาก Gemini ไม่สำเร็จ: %w", closeErr)
		}
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("API %d", resp.StatusCode)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := string(body)
			if len(msg) > 200 {
				msg = msg[:200]
			}
			return nil, fmt.Errorf("Gemini API %d: %s", resp.StatusCode, msg)
		}
		return parseGeminiResponse(body)
	}
	return nil, lastErr
}

func parseGeminiResponse(body []byte) (*analysis, error) {
	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("อ่าน JSON จาก Gemini ไม่สำเร็จ: %w", err)
	}
	var text strings.Builder
	for _, candidate := range out.Candidates {
		for _, part := range candidate.Content.Parts {
			text.WriteString(part.Text)
		}
	}
	if text.Len() == 0 {
		return nil, errors.New("Gemini ไม่ส่งข้อความวิเคราะห์กลับมา")
	}
	return parseAnalysis(text.String())
}

func parseAnalysis(text string) (*analysis, error) {
	i, j := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if i < 0 || j < i {
		return nil, errors.New("ผลวิเคราะห์ไม่ใช่ JSON")
	}
	var r analysis
	if err := json.Unmarshal([]byte(text[i:j+1]), &r); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	tags := make([]string, 0, 8)
	for _, t := range r.Tags {
		t = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if t != "" && !seen[t] && len(tags) < 8 {
			seen[t] = true
			tags = append(tags, t)
		}
	}
	r.Tags, r.Caption = tags, strings.TrimSpace(r.Caption)
	return &r, nil
}

func (a *App) saveAnalysis(id uint, r *analysis) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		tgs := make([]Tag, 0, len(r.Tags))
		for _, n := range r.Tags {
			var t Tag
			if err := tx.Where("name = ?", n).FirstOrCreate(&t, Tag{Name: n}).Error; err != nil {
				return err
			}
			tgs = append(tgs, t)
		}
		if err := tx.Model(&Image{ID: id}).Association("Tags").Replace(tgs); err != nil {
			return err
		}
		return tx.Model(&Image{}).Where("id = ?", id).Updates(map[string]any{"caption": r.Caption, "analyzed": true}).Error
	})
}

type AnalyzeResult struct {
	Total     int      `json:"total"`
	Done      int      `json:"done"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors"`
	ElapsedMs int64    `json:"elapsedMs"`
}

// AnalyzeImages วิเคราะห์ภาพของปลายทางที่ยังไม่เคยวิเคราะห์ (ขนาน 3 งาน, เขียน DB ทีละรายการจาก goroutine เดียว)
func (a *App) AnalyzeImages(dest string) (*AnalyzeResult, error) {
	dest = normDir(dest)
	if dest == "" {
		return nil, errors.New("กรุณาระบุโฟลเดอร์ปลายทางก่อน")
	}
	if a.apiKey() == "" {
		return nil, errNoKey
	}
	start := time.Now()
	var pending []Image
	if err := a.db.Where("dest_dir = ? AND analyzed = ?", dest, false).Find(&pending).Error; err != nil {
		return nil, err
	}
	res := &AnalyzeResult{Errors: []string{}}
	jobs := make([]Image, 0, len(pending))
	for _, im := range pending {
		if !decodableExts[strings.ToLower(filepath.Ext(im.FileName))] {
			continue // เช่น HEIC: สำรองได้แต่วิเคราะห์ไม่ได้
		}
		if _, err := os.Stat(im.Path); err != nil {
			continue
		}
		jobs = append(jobs, im)
	}
	res.Total = len(jobs)

	type out struct {
		im  Image
		r   *analysis
		err error
	}
	jobCh := make(chan Image)
	outCh := make(chan out)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for im := range jobCh {
				jpg, err := loadResized(im.Path, 1024)
				if err != nil {
					outCh <- out{im: im, err: err}
					continue
				}
				r, err := a.callGemini(a.ctx, jpg)
				outCh <- out{im: im, r: r, err: err}
			}
		}()
	}
	go func() {
		for _, im := range jobs {
			jobCh <- im
		}
		close(jobCh)
		wg.Wait()
		close(outCh)
	}()

	n := 0
	for o := range outCh {
		n++
		if o.err == nil {
			o.err = a.saveAnalysis(o.im.ID, o.r)
		}
		if o.err != nil {
			res.Failed++
			res.Errors = append(res.Errors, o.im.FileName+": "+o.err.Error())
		} else {
			res.Done++
		}
		runtime.EventsEmit(a.ctx, "analyze:progress", map[string]any{"done": n, "total": res.Total, "file": o.im.FileName})
	}
	res.ElapsedMs = time.Since(start).Milliseconds()
	return res, nil
}

// AnalyzeImage วิเคราะห์ภาพเดียวใหม่อีกครั้ง
func (a *App) AnalyzeImage(id uint) (*Image, error) {
	im, err := a.imageByID(id)
	if err != nil {
		return nil, err
	}
	jpg, err := loadResized(im.Path, 1024)
	if err != nil {
		return nil, err
	}
	r, err := a.callGemini(a.ctx, jpg)
	if err != nil {
		return nil, err
	}
	if err := a.saveAnalysis(id, r); err != nil {
		return nil, err
	}
	var out Image
	err = a.db.Preload("Tags").First(&out, id).Error
	return &out, err
}
