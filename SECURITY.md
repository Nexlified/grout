# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

We recommend always using the latest version of Grout for the best security posture.

## Reporting a Vulnerability

The Grout team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them through one of the following methods:

1. **GitHub Security Advisories** (Preferred):
   - Go to the [Security tab](https://github.com/Nexlified/grout/security) of this repository
   - Click on "Report a vulnerability"
   - Fill out the form with details about the vulnerability

2. **Email**:
   - Send an email to the project maintainers through GitHub
   - Include detailed information about the vulnerability

### What to Include in Your Report

To help us better understand the nature and scope of the vulnerability, please include as much of the following information as possible:

- Type of vulnerability (e.g., XSS, CSRF, injection, etc.)
- Full paths of source file(s) related to the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

After you submit a vulnerability report, you can expect:

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
2. **Investigation**: We will investigate the issue and determine its impact and severity
3. **Updates**: We will keep you informed about the progress of addressing the vulnerability
4. **Resolution**: Once the vulnerability is confirmed, we will:
   - Develop a fix
   - Prepare a security advisory
   - Release a patched version
   - Publicly disclose the vulnerability (with credit to you, if desired)

### Disclosure Policy

- We ask that you give us reasonable time to address the vulnerability before any public disclosure
- We will work with you to understand and resolve the issue promptly
- We will credit you in the security advisory (unless you prefer to remain anonymous)
- Once a fix is released, we will publish a security advisory detailing the vulnerability

## Security Best Practices for Users

When deploying Grout, we recommend following these security best practices:

### Network Security

1. **Use HTTPS**: Always serve Grout behind a reverse proxy with TLS/SSL enabled
2. **Firewall**: Restrict access to the service port to trusted networks
3. **Rate Limiting**: Implement rate limiting at the reverse proxy level to prevent abuse
4. **DDoS Protection**: Use a CDN or DDoS protection service for public deployments

### Configuration Security

1. **Environment Variables**: Use environment variables for sensitive configuration
2. **Least Privilege**: Run the Grout container with minimal privileges
3. **Resource Limits**: Set appropriate memory and CPU limits for the container
4. **Keep Updated**: Regularly update to the latest version to receive security patches

### Input Validation

Grout includes built-in input validation and security features:

1. **Security Headers**: All HTML responses include security headers (CSP, X-Content-Type-Options, X-Frame-Options, X-XSS-Protection)
2. **Sanitize Inputs**: Validate and sanitize user inputs at the application layer
3. **Size Limits**: Set reasonable limits for image dimensions at the reverse proxy
4. **Content Security**: Built-in Content Security Policy (CSP) headers for HTML endpoints

### Monitoring and Logging

1. **Access Logs**: Enable and monitor access logs for suspicious activity
2. **Error Monitoring**: Set up alerts for unusual error rates
3. **Cache Monitoring**: Monitor cache hit rates and memory usage
4. **Security Scanning**: Regularly scan your deployment for vulnerabilities

### Docker Security

When running Grout in Docker:

1. **Non-Root User**: The Docker image runs as a non-root user by default
2. **Read-Only Filesystem**: Consider mounting the filesystem as read-only
3. **No Privileged Mode**: Never run the container in privileged mode
4. **Security Scanning**: Regularly scan the Docker image for vulnerabilities

### Example Secure Deployment

Here's an example of a secure deployment using Docker Compose behind Nginx:

```yaml
version: '3.8'

services:
  grout:
    image: nexlified/grout:latest
    restart: unless-stopped
    read_only: true
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    environment:
      - ADDR=:8080
      - CACHE_SIZE=2000
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
    networks:
      - internal

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - grout
    networks:
      - internal

networks:
  internal:
    driver: bridge
```

## Known Security Considerations

### Image Generation

- **Memory Usage**: Large image requests can consume significant memory. Use the cache size limit and configure memory limits for the container.
- **CPU Usage**: Complex rendering operations can be CPU-intensive. Consider rate limiting at the reverse proxy level.

### Caching

- **Cache Poisoning**: The ETag-based caching is based on query parameters. Ensure your reverse proxy validates inputs.
- **Memory Exhaustion**: The LRU cache has a fixed size to prevent memory exhaustion. Monitor memory usage in production.

### Input Handling

- **URL Parameters**: All user inputs from URL parameters are validated and sanitized.
- **Dimension Limits**: Invalid dimensions automatically fallback to safe defaults.
- **Color Parsing**: Invalid color values fallback to safe defaults (gray).

## Security Updates

Security updates will be released as soon as possible after a vulnerability is confirmed. Updates will be announced:

- In the [GitHub Security Advisories](https://github.com/Nexlified/grout/security/advisories)
- In the repository's [CHANGELOG.md](CHANGELOG.md)
- As GitHub releases with security tags

## Attribution

We thank the security researchers and contributors who help keep Grout secure. Security researchers who responsibly disclose vulnerabilities will be credited in our security advisories (unless they prefer to remain anonymous).

## Questions

If you have questions about this security policy, please open a discussion in GitHub Discussions or contact the maintainers.

---

Thank you for helping keep Grout and its users safe!
