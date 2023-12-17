# GoBlog-Letterboxd

A [GoBlog](https://github.com/jlelse/GoBlog) plugin that archives Letterboxd diary entries via RSS and Micropub.

## Installation
1. Copy the repository to the GoBlog's plugins folder.
2. Setup [Config](#config) values.
3. Rebuild GoBlog executable.

## Config
```yaml
- path: ./plugins/letterboxd
    import: letterboxd
    config:
      username: "user" # Letterboxd Username
      section: "section" # GoBlog's Watches Section
      token: "MICROPUB-TOKEN" # GoBlog's Micropub Token
```

### Features

- [x] Fetches diary entries from a Letterboxd user RSS feed
- [x] Implement “Rewatched” status
- [x] Saves a copy of the Film poster locally
- [ ] Automatically fetch movie backdrops
- [ ] Fetch Film Directors names