package env

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Debug               bool   `env:"DEBUG,default=false"` // Enable debug logs
	JSON                bool   `env:"JSON,default=false"`  // Enable JSON log output
	NavidromeURL        string `env:"NAVIDROME_URL,required"`
	NavidromeUsername   string `env:"NAVIDROME_USERNAME,required"`
	NavidromePassword   string `env:"NAVIDROME_PASSWORD,required"`
	NavidromeDBPath     string `env:"NAVIDROME_DB_PATH,default=/data/navidrome/navidrome.db"`
	ImportPlaylistsPath string `env:"IMPORT_PLAYLISTS_PATH,default=/data/playlists/input"`
}

func Initialize(ctx context.Context) (*Config, error) {
	var c Config
	err := envconfig.Process(ctx, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
