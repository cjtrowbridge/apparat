//go:build gui && android

// The Activity supplies the clipboard through RuntimeInfo.
// Without an attached Activity there is no foreground clipboard owner.
package gui

func systemCopy(string) bool {
	return false
}
func systemPaste() string {
	return ""
}
