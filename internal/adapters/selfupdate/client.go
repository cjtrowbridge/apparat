// Package selfupdate reads Apparat's public development channel at one immutable revision.
// Network requests are bounded and cancellable; only qualified target metadata is accepted.
package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"
)

const Repository = "cjtrowbridge/apparat"
const MaxBinaryBytes int64 = 256 << 20

var hexRevision = regexp.MustCompile(`^[0-9a-f]{40}$`)
var hexDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Error codes are safe to display. Underlying transport errors may include private paths.
var ErrMetadata = errors.New("update metadata rejected")
var ErrDownload = errors.New("update download failed")
var ErrVerification = errors.New("update verification failed")
var ErrInstall = errors.New("update installation failed")

type Client struct {
	HTTP           *http.Client
	API, Raw, Arch string
}

func NewClient() *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 30 * time.Second
	return &Client{
		HTTP: &http.Client{Transport: transport, Timeout: 15 * time.Minute,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 || req.URL.Scheme != "https" {
					return ErrDownload
				}
				return nil
			}},
		API:  "https://api.github.com/repos/" + Repository,
		Raw:  "https://raw.githubusercontent.com/" + Repository,
		Arch: runtime.GOARCH,
	}
}

type Release struct {
	Revision, Path, SHA256 string
}

type manifest struct {
	Edition  string
	Ref      string
	Accepted bool
	Schema   int `json:"delivery_schema"`
	Targets  []struct{ Name, Output, State, SHA256 string }
}

func (c *Client) request(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, ErrDownload
	}
	req.Header.Set("User-Agent", "Apparat-Linux-Updater")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, ErrDownload
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, ErrDownload
	}
	return resp, nil
}

func (c *Client) readJSON(ctx context.Context, url string, value any) error {
	resp, err := c.request(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
	if err != nil || len(data) > 2<<20 || json.Unmarshal(data, value) != nil {
		return ErrMetadata
	}
	return nil
}

func (c *Client) Latest(ctx context.Context) (Release, error) {
	var commit struct{ SHA string }
	err := c.readJSON(ctx, c.API+"/commits/main", &commit)
	if err != nil {
		return Release{}, err
	}
	if !hexRevision.MatchString(commit.SHA) {
		return Release{}, ErrMetadata
	}
	var m manifest
	err = c.readJSON(ctx, c.Raw+"/"+commit.SHA+"/releases/verified.json", &m)
	if err != nil {
		return Release{}, err
	}
	if !m.Accepted || m.Edition != "ordinary" || m.Ref != "refs/heads/main" || m.Schema != 1 {
		return Release{}, ErrMetadata
	}
	name := "linux/" + c.Arch + "/apparat"
	path := "releases/" + name + "/latest"
	var found []Release
	for _, target := range m.Targets {
		if target.Name == name {
			if target.Output != path || target.State != "passed" || !hexDigest.MatchString(target.SHA256) {
				return Release{}, ErrMetadata
			}
			found = append(found, Release{commit.SHA, path, target.SHA256})
		}
	}
	if len(found) != 1 {
		return Release{}, ErrMetadata
	}
	return found[0], nil
}

func Digest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", ErrVerification
	}
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", ErrVerification
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (r Release) Valid(arch string) bool {
	return hexRevision.MatchString(r.Revision) && hexDigest.MatchString(r.SHA256) &&
		r.Path == fmt.Sprintf("releases/linux/%s/apparat/latest", arch)
}
