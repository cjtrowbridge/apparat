//go:build !speech

// Keeps display-free tests and development builds independent of the native toolchain.
// Qualified GUI releases always select the speech build tag.
package speech

import "errors"

type decoder struct{}
type recorder struct{}

var unavailable = errors.New("Local speech requires the packaged GUI build")

func load(string) (*decoder, error) {
	return nil, unavailable
}
func (*decoder) close() {}
func (*decoder) decode([]float32) (string, error) {
	return "", unavailable
}
func capture() (*recorder, error) {
	return nil, unavailable
}
func (*recorder) stop()  {}
func (*recorder) close() {}
func (*recorder) snapshot() []float32 {
	return nil
}
