package handler

import (
	"strings"
	"testing"
)

func TestArticleSanitizationPolicyPreservesColorClassOnBoldText(t *testing.T) {
	input := `<p><strong class="ql-color-gold">20</strong> <span class="ql-color-gold">ข้อความ</span></p>`
	got := articleSanitizationPolicy().Sanitize(input)

	if !strings.Contains(got, `<strong class="ql-color-gold">20</strong>`) {
		t.Fatalf("bold text color class was removed: %q", got)
	}
	if !strings.Contains(got, `<span class="ql-color-gold">ข้อความ</span>`) {
		t.Fatalf("plain text color class was removed: %q", got)
	}
}

func TestArticleSanitizationPolicyPreservesVideoOrientationClass(t *testing.T) {
	input := `<iframe class="ql-video video-portrait" src="https://www.youtube-nocookie.com/embed/abcdefghijk"></iframe>`
	got := articleSanitizationPolicy().Sanitize(input)

	if !strings.Contains(got, `class="ql-video video-portrait"`) {
		t.Fatalf("video orientation class was removed: %q", got)
	}
}
