package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReadFragmentedAndConsecutiveFrames(t *testing.T) {
	var stream bytes.Buffer
	first := New(Register, "nodo-1")
	first.Capacity = 2
	second := New(Heartbeat, "nodo-1")
	second.TS = 123
	for _, want := range []Message{first, second} {
		if err := Write(&stream, want); err != nil {
			t.Fatal(err)
		}
	}
	reader := iotest.OneByteReader(&stream)
	for _, want := range []Message{first, second} {
		if got, err := Read(reader); got != want || err != nil {
			t.Fatalf("got %+v (%v), want %+v", got, err, want)
		}
	}
	if _, err := Read(reader); err != io.EOF {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestInvalidFrames(t *testing.T) {
	tests := []struct {
		name string
		size uint32
		body string
		want error
	}{
		{"empty", 0, "", ErrFrame},
		{"oversized", MaxBody + 1, "", ErrFrame},
		{"truncated", 2, "{", io.ErrUnexpectedEOF},
		{"malformed", 1, "{", ErrJSON},
		{"invalid UTF-8", 3, "{\xff}", ErrJSON},
		{"unknown fields", 0, `{"proto":1,"type":"BYE","node_id":"a","extra":1}`, ErrJSON},
		{"multiple objects", 0, "{} {}", ErrJSON},
		{"unsupported version", 0, `{"proto":2,"type":"BYE","node_id":"a"}`, ErrVersion},
		{"invalid ID", 0, `{"proto":1,"type":"BYE","node_id":"a b"}`, ErrMessage},
		{"missing capacity", 0, `{"proto":1,"type":"REGISTER","node_id":"a"}`, ErrMessage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size := test.size
			if size == 0 {
				size = uint32(len(test.body))
			}
			var header [4]byte
			binary.BigEndian.PutUint32(header[:], size)
			frame := io.MultiReader(bytes.NewReader(header[:]), strings.NewReader(test.body))
			if _, err := Read(frame); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
