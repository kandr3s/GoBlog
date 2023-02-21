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

- [x] Integrate Icons Plugins
- [] Move custom menus here.