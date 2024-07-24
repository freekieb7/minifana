package wal

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"testing"
)

func TestWal(t *testing.T) {
	tmp, _ := os.CreateTemp("", "example*.bin")

	msg := "I want to write this"
	msg2 := "I want to write this 2"

	write(tmp, []byte(msg))
	write(tmp, []byte(msg2))

	tmp.Seek(0, 0) // Reset pointer

	ret := read(tmp)

	os.Remove(tmp.Name())

	if msg != string(ret[0]) {
		t.Error("Expected to be the same")
	}

	if msg2 != string(ret[1]) {
		t.Error("Expected to be the same")
	}
}

func write(tmp *os.File, msg []byte) {
	msgLengthInBytes := len(msg)

	maxFrameLenInBytes := math.MaxUint16 // 65 KB

	start := 0
	end := msgLengthInBytes

	if end > maxFrameLenInBytes {
		end = maxFrameLenInBytes
	}

	for i := 0; i < math.MaxInt; i++ {
		frame := msg[start:end]
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint16(bs, uint16(len(frame)))

		binary.Write(tmp, binary.LittleEndian, bs)
		tmp.Write(msg[start:end])

		if end < msgLengthInBytes {
			start = end
			end = msgLengthInBytes

			if end > msgLengthInBytes {
				end = maxFrameLenInBytes
			}

			continue
		}

		break
	}
}

func read(tmp *os.File) [][]byte {
	var result [][]byte

	frameSizeBuffer := make([]byte, 4)
	frameBuffer := make([]byte, math.MaxUint16)

	var buf []byte

	for i := 0; i < math.MaxInt; i++ {
		_, err := tmp.Read(frameSizeBuffer)

		if err == io.EOF {
			return result
		}

		frameLength := binary.LittleEndian.Uint16(frameSizeBuffer)

		if frameLength < math.MaxUint16 {
			tBuf := make([]byte, frameLength)
			tmp.Read(tBuf)
			buf = append(buf, tBuf...)

			result = append(result, buf)
			buf = []byte{}
			continue
		}

		tmp.Read(frameBuffer)
		buf = append(buf, frameBuffer...)

	}

	panic("content was bigger then expected")
}
