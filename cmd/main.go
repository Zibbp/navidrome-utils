package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/zibbp/navidrome-utils/internal/database"
	"github.com/zibbp/navidrome-utils/internal/env"
	"github.com/zibbp/navidrome-utils/internal/file"
	"github.com/zibbp/navidrome-utils/internal/navidrome"
	"golang.org/x/exp/slices"
)

func main() {
	ctx := context.Background()

	// Initialize config
	config, err := env.Initialize(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing config")
	}

	// Setup logger
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	if config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	if !config.JSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Initialize DB connection
	db, err := database.Setup(config.NavidromeDBPath)
	if err != nil {
		slog.Error("Error setting up database", "error", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	// Initialize Navidrome API
	navidromeClient := navidrome.NewClient(config.NavidromeURL, 3, (1 * time.Second), http.DefaultClient)
	err = navidromeClient.Login(ctx, config.NavidromeUsername, config.NavidromePassword)
	if err != nil {
		log.Fatal().Err(err).Msg("error authenticating with the navidrome API")
	}

	// Read playlists directory and get all playlists
	playlists, err := file.ReadPlaylistFiles(config.ImportPlaylistsPath)
	if err != nil {
		slog.Error("Error reading playlist files", "error", err)
		os.Exit(1)
	}

	navidromePlaylists, err := navidromeClient.GetPlaylists(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting navidrome playlists")
	}

	for _, playlist := range playlists {
		logger := log.With().Str("playlist_name", playlist.Name).Logger()
		logger.Info().Msg("processing playlist")

		checkString := fmt.Sprintf("%s:%s", playlist.SourceId, playlist.DestinationId)
		playlistComment := fmt.Sprintf("%s\n\n%s", playlist.Description, checkString)

		navidromePlaylistId := ""

		// Check if playlist already exists in Navidrome
		for _, navidromePlaylist := range navidromePlaylists {
			logger.Debug().Msgf("checking if playlist '%s' contains '%s'", navidromePlaylist.ID, checkString)
			if strings.Contains(navidromePlaylist.Comment, checkString) {
				navidromePlaylistId = navidromePlaylist.ID
				logger = logger.With().Str("navidrome_playlist_id", navidromePlaylistId).Logger()
				logger.Info().Msgf("found navidrome playlist for %s", playlist.Name)
				break
			}
		}

		// Create Navidrome playlist if it does not exist
		if navidromePlaylistId == "" {
			logger.Info().Msg("navidrome playlist does not exist...creaing")
			createdId, err := navidromeClient.CreatePlaylist(ctx, &navidrome.CreatePlaylistRequest{
				Name:    playlist.Name,
				Comment: playlistComment,
				Public:  true,
			})
			if err != nil {
				logger.Error().Err(err).Msg("error creating navidrome playlist")
				continue
			}
			navidromePlaylistId = createdId
			logger = logger.With().Str("navidrome_playlist_id", navidromePlaylistId).Logger()
			logger.Info().Msg("created navidrome playlist")
		}

		// Always update playlist with latest data from source
		_, err = navidromeClient.UpdatePlaylist(ctx, navidromePlaylistId, &navidrome.CreatePlaylistRequest{
			Name:    playlist.Name,
			Comment: playlistComment,
			Public:  true,
		})
		if err != nil {
			logger.Error().Err(err).Msg("error updating navidrome playlist")
			continue
		}

		logger.Info().Msg("processing tracks")

		// Get track ids current in playlist
		navidromeTrackIdsInPlaylist, err := navidromeClient.GetPlaylistTracks(ctx, navidromePlaylistId)
		if err != nil {
			logger.Error().Err(err).Msg("error getting tracks in navidrome playlist")
			continue
		}

		log.Info().Msgf("%d tracks in navidrome playlist", len(navidromeTrackIdsInPlaylist))

		// Process each track
		for _, track := range playlist.Tracks {
			trackLogger := logger.With().Str("track_id", track.ID).Str("track_isrc", track.ISRC).Logger()
			trackLogger.Debug().Msg("searching for track in navidrome")
			navidromeTrackId := ""

			// Find track by ISRC or title/artist
			if track.ISRC != "" {
				navidromeTrackId, err = db.GetTrackIDByIsrc(track.ISRC)
				if err != nil {
					trackLogger.Error().Err(err).Msg("error getting navidrome track by isrc")
				}
			} else {
				navidromeTrackId, err = db.GetTrackIDByTitleAndArtist(track.Title, track.Artist)
				if err != nil {
					trackLogger.Error().Err(err).Msg("error getting navidrome track by title and artist")
				}
			}

			if navidromeTrackId == "" {
				trackLogger.Warn().Msg("could not find track in navidrome")
				continue
			}

			trackLogger = trackLogger.With().Str("navidrome_track_id", navidromeTrackId).Logger()
			trackLogger.Debug().Msg("found track in navidrome by isrc")

			// Check if track is already in playlist, if not add to playlist
			if slices.Contains(navidromeTrackIdsInPlaylist, navidromeTrackId) {
				trackLogger.Debug().Msg("track already exists in navidrome playlist")
				continue
			} else {
				err = navidromeClient.AddTrackToPlaylist(ctx, navidromePlaylistId, &navidrome.AddTrackToPlaylistRequest{
					IDs: []string{navidromeTrackId},
				})
				if err != nil {
					trackLogger.Error().Err(err).Msg("error adding track to navidrome playlist")
					continue
				}
				trackLogger.Info().Msg("added track to playlist")
			}
		}
		logger.Info().Msg("finished processing playlist")
	}
}
