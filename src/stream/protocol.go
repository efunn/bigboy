package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Wire protocol, all integers big-endian:
//
//	handshake: "BBOY" | uint32 width | uint32 height
//	frame:     uint32 byteLen | raw RGBA pixels (byteLen == width*height*4)
//
// The 2020 version hardcoded the resolution on both ends; the handshake
// lets the client size its buffers to whatever the host is capturing.

var magic = [4]byte{'B', 'B', 'O', 'Y'}

func writeHandshake(w io.Writer, width, height int) error {
	buf := make([]byte, 12)
	copy(buf[0:4], magic[:])
	binary.BigEndian.PutUint32(buf[4:8], uint32(width))
	binary.BigEndian.PutUint32(buf[8:12], uint32(height))
	_, err := w.Write(buf)
	return err
}

func readHandshake(r io.Reader) (width, height int, err error) {
	buf := make([]byte, 12)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, 0, err
	}
	if [4]byte(buf[0:4]) != magic {
		return 0, 0, fmt.Errorf("bad handshake magic %q", buf[0:4])
	}
	return int(binary.BigEndian.Uint32(buf[4:8])), int(binary.BigEndian.Uint32(buf[8:12])), nil
}

func writeFrame(w io.Writer, pix []byte) error {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(pix)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(pix)
	return err
}

// readFrame reads one frame into dst, which must be exactly the size
// negotiated in the handshake.
func readFrame(r io.Reader, dst []byte) error {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return err
	}
	frameLen := binary.BigEndian.Uint32(lenBuf[:])
	if int(frameLen) != len(dst) {
		return fmt.Errorf("frame length %d does not match negotiated buffer %d", frameLen, len(dst))
	}
	_, err := io.ReadFull(r, dst)
	return err
}
