package protocol

import (
	"encoding/binary"
	"time"
)

const (
	PacketTypeProbe    = 1
	PacketTypeFeedback = 2
	PacketHeaderSize   = 17
)

type Packet struct {
	Type      uint8
	Sequence  uint64
	Timestamp int64
}

func NewProbePacket(seq uint64) *Packet {
	return &Packet{
		Type:      PacketTypeProbe,
		Sequence:  seq,
		Timestamp: time.Now().UnixNano(),
	}
}

func NewFeedbackPacket(seq uint64) *Packet {
	return &Packet{
		Type:      PacketTypeFeedback,
		Sequence:  seq,
		Timestamp: time.Now().UnixNano(),
	}
}

func (p *Packet) Marshal() []byte {
	buf := make([]byte, PacketHeaderSize)
	buf[0] = p.Type
	binary.BigEndian.PutUint64(buf[1:9], p.Sequence)
	binary.BigEndian.PutUint64(buf[9:17], uint64(p.Timestamp))
	return buf
}

func Unmarshal(data []byte) (*Packet, error) {
	if len(data) < PacketHeaderSize {
		return nil, ErrInvalidPacket
	}

	return &Packet{
		Type:      data[0],
		Sequence:  binary.BigEndian.Uint64(data[1:9]),
		Timestamp: int64(binary.BigEndian.Uint64(data[9:17])),
	}, nil
}
