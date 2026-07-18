package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/kbinani/screenshot"
)

// frameSource produces raw RGBA frames of a fixed size.
type frameSource interface {
	next() ([]byte, error)
	size() (width, height int)
}

// runHost streams frames (screen capture, or a synthetic test pattern) to
// every connected client. Blocking writes to a client naturally throttle
// that client's frame loop; each connection gets its own source.
func runHost(ctx context.Context, addr string, display int, pattern bool) error {
	if !pattern && screenshot.NumActiveDisplays() <= display {
		return fmt.Errorf("display %d not found (%d active)", display, screenshot.NumActiveDisplays())
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// unblock Accept when the context is cancelled
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	log.Printf("host: serving on %s (pattern=%v)", addr, pattern)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go func() {
			defer conn.Close()
			src := newSource(display, pattern)
			if err := streamTo(ctx, conn, src); err != nil && ctx.Err() == nil {
				log.Printf("host: client %s dropped: %v", conn.RemoteAddr(), err)
			}
		}()
	}
}

func newSource(display int, pattern bool) frameSource {
	bounds := screenshot.GetDisplayBounds(display)
	if pattern {
		return newPatternSource(bounds.Dx(), bounds.Dy())
	}
	return &screenSource{display: display}
}

func streamTo(ctx context.Context, conn net.Conn, src frameSource) error {
	width, height := src.size()
	w := bufio.NewWriterSize(conn, 1<<20)

	if err := writeHandshake(w, width, height); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	for ctx.Err() == nil {
		pix, err := src.next()
		if err != nil {
			return err
		}
		if err := writeFrame(w, pix); err != nil {
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// screenSource captures the local display.
type screenSource struct {
	display int
}

func (s *screenSource) size() (int, int) {
	bounds := screenshot.GetDisplayBounds(s.display)
	return bounds.Dx(), bounds.Dy()
}

func (s *screenSource) next() ([]byte, error) {
	img, err := screenshot.CaptureRect(screenshot.GetDisplayBounds(s.display))
	if err != nil {
		if runtime.GOOS == "darwin" {
			return nil, fmt.Errorf("capture failed: %w (on macOS, grant Screen Recording "+
				"permission in System Settings > Privacy & Security, or use -pattern)", err)
		}
		return nil, fmt.Errorf("capture failed: %w", err)
	}
	return img.Pix, nil
}

// patternSource generates an animated gradient with a sweeping bar — a
// capture-free way to exercise the full pipeline and eyeball latency or
// dropped frames (the bar sweeps the full width once per second).
type patternSource struct {
	width, height int
	pix           []byte
	start         time.Time
}

func newPatternSource(width, height int) *patternSource {
	return &patternSource{
		width:  width,
		height: height,
		pix:    make([]byte, width*height*4),
		start:  time.Now(),
	}
}

func (p *patternSource) size() (int, int) { return p.width, p.height }

func (p *patternSource) next() ([]byte, error) {
	elapsed := time.Since(p.start).Seconds()
	phase := int(elapsed*256) % 256
	barX := int(elapsed*float64(p.width)) % p.width
	for y := 0; y < p.height; y++ {
		row := p.pix[y*p.width*4:]
		rowG := byte((y*256/p.height + phase) % 256)
		for x := 0; x < p.width; x++ {
			r := byte((x*256/p.width + phase) % 256)
			g := rowG
			b := byte(255 - r)
			if x >= barX-4 && x <= barX+4 {
				r, g, b = 255, 255, 255
			}
			row[x*4+0] = r
			row[x*4+1] = g
			row[x*4+2] = b
			row[x*4+3] = 255
		}
	}
	// pace the pattern at ~60fps; real capture is naturally slower
	time.Sleep(16 * time.Millisecond)
	return p.pix, nil
}
