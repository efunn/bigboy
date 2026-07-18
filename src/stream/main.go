// bigboy stream: continuous desktop streaming, end to end.
//
// Modes:
//   host     - capture the local screen and serve frames over TCP
//   client   - connect to a host and render the incoming frames with OpenGL
//   loopback - run both in one process (the original end_to_end demo, but continuous)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	mode := flag.String("mode", "loopback", "host, client, or loopback")
	addr := flag.String("addr", "127.0.0.1:8080", "address to serve on (host) or connect to (client)")
	display := flag.Int("display", 0, "display index to capture (host)")
	pattern := flag.Bool("pattern", false, "stream a synthetic test pattern instead of the screen")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch *mode {
	case "host":
		if err := runHost(ctx, *addr, *display, *pattern); err != nil {
			log.Fatal(err)
		}
	case "client":
		if err := runClient(ctx, cancel, *addr); err != nil {
			log.Fatal(err)
		}
	case "loopback":
		go func() {
			if err := runHost(ctx, *addr, *display, *pattern); err != nil && ctx.Err() == nil {
				log.Fatal(err)
			}
		}()
		if err := runClient(ctx, cancel, *addr); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		flag.Usage()
		os.Exit(2)
	}
}
