// Stages verified ELF bytes alongside the installed binary and promotes by atomic rename.
// Keeps one previous executable; database rollback remains an offline backup operation.
package selfupdate

import (
	"context"
	"debug/elf"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Candidate struct {
	Path, Target, Original, SHA256 string
}

func (c *Client) Download(ctx context.Context, r Release, target string) (*Candidate, error) {
	return c.DownloadWithProgress(ctx, r, target, nil)
}

func (c *Client) DownloadWithProgress(ctx context.Context, r Release, target string, report func(int64, int64)) (*Candidate, error) {
	if !hexRevision.MatchString(r.Revision) || !hexDigest.MatchString(r.SHA256) || r.Path != "releases/linux/"+c.Arch+"/"+c.program()+"/latest" {
		return nil, ErrMetadata
	}
	target, err := filepath.EvalSymlinks(target)
	if err != nil {
		return nil, ErrInstall
	}
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return nil, ErrInstall
	}
	original, err := Digest(target)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".apparat-update-*")
	if err != nil {
		return nil, ErrInstall
	}
	path := file.Name()
	keep := false
	defer func() {
		_ = file.Close()
		if !keep {
			_ = os.Remove(path)
		}
	}()
	resp, err := c.request(ctx, c.Raw+"/"+r.Revision+"/"+r.Path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.ContentLength > MaxBinaryBytes {
		return nil, ErrVerification
	}
	var dest io.Writer = file
	if report != nil {
		report(0, resp.ContentLength)
		dest = &progressWriter{Writer: file, total: resp.ContentLength, report: report, last: time.Now()}
	}
	n, err := io.Copy(dest, io.LimitReader(resp.Body, MaxBinaryBytes+1))
	if err != nil {
		return nil, ErrDownload
	}
	if n == 0 || n > MaxBinaryBytes {
		return nil, ErrVerification
	}
	if report != nil {
		report(n, resp.ContentLength)
	}
	err = file.Chmod(0755)
	if err != nil {
		return nil, ErrInstall
	}
	err = file.Sync()
	if err != nil {
		return nil, ErrInstall
	}
	err = file.Close()
	if err != nil {
		return nil, ErrInstall
	}
	digest, err := Digest(path)
	if err != nil || digest != r.SHA256 || !nativeELF(path, c.Arch) {
		return nil, ErrVerification
	}
	if ctx.Err() != nil {
		return nil, ErrDownload
	}
	keep = true
	return &Candidate{Path: path, Target: target, Original: original, SHA256: r.SHA256}, nil
}

func nativeELF(path, arch string) bool {
	f, err := elf.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	machine := map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}[arch]
	return machine != 0 && f.Class == elf.ELFCLASS64 && f.Data == elf.ELFDATA2LSB &&
		f.Machine == machine && (f.Type == elf.ET_EXEC || f.Type == elf.ET_DYN)
}

func (c *Candidate) Discard() {
	_ = os.Remove(c.Path)
}

func (c *Candidate) Install(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrInstall
	}
	current, err := Digest(c.Target)
	if err != nil || current != c.Original {
		return ErrVerification
	}
	staged, err := Digest(c.Path)
	if err != nil || staged != c.SHA256 {
		return ErrVerification
	}
	// Link through a unique name so an existing previous backup is replaced atomically.
	backup, err := os.CreateTemp(filepath.Dir(c.Target), ".apparat-previous-*")
	if err != nil {
		return ErrInstall
	}
	name := backup.Name()
	_ = backup.Close()
	_ = os.Remove(name)
	defer os.Remove(name)
	err = os.Link(c.Target, name)
	if err != nil {
		return ErrInstall
	}
	err = os.Rename(name, c.Target+".previous")
	if err != nil {
		return ErrInstall
	}
	if ctx.Err() != nil {
		return ErrInstall
	}
	err = os.Rename(c.Path, c.Target)
	if err != nil {
		return ErrInstall
	}
	// The replacement is complete even if directory durability cannot be confirmed.
	// Reporting failure after promotion would incorrectly claim the old bytes remain.
	dir, err := os.Open(filepath.Dir(c.Target))
	if err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

type progressWriter struct {
	io.Writer
	received, total int64
	report          func(int64, int64)
	last            time.Time
}

func (w *progressWriter) Write(data []byte) (int, error) {
	n, err := w.Writer.Write(data)
	w.received += int64(n)
	if time.Since(w.last) >= 250*time.Millisecond {
		w.report(w.received, w.total)
		w.last = time.Now()
	}
	return n, err
}
