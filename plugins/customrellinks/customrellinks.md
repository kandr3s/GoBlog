A plugin that adds Microsub/Manifest `rel` links to GoBlog.

```YAML
- path: ./plugins/customrellinks
    import: customrellinks
    config:
      blog: # Name of the blog
        - name: microsub # Name of Rel-Link
          link: https://mymicrosub.tld/user # Rel-Link
        - name: manifest # Defaults to /manifest.json
```