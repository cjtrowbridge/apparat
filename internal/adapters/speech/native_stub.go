//go:build !speech

// Keeps display-free tests and development builds independent of the native toolchain.
// Qualified GUI releases always select the speech build tag.
package speech

import "errors"

type decoder struct{}
type recorder struct{}

var errUnavailable = errors.New("local speech requires the packaged GUI build")

func load(string) (*decoder, error) {
	return nil, errUnavailable
}
func (*decoder) close() {}
func (*decoder) decode([]float32) (string, error) {
	return "", errUnavailable
}
func capture() (*recorder, error) {
	return nil, errUnavailable
}
func (*recorder) stop()  {}
func (*recorder) close() {}
func (*recorder) count() int {
	return 0
}
func (*recorder) snapshot(int, bool) []float32 {
	return nil
}
