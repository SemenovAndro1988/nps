# Basic usage
## Ad-hoc mode (no config file)
In this mode, every setting is configured in the server web UI. The client only needs to run a single command and requires no extra configuration.
```
 ./npc -server=ip:port -vkey=<verify key shown in the web UI>
```
## Register as a system service (auto-start, daemon)
On Linux / macOS:
- Register: `sudo ./npc install <other flags, e.g. -server=xx -vkey=xx or -config=xxx>`
- Start: `sudo npc start`
- Stop: `sudo npc stop`
- To change the launch arguments, run `./npc uninstall` first and then re-register.

On Windows, run cmd as administrator:

- Register: `npc.exe install <other flags, e.g. -server=xx -vkey=xx or -config=xxx>`
- Start: `npc.exe start`
- Stop: `npc.exe stop`
- To change the launch arguments, run `npc.exe uninstall` first and then re-register.
- To restart the client automatically when it exits, configure it as shown below.
![image](https://github.com/ehang-io/nps/blob/master/docs/windows_client_service_configuration.png?raw=true)

After being registered as a service, the log file is created in the current directory on Windows; on Linux / macOS it lives at /var/log/npc.log.

## Update the client
First, `cd` to the directory that contains the npc binary.

Stop the running service with `sudo npc stop` or `npc.exe stop`, then:

On Linux:
```shell
 sudo npc-update update
```
On Windows:
```shell
npc-update.exe update
```

Once the update finishes, run `sudo npc start` or `npc.exe start` again to complete the upgrade.

If the update is unsuccessful, download the release archive manually and replace the existing npc binary.

## Config-file mode
This mode authenticates with the NPS public key or the client private key. Settings are configured on the client side and can also be managed from the server's web UI.
```
 ./npc -config=<path to npc config file>
```
## Config file reference
[Sample config file](https://github.com/ehang-io/nps/tree/master/conf/npc.conf)
#### Global section
```ini
[common]
server_addr=1.1.1.1:8024
conn_type=tcp
vkey=123
username=111
password=222
compress=true
crypt=true
rate_limit=10000
flow_limit=100
remark=test
max_conn=10
#pprof_addr=0.0.0.0:9999
```
Item | Description
---|---
server_addr | server ip/domain:port
conn_type | communication mode with the server (`tcp` or `kcp`)
vkey | verify key from the server config file (not the web UI)
username | basic-auth username for SOCKS5 / HTTP(S) (optional)
password | basic-auth password for SOCKS5 / HTTP(S) (optional)
compress | enable transport compression (`true`, `false`, or omit)
crypt | enable transport encryption (`true`, `false`, or omit)
rate_limit | rate limit (optional)
flow_limit | traffic limit (optional)
remark | client remark (optional)
max_conn | maximum concurrent connections (optional)
pprof_addr | debug pprof ip:port

#### Domain proxy

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[web1]
host=a.proxy.com
target_addr=127.0.0.1:8080,127.0.0.1:8082
host_change=www.proxy.com
header_set_proxy=nps
```
Item | Description
---|---
web1 | section name / remark
host | domain (resolves both http and https)
target_addr | intranet targets; comma-separated for load balancing
host_change | rewrite the request `Host` header
header_xxx | add or modify a request header. `header_proxy` adds the header `proxy: nps`.

#### TCP tunnel

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[tcp]
mode=tcp
target_addr=127.0.0.1:8080
server_port=9001
```
Item | Description
---|---
mode | tcp
server_port | proxy port on the server
target_addr | intranet target

#### UDP tunnel

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[udp]
mode=udp
target_addr=127.0.0.1:8080
server_port=9002
```
Item | Description
---|---
mode | udp
server_port | proxy port on the server
target_addr | intranet target
#### HTTP proxy mode

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[http]
mode=httpProxy
server_port=9003
```
Item | Description
---|---
mode | httpProxy
server_port | proxy port on the server
#### SOCKS5 proxy mode

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[socks5]
mode=socks5
server_port=9004
multi_account=multi_account.conf
```
Item | Description
---|---
mode | socks5
server_port | proxy port on the server
multi_account | path to a SOCKS5 multi-account file (optional). When set, `basic_username` / `basic_password` cannot pass authentication.
#### Secret proxy mode

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[secret_ssh]
mode=secret
password=ssh2
target_addr=10.1.50.2:22
```
Item | Description
---|---
mode | secret
password | unique password
target_addr | intranet target

#### P2P proxy mode

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[p2p_ssh]
mode=p2p
password=ssh2
target_addr=10.1.50.2:22
```
Item | Description
---|---
mode | p2p
password | unique password
target_addr | intranet target


#### File access mode
NPS provides a publicly accessible local file service. This mode is only available when the client is started in config-file mode.

```ini
[common]
server_addr=1.1.1.1:8024
vkey=123
[file]
mode=file
server_port=9100
local_path=/tmp/
strip_pre=/web/
````

Item | Description
---|---
mode | file
server_port | port to open on the server
local_path | local directory served by the client
strip_pre | URL prefix that maps to `local_path`

With `strip_pre`, accessing `ip:9100/web/` from the public network is equivalent to browsing the `/tmp/` directory.

#### Auto-reconnect
```ini
[common]
auto_reconnection=true
```
