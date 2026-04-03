package swarm

/*1. The "Life & Death" Functions (Management)
These functions handle the existence of a file in your node's memory.

Swarm Registration: This takes an InfoHash and metadata (piece count, SHA1 hashes) from Postgres and creates a "Room" for that file. It’s like turning on the "Open" sign for a specific asset.

Swarm Shutdown: When a file is no longer needed or deleted from the CDN, this function clears the bitfield from RAM, closes all active peer connections for that file, and removes the entry from the Manager’s map.

Active Participation Check: The "Gatekeeper" function. When a stranger dials in, this function quickly checks the map to see if we are hosting that specific file. If not, it drops the call.

2. The "Personnel" Functions (Peer Handling)
These functions manage the people coming in and out of those rooms.

Peer Registration: Once a handshake is verified, this function creates a "Peer Profile," attaches it to the correct Swarm, and initializes their status (setting them to "choked" until we're ready to trade).

Peer Cleanup: When a connection breaks, this function removes the peer from the map. Crucially, it must also notify the "Piece Picker" to stop expecting any data that was promised by that specific peer.

Bitfield Synchronization: This function handles the "I'll show you mine if you show me yours" part. It reads our local progress and sends it to new peers, while also saving the map of what they have into our RAM.

3. The "Strategy" Functions (Orchestration)
This is where the CDN efficiency happens. These functions decide how to move the data.

The Piece Picker: The "Brain." It compares what we need (our Bitfield) against what the peers have. It decides whether to use Rarest-First (to help the CDN health) or Sequential (to get the start of the file fast).

Request Queue Manager: To avoid "requesting the same thing twice," this function keeps track of "Pending Requests." It ensures we don't ask Peer A and Peer B for Piece #10 at the same exact time.

The Broadcast (HAVE) Engine: When we finish a piece, this function loops through every other peer in the room and shouts: "I have Piece #10 now!" This turns your node from a "Leecher" into a "Source."

4. The "Integrity" Functions (Verification)
Because this is a CDN, you cannot trust anyone. You must verify everything.

Piece Validator: This function takes a completed chunk of data from the network and calculates its SHA1 hash. It compares the result against the "Golden Hash" we got from Postgres.

Persistence Updater: Once a piece is validated, this function writes the data to the SSD and updates Postgres to mark that piece as "Verified." This ensures that if the server reboots, we don't have to download it again.

5. The "Origin Fallback" Functions (The CDN Safety Net)
Since you are centralized, you have a "Source of Truth" that a standard torrent doesn't have.

Health Monitor: This function constantly checks the "Availability" of pieces. If it sees that a piece is missing from all peers (or if all peers are too slow), it raises a "Red Flag."

Origin Requester: When the Red Flag is raised, this function stops looking at the P2P network and dials your Origin Server directly (via HTTP or a dedicated connection) to fetch the missing bytes. It then feeds those bytes back into the P2P swarm so other peers can benefit.

6. The "Reporting" Functions (Stats)
Node Pulse: This function calculates the total upload/download speeds and peer counts. It provides the data for your CDN dashboard so you can see if the system is actually saving you bandwidth on your Origin.
*/
import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NeerajRijhwani/peer-cdn/internal"
	"github.com/NeerajRijhwani/peer-cdn/internal/peer"
	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
)

type SwarmManager struct {
	mu     sync.RWMutex
	Swarms map[[20]byte]*Swarm // Key is Hex InfoHash
}

type Swarm struct {
	InfoHash    [20]byte
	File        *os.File
	Pool        *peer.PeerPool
	Bitfield    []byte // Bitfield i have
	BitMu       sync.RWMutex
	PieceHashes [][20]byte
	PieceLength int
	TotalLength int64

	Downloaded uint64 // Stats for this specific file
	Uploaded   uint64

	OriginURL       string
	FallbackTimeout time.Duration

	AvailabilityMap []int
	AvailabilityMu  sync.Mutex

	InFlightRequests map[int]bool
	RequestMu        sync.Mutex
}

func Initalize_SwarmManager() *SwarmManager {
	return &SwarmManager{
		Swarms: make(map[[20]byte]*Swarm),
	}
}

