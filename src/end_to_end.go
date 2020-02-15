package main

import (
	"encoding/binary"
	"context"
	"fmt"
	"log"
	"net"
	"bufio"

	"github.com/cretz/go-scrap"

	"image"
	// "image/draw"

	// "time"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/widget"
)

// This example records the current screen
// and streams it to a shiny window

func main() {
	// Stream to shiny window and wait for enter key asynchronously

	fmt.Println("Starting stream... press enter to exit...")
	errCh := make(chan error, 2)
	ctx, cancelFn := context.WithCancel(context.Background())

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
			go handleConnection(conn)
		}
	}()

	// Record
	go func() { errCh <- recordToStream(ctx) }()
	// Wait for enter
	go func() {
		fmt.Scanln()
		errCh <- nil
	}()
	err = <-errCh
	cancelFn()
	if err != nil && err != context.Canceled {
		log.Fatalf("Execution failed: %v", err)
	}
}

func recordToStream(ctx context.Context) error {
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

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
		// Check if we're done, otherwise go again
		select {
		case <-ctx.Done():
			return ctx.Err()
		// case err := <-errCh:
		// 	return err
		default:
		}
	}
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

func handleConnection(conn net.Conn) {
	screenWidth  := 1920
	screenHeight := 1200
	screenRect := image.Rect(0,0,screenWidth,screenHeight)
	screenImage := image.NewRGBA(screenRect)
	bufReader := bufio.NewReader(conn)

	// read the frame length
	frameLenLen := 4
	frameLenBuf := make([]byte, frameLenLen) // 32 bit uint
	frameLenBytesRead := 0
	for {
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
	readFrameBuf := make([]byte, readFrameLen)
	frameBytesRead := 0
	for {
		bytesRead, _ := bufReader.Read(readFrameBuf[frameBytesRead : readFrameLen])
		frameBytesRead += bytesRead
		if uint32(frameBytesRead) == readFrameLen {
			break
		}
	}
	screenImage.Pix = []uint8(readFrameBuf)

	driver.Main(func(s screen.Screen) {
		w := widget.NewSheet(widget.NewImage(screenImage, screenImage.Bounds()))
		if err := widget.RunWindow(s, w, &widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Width: screenWidth,
				Height: screenHeight,
				Title: "Scrap Shiny Example",
			},
		}); err != nil {
			log.Fatal(err)
		}
	})
}



