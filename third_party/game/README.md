# Game And HUD Sources

These sources support the controller-first cross-platform HUD. They are pinned source trees; activation in the Apparat Go module remains separate work.

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `ebiten` | `https://github.com/hajimehoshi/ebiten.git` | `5f363d17271b029d43e17d17dc00da34f43e2a86` (`v2.10.6`) | Apache-2.0 | Rendering, input, audio, window lifecycle, and mobile binding foundation | Active local replacement in root `go.mod`; matches the nominal `v2.10.6` requirement; Linux and Android build validation recorded in the remediation plans |
| `ebitenui` | `https://github.com/ebitenui/ebitenui.git` | `a9fa18fc48cda29447f392a4afca584359b9f81f` (`v0.7.3-4-ga9fa18f`) | MIT | Retained-mode forms, lists, layout, and focus-capable HUD controls | Active root Go dependency at `v0.7.4-0.20260422023258-b1c31d852489` |
| `debugui` | `https://github.com/ebitengine/debugui.git` | `c0db3047c9344d406475b7ef9c97bfa0a01db79d` (`v0.4.0-alpha-9-gc0db304`) | Apache-2.0 | Developer-only runtime diagnostics and overlays | Reference pending a focused developer-tool integration; never the primary HUD |

## Selection Notes

- The initial upstream refresh reached 2.11 development head. Version-policy review selected stable v2.10.6: the old 2.9 restriction expressed stability intent, not an enduring API requirement. No required 2.11-only capability was found.
- Root `go.mod` replaces Ebitengine with this exact checkout. The module graph and gomobile tool pin are reconciled; mobile builds apply the display guard to an ignored source export, never to this checkout.
- EbitenUI is the default for conventional controls where retained widgets improve focus, forms, lists, and layout.
- Raw Ebitengine rendering remains appropriate for dense custom surfaces.
- Debugui is restricted to development diagnostics and must not define user-facing product workflows.
- Debugui remains reference-only; its source refresh does not activate an application dependency.
- Adding the source trees does not validate Steam Deck, Windows, macOS, or Android packaging.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request, then selected stable Ebitengine v2.10.6 during the authorized compatibility remediation. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Reference-only updates do not activate integrations. The engine/module/tool reconciliation is recorded in the [dependency plan](../../plans/past/2026-11-16-11-13-18_reconcile-engine-and-mobile-dependencies.md).
