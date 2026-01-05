# Static Files Directory

This directory contains static files that are served by the Grout application.

## Files in this directory:

- `robots.txt` - Robots exclusion file for web crawlers
- `sitemap.xml` - Sitemap for search engines

## Template Variables:

Both `robots.txt` and `sitemap.xml` support the `{{DOMAIN}}` placeholder, which will be replaced with the actual domain configured via the `DOMAIN` environment variable or `-domain` flag.

## Customization:

You can customize these files by editing them directly. Changes will be picked up on the next request (files are read on each request, not cached).

## Docker Deployment:

When deploying with Docker, this directory should be mounted as a volume to persist changes:

```yaml
volumes:
  - ./static:/app/static
```

This ensures your customizations persist across container restarts and updates.