func RegisterSwarm(db *storage.Postgres, infohash [20]byte, bitfield []byte, Filepath string, Filename string) (*Swarm, error) {
	s := &Swarm{}
	s.Bitfield = bitfield
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	details, err := db.GetDetails(ctx, hex.EncodeToString(infohash[:]))
	if err != nil {
		return nil, fmt.Errorf("Cannot register swarm : %w", err)
	}

	hashes, err := db.GetPieceHashes(ctx, infohash[:])
	if err != nil {
		return nil, fmt.Errorf("could not get piece hashes: %w", err)
	}
	expectedBits := len(hashes)
	expectedBytes := (expectedBits + 7) / 8

	if len(s.Bitfield) != expectedBytes {
		s.Bitfield = make([]byte, expectedBytes)
	}
	complete_path := filepath.Join(Filepath, Filename)
	s.File, err = os.OpenFile(complete_path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	s.InfoHash = infohash
	s.PieceHashes = hashes
	s.PieceLength = details.Piece_length
	s.TotalLength = details.Total_size
	s.OriginURL = details.Origin_url
	s.FallbackTimeout = 10 * time.Second
	s.Pool = peer.InitializePeerPool()
	s.InFlightRequests = make(map[int]bool)

	return s, nil
}

func (s *Swarm) Shutdownswarm(db *storage.Postgres) error {
	err := s.Pool.Remove_all_Cons()
	if err != nil {
		return err
	}
	s.BitMu.Lock()
	bitfieldCopy := make([]byte, len(s.Bitfield))
	copy(bitfieldCopy, s.Bitfield)
	s.Bitfield = nil
	s.PieceHashes = nil
	s.InFlightRequests = nil
	s.BitMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = db.UpdateBitfield(ctx, hex.EncodeToString(s.InfoHash[:]), bitfieldCopy)
	if err != nil {
		return fmt.Errorf("failed to save progress during shutdown: %w", err)
	}
	s.File.Close()
	return nil
}

func (sm *SwarmManager) GateKeeper(infohash [20]byte) (*Swarm, error) {
	sm.mu.RLock()
	swarm, exists := sm.Swarms[infohash]
	sm.mu.RUnlock()
	if !exists {
		return nil, errors.New("Swarm Not Found")
	}
	return swarm, nil
}

func (sm *SwarmManager) Peer_Registeration(infohash [20]byte, peerid string, bitfield []byte, amchoke, ischoke bool, conn net.Conn) error {
	sm.mu.RLock()
	swarm, exists := sm.Swarms[infohash]
	sm.mu.RUnlock()

	if !exists {
		return errors.New("Swarm Not Found")
	}

	swarm.Pool.AddConn(peerid, bitfield, amchoke, ischoke, conn)
	return nil
}

func (sm *SwarmManager) Peer_Cleanup(infohash [20]byte, peerid string) error {
	sm.mu.RLock()
	swarm, exists := sm.Swarms[infohash]
	sm.mu.RUnlock()

	if !exists {
		return errors.New("Swarm Not Found")
	}

	swarm.Pool.RemoveConn(peerid)
	return nil
}

func (swarm *Swarm) Bitfield_Sync(peerid string, bitfield []byte) ([]byte, error) {
	swarm.BitMu.RLock()
	my_bitfield := make([]byte, len(swarm.Bitfield))
	copy(my_bitfield, swarm.Bitfield)
	swarm.BitMu.RUnlock()

	swarm.Pool.RLock()
	peer, exists := swarm.Pool.Conns[peerid]
	swarm.Pool.RUnlock()
	if !exists {
		return nil, errors.New("Peer Not Found")
	}
	peer.Lock()
	peer.Bitfield = make([]byte, len(bitfield))
	copy(peer.Bitfield, bitfield)
	peer.Unlock()

	return my_bitfield, nil
}

// Interested in exchaning info i.e the we dont have a peice which peer have
func IsInterested(seeder_bitfield, leecher_bitfield []byte) bool {
	for i := range seeder_bitfield {
		if ((^seeder_bitfield[i]) & leecher_bitfield[i]) != 0 {
			return true
		}
	}
	return false
}
func (s *Swarm) UpdateAvailability(peerBitfield []byte, increment bool) {
	s.AvailabilityMu.Lock()
	defer s.AvailabilityMu.Unlock()

	for i := 0; i < len(s.PieceHashes); i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		if (peerBitfield[byteIdx]>>bitIdx)&1 == 1 {
			if increment {
				s.AvailabilityMap[i]++
			} else {
				s.AvailabilityMap[i]--
			}
		}
	}
}

func (s *Swarm) HasPiece(index int) bool {
	s.BitMu.RLock()
	defer s.BitMu.RUnlock()
	byteIdx := index / 8
	bitIdx := 7 - (index % 8)
	return (s.Bitfield[byteIdx]>>bitIdx)&1 == 1
}

func (s *Swarm) PeicePicker() (int, error) {
	// Rarest Peice Picker

	s.BitMu.RLock()
	myBF := make([]byte, len(s.Bitfield))
	copy(myBF, s.Bitfield)
	s.BitMu.RUnlock()

	s.AvailabilityMu.Lock()
	availability := make([]int, len(s.AvailabilityMap))
	copy(availability, s.AvailabilityMap)
	s.AvailabilityMu.Unlock()

	s.RequestMu.Lock()
	defer s.RequestMu.Unlock()

	best_piece := -1
	min := math.MaxInt32
	for i, count := range s.AvailabilityMap {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		if (myBF[byteIdx]>>bitIdx)&1 == 1 || s.InFlightRequests[i] {
			continue
		}
		if count > 0 && count < min {
			min = count
			best_piece = i
		}
	}
	if best_piece != -1 {
		s.InFlightRequests[best_piece] = true
		return best_piece, nil
	}

	return -1, nil
}

func (s *Swarm) Delete_Inflight_Request(bitidx int) {
	s.RequestMu.Lock()
	defer s.RequestMu.Unlock()

	s.InFlightRequests[bitidx] = false
}

func (s *Swarm) Piece_Announcer(i int) error {
	s.Pool.RLock()
	activeConns := make([]net.Conn, 0, len(s.Pool.Conns))
	for _, p := range s.Pool.Conns {
		activeConns = append(activeConns, p.Conn)
	}
	s.Pool.RUnlock()

	buffer_payload := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer_payload, uint32(i))
	payload := internal.MessageProtocol{Id: Have, Payload: buffer_payload}
	buf := internal.Seralize(&payload)
	for _, peerconn := range activeConns {
		go func(c net.Conn) {
			c.SetWriteDeadline(time.Now().Add(time.Second * 2))
			c.Write(buf)
		}(peerconn)
	}
	return nil
}

func Write_Block_To_File(index, begin, pieceLen int, block []byte, file *os.File) error {
	absoluteOffset := int64(index*pieceLen + begin)

	n, err := file.WriteAt(block, absoluteOffset)
	if err != nil {
		return err
	}
	if n != len(block) {
		return io.ErrShortWrite
	}
	return nil
}
