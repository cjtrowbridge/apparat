// Installs verified model bytes in device-local storage and describes downloadable variants.
// Sizes, digests and immutable revisions belong to the embedded release catalog.
package speech

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed catalog.json languages.txt ggml-tiny.en-q5_1.bin LICENSE-*
var assets embed.FS

type Model struct {
	ID, Label, File, Revision, SHA256 string
	Bytes                             int64
}

const Default = "tiny.en"

var Models []Model
var Languages string

func init() {
	b, err := assets.ReadFile("catalog.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &Models)
	if err != nil {
		panic(err)
	}
	b, err = assets.ReadFile("languages.txt")
	if err != nil {
		panic(err)
	}
	Languages = string(b)
}
func lookup(id string) (Model, error) {
	for _, m := range Models {
		if m.ID == id {
			return m, nil
		}
	}
	return Model{}, errors.New("unknown speech model")
}
func Install(ctx context.Context, dir string, m Model, progress func(int64)) (string, error) {
	target := filepath.Join(dir, m.File)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}
	file, err := os.Open(target)
	if err == nil {
		h := sha256.New()
		n, e := io.Copy(h, io.LimitReader(file, m.Bytes+1))
		_ = file.Close()
		if e == nil && n == m.Bytes && fmt.Sprintf("%x", h.Sum(nil)) == m.SHA256 {
			return target, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	var src io.ReadCloser
	if m.ID == Default {
		src, err = assets.Open(m.File)
	} else {
		req, e := http.NewRequestWithContext(ctx, "GET", "https://huggingface.co/ggerganov/whisper.cpp/resolve/"+m.Revision+"/"+m.File, nil)
		if e != nil {
			return "", e
		}
		client := http.Client{Timeout: 30 * time.Minute}
		res, e := client.Do(req)
		if e != nil {
			return "", errors.New("model download failed; check the connection and retry")
		}
		src = res.Body
		if res.StatusCode != 200 {
			_ = src.Close()
			return "", fmt.Errorf("model download returned HTTP %d", res.StatusCode)
		}
	}
	if err != nil {
		return "", err
	}
	defer func() {
		_ = src.Close()
	}()
	return installStream(ctx, target, m, src, progress)
}
func installStream(ctx context.Context, target string, m Model, src io.Reader, progress func(int64)) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(target), ".model-*")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	h := sha256.New()
	writer := io.MultiWriter(h, file)
	buf := make([]byte, 256*1024)
	var total int64
	for {
		err = ctx.Err()
		if err != nil {
			return "", err
		}
		n, e := src.Read(buf)
		total += int64(n)
		if total > m.Bytes {
			return "", errors.New("model exceeds catalog size")
		}
		_, err = writer.Write(buf[:n])
		if err != nil {
			return "", err
		}
		if progress != nil {
			progress(total)
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", errors.New("incomplete model download")
		}
	}
	if total != m.Bytes || fmt.Sprintf("%x", h.Sum(nil)) != m.SHA256 {
		return "", errors.New("model verification failed")
	}
	err = file.Sync()
	if err == nil {
		err = file.Close()
	}
	if err == nil {
		err = os.Rename(file.Name(), target)
	}
	return target, err
}
