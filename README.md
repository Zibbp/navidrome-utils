# Important

This repository will not receive any updates until `isrc` tags are supportd in Navidrome. Searching by track titles, artists, and albums is not the best. The [maintainer is working on a scanner refactor](https://github.com/navidrome/navidrome/pull/2709) which will support `isrc` tags. Once this is released the logic can be simiplied by searching for the `isrc` tag.

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
