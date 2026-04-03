package api

import (
	"net"
	"net/http"
	"strconv"

	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
	"github.com/NeerajRijhwani/peer-cdn/internal/tracker"
)

func Announce(w http.ResponseWriter, r *http.Request) (apiresponse, apierror) {
	query := r.URL.Query()
	tracker := r.Context().Value("tracker").(tracker.Tracker)
	port, err := strconv.Atoi(query.Get("port"))
	if err != nil {
		e := ApiError(400, "invalid port", err)
		return apiresponse{}, e
	}
	uploaded, err := strconv.Atoi(query.Get("uploaded"))
	if err != nil {
		e := ApiError(400, "invalid port", err)
		return apiresponse{}, e
	}
	downloaded, err := strconv.Atoi(query.Get("downloaded"))
	if err != nil {
		e := ApiError(400, "invalid port", err)
		return apiresponse{}, e
	}
	left, err := strconv.Atoi(query.Get("left"))
	if err != nil {
		e := ApiError(400, "invalid port", err)
		return apiresponse{}, e
	}
	peerannounce := storage.AnnounceRequest{
		Type:       "type",
		InfoHash:   query.Get("info_hash"),
		PeerID:     query.Get("peer_id"),
		Port:       port,
		Uploaded:   int64(uploaded),
		Downloaded: int64(downloaded),
		Left:       int64(left),
		Event:      query.Get("event"),
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	Peerinfos, err := tracker.HandleAnnounce(r.Context(), &peerannounce, ip)
	if err != nil {
		e := ApiError(400, "invalid port", err)
		return apiresponse{}, e
	}
	res := storage.AnnounceResponse{
		Type:     peerannounce.Type,
		Interval: 1800,
		Peers:    Peerinfos,
	}
	var list [2]any
	list[0] = peerannounce
	list[1] = res
	return apiresponse{
		status_code: 200,
		message:     "Announced Successfully",
		data:        list,
	}, apierror{}
}

func Checkhash(w http.ResponseWriter, r *http.Request) (apiresponse, apierror) {
	query := r.URL.Query()
	tracker := r.Context().Value("tracker").(tracker.Tracker)
	check := tracker.CheckInfoHash(r.Context(), query.Get("infohash"), query.Get("peerid"))
	if !check {
		return apiresponse{}, apierror{
			status_code: 400,
			message:     "InfoHash Requested not found",
			err:         nil,
		}
	}
	return apiresponse{
		status_code: 400,
		message:     "InfoHash Requested not found",
		data:        nil,
	}, apierror{}
}
