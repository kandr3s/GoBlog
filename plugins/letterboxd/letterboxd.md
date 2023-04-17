# GoBlog-Letterboxd

A [GoBlog](https://github.com/jlelse/GoBlog) plugin that uses RSS and Micropub to create an archive copy on your site of a Letterboxd film diary.

## Config
```yaml
- path: ./plugins/letterboxd
    import: letterboxd
    config:
      username: "user" # Letterboxd Username
      blogURL: "http://goblog.url" # GoBlog instance URL
      section: "section" # GoBlog's Watches Section
      token: "MICROPUB-TOKEN" # GoBlog's Micropub Token
```

**Demo:** 📺 [Watches](https://kandr3s.co/watches)

---

### TO-DO

- [x] Add Microformats in Watches (Fetched from Letterboxd)
- [x] Fetch data directly from the Letterboxd feed
- [x] Implement “Rewatched”
- [x] Set up variables in GoBlog's config
- [ ] Automatically fetch movie backdrops 
