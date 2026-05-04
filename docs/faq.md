# FAQ

- The server does not start.
```
By default, the server uses ports 8024, 8080, 80 and 443. If any port conflicts, the server will not start. Adjust the configuration accordingly.
```
- The client cannot connect to the server.
```
Make sure all ports in the config file are allowed by your security groups and firewall.
Make sure the vkey matches.
Make sure the client and server versions match.
```
- Changes to the server config file have no effect.
```
After "install", the Linux config file is located at /etc/nps.
```
- P2P punching fails [P2P service](https://ehang-io.github.io/nps/#/example)
```
If both peers are behind a Symmetric NAT, P2P will not succeed. Please check the NAT type first and follow the documentation linked above.
```
