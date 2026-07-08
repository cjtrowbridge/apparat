# `apparat`

`cmd/apparat` is the GUI console entrypoint.

It loads configuration with binary name `apparat`, creates the shared application runtime, supports `--smoke-test` and `--doctor`, and starts the GUI adapter during normal execution.

At startup it prints the selected runtime root and `last_run.log` path before entering the Ebitengine loop. By default on Linux, the GUI runtime root is `~/.local/share/apparat/apparat` unless `XDG_DATA_HOME`, `APPARAT_RUNTIME_DIR`, or `--runtime-dir` overrides it.

The release pipeline builds this command with the `gui` build tag into:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
```

On Linux, normal GUI builds require Ebitengine native desktop dependencies such as X11, cursor, randr, xinerama, xi, OpenGL, xxf86vm, and ALSA development headers.
