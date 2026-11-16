# Tab emoji artwork

Four unmodified Twemoji v16.0.1 72x72 PNGs, by Twitter, Inc. and other contributors, licensed under CC BY 4.0; see [LICENSE-GRAPHICS](LICENSE-GRAPHICS).

Source: https://github.com/jdecked/twemoji/tree/v16.0.1/assets/72x72

- `1f528.png`: hammer, Projects.
- `1f5a5.png`: desktop computer, Cluster.
- `1f9ea.png`: test tube, Research.
- `2699.png`: gear, Settings; copied from the same attributed asset set in STMPCK.

`tab_icons.go` embeds these images and scales them once to 28 logical pixels on the GUI thread. No font fallback or runtime network access is required. Missing/corrupt files fail at build/load rather than silently producing blank tabs. GUI tests instantiate every tab; device screenshots verify actual appearance. Comrades remains the font-rendered ☭ symbol, not a bundled image.
