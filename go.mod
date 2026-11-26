module github.com/cjtrowbridge/apparat

go 1.26.11

require (
	github.com/bytemare/opaque v0.18.0
	github.com/ebitengine/gomobile v0.0.0-20260820040257-d11f821a26a6
	github.com/ebitenui/ebitenui v0.7.4-0.20260422023258-b1c31d852489
	github.com/hajimehoshi/ebiten/v2 v2.10.6
	github.com/pmezard/go-difflib v1.0.0
	github.com/yuin/goldmark v1.7.16
	golang.org/x/crypto v0.56.0
	golang.org/x/image v0.45.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.53.1-0.20260625155647-5d243466fa05
)

require (
	dario.cat/mergo v1.0.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	filippo.io/nistec v0.0.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/bytemare/ecc v0.9.0 // indirect
	github.com/bytemare/hash v0.6.2 // indirect
	github.com/bytemare/hash2curve v0.5.4 // indirect
	github.com/bytemare/ksf v0.5.0 // indirect
	github.com/bytemare/secp256k1 v0.3.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/creack/pty v1.1.24
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/purego v0.11.3 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/frustra/bbcode v0.0.0-20201127003707-6ef347fbe1c8 // indirect
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.9.2 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gtank/ristretto255 v0.2.0 // indirect
	github.com/hinshun/vt10x v0.0.0-20220301184237-5011da428d02
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/pjbgf/sha1cd v0.7.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	modernc.org/libc v1.73.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/hajimehoshi/ebiten/v2 => ./third_party/game/ebiten

require (
	ebe-boilerplate v0.0.0
	github.com/go-git/go-git/v5 v5.19.3
	golang.org/x/sys v0.47.0
)

require (
	golang.design/x/clipboard v0.7.0
	golang.org/x/exp/shiny v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/mobile v0.0.0-20231127183840-76ac6878050a // indirect
)

replace ebe-boilerplate => ./third_party/stmpck
