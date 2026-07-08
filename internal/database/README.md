# Database Package

This package owns SQLite connection lifecycle and forward migrations.

It opens local SQLite databases, enables required pragmas, applies checksumed migrations, exposes diagnostics, and keeps SQL details away from HUD packages.

Repository packages may depend on this package. GUI and command entrypoints should use it through the shared application runtime.
