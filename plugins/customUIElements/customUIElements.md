A GoBlog plugin to manage custom UI elements.

```YAML
- path: ./plugins/customUIElements
    import: customUIElements
    config:
      blog: # Name of the blog
        - name: microsub # Element name
          link: https://mymicrosub.tld/user # Microsub element link
        - name: manifest # Defaults to /manifest.json
        - name: socialicons
        - name: indiewebicons
```

### TODO

- [x] Integrate Social Icons Plugins
- [x] Timeline style
- [x] Section Emoji/Icons
- [] Move custom menus here.