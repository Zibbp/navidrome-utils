# Navidrome Utils

Create M3U playlists from a JSON input using tracks in the Navidrome database.

The `navidrome-isrc` branch will not function on all navidrome instances, it requires a modification to parse ISRC tags.

## Paths

- `/data/navidrome`: mount to navidrome's data directory that contains the sqlite database
- `/data/playlists/input`: mount to JSON playlist files
- `/data/playlists/output`: mount to export of M3U files
