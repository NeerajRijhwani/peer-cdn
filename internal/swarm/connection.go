package swarm

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/NeerajRijhwani/peer-cdn/internal"
	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
)

const (
	Choke          = iota // not sending any data right now."
	Unchoke               // ready to fulfill requests
	Interested            // You have pieces I want to download
	Not_Interested        // You don't have any pieces I need
	Have                  // The index of the piece just successfully downloaded.
	Bitfield              // A bit-map of all pieces the sender has.
	Request               // 12 bytes: Index (4), Begin/Offset (4), and Length (4).
	Piece                 // 8+ bytes: Index (4), Begin (4), and the actual File Data.
	Cancel                // Used to stop a pending request.
)

// [Peer A] (initiator)                  [Peer B] (receiver)

//     |                                       |
//     | ---- TCP CONNECT --------------------> |
//     |                                       |
//     | ---- HANDSHAKE ---------------------> |
//     |                                       |
//     | <--- HANDSHAKE ---------------------- |
//     |                                       |
//     | ---- BITFIELD ----------------------> |
//     |                                       |
//     | <--- BITFIELD ----------------------- |
//     |                                       |
//     | ---- INTERESTED --------------------> |
//     |                                       |
//     | <--- UNCHOKE ------------------------ |
//     |                                       |
//     | ---- REQUEST -----------------------> |
//     |                                       |
//     | <--- PIECE -------------------------- |
// Diagram of flow of application

// func Handleconection(conn net.Conn, swarmManager *SwarmManager, db *storage.Postgres, peerid, infoHash string) {
// 	defer conn.Close()
// 	conn.SetDeadline(time.Now().Add(5 * time.Second))
// 	var swarm *Swarm
// 	//request is incoming i.e we dont know what file and peerid is requesting
// 	var infodata, peer_id [20]byte

// 	if infoHash == "" {
// 		buf := make([]byte, 68)
// 		_, err := io.ReadFull(conn, buf)
// 		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
// 			return
// 		}
// 		copy(infodata[:], buf[28:48])
// 		copy(peer_id[:], buf[48:])
// 		swarm, err = swarmManager.GateKeeper(infodata)
// 		if err != nil {
// 			// swarm does not exists
// 			return
// 		}
// 		handshake := internal.NewProtocolHandshake(infodata, peer_id)
// 		seralizedhandshake := handshake.Seralize()
// 		conn.Write(seralizedhandshake)
// 	}
// 	// request is outgoing i.e we know the infohash of the file we need
// 	if infoHash != "" {
// 		copy(infodata[:], []byte(infoHash))
// 		copy(peer_id[:], []byte(peerid))
// 		handshake := internal.NewProtocolHandshake(infodata, peer_id)
// 		seralizedhandshake := handshake.Seralize()
// 		conn.Write(seralizedhandshake)
// 		buf := make([]byte, 68)
// 		_, err := io.ReadFull(conn, buf)
// 		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
// 			return
// 		}
// 		var receivedinfodata [20]byte
// 		copy(receivedinfodata[:], buf[28:48])
// 		isverified := handshake.Verifyhandshake(receivedinfodata)
// 		if !isverified {
// 			return
// 		}
// 		swarm, err = RegisterSwarm(db, infodata, nil)
// 	}
// 	// Connection is verified
// 	conn.SetDeadline(time.Time{})
// 	// Add Peer to the pool
// 	swarmManager.Peer_Registeration(swarm.InfoHash, peerid, swarm.Bitfield, true, false, conn)
// 	// Read Messages
// 	for {

// 	}

// }

func IncomingRequest(conn net.Conn, swarmManager *SwarmManager, db *storage.Postgres, peerid, infoHash string, file *os.File) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	var infodata, peer_id [20]byte

	buf := make([]byte, 68)
	_, err := io.ReadFull(conn, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return
	}
	copy(infodata[:], buf[28:48])
	copy(peer_id[:], buf[48:])
	_, err = swarmManager.GateKeeper(infodata)
	if err != nil {
		// swarm does not exists
		return
	}
	handshake := internal.NewProtocolHandshake(infodata, peer_id)
	seralizedhandshake := handshake.Seralize()
	conn.Write(seralizedhandshake)

	sw := swarmManager.Swarms[infodata]
	for {
		conn.SetDeadline(time.Now().Add(30 * time.Second))
		msg, err := internal.ReadData(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Peer %s timed out. Closing Conection.", peerid)
				break
			}

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("Peer %s disconnected.", peerid)
				break
			}
			log.Printf("Error from %s: %v", peerid, err)
			break
		}
		if msg == nil {
			//keep alive condition
			continue
		}
		switch msg.Id {
		case Choke:
			// choke condition
			sw.Pool.Conns[string(peer_id[:])].Update_IsChoking(true)
		case Unchoke:
			//unchoke condition
			sw.Pool.Conns[string(peer_id[:])].Update_IsChoking(false)
		case Interested:
			// Interested
			sw.Pool.Conns[string(peer_id[:])].SetPeerInterested(true)
		case Not_Interested:
			// not Interested
			sw.Pool.Conns[string(peer_id[:])].SetPeerInterested(false)
		case Have:
			//Have (peer downloaded a piece)
			sw.Pool.Conns[string(peer_id[:])].UpdateBitfield(binary.BigEndian.Uint32(msg.Payload[:]))
		case Bitfield:
			// peer Send Bitfield
			sw.Pool.Conns[string(peer_id[:])].AddBitfield(msg.Payload)
		case Request:
			// request for a piece
			index := binary.BigEndian.Uint32(msg.Payload[0:4])
			offset := binary.BigEndian.Uint32(msg.Payload[4:8])
			lengthPiece := binary.BigEndian.Uint32(msg.Payload[8:12])
			sw.Pool.Conns[string(peer_id[:])].SendPiece(index, offset, lengthPiece)
		case Piece:
			//block piece sent to me through network
			idx := binary.BigEndian.Uint32(msg.Payload[0:4])
			offset := binary.BigEndian.Uint32(msg.Payload[4:8])
			block := msg.Payload[8:]
			Write_Block_To_File(int(idx), int(offset), sw.PieceLength, block, file)

		case Cancel:
			//Cancel any ongoing request

		default:

		}
	}
}

func OutgoingRequest(conn net.Conn, swarmManager *SwarmManager, db *storage.Postgres, peerid, infoHash string) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	//request is incoming i.e we dont know what file and peerid is requesting
	var infodata, peer_id [20]byte

	copy(infodata[:], []byte(infoHash))
	copy(peer_id[:], []byte(peerid))
	handshake := internal.NewProtocolHandshake(infodata, peer_id)
	seralizedhandshake := handshake.Seralize()
	conn.Write(seralizedhandshake)
	buf := make([]byte, 68)
	_, err := io.ReadFull(conn, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return
	}
	var receivedinfodata [20]byte
	copy(receivedinfodata[:], buf[28:48])
	isverified := handshake.Verifyhandshake(receivedinfodata)
	if !isverified {
		return
	}
}

func getBitfield(sm *SwarmManager, infohash []byte) ([]byte, error) {
	swarm := sm.Swarms[[20]byte(infohash)]
	swarm.BitMu.RLock()
	bitfield := swarm.Bitfield
	swarm.BitMu.RUnlock()
	return bitfield, nil
}
