package main

import (
	"os"

	"log/slog"

	"github.com/zibbp/navidrome-utils/internal/database"
	"github.com/zibbp/navidrome-utils/internal/file"
)

func main() {
	// DB
	db, err := database.Setup()
	if err != nil {
		slog.Error("Error setting up database", "error", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	playlists, err := file.ReadPlaylistFiles()
	if err != nil {
		slog.Error("Error reading playlist files", "error", err)
		os.Exit(1)
	}
	// For loop each playlist
	for _, playlist := range playlists {
		// Create the M3U playlist
		err := file.CreateM3UPlaylistFile(playlist.Name)
		if err != nil {
			slog.Error("Error creating M3U playlist file", "error", err)
			os.Exit(1)
		}
		slog.Info("Created M3U playlist file", "name", playlist.Name)
		// For loop each track
		for _, track := range playlist.Tracks {

			navidromeTrack := ""
			if track.ISRC != "" {
				navidromeTrack, err = db.FindTrackByISRC(track.ISRC)
				if err != nil {
					slog.Error("Error finding track by ISRC", "isrc", track.ISRC, "error", err)
				}
			}
			// try alternate search to find track
			if navidromeTrack == "" {
				navidromeTrack, err = db.FindTrackByTitle(track.Title, track.Artist)
				if err != nil {
					slog.Error("Error finding track by title and artist", "title", track.Title, "artist", track.Artist, "error", err)
				}
			}
			if navidromeTrack != "" {
				err := file.CheckTrackInM3UPlaylist(navidromeTrack, playlist.Name)
				if err != nil {
					slog.Error("Error adding track to M3U playlist", "track", navidromeTrack, "error", err)
				}
				continue
			} else {
				slog.Info("Track not found", "title", track.Title)
			}
		}

		slog.Info("Finished processing playlist", "playlist", playlist.Name)
	}

}
