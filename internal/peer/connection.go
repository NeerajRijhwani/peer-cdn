package peer

import (
	"fmt"
	"net"
	"sync"
)

type PeerPool struct {
	sync.RWMutex
	Conns map[string]*PeerConnection
}

type PeerConnection struct {
	sync.RWMutex
	Conn           net.Conn
	Bitfield       []byte // What pieces do they have?
	AmChoking      bool   // Am I refusing to send them data?
	PeerInterested bool
	IsChoking      bool // Are they refusing to send me data?
	AmInterested   bool

	RequestMap    map[string]bool
	RequestSignal chan struct{}
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
		Conn:           conn,
		Bitfield:       bitfield,
		AmChoking:      amchoke,
		IsChoking:      ischoke,
		PeerInterested: false,
		AmInterested:   false,
	}
	p.Conns[peerid] = &peercon
	fmt.Printf("Connection Added :%s", peerid)
}

func (p *PeerPool) Remove_all_Cons() error {
	p.Lock()
	defer p.Unlock()
	for key, value := range p.Conns {
		if err := value.Conn.Close(); err != nil {
			fmt.Println("Connection cannot be closed")
		}
		delete(p.Conns, key)
	}
	return nil
}

// removing a tcp connection in peer pool
func (p *PeerPool) RemoveConn(peerid string) {
	p.Lock()
	defer p.Unlock()

	if peercon, ok := p.Conns[peerid]; ok {
		(*peercon).Conn.Close()
		delete(p.Conns, peerid)
	}

	fmt.Printf("Connection Deleted: %s", peerid)
}

func (p *PeerConnection) Update_AmChoking(value bool) {
	p.Lock()
	p.AmChoking = value
	p.Unlock()
}

func (p *PeerConnection) Update_IsChoking(value bool) {
	p.Lock()
	p.AmChoking = value
	p.Unlock()
}

func (pc *PeerConnection) SetPeerInterested(state bool) {
	pc.Lock()
	defer pc.Unlock()
	pc.PeerInterested = state
}

func (pc *PeerConnection) SetAmInterested(state bool) {
	pc.Lock()
	defer pc.Unlock()
	pc.AmInterested = state
}

func (pc *PeerConnection) UpdateBitfield(index uint32) {
	pc.Lock()
	defer pc.Unlock()
	byteIndex := index / 8

	if byteIndex >= uint32(len(pc.Bitfield)) {
		return
	}

	pc.Bitfield[byteIndex] |= byte(1 << (7 - uint(index%8)))
}

func (pc *PeerConnection) AddBitfield(bitfield []byte) {
	pc.Lock()
	defer pc.Unlock()

	pc.Bitfield = bitfield
}

func (pc *PeerConnection) SendPiece(index, offset, length uint32) {

}
