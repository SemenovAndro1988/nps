# Advanced features
## Use HTTPS

**Option 1:** Terminate HTTPS at NPS, similar to nginx.

In the configuration file, set `https_proxy_port` to 443 (or any other port) and set `https_just_proxy=false`. Restart NPS, then upload the certificate and key for each domain via the host add/edit page in the web UI.

**In addition:** you can specify a default HTTPS configuration in `nps.conf`. When a request hits a domain that has no certificate configured, NPS uses the default certificate. The default certificate is also used when the client's hello does not carry an SNI extension.


**Option 2:** Run HTTPS on the intranet server.

In `nps.conf`, set `https_just_proxy=true` and open `https_proxy_port`. NPS will forward the HTTPS request directly to the intranet server, which performs the TLS handshake itself.

## Combine with nginx

Sometimes you want to keep nginx on the cloud server (for example, for static-file caching). NPS works well together with nginx: set `httpProxyPort` to a non-80 port and configure nginx as a reverse proxy. With `httpProxyPort=8010`:
```
server {
    listen 80;
    server_name *.proxy.com;
    location / {
        proxy_set_header Host  $http_host;
        proxy_pass http://127.0.0.1:8010;
    }
}
```
For HTTPS, listen on 443 in nginx, configure SSL, and disable HTTPS in NPS by leaving `httpsProxyPort` empty. With `httpProxyPort=8020`:

```
server {
    listen 443;
    server_name *.proxy.com;
    ssl on;
    ssl_certificate  certificate.crt;
    ssl_certificate_key private.key;
    ssl_session_timeout 5m;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;
    ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
    ssl_prefer_server_ciphers on;
    location / {
        proxy_set_header Host  $http_host;
        proxy_pass http://127.0.0.1:8020;
    }
}
```
## HTTPS for the web UI
If the web UI itself should be served over HTTPS, set `web_open_ssl=true` in `nps.conf` and configure `web_cert_file` and `web_key_file`.
## Front the web UI with Caddy

You can front the NPS web UI with Caddy on a sub-path.

To make `http://caddy_ip:caddy_port/nps` reach the NPS UI, configure Caddyfile:

```Caddyfile
caddy_ip:caddy_port/nps {
  ## server_ip is the NPS server IP
  ## web_port is the NPS web port
  proxy / http://server_ip:web_port/nps {
	transparent
  }
}
```

Then set `web_base_url=/nps` in `nps.conf`:
```
web_base_url=/nps
```


## Disable a proxy

Leave `http_proxy_port` empty to disable the HTTP proxy and `https_proxy_port` empty to disable the HTTPS proxy.

## Persist traffic data
The server can persist traffic data to disk. This is disabled by default. Set the `flow_store_interval` parameter (in minutes) in `nps.conf` to enable it.

**Note:** NPS does not persist clients that connected through the public key.
## System information display
NPS can display server statistics in the web UI, but some charts are disabled by default. To enable them, set `system_info_display=true` in `nps.conf`.

## Custom client connection key
Each client can have a custom verify key configured in the web UI, as long as it is unique.
## Disable public-key access
Leave `public_vkey` empty in `nps.conf` (or delete it).

## Disable the web UI
Leave `web_port` empty in `nps.conf` (or delete it).

## Multi-user login on the server
If `allow_user_login=true` in `nps.conf`, the server's web UI supports multi-user login. The login username is `user` and the default password is the verify key of each client. After logging in, the user can enter the client edit page and change the web login username and password. The feature is disabled by default.

## User registration
The NPS server supports user registration. Set `allow_user_register=true` in `nps.conf` to add a registration form to the login page.

## Listen on a specific IP

NPS supports listening on a different server IP per tunnel. Set `allow_multi_ip=true` in `nps.conf`. The IP can then be controlled in the web UI, or in the npc config file (optional; defaults to `0.0.0.0`):
```ini
server_ip=xxx
```
## Proxy to the server's local services
When NPS listens on port 80 or 443, every request is forwarded to the intranet by default. Sometimes NPS-hosted services on the same VPS also need those ports. NPS supports forwarding to local services, similar to nginx's `proxy_pass`. This feature works with domain proxies as well as TCP and UDP tunnels and is disabled by default.

**Example:** there is a service on the NPS VPS that listens on port 5000, but NPS occupies ports 80 and 443. You want to reach the service on port 5000 over HTTP(S) using a domain name.

**Usage:** set `allow_local_proxy=true` in `nps.conf`, then in the web UI configure the tunnel or host you want to forward and enable the "proxy to local" option.
