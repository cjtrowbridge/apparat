stmpck-embed.patch exposes the embedded surface and adapts input and save ownership against the pinned private STMPCK module. Refresh against that exact revision and verify the canonical Linux and Android builds.

The embedding patch also moves only the embedded gallery settings gear after discovery in its centered toolbar, updating draw and hit positions together. Standalone controls remain at their original positions.

The scoped `.gitattributes` rule allows unified-diff context syntax (one space before a tab) while retaining other whitespace checks. It does not change generated Go indentation.

The embedded menu delegates directional keys, D-pad, activation and Back to the
host while it shows the gallery or settings. The host uses the surface's tile
geometry and pointer activation path. Standalone navigation and active gameplay
retain their original input handling. Exercise the gallery focus and Merge save
checks whenever this patch or its pinned upstream changes.

The STMPCK surface exposes its game-name table in stable game-ID order so the Actions catalog and gallery share the same destinations.
