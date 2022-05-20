package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"

	"github.com/zibbp/navidrome-utils/internal/navidrome"
)

func ReadPlaylistFiles() ([]navidrome.Playlist, error) {
	var playlists []navidrome.Playlist

	// Get all files in the playlist directory
	files, err := ioutil.ReadDir("/data/navidrome/playlists")
	if err != nil {
		log.Fatal("Error reading playlist directory: ", err)
		return nil, err
	}
	// Loop over each file
	for _, file := range files {
		// Read the file
		playlist, err := ReadPlaylistFile(file.Name())
		if err != nil {
			log.Fatal("Error reading playlist file: ", err)
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
	data, err := ioutil.ReadFile("/data/navidrome/playlists/" + name)
	if err != nil {
		log.Fatal("Error reading playlist file: ", err)
		return playlist, err
	}

	// Unmarshal the data into the playlist
	err = json.Unmarshal(data, &playlist)
	if err != nil {
		log.Fatal("Error unmarshalling playlist: ", err)
		return playlist, err
	}

	return playlist, nil
}

func CreateM3UPlaylistFile(name string) error {
	// check if file exits if not create it
	safeFileName, err := filenamify.Filenamify(name, filenamify.Options{Replacement: "-"})
	if err != nil {
		log.Fatal("Error creating safe file name: ", err)
	}
	filePath := fmt.Sprintf("/playlists/%s.m3u", safeFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal("Error creating m3u playlist file: ", err)
			return err
		}
		if _, err := file.WriteString("#EXTM3U\n"); err != nil {
			log.Fatal("Error writing to m3u playlist file: ", err)
		}
		defer file.Close()
	}

	return nil
}

func CheckTrackInM3UPlaylist(track string, playlist string) error {
	// Append track to file if not already in file
	safeFileName, err := filenamify.Filenamify(playlist, filenamify.Options{Replacement: "-"})
	if err != nil {
		log.Fatal("Error creating safe file name: ", err)
	}
	filePath := fmt.Sprintf("/playlists/%s.m3u", safeFileName)

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading m3u playlist file: ", err)
		return err
	}

	s := string(file)

	if !strings.Contains(s, track) {
		log.Debugf("Adding track %s to playlist %s", track, playlist)
		err := addTrackToM3UPlaylist(track, safeFileName)
		if err != nil {
			log.Fatal("Error adding track to m3u playlist file: ", err)
			return err
		}
	}
	log.Debugf("Found track %s in playlist %s", track, playlist)
	return nil
}

func addTrackToM3UPlaylist(track string, playlist string) error {
	// Append track to file
	filePath := fmt.Sprintf("/playlists/%s.m3u", playlist)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal("Error opening m3u playlist file: ", err)
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(track + "\n"); err != nil {
		log.Fatal("Error writing to m3u playlist file: ", err)
		return err
	}

	return nil
}
