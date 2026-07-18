package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// GLFW/OpenGL must stay on the main OS thread
	runtime.LockOSThread()
}

// runClient connects to a host, then renders incoming frames into a single
// reused OpenGL texture. The 2020 version created (and never freed) a new
// texture object per frame; here the pixels are uploaded with TexSubImage2D.
func runClient(ctx context.Context, cancel context.CancelFunc, addr string) error {
	var conn net.Conn
	var err error
	// in loopback mode the host goroutine may not be listening yet
	for attempt := 0; ; attempt++ {
		conn, err = net.Dial("tcp", addr)
		if err == nil || attempt >= 20 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	reader := bufio.NewReaderSize(conn, 1<<20)
	width, height, err := readHandshake(reader)
	if err != nil {
		return err
	}
	log.Printf("client: receiving %dx%d from %s", width, height, addr)

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("could not initialize glfw: %w", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	winW, winH := fitToMonitor(width, height)
	win, err := glfw.CreateWindow(winW, winH, "bigboy stream", nil, nil)
	if err != nil {
		return fmt.Errorf("could not create window: %w", err)
	}
	win.MakeContextCurrent()
	glfw.SwapInterval(1) // vsync

	if err := gl.Init(); err != nil {
		return err
	}

	// captured pixels are sRGB-encoded; sample them as sRGB and let the
	// sRGB framebuffer re-encode on output (gamma-correct passthrough)
	gl.Enable(gl.FRAMEBUFFER_SRGB)

	program, err := newProgram(vertexShaderSrc, fragmentShaderSrc)
	if err != nil {
		return err
	}
	defer gl.DeleteProgram(program)

	vao := createQuadVAO()
	texture := createStreamTexture(width, height)

	win.SetFramebufferSizeCallback(func(_ *glfw.Window, w, h int) {
		gl.Viewport(0, 0, int32(w), int32(h))
	})
	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})

	// Network reader: double-buffered so the reader fills one frame while
	// the render loop uploads the other. The unbuffered channel hands a
	// full buffer to the renderer and provides backpressure to the host.
	frames := make(chan []byte)
	readErr := make(chan error, 1)
	go func() {
		bufs := [2][]byte{
			make([]byte, width*height*4),
			make([]byte, width*height*4),
		}
		for i := 0; ; i ^= 1 {
			if err := readFrame(reader, bufs[i]); err != nil {
				readErr <- err
				return
			}
			select {
			case frames <- bufs[i]:
			case <-ctx.Done():
				return
			}
		}
	}()

	frameCount := 0
	lastTitle := time.Now()
	for !win.ShouldClose() && ctx.Err() == nil {
		glfw.PollEvents()

		// upload a new frame if one has arrived; otherwise redraw the last
		select {
		case pix := <-frames:
			gl.BindTexture(gl.TEXTURE_2D, texture)
			gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(width), int32(height),
				gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pix))
			frameCount++
		case err := <-readErr:
			log.Printf("client: stream ended: %v", err)
			win.SetShouldClose(true)
		default:
		}

		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(program)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.BindVertexArray(vao)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)

		win.SwapBuffers()

		if now := time.Now(); now.Sub(lastTitle) >= time.Second {
			win.SetTitle(fmt.Sprintf("bigboy stream — %dx%d @ %d fps", width, height, frameCount))
			log.Printf("client: %d fps", frameCount)
			frameCount = 0
			lastTitle = now
		}
	}

	// stop the host capture loop and unblock the reader goroutine
	cancel()
	conn.Close()
	return nil
}

// fitToMonitor shrinks the window (preserving aspect) if the stream is
// larger than 90% of the primary monitor.
func fitToMonitor(width, height int) (int, int) {
	if err := glfw.Init(); err != nil {
		return width, height
	}
	mode := glfw.GetPrimaryMonitor().GetVideoMode()
	scale := 1.0
	if s := 0.9 * float64(mode.Width) / float64(width); s < scale {
		scale = s
	}
	if s := 0.9 * float64(mode.Height) / float64(height); s < scale {
		scale = s
	}
	return int(float64(width) * scale), int(float64(height) * scale)
}

func createStreamTexture(width, height int) uint32 {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// allocate storage once; frames are uploaded with TexSubImage2D
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.SRGB8_ALPHA8, int32(width), int32(height),
		0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	return tex
}

// createQuadVAO builds a fullscreen quad. Texture v=0 is mapped to the top
// of the quad so image row 0 (top of the captured screen) renders at the top.
func createQuadVAO() uint32 {
	vertices := []float32{
		// x, y, z, u, v
		-1.0, 1.0, 0.0, 0.0, 0.0, // top left
		1.0, 1.0, 0.0, 1.0, 0.0, // top right
		1.0, -1.0, 0.0, 1.0, 1.0, // bottom right
		-1.0, -1.0, 0.0, 0.0, 1.0, // bottom left
	}
	indices := []uint32{
		0, 1, 2,
		0, 2, 3,
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	var stride int32 = 5 * 4
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, stride, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
	return vao
}
