package tracker

import (
	"context"
	"fmt"
	"time"

	"github.com/NeerajRijhwani/peer-cdn/internal/peer"
	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
	"go.uber.org/zap"
)

type Peer struct {
	IP       string
	Port     int
	LastSeen time.Time
}

type Tracker struct {
	peerManager *peer.Manager
	logger      *zap.Logger
}

func NewTracker(peerManager *peer.Manager, logger *zap.Logger) *Tracker {
	return &Tracker{
		peerManager: peerManager,
		logger:      logger,
	}
}

func (t *Tracker) CheckInfoHash(ctx context.Context, infohash, peer_id string) bool {
	result, err := t.peerManager.GetCurrentPeer(ctx, peer_id, infohash)
	if (err != nil || result == storage.PeerInfo{}) {
		return false
	}
	return true
}

func (t *Tracker) HandleAnnounce(ctx context.Context, req *storage.AnnounceRequest, clientIP string) ([]storage.PeerInfoJson, error) {
	peerinfo := storage.PeerInfo{
		PeerID:     req.PeerID,
		IP:         clientIP,
		Port:       req.Port,
		InfoHash:   req.InfoHash,
		Uploaded:   req.Uploaded,
		Downloaded: req.Downloaded,
		Left:       req.Left,
		LastSeen:   time.Now(),
	}
	switch req.Event {
	case "stopped":
		err := t.peerManager.RemovePeer(ctx, req.InfoHash, req.PeerID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "started", "completed", "":
		// Store/update peer
		err := t.peerManager.StorePeer(ctx, &peerinfo)
		if err != nil {
			return nil, err
		}
		// get swarm of peers
		peers, err := t.peerManager.GetPeers(ctx, req.InfoHash, req.PeerID)
		if err != nil {
			return nil, err
		}
		peerlist := make([]storage.PeerInfoJson, len(peers))
		for i, peer := range peers {
			peerlist[i] = storage.PeerInfoJson{
				PeerID: peer.PeerID,
				IP:     peer.IP,
				Port:   peer.Port,
			}
		}
		return peerlist, nil
	default:
		return nil, fmt.Errorf("unknown event: %s", req.Event)
	}
}
