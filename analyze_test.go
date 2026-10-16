package main

import "testing"

func TestParseGeminiResponse(t *testing.T) {
	body := []byte(`{
		"candidates": [{
			"content": {
				"parts": [
					{"text": "{\"caption\":\"ชายทะเล\",\"tags\":[\"ทะเล\",\"ฟ้า\",\"ทะเล\"]}"}
				]
			}
		}]
	}`)

	got, err := parseGeminiResponse(body)
	if err != nil {
		t.Fatalf("parseGeminiResponse() error = %v", err)
	}
	if got.Caption != "ชายทะเล" {
		t.Errorf("Caption = %q, want %q", got.Caption, "ชายทะเล")
	}
	if len(got.Tags) != 2 || got.Tags[0] != "ทะเล" || got.Tags[1] != "ฟ้า" {
		t.Errorf("Tags = %#v, want [ทะเล ฟ้า]", got.Tags)
	}
}

func TestParseGeminiResponseWithoutText(t *testing.T) {
	if _, err := parseGeminiResponse([]byte(`{"candidates":[]}`)); err == nil {
		t.Fatal("parseGeminiResponse() expected an error when Gemini returns no text")
	}
}
