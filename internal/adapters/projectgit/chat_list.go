// Builds the project chat directory from versioned Markdown front matter.
// Reads only the bounded header of each selected file and rejects link substitution.
package projectgit

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func ChatHeaders(location, before string) ([]domain.ChatHeader, error) {
	if before != "" && !domain.ValidID(before) {
		return nil, errors.New("invalid chat cursor")
	}
	dir, err := chatDirectory(location, false)
	if errors.Is(err, os.ErrNotExist) {
		return []domain.ChatHeader{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = dir.Close()
	}()
	files, err := fs.ReadDir(dir.FS(), ".")
	if err != nil {
		return nil, err
	}
	headers := []domain.ChatHeader{}
	for i := len(files) - 1; i >= 0 && len(headers) < 32; i-- {
		entry := files[i]
		name := entry.Name()
		if len(name) < 29 || !strings.HasSuffix(name, ".md") {
			continue
		}
		id := name[len(name)-29 : len(name)-3]
		expected, err := ChatFile(id)
		if err != nil || expected != name || before != "" && id >= before {
			continue
		}
		f, err := openChatLeaf(dir, name)
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(f, 8192))
		err = errors.Join(readErr, f.Close())
		if err != nil {
			return nil, err
		}
		end := bytes.Index(data, []byte("\n---\n"))
		if end < 0 {
			return nil, errors.New("missing chat header")
		}
		s, err := parseChat(data[:end+5], id)
		if err != nil {
			return nil, err
		}
		if len(data) > end+5 {
			headers = append(headers, s.ChatHeader)
		}
	}
	return headers, nil
}
