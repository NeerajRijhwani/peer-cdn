package storage

type AnnounceRequest struct {
	Type       string `json:"type"`
	InfoHash   string `json:"info_hash"`  // infohash of the file requested by peer
	PeerID     string `json:"peer_id"`    // unique peer id
	Port       int    `json:"port"`       //port of the peer listening
	Uploaded   int64  `json:"uploaded"`   // total bytes sent to other peers since started
	Downloaded int64  `json:"downloaded"` // total bytes received from other peers since started
	Left       int64  `json:"left"`       // total bytes left to finish the file
	Event      string `json:"event"`      // event- started,completed,stopped
}

type AnnounceResponse struct {
	Type     string         `json:"type"`
	Interval int            `json:"interval"` // seconds until next announce
	Peers    []PeerInfoJson `json:"peers"`    // list of all peers
}
type PeerInfoJson struct {
	PeerID string `json:"peer_id"`
	IP     string `json:"ip"`
	Port   int    `json:"port"`
}
type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}