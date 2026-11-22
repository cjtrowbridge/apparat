// Dispatches local identity prompts to bundled native-dialog helpers.
// Sensitive inputs and answers cross private pipes and never command arguments.
package keydialog

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf16"
)

//go:embed linux.py
var linux string

//go:embed windows.ps1
var windows string

//go:embed macos.js
var macos string

func Run(ctx context.Context, request string) string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "/usr/bin/python3", "-I", "-c", linux)
	case "windows":
		chars := utf16.Encode([]rune(windows))
		raw := make([]byte, len(chars)*2)
		for i, v := range chars {
			binary.LittleEndian.PutUint16(raw[2*i:], v)
		}
		cmd = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", base64.StdEncoding.EncodeToString(raw))
	case "darwin":
		result, handled := macCredential(request)
		if handled {
			return result
		}
		cmd = exec.CommandContext(ctx, "/usr/bin/osascript", "-l", "JavaScript", "-e", macos)
	default:
		return "null"
	}
	cmd.Stdin = strings.NewReader(request)
	out, err := cmd.Output()
	if err != nil {
		return "null"
	}
	return string(out)
}
