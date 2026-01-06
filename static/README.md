# Static Files Directory

This directory is used for static files that are served by the Grout application.

## Common files you may place in this directory

- `robots.txt` - Robots exclusion file for web crawlers
- `sitemap.xml` - Sitemap for search engines

> Note: This repository only tracks this `README.md` in the `static/` directory (see `.gitignore`). The `robots.txt` and `sitemap.xml` files are **not** included by default — you must create them yourself if you want to use them. If these files don't exist, Grout will serve embedded default versions.

## Template Variables:

If you create `robots.txt` and/or `sitemap.xml` here, they support the `{{DOMAIN}}` placeholder, which will be replaced with the actual domain configured via the `DOMAIN` environment variable or `-domain` flag.

## Customization:

After you create these files, you can customize them by editing them directly. Changes will be picked up on the next request (files are read on each request, not cached) to allow hot-reloading without restarting the server. These files are small and typically change rarely, so this trade-off keeps configuration simple; for high-traffic deployments, you should usually place Grout behind a CDN or reverse proxy that caches `robots.txt` and `sitemap.xml` responses.

## Docker Deployment:

When deploying with Docker, this directory should be mounted as a volume to persist changes:

```yaml
volumes:
  - ./static:/app/static
```

This ensures your customizations persist across container restarts and updates.
