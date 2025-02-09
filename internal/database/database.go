package database

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func Setup(path string) (*Database, error) {
	log.Debug().Msgf("using database path: %s", path)
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?mode=ro&_foreign_keys=on&_busy_timeout=5000", path))
	if err != nil {
		return nil, err
	}

	log.Info().Msg("conneted to navidrome database")
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Set PRAGMA read_uncommitted = 1 once at startup
	_, err = db.Exec("PRAGMA read_uncommitted = 1;")
	if err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (d *Database) GetTrackIDByIsrc(isrc string) (string, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRow("SELECT id FROM media_file WHERE json_extract(tags, '$.isrc[0].value') = ?", isrc).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, tx.Commit()
}

func (d *Database) GetTrackIDsInPlaylist(playlistId string) ([]string, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var ids []string
	rows, err := tx.Query("SELECT media_file_id FROM playlist_tracks WHERE playlist_id = ?", playlistId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, tx.Commit()
}

func (d *Database) GetTrackIDByTitleAndArtist(title, artist string) (string, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var path string
	err = tx.QueryRow("SELECT media_file_id FROM media_file WHERE title LIKE ? AND artist LIKE ?", "%"+title+"%", "%"+artist+"%").Scan(&path)
	if err != nil {
		log.Print("Error finding track: ", err)
		return "", err
	}

	return path, tx.Commit()
}
