List clients

```
POST /client/list/
```


| Parameter | Description |
| --- | --- |
| search | search keyword |
| order | sort: `asc` ascending, `desc` descending |
| offset | page number |
| limit | items per page |

***
Get a single client

```
POST /client/getclient/
```


| Parameter | Description |
| --- | --- |
| id | client id |

***
Add a client

```
POST /client/add/
```

| Parameter | Description |
| --- | --- |
| remark | remark |
| u | HTTP basic auth username |
| p | HTTP basic auth password |
| limit | items per page |
| vkey | client verify key |
| config\_conn\_allow | allow connection in config-file mode (1 = yes, 0 = no) |
| compress | enable compression (1 = yes, 0 = no) |
| crypt | enable encryption (1 = yes, 0 = no) |
| rate\_limit | bandwidth limit in KB/s; empty means unlimited |
| flow\_limit | traffic limit in MB; empty means unlimited |
| max\_conn | maximum concurrent connections; empty means unlimited |
| max\_tunnel | maximum tunnels; empty means unlimited |

***
Edit a client

```
POST /client/edit/
```

| Parameter | Description |
| --- | --- |
| remark | remark |
| u | HTTP basic auth username |
| p | HTTP basic auth password |
| limit | items per page |
| vkey | client verify key |
| config\_conn\_allow | allow connection in config-file mode (1 = yes, 0 = no) |
| compress | enable compression (1 = yes, 0 = no) |
| crypt | enable encryption (1 = yes, 0 = no) |
| rate\_limit | bandwidth limit in KB/s; empty means unlimited |
| flow\_limit | traffic limit in MB; empty means unlimited |
| max\_conn | maximum concurrent connections; empty means unlimited |
| max\_tunnel | maximum tunnels; empty means unlimited |
| id | id of the client to edit |

***
Delete a client

```
POST /client/del/
```

| Parameter | Description |
| --- | --- |
| id | id of the client to delete |

***
List host (domain) rules

```
POST /index/hostlist/
```

| Parameter | Description |
| --- | --- |
| search | search keyword (matches domain or remark) |
| offset | page number |
| limit | items per page |

***
Add a host (domain) rule

```
POST /index/addhost/
```


| Parameter | Description |
| --- | --- |
| remark | remark |
| host | domain |
| scheme | protocol type (`all`, `http`, or `https`) |
| location | URL routing prefix; empty means no restriction |
| client\_id | client id |
| target | intranet target (ip:port) |
| header | request header |
| hostchange | rewritten request host |

***
Edit a host (domain) rule

```
POST /index/edithost/
```

| Parameter | Description |
| --- | --- |
| remark | remark |
| host | domain |
| scheme | protocol type (`all`, `http`, or `https`) |
| location | URL routing prefix; empty means no restriction |
| client\_id | client id |
| target | intranet target (ip:port) |
| header | request header |
| hostchange | rewritten request host |
| id | id of the host rule to edit |

***
Delete a host (domain) rule

```
POST /index/delhost/
```

| Parameter | Description |
| --- | --- |
| id | id of the host rule to delete |

***
Get a single tunnel

```
POST /index/getonetunnel/
```

| Parameter | Description |
| --- | --- |
| id | tunnel id |

***
List tunnels

```
POST /index/gettunnel/
```

| Parameter | Description |
| --- | --- |
| client\_id | id of the client owning the tunnel |
| type | type: `tcp`, `udp`, `httpProx`, `socks5`, `secret`, `p2p` |
| search | search keyword |
| offset | page number |
| limit | items per page |

***
Add a tunnel

```
POST /index/add/
```

| Parameter | Description |
| --- | --- |
| type | type: `tcp`, `udp`, `httpProx`, `socks5`, `secret`, `p2p` |
| remark | remark |
| port | server port |
| target | target (ip:port) |
| client\_id | client id |

***
Edit a tunnel

```
POST /index/edit/
```

| Parameter | Description |
| --- | --- |
| type | type: `tcp`, `udp`, `httpProx`, `socks5`, `secret`, `p2p` |
| remark | remark |
| port | server port |
| target | target (ip:port) |
| client\_id | client id |
| id | tunnel id |

***
Delete a tunnel

```
POST /index/del/
```

| Parameter | Description |
| --- | --- |
| id | tunnel id |

***
Stop a tunnel

```
POST /index/stop/
```

| Parameter | Description |
| --- | --- |
| id | tunnel id |

***
Start a tunnel

```
POST /index/start/
```

| Parameter | Description |
| --- | --- |
| id | tunnel id |
