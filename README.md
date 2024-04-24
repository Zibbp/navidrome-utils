# Navidrome Utils

Currently has one function which is creating M3U playlist files from Navidrome's database.

### How It Works

JSON playlist files are read which contain tracks for that playlist. Each trach is searched against the Navidrome database and if it is found, the track is added to the M3U playlist file.

### How to run
Check out this repository. Grab your navidrome.db (SQLite) file and place it inside the navidrome folder in the repository.
Copy the generated data directory from tidal-utils inside the repository's directory.

Start the software via `docker-compose up`.
Watch it go to work.

The playlists directory will contain your playlist files once it's done.