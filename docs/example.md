# Examples
## Common preparation (required for all modes)
- Start the server. Assume the public server IP is 1.1.1.1, `bridge_port` is 8024 in the config file, and `web_port` is 8080.
- Open 1.1.1.1:8080.
- In the client management page, create a client and write down the verify key.
- On the intranet, run the client (on Windows use cmd and the `.exe` binary):

```shell
./npc -server=1.1.1.1:8024 -vkey=<client verify key>
```
**Note:** after the server is started, make sure that the `bridge_port` configured on the server is reachable from the client device (verify with `telnet`, `netcat`, etc.).

## Domain proxy

**Use case:** mini-program development, WeChat public account development, product demo.

**Note:** the domain proxy mode is an HTTP reverse proxy (it is not a DNS server) and can be configured easily through the web UI.

**Scenario:**
- You own the domain `proxy.com` and a public server with IP 1.1.1.1.
- Two intranet sites: 127.0.0.1:81 and 127.0.0.1:82.
- You want `(http|https://)a.proxy.com` to reach 127.0.0.1:81 and `(http|https://)b.proxy.com` to reach 127.0.0.1:82.

**Steps**
- Resolve `*.proxy.com` to the public server 1.1.1.1.
- Open the host management page of the client you just created and add two rules: 1) host: `a.proxy.com`, target: `127.0.0.1:81`; 2) host: `b.proxy.com`, target: `127.0.0.1:82`.

You can now successfully visit `(http|https://)a.proxy.com` and `b.proxy.com`.

**HTTPS:** if you want to use HTTPS, please follow the additional configuration in [Use HTTPS](/nps_extend).

## TCP tunnel


**Use case:** SSH, remote desktop, and other TCP scenarios.

**Scenario:**
You want to access port 22 of intranet machine 10.1.50.101 by connecting to port 8001 of public server 1.1.1.1 to perform SSH access.

**Steps**
- In the tunnel management of the client you just created, add a TCP tunnel: listen port (8001), target IP and port (10.1.50.101:22), then save.
- Connect to public server 1.1.1.1 on the listen port (8001), which is equivalent to connecting to 10.1.50.101:22, e.g. `ssh -p 8001 root@1.1.1.1`.

## UDP tunnel

**Use case:** intranet DNS resolution and other UDP scenarios.

**Scenario:**
There is an intranet DNS server (10.1.50.102:53) that you want to use from outside the intranet. The public server is 1.1.1.1.

**Steps**
- In the tunnel management of the client you just created, add a UDP tunnel: listen port (53), target IP and port (10.1.50.102:53), then save.
- Set the local DNS server to 1.1.1.1, which is equivalent to using 10.1.50.102 as the DNS server.

## SOCKS5 proxy


**Use case:** access intranet devices or resources from outside as if you were on a VPN.

**Scenario:**
Use port 8003 of public server 1.1.1.1 as a SOCKS5 proxy to access any intranet device or resource.

**Steps**
- In the tunnel management of the client you just created, add a SOCKS5 proxy: listen port (8003), then save.
- On the external machine, configure a SOCKS5 proxy (e.g. with Proxifier as a global proxy), pointing at the public server IP (1.1.1.1) and the listen port (8003); you now have full intranet access.

**Note**
After the SOCKS5 proxy accepts a SOCKS5 packet, the underlying socket is already in `accept` state. As a result, port scans show every port as open and connections are closed shortly after. If you need behaviour identical to running on the intranet, connect to a remote device.

## HTTP forward proxy

**Use case:** access intranet websites from outside via an HTTP forward proxy.

**Scenario:**
Use port 8004 of public server 1.1.1.1 as an HTTP proxy to access intranet websites.

**Steps**

- In the tunnel management of the client you just created, add an HTTP proxy: listen port (8004), then save.
- On the external machine, configure an HTTP proxy with IP set to the public server IP (1.1.1.1) and the listen port (8004) to start browsing.

**Note: for the secret proxy and the P2P proxy, in addition to the unified server/client configuration, you also need a client acting as the access side that exposes a port for connections.**

## Secret proxy

**Use case:** TCP services with high security requirements that should not occupy extra ports and that should prevent other people from connecting, e.g. SSH.

**Scenario:**
Reach port 22 of intranet server 10.1.50.2 without opening a new port.

**Steps**
- Add a secret proxy entry to the client you just created and set a unique password `secrettest` and the intranet target `10.1.50.2:22`.
- On the machine that needs the SSH connection, run

```
./npc -server=1.1.1.1:8024 -vkey=vkey -type=tcp -password=secrettest -local_type=secret
```
To use a custom local port, append `-local_port=xx`. The default is 2000.

**Note:** `password` is the unique key configured in the web UI; the exact command is shown in the web UI command hint.

Assuming the SSH user on 10.1.50.2 is `root`, run `ssh -p 2000 root@127.0.0.1` to reach SSH.


## P2P service

**Use case:** large traffic, where data does not pass through the public server. Because P2P punching depends heavily on the NAT type, success is not guaranteed for every NAT, but most NAT types are supported. [NAT type detection](/npc_extend)

**Scenario:**

You want to access port 22 of intranet machine 10.2.50.2 from your local (access) machine on port 2000.

**Steps**
- In `nps.conf`, set `p2p_ip` (NPS server IP) and `p2p_port` (NPS server UDP port).
> Note: if `p2p_port` is 6000, open UDP ports 6000-6002 (two extra ports) on the firewall.
- Add a P2P proxy entry to the client you just created and set a unique password `p2pssh`.
- On the access machine, run

```
./npc -server=1.1.1.1:8024 -vkey=123 -password=p2pssh -target=10.2.50.2:22
```
To use a custom local port, append `-local_port=xx`. The default is 2000.

**Note:** `password` is the unique key configured in the web UI; the exact command is shown in the web UI command hint.

Assuming the SSH user on intranet machine 10.2.50.2 is `root`, run `ssh -p 2000 root@127.0.0.1` to reach machine 2's SSH. For a website you can visit `127.0.0.1:2000` in your browser.
