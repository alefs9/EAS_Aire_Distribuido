package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

const (
	Version = 1
	MaxBody = 1 << 20

	Register     = "REGISTER"
	Registered   = "REGISTERED"
	Heartbeat    = "HEARTBEAT"
	HeartbeatAck = "HEARTBEAT_ACK"
	Bye          = "BYE"
	ByeAck       = "BYE_ACK"
	Error        = "ERROR"
)

var (
	ErrFrame   = errors.New("invalid frame length")
	ErrJSON    = errors.New("invalid JSON message")
	ErrVersion = errors.New("unsupported protocol version")
	ErrMessage = errors.New("invalid message fields")
	nodeID     = regexp.MustCompile("^[A-Za-z0-9_-]{1,64}$")
)

type Message struct {
	Proto    int         `json:"proto"`
	Type     string      `json:"type"`
	NodeID   string      `json:"node_id,omitempty"`
	Capacity int         `json:"capacity,omitempty"`
	TS       int64       `json:"ts,omitempty"`
	Code     string      `json:"code,omitempty"`
	Detail   string      `json:"message,omitempty"`
	Work     bool        `json:"work,omitempty"`
	Job      *JobMessage `json:"job,omitempty"`
}

func New(kind, id string) Message {
	return Message{Proto: Version, Type: kind, NodeID: id}
}

func ValidID(id string) bool { return nodeID.MatchString(id) }

func (m Message) validate() error {
	if m.Proto != Version {
		return ErrVersion
	}
	if m.Type == Error {
		if (m.NodeID == "" || ValidID(m.NodeID)) && m.Code != "" && m.Detail != "" && m.Capacity == 0 && m.TS == 0 && m.Job == nil && !m.Work {
			return nil
		}
		return ErrMessage
	}
	if !ValidID(m.NodeID) || m.Code != "" || m.Detail != "" {
		return ErrMessage
	}
	if m.Job != nil {
		return m.validateJob()
	}
	if m.Work && m.Type != HeartbeatAck {
		return ErrMessage
	}
	switch m.Type {
	case Register:
		if m.Capacity > 0 && m.TS == 0 && m.Job == nil && !m.Work {
			return nil
		}
	case Heartbeat, HeartbeatAck:
		if m.TS > 0 && m.Capacity == 0 {
			return nil
		}
	case Registered, Bye, ByeAck:
		if m.Capacity == 0 && m.TS == 0 && m.Job == nil && !m.Work {
			return nil
		}
	}
	return ErrMessage
}

func Read(r io.Reader) (Message, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return Message{}, err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 || size > MaxBody {
		return Message{}, ErrFrame
	}
	body := make([]byte, size)
	if _, err := io.ReadFull(r, body); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Message{}, err
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 || body[0] != '{' || !utf8.Valid(body) || !json.Valid(body) {
		return Message{}, ErrJSON
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var m Message
	if err := decoder.Decode(&m); err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrJSON, err)
	}
	return m, m.validate()
}

func Write(w io.Writer, m Message) error {
	if err := m.validate(); err != nil {
		return err
	}
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if len(body) > MaxBody {
		return ErrFrame
	}
	frame := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(frame, uint32(len(body)))
	copy(frame[4:], body)
	n, err := w.Write(frame)
	if err == nil && n != len(frame) {
		return io.ErrShortWrite
	}
	return err
}
