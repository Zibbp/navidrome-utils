package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func Setup() (*Database, error) {
	log.Print("Connecting to Navidrome database.")
	db, err := sql.Open("sqlite3", "/navidrome/navidrome.db")

	if err != nil {
		log.Fatal("Error opening database: ", err)
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (d *Database) FindTrack(title, artist string) (string, error) {
	var path string
	err := d.DB.QueryRow("SELECT path FROM media_file WHERE title LIKE ? AND artist LIKE ?", "%"+title+"%", "%"+artist+"%").Scan(&path)
	if err != nil {
		log.Print("Error finding track: ", err)
		return "", err
	}
	return path, nil
}
