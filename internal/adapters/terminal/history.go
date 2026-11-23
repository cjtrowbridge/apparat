// Persists terminal bytes in numbered chunks under the session's private directory.
// Readers page backward from a byte offset; retained data has a fixed per-session ceiling.
package terminal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const chunkSize = 32768
const chunkLimit = 256

type scrollback struct {
	dir    string
	offset int64
}

func (b *scrollback) Write(data []byte) (int, error) {
	count := len(data)
	for len(data) > 0 {
		index := b.offset / chunkSize
		used := int(b.offset % chunkSize)
		size := min(len(data), chunkSize-used)
		name := filepath.Join(b.dir, fmt.Sprintf("output-%016x", index))
		if used == 0 && index >= chunkLimit {
			_ = os.Remove(filepath.Join(b.dir, fmt.Sprintf("output-%016x", index-chunkLimit)))
		}
		f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return count, nil
		}
		n, err := f.Write(data[:size])
		_ = f.Close()
		b.offset += int64(n)
		if err != nil || n != size {
			return count, nil
		}
		data = data[size:]
	}
	return count, nil
}
func readOutput(dir, cursor string) (string, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	var pages []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "output-") {
			pages = append(pages, e.Name())
		}
	}
	if len(pages) == 0 {
		return "No retained output", "", nil
	}
	sort.Strings(pages)
	first, _ := strconv.ParseInt(strings.TrimPrefix(pages[0], "output-"), 16, 64)
	last, _ := strconv.ParseInt(strings.TrimPrefix(pages[len(pages)-1], "output-"), 16, 64)
	info, err := os.Stat(filepath.Join(dir, pages[len(pages)-1]))
	if err != nil {
		return "", "", err
	}
	end := last*chunkSize + info.Size()
	if cursor != "" {
		end, err = strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return "", "", err
		}
	}
	if end <= first*chunkSize {
		return "End of retained output", "", nil
	}
	start := max(first*chunkSize, end-4096, (end-1)/chunkSize*chunkSize)
	f, err := os.Open(filepath.Join(dir, fmt.Sprintf("output-%016x", start/chunkSize)))
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	data := make([]byte, end-start)
	n, err := f.ReadAt(data, start%chunkSize)
	if err != nil && err != io.EOF {
		return "", "", err
	}
	// Render inert text. Control data is never interpreted by the host console or the clipboard.
	value := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return -1
	}, string(data[:n]))
	next := ""
	if start > first*chunkSize {
		next = strconv.FormatInt(start, 10)
	}
	return value, next, nil
}
