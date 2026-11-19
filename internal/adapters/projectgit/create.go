// Initializes a new offline working tree on main in a private staging directory.
// Reserves the new folder exclusively before moving the completed Git metadata into it.
package projectgit

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func Create(ctx context.Context, base, parent, name string) (domain.Project, error) {
	var empty domain.Project
	if name == "" || name != strings.TrimSpace(name) || strings.HasSuffix(name, ".") || strings.EqualFold(name, ".git") || !filepath.IsLocal(name) || strings.ContainsAny(name, `/\:*?"<>|`) || strings.IndexFunc(name, unicode.IsControl) >= 0 || len(name) > 255 {
		return empty, errors.New("enter one folder name without path separators or reserved characters")
	}
	if parent == "" {
		parent = filepath.Join(base, "data", "repos")
		err := os.MkdirAll(parent, 0700)
		if err != nil {
			return empty, err
		}
	}
	if !filepath.IsAbs(parent) {
		return empty, errors.New("choose an absolute parent folder")
	}
	dest := filepath.Join(parent, name)
	stage, err := os.MkdirTemp(parent, ".apparat-create-")
	if err != nil {
		return empty, errors.New("cannot write to the chosen parent folder")
	}
	defer func() {
		_ = os.RemoveAll(stage)
	}()
	_, err = git.PlainInitWithOptions(stage, &git.PlainInitOptions{InitOptions: git.InitOptions{DefaultBranch: plumbing.Main}})
	if err != nil {
		return empty, err
	}
	p, err := Inspect(stage)
	if err != nil {
		return empty, err
	}
	err = ctx.Err()
	if err != nil {
		return empty, err
	}
	err = os.Mkdir(dest, 0700)
	if errors.Is(err, os.ErrExist) {
		return empty, errors.New("that folder already exists; choose a new name or open the existing repo")
	}
	if err != nil {
		return empty, err
	}
	// Only the initialized metadata belongs in an empty working tree.
	err = os.Rename(filepath.Join(stage, ".git"), filepath.Join(dest, ".git"))
	if err != nil {
		_ = os.Remove(dest)
		return empty, err
	}
	p.Location = filepath.Join(filepath.Dir(p.Location), name)
	p.Name = name
	return p, nil
}
