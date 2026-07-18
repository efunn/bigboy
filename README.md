# bigboy
large computer does large things

## stream

`src/stream` is the continuous version of the old `end_to_end` demos: it
captures a display and streams raw RGBA frames over TCP to an OpenGL client.

```sh
cd src/stream
go build .

# everything in one process (capture + render):
./stream -mode loopback

# or across machines:
./stream -mode host -addr 0.0.0.0:8080          # on the big computer
./stream -mode client -addr bigboy.local:8080   # on the thin client

# no screen-recording permission handy? exercise the pipeline with a
# synthetic animated test pattern (also useful for eyeballing latency —
# the white bar sweeps the full width once per second):
./stream -mode loopback -pattern
```

On macOS the capturing process needs Screen Recording permission
(System Settings > Privacy & Security). Esc or closing the window exits.

The old single-frame demos (`end_to_end.go`, `opengl_end_to_end.go`) predate
Go modules and depend on the Rust-backed `go-scrap`; `stream` replaces that
with `kbinani/screenshot`, so it builds with a plain `go build`.
