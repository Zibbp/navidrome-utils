package file

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"log/slog"

	"github.com/flytam/filenamify"

	"github.com/zibbp/navidrome-utils/internal/navidrome"
)

func ReadPlaylistFiles() ([]navidrome.Playlist, error) {
	var playlists []navidrome.Playlist

	// Get all files in the playlist directory
	files, err := os.ReadDir("/data/playlists/input")
	if err != nil {
		slog.Error("Error reading playlist files", "error", err)
		return nil, err
	}
	// Loop over each file
	for _, file := range files {
		// Read the file
		playlist, err := ReadPlaylistFile(file.Name())
		if err != nil {
			slog.Error("Error reading playlist file", "error", err)
			return nil, err
		}
		// Add the playlist to the list
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func ReadPlaylistFile(name string) (navidrome.Playlist, error) {
	var playlist navidrome.Playlist

	// Read the file
	data, err := os.ReadFile("/data/playlists/input/" + name)
	if err != nil {
		slog.Error("Error reading playlist file", "error", err)
		return playlist, err
	}

	// Unmarshal the data into the playlist
	err = json.Unmarshal(data, &playlist)
	if err != nil {
		slog.Error("Error unmarshalling playlist file", "error", err)
		return playlist, err
	}

	return playlist, nil
}

func CreateM3UPlaylistFile(name string) error {
	// check if file exits if not create it
	safeFileName, err := filenamify.Filenamify(name, filenamify.Options{Replacement: "-"})
	if err != nil {
		slog.Error("Error creating safe file name", "error", err)
	}
	filePath := fmt.Sprintf("/data/playlists/output/%s.m3u", safeFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			slog.Error("Error creating m3u playlist file", "error", err)
			return err
		}
		if _, err := file.WriteString("#EXTM3U\n"); err != nil {
			slog.Error("Error writing to m3u playlist file", "error", err)
		}
		defer file.Close()
	}

	return nil
}

func CheckTrackInM3UPlaylist(track string, playlist string) error {
	// Append track to file if not already in file
	safeFileName, err := filenamify.Filenamify(playlist, filenamify.Options{Replacement: "-"})
	if err != nil {
		slog.Error("Error creating safe file name", "error", err)
	}
	filePath := fmt.Sprintf("/data/playlists/output/%s.m3u", safeFileName)

	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading m3u playlist file", "error", err)
		return err
	}

	s := string(file)

	if !strings.Contains(s, track) {
		slog.Info("Adding track to m3u playlist", track, playlist)
		err := addTrackToM3UPlaylist(track, safeFileName)
		if err != nil {
			slog.Error("Error adding track to m3u playlist", "error", err)
			return err
		}
	}
	return nil
}

func addTrackToM3UPlaylist(track string, playlist string) error {
	// Append track to file
	filePath := fmt.Sprintf("/data/playlists/output/%s.m3u", playlist)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		slog.Error("Error opening m3u playlist file", "error", err)
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(track + "\n"); err != nil {
		slog.Error("Error writing to m3u playlist file", "error", err)
		return err
	}

	return nil
}
