package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"bufio"

	"github.com/cretz/go-scrap"

	"image"

	// OpenGL and GLFW requirements
	"unsafe"
    "runtime"
    "github.com/go-gl/glfw/v3.3/glfw"
    "github.com/go-gl/gl/v4.1-core/gl"

    // sample shader/texture stuff
	"github.com/cstegel/opengl-samples-golang/basic-textures/gfx"
)

// This example records the current screen
// and sends a single frame to a shiny window

func main() {
    // lock thread for GLWF/OpenGL (must be at top of main)
    runtime.LockOSThread()

	// set up graphics (hardcoded dimensions for now)
	// change dimensions to match primary display
	screenWidth  := 1920
	screenHeight := 1080
	screenRect := image.Rect(0,0,screenWidth,screenHeight)
	screenImage := image.NewRGBA(screenRect)
	frameStatusChan := make(chan bool)

	// should add exit condition for recordToStream() here
	fmt.Println("Starting application... ctrl-c to exit...")

	// set up TCP server for receiving data
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	// listen for connections in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			go handleConnection(conn, frameStatusChan, screenImage.Pix)
		}
	}()

	// start recording
	go recordToStream()

	// GLFW with OpenGL graphics

    // initialize GLFW library
    if err := glfw.Init(); err != nil {  
        panic(fmt.Errorf("could not initialize glfw: %v", err)) 
    }

    // OpenGL context hints
    glfw.WindowHint(glfw.ContextVersionMajor, 4) 
    glfw.WindowHint(glfw.ContextVersionMinor, 1) 
    glfw.WindowHint(glfw.Resizable, glfw.True) 
    glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile) 
    glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

    // create GLFW window and context
    win, err := glfw.CreateWindow(screenWidth, screenHeight, "Scrap OpenGL example", nil, nil)
    if err != nil {  
        panic(fmt.Errorf("could not create opengl renderer: %v", err))
    }
    win.MakeContextCurrent()

    // initialize OpenGL
    if err := gl.Init(); err != nil {
       panic(err)
    }

	// the linked shader program determines how the data will be rendered
	vertShader, err := gfx.NewShaderFromFile("shaders/basic.vert", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/basic.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	shaderProgram, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		panic(err)
	}
	defer shaderProgram.Delete()

	vertices := []float32{
		// top left
		1.0, 1.0, 0.0,   // position
		1.0, 1.0, 1.0,    // Color
		1.0, 0.0,         // texture coordinates

		// top right
		-1.0, 1.0, 0.0,
		1.0, 1.0, 1.0,
		0.0, 0.0,

		// bottom right
		-1.0, -1.0, 0.0,
		1.0, 1.0, 1.0,
		0.0, 1.0,

		// bottom left
		1.0, -1.0, 0.0,
		1.0, 1.0, 1.0,
		1.0, 1.0,
	}

	indices := []uint32{
		// rectangle
		0, 1, 2,  // top triangle
		0, 2, 3,  // bottom triangle
	}

	VAO := createVAO(vertices, indices)

	for !win.ShouldClose() {
		// poll events and call their registered callbacks
		glfw.PollEvents()

		// wait for frame status from reader (single frame)
		<-frameStatusChan

		// swap (swizzle) blue and red pixel values
		ConvertBGRA(screenImage.Pix)

		// create texture from image
		texture0, err := gfx.NewTexture(screenImage, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
		if err != nil {
			panic(err.Error())
		}

		// draw vertices
		shaderProgram.Use()

		// set texture0 to uniform0 in the fragment shader
		texture0.Bind(gl.TEXTURE0)
		texture0.SetUniform(shaderProgram.GetUniformLocation("ourTexture0"))

		gl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		gl.BindVertexArray(0)

		texture0.UnBind()

		// end of draw loop

		// swap in the rendered buffer
		win.SwapBuffers()
	}

}

func recordToStream() error {
	// Create the capturer
	cap, err := capturer()
	if err != nil {
		return err
	}

	// set up TCP connection for sending data
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		return err
	}

	// Just start sending a bunch of frames
	frameLenEncodingBuf := make([]byte, 4) // 32 bit value
	for {
		// Get the frame...
		if pix, _, err := cap.Frame(); err != nil {
			return err
		} else if pix != nil {
			frameLen := uint32(4 * cap.Width() * cap.Height())
			binary.BigEndian.PutUint32(frameLenEncodingBuf, frameLen)
			conn.Write(frameLenEncodingBuf)
			// Send a row at a time
			stride := len(pix) / cap.Height()
			rowLen := 4 * cap.Width() // RGBA = 4 * width
			for i := 0; i < len(pix); i += stride {
				if _, err = conn.Write(pix[i : i+rowLen]); err != nil {
					break
				}
			}
			if err != nil {
				panic(err)
			}
		}
		// should add exit condition here
		// (was done with case <-ctx.Done() previously)
	}
	return(err)
}

func capturer() (*scrap.Capturer, error) {
	if d, err := scrap.PrimaryDisplay(); err != nil {
		return nil, err
	} else if c, err := scrap.NewCapturer(d); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func handleConnection(conn net.Conn, frameStatusChan chan bool, frameBuffer []uint8) {
	bufReader := bufio.NewReader(conn)

	// loop forever (fix me eventually!)
	for {
		// read the frame length
		frameLenLen := 4
		frameLenBuf := make([]byte, frameLenLen) // 32 bit uint
		frameLenBytesRead := 0
		for {
			// this breaks if there is nothing in the buffer (EOF error)
			bytesRead, err := bufReader.Read(frameLenBuf[frameLenBytesRead : frameLenLen])
			if err != nil{
				panic(err)
			}
			frameLenBytesRead += bytesRead
			if frameLenBytesRead == frameLenLen {
				break
			}
		}

		// read the frame according to frame length
		readFrameLen := binary.BigEndian.Uint32(frameLenBuf)
		frameBytesRead := 0
		for {
			bytesRead, _ := bufReader.Read(frameBuffer[frameBytesRead : readFrameLen])
			frameBytesRead += bytesRead
			if uint32(frameBytesRead) == readFrameLen {
				break
			}
		}
		frameStatusChan <- true
	}
}

func ConvertBGRA(p []uint8) {
	if len(p)%4 != 0 {
		panic("input slice length is not a multiple of 4")
	}
	for i := 0; i < len(p); i += 4 {
		p[i+0], p[i+2] = p[i+2], p[i+0]
	}
}

func createVAO(vertices []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32;
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// copy indices into element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 3*4 + 2*4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3*4

	// color
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3*4

	// texture position
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(2)
	offset += 2*4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}
