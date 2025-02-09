# Navidrome Utils

Creates playlists in Navidrome from a JSON file. This does not create a m3u playlist file, instead it uses the API to create and add tracks to playlists. This is so tracks can be added to the playlist by a human without it being overridden when run again.

## Docker

To run this you will need the following.

- JSON playlist files.
  - In the format in `internal/navidrome/navidrome.go`.
- Your Navidrome credentials.
  - I recommened using a different account unless you want the playlists to be owned by you.

Use the provided `docker-compose.yml` file to get started. Update the volumes to point to the Navidrome `data` directory (the one with the sqlite database) and the path to your JSON playlists. Update the environment variable with your Navidrome credentials.
