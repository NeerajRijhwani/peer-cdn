package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Postgres struct {
	conn *pgx.Conn
}

func Connect() (*Postgres, error) {
	conn, err := pgx.Connect(context.Background(), "postgres://username:password@localhost:5432/mydb")
	if err != nil {
		return nil, err
	}
	connection := Postgres{
		conn,
	}
	return &connection, nil
}

func (p *Postgres) GetDetails(ctx context.Context, infohash string) (File, error) {
	rows, err := p.conn.Query(ctx, "SELECT info_hash,file_name, total_size, piece_length, origin_url,is_active FROM FILE WHERE infohash=$1", infohash)
	if err != nil {
		return File{}, fmt.Errorf("query execution failed: %w", err)
	}
	filedetails, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[File])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return File{}, err
		}
		return File{}, fmt.Errorf("failed to scan file: %w", err)
	}
	return filedetails, nil
}

func CheckInfoHash(infohash string) bool {
	return true
}

func (p *Postgres) UpdateBitfield(ctx context.Context, infohash string, bitfield []byte) error {
	return nil
}

func (p *Postgres) GetPieceHashes(ctx context.Context, infohash []byte) ([][20]byte, error) {
	return make([][20]byte, 0), nil
}
