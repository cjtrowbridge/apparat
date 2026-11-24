# Tab emoji artwork

Unmodified Twemoji v16.0.1 72x72 PNGs, by Twitter, Inc. and other contributors, licensed under CC BY 4.0; see [LICENSE-GRAPHICS](LICENSE-GRAPHICS).

Source: https://github.com/jdecked/twemoji/tree/v16.0.1/assets/72x72

- `1f528.png`: hammer, Projects.
- `1f5a7.png`: networked computers (U+1F5A7), Cluster. Rendered from Noto Sans Symbols 2 (Google, SIL Open Font License 1.1), independent of device fonts; this image is not Twemoji.
- `1f9ea.png`: test tube, Research.
- `2699.png`: gear, Settings; copied from the same attributed asset set in STMPCK.

`tab_icons.go` embeds these images and scales them once to 28 logical pixels on the GUI thread. No font fallback or runtime network access is required. Missing/corrupt files fail at build/load rather than silently producing blank tabs. GUI tests instantiate every tab; device screenshots verify actual appearance. Comrades remains the font-rendered ☭ symbol, not a bundled image.

- `1f9e0.png`: Twemoji brain for Global Agentic Session and for marking project repos with configured actions.
- `1f489.png`: STMPCK syringe from the same attributed Twemoji release.

- `1f519.png`: 🔙 workspace Back, from the pinned Twemoji release above. Its dark monochrome ink is tinted white when drawn on the dark navigation button.

- `2795.png`: ➕ prefix for pinned repository actions, from the same Twemoji release. The renderer tints it white for contrast.

- `270f.png`: Twemoji pencil, for editing the displayed personal screen name.

- `1f4f1.png` and `1f4bb.png` are the mobile phone and laptop from the same Twemoji release. They identify the operating system reported by an authenticated GUI installation.

- `1f3af.png`: Twemoji bullseye marking the local fingerprint at the start of its device row.

Pool tabs use ☑️ (2611.png from Twemoji v16.0.1), the existing networked computers icon, and 👥 (1f465.png from the STMPCK Twemoji set).

Workspace navigation uses the pinned Twemoji speech balloon (1f4ac.png) and folder (1f4c1.png), plus the existing pencil, laptop and checklist artwork.

Files uses the pinned Twemoji page (1f4c4.png) and folder (1f4c1.png) before each row’s Git marker and name.

Inference types use the speech balloon for text, framed picture (1f5bc), speaker (1f50a), microphone (1f3a4) and clapper board (1f3ac). The four additional PNGs come from the same pinned Twemoji release.
