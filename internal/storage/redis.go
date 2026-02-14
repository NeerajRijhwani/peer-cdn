package storage

import "time"

type PeerInfo struct {
	PeerID     string
	IP         string
	Port       int
	InfoHash   string
	Uploaded   int64
	Downloaded int64
	Left       int64
	LastSeen   time.Time
}