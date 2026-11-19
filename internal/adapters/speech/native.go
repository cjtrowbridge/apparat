//go:build speech

// Bridges the capture device and Whisper decoder to the single owning Go worker.
// The build pipeline links static CPU libraries for the current target.
package speech

/*
#cgo CPPFLAGS: -I${SRCDIR}/../../../third_party/speech/whisper.cpp/include -I${SRCDIR}/../../../third_party/speech/whisper.cpp/ggml/include
#cgo CXXFLAGS: -std=c++17 -O2
#cgo !android LDFLAGS: -L${SRCDIR}/../../../.tmp/whisper-host/lib
#cgo android LDFLAGS: -L${SRCDIR}/../../../.tmp/whisper-android/lib
#cgo LDFLAGS: -lwhisper -lggml -lggml-cpu -lggml-base
#cgo linux,!android LDFLAGS: -lm -ldl -lpthread
#cgo android LDFLAGS: -lm -ldl -llog -landroid -static-libstdc++
#cgo darwin LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework CoreFoundation
#include <stdlib.h>
#include <stdbool.h>
void *speech_load(const char *);
void speech_free(void *);
char *speech_decode(void *,const float *,int);
void *speech_capture(void);
void speech_stop(void *);
void speech_close_capture(void *);
int speech_snapshot(void *,float *,int,bool);
*/
import "C"
import (
	"errors"
	"strings"
	"unsafe"
)

type decoder struct{ ptr unsafe.Pointer }

func load(path string) (*decoder, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))
	h := C.speech_load(p)
	if h == nil {
		return nil, errors.New("Model could not load on this device")
	}
	return &decoder{h}, nil
}
func (d *decoder) close() {
	C.speech_free(d.ptr)
}
func (d *decoder) decode(pcm []float32) (string, error) {
	if len(pcm) == 0 {
		return "", nil
	}
	result := C.speech_decode(d.ptr, (*C.float)(unsafe.Pointer(&pcm[0])), C.int(len(pcm)))
	if result == nil {
		return "", errors.New("Speech decoding failed")
	}
	defer C.free(unsafe.Pointer(result))
	return strings.TrimSpace(C.GoString(result)), nil
}

type recorder struct{ ptr unsafe.Pointer }

func capture() (*recorder, error) {
	ptr := C.speech_capture()
	if ptr == nil {
		return nil, errors.New("Microphone unavailable. Check permission and the selected input device.")
	}
	return &recorder{ptr}, nil
}
func (r *recorder) stop() {
	C.speech_stop(r.ptr)
}
func (r *recorder) close() {
	C.speech_close_capture(r.ptr)
}
func (r *recorder) count() int {
	return int(C.speech_snapshot(r.ptr, nil, 0, false))
}
func (r *recorder) snapshot(n int, consume bool) []float32 {
	if n <= 0 {
		return nil
	}
	pcm := make([]float32, n)
	size := C.speech_snapshot(r.ptr, (*C.float)(unsafe.Pointer(&pcm[0])), C.int(n), C.bool(consume))
	if size < 0 {
		return nil
	}
	return pcm[:int(size)]
}
