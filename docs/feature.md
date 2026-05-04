# Advanced features
## Cache
For websites, static files often consume the most bandwidth, and in intranet penetration scenarios static files have to be fetched from the client first, which costs even more traffic. NPS supports caching of static files in the domain proxy mode.

For example, if a site has `a.css`, NPS reads the file from the npc client only once, stores it in memory and returns the cached content for subsequent requests instead of asking the client again. The feature is disabled by default. To enable it, set `http_cache=true` in `nps.conf` and configure `http_cache_length` (the number of cached files; this consumes memory, so do not set it too high; `0` means unlimited).

## Compression

Because of intranet penetration, a lot of data is exchanged between the intranet client and the server. To reduce traffic and speed up transfer, NPS supports SNAPPY compression.

- Compression is supported in every mode.
- Configure it in the web UI or in the client config file.


## Encrypted transport

If your corporate firewall identifies and blocks certain protocols (e.g. SSH), enabling encrypted transport between the server and the client can effectively prevent traffic from being intercepted.
- NPS now generates a random TLS certificate at startup by default for encrypted transport.



## Site protection
In the domain proxy mode, every client shares the same HTTP service port and anyone who knows the domain can access it. Some development or test environments require confidentiality, so you can configure a username and password and NPS will protect access through HTTP Basic Auth.


- Configure it in the web UI or in the client config file.

## Host modification

Because the host expected by an intranet site may differ from the public domain, the domain proxy supports host modification, which rewrites the `Host` header of the request.

**Usage:** configure it in the web UI.

## Custom headers

Headers can be added or modified to suit the service.

## 404 page configuration
The domain proxy supports a custom 404 page; just edit `/web/static/page/error.html`. Static assets in that page are not currently supported.

## Traffic limit

Per-client traffic limits are supported. When the sum of inbound and outbound traffic for a client reaches the configured quota, the client is rejected: domain proxies serve the 404 page and other proxies refuse the connection. To enable this, set `allow_flow_limit=true` in `nps.conf` (disabled by default).

## Bandwidth limit

Per-client bandwidth limits are supported. The bandwidth is the sum of inbound and outbound traffic with equal weight. To enable this, set `allow_rate_limit=true` in `nps.conf` (disabled by default).

## Load balancing
NPS supports load balancing for the domain proxy and for the TCP proxy. In the host or tunnel form in the web UI, fill in multiple targets on separate lines to enable round-robin load balancing.

## Allowed-port whitelist
To prevent ports on the server from being abused, you can configure `allow_ports` in `nps.conf` to restrict which ports may be opened. Omitting or leaving the value empty disables the limit. Format:

```ini
allow_ports=9001-9009,10001,11000-12000
```

## Port range mapping
When the client is started in config-file mode, you can map a range of local ports. Only TCP and UDP modes are supported, e.g.:

```ini
[tcp]
mode=tcp
server_port=9001-9009,10001,11000-12000
target_port=8001-8009,10002,13000-14000
```

The values are comma-separated and may be single ports or ranges. The lengths must match exactly between server and target, otherwise the mapping will fail.
## Port range mapping to other machines
```ini
[tcp]
mode=tcp
server_port=9001-9009,10001,11000-12000
target_port=8001-8009,10002,13000-14000
target_ip=10.1.50.2
```
When `target_ip` is set, the ports map to that machine; if it is omitted, NPS maps to the local 127.0.0.1. Only valid for range mappings.

## KCP support

When the network quality is excellent (e.g. dedicated lines, intranet), enabling KCP can slightly reduce latency. To use it, set `bridge_type=kcp` in `nps.conf`. NPS will then open a UDP port (`bridge_port`).

Note: when the server uses KCP, the client must use the same setting. In ad-hoc mode, append `type=kcp`; in config-file mode, set `tp=kcp` in the config file.

## Wildcard domain
Wildcard domains are supported. For instance, set the host to `*.proxy.com` and `a.proxy.com`, `b.proxy.com`, etc., will all resolve to the same target. Configure this in the web UI or in the client config file.

## URL routing
NPS can forward requests with the same domain to different intranet servers based on the URL. Configure this in the web UI or in the client config file (the parameter is optional). Example:

```ini
[web1]
host=a.proxy.com
target_addr=127.0.0.1:7001
location=/test
[web2]
host=a.proxy.com
target_addr=127.0.0.1:7002
location=/static
```
Requests to `a.proxy.com/test` are forwarded to `web1`, and `a.proxy.com/static` to `web2`.

## IP access restriction
Exposing high-risk ports such as SSH on the public internet introduces risk. NPS supports restricting access by IP.

**Usage:** set `ip_limit=true` in `nps.conf`. After enabling it, only registered IPs may access NPS proxies.

**Register IPs:**

**Method 1:**
On the machine that needs access, run the client:

```
./npc register -server=ip:port -vkey=<public or client vkey> time=2
```

`time` is the number of hours the access is valid. For instance `time=2` allows the current public IP to access NPS proxies for the next two hours.

**Method 2:**
Logging in to the NPS web UI also acts as an authentication step: a successful login grants the logged-in IP two hours of access.


