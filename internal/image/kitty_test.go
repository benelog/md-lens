package image

import (
	"encoding/base64"
	"strings"
	"testing"
)

const esc = "\x1b" // st (ESC \) is already defined in the package

func TestSmallPayloadIsASingleFrame(t *testing.T) {
	png := make([]byte, 30)
	var sb strings.Builder
	EmitKitty(png, 10, 5, &sb)
	s := sb.String()
	if c := strings.Count(s, esc+"_G"); c != 1 {
		t.Errorf("frame count = %d, want 1", c)
	}
	if !strings.Contains(s, "f=100,a=T,t=d,c=10,r=5,C=1,m=0;") {
		t.Error("missing control sequence")
	}
	if !strings.HasSuffix(s, st) {
		t.Error("must end with ST")
	}
}

func TestLargePayloadIsChunkedOnFourKilobyteBoundaries(t *testing.T) {
	png := make([]byte, 5000)
	base64Len := len(base64.StdEncoding.EncodeToString(png)) // 6668

	var sb strings.Builder
	EmitKitty(png, 20, 8, &sb)
	s := sb.String()

	if c := strings.Count(s, esc+"_G"); c != 2 {
		t.Errorf("frame count = %d, want 2", c)
	}
	if c := strings.Count(s, "f=100,a=T"); c != 1 {
		t.Errorf("control count = %d, want 1 (control only on first frame)", c)
	}
	if !strings.Contains(s, "m=1;") {
		t.Error("non-final chunk must be marked m=1")
	}
	if !strings.Contains(s, "m=0;") {
		t.Error("final chunk must be marked m=0")
	}

	frames := strings.Split(s, st)
	payload1 := frames[0][strings.IndexByte(frames[0], ';')+1:]
	payload2 := frames[1][strings.IndexByte(frames[1], ';')+1:]

	if len(payload1) != KittyChunk {
		t.Errorf("payload1 length = %d, want %d", len(payload1), KittyChunk)
	}
	if len(payload1)%4 != 0 {
		t.Error("chunk length must be a multiple of 4")
	}
	if len(payload2) != base64Len-KittyChunk {
		t.Errorf("payload2 length = %d, want %d", len(payload2), base64Len-KittyChunk)
	}
	if !strings.HasPrefix(frames[1], esc+"_Gm=0;") {
		t.Error("second frame must start with ESC _G m=0;")
	}
}
