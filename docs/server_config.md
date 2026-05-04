# Server configuration file
- /etc/nps/conf/nps.conf

Name | Description
---|---
web_port | Web management port
web_password | Web UI management password
web_username | Web UI management account
web_base_url | Web management base path, used when the web UI lives behind a sub-path on a reverse proxy
bridge_port  | Server / client communication port
https_proxy_port | HTTPS listening port for domain proxies
http_proxy_port | HTTP listening port for domain proxies
auth_key | Web API key
bridge_type | Connection method between client and server: `kcp` or `tcp`
public_vkey | Verify key used when the client starts in config-file mode. Empty disables client config-file mode.
ip_limit | Whether to restrict IP access: `true`, `false`, or omit
flow_store_interval | Server traffic data persistence interval (minutes); omit to disable
log_level | Log output level
auth_crypt_key | AES key used to encrypt the server authKey, 16 bytes
p2p_ip | Server IP, required when using P2P mode
p2p_port | UDP port opened in P2P mode
pprof_ip | Debug pprof server IP
pprof_port | Debug pprof port
disconnect_timeout | Client connection timeout, in 5-second units. Default 60, i.e. 300s = 5 mins.
