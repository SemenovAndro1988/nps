# Description
## Get the real user IP
To enable this, set `http_add_origin_header=true` in `nps.conf`.

In domain proxy mode, the real user IP can be obtained from the `X-Forwarded-For` and `X-Real-IP` request headers.

**The proxy automatically adds these two headers to every http(s) request before forwarding.**

## Hot reload
Most configuration changes made in the web UI take effect immediately and require no restart of either the client or the server.

## Client address display
The connection address of each client is displayed in the web UI.

## Traffic statistics
Traffic used by each proxy is tracked and displayed. Numbers may differ slightly from the actual figures due to compression and encryption.

## Current client bandwidth
The current bandwidth of each client is reported as a reference; small deviations from the real value are possible.

## Client / server version compatibility
The core version of the client and the server must match, otherwise the client will fail to connect to the server.

## Linux limitations
By default, Linux limits the number of connections. On a high-performance machine you can tune kernel parameters to handle more connections.
`tcp_max_syn_backlog` `somaxconn`
Adjust the values according to your needs to improve network performance.

## Web management protection
If an IP fails to log in 10 times in a row, it will be blocked from further attempts for one minute.
