//go:build !darwin || !cgo

// Allows password dialogs without linking Apple Security.framework.
// Other operating systems implement credential storage in their native dialog helper.
package keydialog

func macCredential(string) (string, bool) {
	return "", false
}
