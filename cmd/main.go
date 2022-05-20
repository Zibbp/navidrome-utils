package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zibbp/navidrome-utils/internal/database"
	"github.com/zibbp/navidrome-utils/internal/file"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}
	// DB
	db, err := database.Setup()

	playlists, err := file.ReadPlaylistFiles()
	if err != nil {
		log.Fatal("Error reading playlist files: ", err)
	}
	// For loop each playlist
	for _, playlist := range playlists {
		// Create the M3U playlist
		err := file.CreateM3UPlaylistFile(playlist.Name)
		if err != nil {
			log.Fatal("Error creating M3U playlist file: ", err)
		}
		log.Infof("Processing playlist %s which has %d tracks", playlist.Name, len(playlist.Tracks))
		// For loop each track
		for _, track := range playlist.Tracks {

			foundTrack, err := db.FindTrack(track.Title, track.Artist)
			if err != nil {
				log.Errorf("Error finding track: %s - %s - with error: %s", track.Title, track.Artist, err)
			}
			if foundTrack != "" {
				log.Debugf("Found track in Navidrome database %s", track)

				err := file.CheckTrackInM3UPlaylist(foundTrack, playlist.Name)
				if err != nil {
					log.Fatal("Error adding track to M3U playlist: ", err)
				}
			}
		}
		log.Infof("Finished processing playlist %s", playlist.Name)
	}

}
