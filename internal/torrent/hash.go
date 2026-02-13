package torrent

import (
	"bytes"
	"crypto/sha1"
	"io"
	"math"
)

// Torrent MetaData Struct
type MetaData struct {
	Name        string
	FileURL     string
	FileSize    int64
	PieceLength int
	Pieces      [][]byte
	InfoHash    []byte
	
}

func calculateInfoHash(info map[string]BencodeValue) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, info); err != nil {
		return nil, err
	}
	hashed:=sha1.Sum(buf.Bytes())
	return hashed[:],nil
}

func CreateMetaData(fileurl,filename string,filesize,piecelength int64,r io.Reader)(*MetaData,error){
	//Loop through r reading the file at peicelength
	//hash the chunks obtained
	//Create infohash of the information
	//return the metaData
	
	if(piecelength==0){
		piecelength=1024*256}
	numberofpieces:=int64(math.Ceil(float64(filesize)/float64(piecelength)))
	pieces:=make([][]byte,0,numberofpieces)
	piece:=make([]byte,piecelength)
	for{
		no,err:=io.ReadFull(r,piece);
		if no>0 {
			hashed:=sha1.Sum(piece[:no])
			pieces = append(pieces, hashed[:])
		}
		
		if err!=nil{
			if(err==io.EOF|| err == io.ErrUnexpectedEOF){
				break;
			}
			return nil,err
		}
	}
	
	var piecesbuffer bytes.Buffer
	
	for _,p :=range(pieces){
		piecesbuffer.Write(p)
	}
	
	info:=map[string]BencodeValue{
		"length":       filesize,
		"name":         filename,
		"piece length": int64(piecelength),
		"pieces":piecesbuffer.Bytes(),
	}

	infoHash, err := calculateInfoHash(info)
	if err != nil {
		return nil, err
	}

	meta:=MetaData{
		Name:filename,
		FileURL: fileurl,
		FileSize: filesize,
		PieceLength:int(piecelength),
		Pieces:pieces,
		InfoHash:infoHash,
	}
	return &meta,nil
}