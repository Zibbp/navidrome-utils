# Navidrome Utils

Currently has one function which is creating M3U playlist files from Navidrome's database. 

### How It Works

JSON playlist files are read which contain tracks for that playlist. Each trach is searched against the Navidrome database and if it is found, the track is added to the M3U playlist file.