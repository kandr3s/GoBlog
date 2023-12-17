## GoBlog NowPlaying Plugin

Plugin for GoBlog blogging system that displays data from Last.FM.

* Display currently playing song
* Automatically post Loved Tracks (#NowPlaying)
* Generate Top Albums in the Month chart

## Installation

1.  Copy the “nowplaying” folder into your plugins folder.
2.  Add the following config to your config.yml

```yaml
plugins:
  - path: ./plugins/nowplaying
    import: nowplaying
    config:
      user: yourLastFMNick
      key: yourLastFMAPIKey
      favorites: false // Saves a #NowPlaying post for Loved tracks on Last.FM
      topalbums: true // Display Top Played Albums in the last month. 
```

Demo: 🎧 [kandr3s' #NowPlaying](https://kandr3s.co/listens#nowplaying)
