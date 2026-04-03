package internal

import (
	"encoding/binary"
	"io"
	"net"
)

type MessageProtocol struct {
	Id      byte
	Payload []byte
}

// Seralizing Message Protocol to Tcp Message
func Seralize(msg *MessageProtocol) []byte {
	if msg == nil {
		return make([]byte, 4) // Returns [0, 0, 0, 0]
	}
	length := uint32(len(msg.Payload) + 1)
	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = msg.Id

	copy(buf[5:], msg.Payload)

	return buf
}

// desearilizing the tcp Message into Message Protocol
func Deseralize(r io.Reader) (*MessageProtocol, error) {
	//read the length of Payload
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	//length =0 that means keep Alive condition
	if length == 0 {
		return nil, nil
	}

	// Read the rest of the Payload
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	msg := &MessageProtocol{
		Id:      messageBuf[0],
		Payload: messageBuf[1:],
	}
	return msg, nil
}

// handshake Protocol = Pstrlength(1):Pstr(19):Reserved Byte for file extenson(8):Infohash(20):peerId(20)
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewProtocolHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// seralizing the handshake struct to byte array
func (h *Handshake) Seralize() []byte {
	buf := make([]byte, 68)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], []byte(h.Pstr))
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func (h *Handshake) Verifyhandshake(expectedHash [20]byte) bool {
	return h.InfoHash == expectedHash && h.Pstr == "BitTorrent protocol"
}

func ReadData(r net.Conn) (*MessageProtocol, error) {
	msg, err := Deseralize(r)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func WriteData(w net.Conn, msg *MessageProtocol) error {
	buffer := Seralize(msg)
	if _, err := w.Write(buffer); err != nil {
		return err
	}
	return nil
}
