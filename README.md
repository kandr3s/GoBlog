
# GoBlog

[GoBlog](https://github.com/jlelse/GoBlog) flavor powering my IndieWeb presence at [kandr3s.co](https://kandr3s.co)

## Custom Plugins

- Syndication: Added icons.
- NowPlaying: Displays the currently playing song via the Last.FM API.
- CustomUIElements: Theme, styling and other UI elements.
- Letterboxd: OwnYour[Watches](https://indieweb.org/watch) crossposting Letterboxd diary entries via RSS and Micropub.
- [💿 Discoteca](https://kandr3s.co/discoteca): Templates for album-focused [Listens](https://indieweb.org/listen).

## Updated Files

While the CustomUIElements plugin attempts to keep all the customization in a single file, the following GoBlog core files were slightly modified:

- `config.go`: Added "Description" field.
- `ui.go`: Added custom classes and other UI elements.
- `uiComponents.go`: Updated visibility and updated times markup. 

Read more on [how I use GoBlog](https://kandr3s.co/colophon).
