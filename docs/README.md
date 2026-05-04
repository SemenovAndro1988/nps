# nps
![](https://img.shields.io/github/stars/ehang-io/nps.svg)   ![](https://img.shields.io/github/forks/ehang-io/nps.svg)
[![Gitter](https://badges.gitter.im/ehang-io-nps/community.svg)](https://gitter.im/ehang-io-nps/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Build Status](https://travis-ci.org/ehang-io/nps.svg?branch=master)](https://travis-ci.org/ehang-io/nps)

NPS is a lightweight, high-performance and powerful **intranet penetration** proxy server. It currently supports **tcp and udp traffic forwarding** and can be used with any **tcp/udp** upper-layer protocol (intranet website access, local payment interface debugging, SSH access, remote desktop, intranet DNS resolution, etc.). It also **supports an intranet HTTP proxy, intranet SOCKS5 proxy**, **P2P and more**, with a powerful web management UI.


## Background
![image](https://github.com/ehang-io/nps/blob/master/image/web.png?raw=true)

1. WeChat public account or mini-program development, etc. ----> domain proxy mode


2. Reach intranet machines over SSH from the public network, mapping cloud server ports to intranet ports ----> TCP proxy mode

3. Use an intranet DNS server from outside, or access intranet machines over UDP ----> UDP proxy mode

4. Access intranet websites from the public network through an HTTP proxy ----> HTTP proxy mode

5. Set up an intranet-penetrating SSH and access intranet resources or devices from outside as if you were on a VPN ----> SOCKS5 proxy mode
