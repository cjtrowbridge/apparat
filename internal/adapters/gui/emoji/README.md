# Tab emoji artwork

Unmodified Twemoji v16.0.1 72x72 PNGs, by Twitter, Inc. and other contributors, licensed under CC BY 4.0; see [LICENSE-GRAPHICS](LICENSE-GRAPHICS).

Source: https://github.com/jdecked/twemoji/tree/v16.0.1/assets/72x72

- `1f528.png`: hammer, Projects.
- `1f5a7.png`: networked computers (U+1F5A7), Cluster. Rendered from Noto Sans Symbols 2 (Google, SIL Open Font License 1.1), independent of device fonts; this image is not Twemoji.
- `1f9ea.png`: test tube, Research.
- `2699.png`: gear, Settings; copied from the same attributed asset set in STMPCK.

`tab_icons.go` embeds these images and scales them once to 28 logical pixels on the GUI thread. No font fallback or runtime network access is required. Missing/corrupt files fail at build/load rather than silently producing blank tabs. GUI tests instantiate every tab; device screenshots verify actual appearance. Comrades remains the font-rendered ☭ symbol, not a bundled image.

- `1f9e0.png`: Twemoji brain, marking project repos with configured actions.
- `1f489.png`: STMPCK syringe from the same attributed Twemoji release.

- `1f519.png`: 🔙 workspace Back, from the pinned Twemoji release above. Its dark monochrome ink is tinted white when drawn on the dark navigation button.

- `2795.png`: ➕ prefix for pinned repository actions, from the same Twemoji release. The renderer tints it white for contrast.

- `270f.png`: Twemoji pencil, for editing the displayed personal screen name.

- `1f4f1.png` and `1f4bb.png` are the mobile phone and laptop from the same Twemoji release. They identify the operating system reported by an authenticated GUI installation.
