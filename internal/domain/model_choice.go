// Identifies a chat model independently of the worker chosen to serve a turn.
// Older workers without digest metadata retain a worker-qualified identity.
package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
)

type ModelRef struct {
	Provider string `json:"provider" yaml:"provider"`
	Name     string `json:"name" yaml:"name"`
	Digest   string `json:"digest,omitempty" yaml:"digest,omitempty"`
	Worker   string `json:"worker,omitempty" yaml:"worker,omitempty"`
}

func (m ModelRef) Valid() bool {
	if m.Provider == "" || m.Name == "" {
		return false
	}
	for _, s := range []string{m.Provider, m.Name, m.Digest, m.Worker} {
		if len(s) > 256 || strings.TrimSpace(s) != s || strings.IndexFunc(s, unicode.IsControl) >= 0 {
			return false
		}
	}
	return true
}

var ErrModelConflict = errors.New("the default changed on another device; review the new default and retry")

type ModelOption struct {
	Model ModelRef
	Nodes []string
}
type ModelDefault struct {
	Proof     string `json:",omitempty"`
	Model     ModelRef
	Revision  int64
	Operation string
}
type ModelSettings struct {
	CanDefault bool
	Default    ModelDefault
	Session    ModelRef
}
type ModelEdit struct {
	Proof            string `json:",omitempty"`
	Project, Session string
	Model            ModelRef
	Default          bool
	Expected         ModelDefault
	Operation        string
}

func ChatModelCatalog(local WorkerInventory, device PoolDevice, peers []ClusterDevice, now time.Time) []ModelOption {
	groups := map[ModelRef][]string{}
	add := func(v WorkerInventory, id, host string) {
		if !v.HasText(now) {
			return
		}
		for _, m := range v.Models {
			if !m.Text {
				continue
			}
			ref := ModelRef{Provider: "ollama", Name: m.Name, Digest: m.Digest}
			if ref.Digest == "" {
				ref.Worker = id
			}
			if !ref.Valid() {
				continue
			}
			nodes := groups[ref]
			found := false
			for _, n := range nodes {
				if n == host {
					found = true
					break
				}
			}
			if !found {
				groups[ref] = append(nodes, host)
			}
		}
	}
	add(local, device.ID, device.DisplayName(""))
	for _, p := range peers {
		if p.Inventory != nil {
			add(*p.Inventory, p.Fingerprint, p.Host)
		}
	}
	var result []ModelOption
	for m, n := range groups {
		sort.Strings(n)
		result = append(result, ModelOption{m, n})
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i].Model, result[j].Model
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if a.Provider != b.Provider {
			return a.Provider < b.Provider
		}
		if a.Digest != b.Digest {
			return a.Digest < b.Digest
		}
		return a.Worker < b.Worker
	})
	return result
}
