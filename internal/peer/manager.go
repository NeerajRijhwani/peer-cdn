package peer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	// Peer TTL in Redis (30 minutes)
	PeerTTL = 1800 * time.Second

	// Maximum peers to return
	MaxPeersReturned = 50
)
type Manager struct {
	redis  *redis.Client
	logger *zap.Logger
}

//Initialize new manager
func NewManager(redisClient *redis.Client, logger *zap.Logger) *Manager {
	return &Manager{
		redis:  redisClient,
		logger: logger,
	}
}

//Store peers in redis 
func (m *Manager) StorePeer(ctx context.Context, peer *storage.PeerInfo) error {

	key := fmt.Sprintf("peer:%s:%s", peer.InfoHash, peer.PeerID)
	data, err := json.Marshal(peer)
	if err != nil {
		return fmt.Errorf("failed to marshal peer: %w", err)
	}
	if err:=m.redis.Set(ctx,key,data,PeerTTL).Err(); err != nil {
		return fmt.Errorf("failed to store peer in Redis: %w", err)
	}
	
	m.logger.Debug("stored peer",
		zap.String("info_hash", peer.InfoHash),
		zap.String("peer_id", peer.PeerID),
	)
	return nil
}

//get peers with excluding the user peer
func (m *Manager) GetPeers(ctx context.Context,infoHash string, excludePeerID string) ([]storage.PeerInfo, error) {
	pattern:=fmt.Sprintf("peer:%s:*",infoHash)
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err!=nil{
		return nil,fmt.Errorf("failed to get peer keys: %w", err)
	}
	var peers []storage.PeerInfo
	for _,key :=range(keys){
		data, err := m.redis.Get(ctx, key).Result()
		if err!=nil{
			
			m.logger.Warn("failed to get peer data", zap.String("key", key), zap.Error(err))
			continue
		}
		var peer storage.PeerInfo
		if err := json.Unmarshal([]byte(data), &peer); err != nil {
			m.logger.Warn("failed to unmarshal peer", zap.Error(err))
			continue
		}
		if peer.PeerID!=excludePeerID{
			peers = append(peers, peer)
		}
		if len(peers) >= MaxPeersReturned {
			break
		}
	}
	return peers, nil
}

//remove a peer from redis 
func (m *Manager) RemovePeer(ctx context.Context, infoHash, peerID string) error {
	key:=fmt.Sprintf("peer:%s:%s",infoHash,peerID)
	if _,err:=m.redis.Del(ctx,key).Result(); err!=nil{
		return fmt.Errorf("failed to remove peer: %w", err)
	}
	return nil
}

// Get swarm size of peers connected 
func (m *Manager) GetSwarmSize(ctx context.Context, infoHash string) (int, error) {
	pattern := fmt.Sprintf("peer:%s:*", infoHash)
	
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get swarm size: %w", err)
	}

	return len(keys), nil
}