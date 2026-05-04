# Advanced features
## NAT type detection
```
 ./npc nat -stun_addr=stun.stunprotocol.org:3478
```
P2P will not work between two Symmetric NATs; other combinations have a high success rate. The `stun_addr` flag specifies the STUN server.
## Status check
```
 ./npc status -config=<path to npc config>
```
## Reload the config file
```
 ./npc restart -config=<path to npc config>
```

## Connect to NPS through a proxy
Sometimes the intranet machine running npc cannot reach the public internet directly. In that case you can connect to NPS through a SOCKS5 proxy.

In config-file mode:
```ini
[common]
proxy_url=socks5://111:222@127.0.0.1:8024
```
In ad-hoc mode, use the flag:

```
-proxy=socks5://111:222@127.0.0.1:8024
```
Both SOCKS5 and HTTP proxies are supported.

That is `socks5://username:password@ip:port`

or `http://username:password@ip:port`.
