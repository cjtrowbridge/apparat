//go:build speech

// Exercises packaged weights against silence and the upstream PCM speech fixture.
package speech

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"strings"
	"testing"
)

func TestBundledDecoder(t *testing.T) {
	m, err := lookup(Default)
	if err != nil {
		t.Fatal(err)
	}
	path, err := Install(context.Background(), t.TempDir(), m, nil)
	if err != nil {
		t.Fatal(err)
	}
	d, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	defer d.close()
	text, err := d.decode(make([]float32, 16000))
	if err != nil || text != "" {
		t.Fatal("silence produced a result", err)
	}
	wav, err := os.ReadFile("../../../third_party/speech/whisper.cpp/samples/jfk.wav")
	if err != nil {
		t.Fatal(err)
	}
	off := bytes.Index(wav, []byte("data"))
	if off < 0 || len(wav) < off+8 {
		t.Fatal("invalid PCM fixture")
	}
	audio := wav[off+8:]
	pcm := make([]float32, len(audio)/2)
	for i := range pcm {
		pcm[i] = float32(int16(binary.LittleEndian.Uint16(audio[2*i:]))) / 32768
	}
	text, err = d.decode(pcm)
	if err != nil || !strings.Contains(strings.ToLower(text), "country") {
		t.Fatal("known phrase was not recognized", err)
	}
}
