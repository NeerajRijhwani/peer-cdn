package peer

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type PeerPool struct {
	sync.RWMutex
	Conns map[string]*PeerConnection
}

type PeerConnection struct {
	conn      net.Conn
	Bitfield  []byte // What pieces do they have?
	AmChoking bool   // Am I refusing to send them data?
	IsChoking bool   // Are they refusing to send me data?
}

// Initalize a new peer pool to manage tcp connections of peers
func InitializePeerPool() *PeerPool {
	return &PeerPool{
		Conns: make(map[string]*PeerConnection),
	}
}

// Adding a tcp connection in peer pool
func (p *PeerPool) AddConn(peerid string, bitfield []byte, amchoke, ischoke bool, conn net.Conn) {
	p.Lock()
	defer p.Unlock()
	peercon := PeerConnection{
		conn:      conn,
		Bitfield:  bitfield,
		AmChoking: amchoke,
		IsChoking: ischoke,
	}
	p.Conns[peerid] = &peercon
	fmt.Printf("Connection Added :%s", peerid)
}

// removing a tcp connection in peer pool
func (p *PeerPool) RemoveConn(peerid string) {
	p.Lock()
	defer p.Unlock()

	if peercon, ok := p.Conns[peerid]; ok {
		(*peercon).conn.Close()
		delete(p.Conns, peerid)
	}

	fmt.Printf("Connection Deleted: %s", peerid)
}

type MessageProtocol struct {
	id      byte
	payload []byte
}

const (
	// KeepAlive      = 0     Used to keep the connection from timing out.
	Choke          = iota // not sending any data right now."
	Unchoke               //  ready to fulfill requests
	Interested            //You have pieces I want to download
	Not_Interested        // You don't have any pieces I need
	Have                  // The index of the piece just successfully downloaded.
	Bitfield              // A bit-map of all pieces the sender has.
	Request               // 12 bytes: Index (4), Begin/Offset (4), and Length (4).
	Peice                 // 8+ bytes: Index (4), Begin (4), and the actual File Data.
	Cancel                // Used to stop a pending request.
)

// Seralizing Message Protocol to Tcp Message
func Seralize(msg *MessageProtocol) []byte {
	if msg == nil {
		return make([]byte, 4) // Returns [0, 0, 0, 0]
	}
	length := uint32(len(msg.payload) + 1)
	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = msg.id

	copy(buf[5:], msg.payload)

	return buf
}

// desearilizing the tcp Message into Message Protocol
func Deseralize(r io.Reader) (*MessageProtocol, error) {
	//read the length of payload
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

	// Read the rest of the payload
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	msg := &MessageProtocol{
		id:      messageBuf[0],
		payload: messageBuf[1:],
	}
	return msg, nil
}

// handshake Protocol = Pstrlength:Pstr:Reserved Byte for file extenson:Infohash:peerid
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
	return h.InfoHash == expectedHash
}
