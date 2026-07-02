# Game And HUD Sources

These sources support the controller-first cross-platform HUD. They are pinned source trees; activation in the Apparat Go module remains separate work.

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `ebiten` | `https://github.com/hajimehoshi/ebiten.git` | `v2.9.9` (`f65118d0bf2d4cdc35b18d661e0a9f2bf9f3e81e`) | Apache-2.0 | Rendering, input, audio, window lifecycle, and mobile binding foundation | Planned active dependency; not wired into a root `go.mod` yet |
| `ebitenui` | `https://github.com/ebitenui/ebitenui.git` | `b1c31d852489cc8b4bc1d9581eaa75686875e5a7` (`v0.7.3-3-gb1c31d8`) | MIT | Retained-mode forms, lists, layout, and focus-capable HUD controls | Planned active dependency; not wired yet |
| `debugui` | `https://github.com/ebitengine/debugui.git` | `19edc7c03832136c85f3b44f1b05e8c997b4836f` (`v0.3.0-alpha-15-g19edc7c`) | Apache-2.0 | Developer-only runtime diagnostics and overlays | Reference pending a focused developer-tool integration; never the primary HUD |

## Selection Notes

- Ebitengine is deliberately pinned to the newest stable 2.9.x release found during this checkpoint.
- Ebitengine source may later be used through a local `replace` during engine-level debugging, but normal application builds should prefer a matching `go.mod` version.
- EbitenUI is the default for conventional controls where retained widgets improve focus, forms, lists, and layout.
- Raw Ebitengine rendering remains appropriate for dense custom surfaces.
- Debugui is restricted to development diagnostics and must not define user-facing product workflows.
- Adding the source trees does not validate Steam Deck, Windows, macOS, or Android packaging.