**Note:** the public IP is not permanent. Choose a sensible validity period, and bear in mind that several people on the same network may share a single public IP.
## Maximum connections per client
To prevent malicious long-lived connections from affecting server stability, the maximum number of concurrent connections can be configured per client in the web UI or client config file. The limit applies to `socks5`, the `HTTP forward proxy`, the `domain proxy`, the `TCP proxy`, the `UDP proxy` and the `secret proxy`. To enable this, set `allow_connection_num_limit=true` in `nps.conf` (disabled by default).

## Maximum tunnels per client
NPS supports limiting the number of tunnels a single client can create. This is disabled by default; to enable it, set `allow_tunnel_num_limit=true` in `nps.conf`.
## Port reuse
In strict network environments where the number of available ports is small, NPS supports powerful port-reuse capabilities. `bridge_port`, `http_proxy_port`, `https_proxy_port` and `web_port` can be configured to share the same port and still work correctly.

- Set the ports you want to reuse to the same value as `bridge_port`. NPS auto-detects the protocol.
- If you also want to reuse the web management port, set `web_host` (a sub-domain) so the requests can be distinguished.

## Multiplexing

Multiplexing is enabled by default for the main NPS communication channel; no extra configuration is required.

The multiplexing implementation is inspired by the TCP sliding-window mechanism: it dynamically computes the latency and bandwidth to decide how much data to push down the network pipe.
Because the main channel is mostly TCP, NPS cannot directly observe packet loss in real time. To allow some loss-induced retransmissions, NPS uses a 5-minute tolerance: if it does not see traffic within that window, the current tunnel is closed and re-established, dropping all current connections.
On Linux, you can tune kernel parameters to adapt to different scenarios.

For workloads that need high bandwidth despite some packet loss, keep the defaults to reduce dropped connections.
For high concurrency, follow the [Linux limitations](## Linux limitations) tuning advice.

For latency-sensitive workloads with some loss, adjust the TCP retry counts:
`tcp_syn_retries`, `tcp_retries1`, `tcp_retries2`.
For high concurrency, follow the same advice as above.
NPS detects the error returned when the system actively closes a connection and re-establishes the tunnel.

## Environment variable templating
npc supports rendering environment variables to fit certain special scenarios.

**In ad-hoc (no config file) mode:**
Set environment variables:
```
export NPC_SERVER_ADDR=1.1.1.1:8024
export NPC_SERVER_VKEY=xxxxx
```
Run `./npc` directly.

**In config-file mode:**
```ini
[common]
server_addr={{.NPC_SERVER_ADDR}}
conn_type=tcp
vkey={{.NPC_SERVER_VKEY}}
auto_reconnection=true
[web]
host={{.NPC_WEB_HOST}}
target_addr={{.NPC_WEB_TARGET}}
```
Reference environment variables in the config file and npc will automatically substitute them at startup.

## Health check

When the client is started in config-file mode, multi-node health checks are supported. Example:

```ini
[health_check_test1]
health_check_timeout=1
health_check_max_failed=3
health_check_interval=1
health_http_url=/
health_check_type=http
health_check_target=127.0.0.1:8083,127.0.0.1:8082

[health_check_test2]
health_check_timeout=1
health_check_max_failed=3
health_check_interval=1
health_check_type=tcp
health_check_target=127.0.0.1:8083,127.0.0.1:8082
```
**The keyword `health` must appear at the start of the section name.**

The first variant is HTTP mode: NPS performs a GET to target+url and a 200 response counts as success.

The second variant is TCP mode: NPS opens a TCP connection to the target and a successful connection counts as success.

If the number of failures exceeds `health_check_max_failed`, NPS removes the target from the npc; once the target comes back online, NPS automatically re-adds it.

Item | Description
---|---
health_check_timeout | Health check timeout
health_check_max_failed | Allowed number of failures before removal
health_check_interval | Health check interval
health_check_type | Health check type
health_check_target | Health check targets, comma separated
health_check_type | Health check type
health_http_url | Health check URL (only used in HTTP mode)

## Logging

Log level

**For npc:**
```
-log_level=0~7 -log_path=npc.log
```
```
LevelEmergency->0  LevelAlert->1

LevelCritical->2 LevelError->3

LevelWarning->4 LevelNotice->5

LevelInformational->6 LevelDebug->7
```
The default level is fully verbose (0 to 7).

**For nps:**

Configure it in `nps.conf`.

## pprof profiling and debugging

You can enable pprof ports in the server and client configurations for profiling and debugging. Comment them out or leave them empty to disable.

Disabled by default.

## Custom client disconnect timeout

The client and server exchange a latency measurement packet every 5 seconds; this interval is fixed.
The number of consecutive missing replies before the client connection is closed can be tuned (default 60, i.e. five minutes without a reply).
Note that the client must close its socket before reconnecting; if the client cannot receive the server's FIN packet, only the client itself can close the socket.
For example, if the server uses a low value while the client uses a high one, and the server closes the connection but the client never receives the FIN, the client keeps waiting until its own timeout fires.

Set `disconnect_timeout` in `nps.conf` or `npc.conf`. The client also accepts a `-disconnect_timeout=60` flag.
