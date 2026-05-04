# Web API

To enable the API, uncomment `auth_key` in `nps.conf` and configure a suitable secret.
## Web API authentication
- Authentication is based on `auth_key`.
- Append two parameters, `auth_key` and `timestamp`, to every request you send.

```
auth_key is generated as: md5(auth_key from the configuration file + current timestamp)
```

```
timestamp is the current Unix timestamp
```
```
curl --request POST \
  --url http://127.0.0.1:8080/client/list \
  --data 'auth_key=2a0000d9229e7dbcf79dd0f5e04bb084&timestamp=1553045344&start=0&limit=10'
```
**Note:** for security reasons, the timestamp is only valid for 20 seconds, so it must be regenerated for every request.

## Get the server time
Since the time difference between the server and the API client must be small, a dedicated endpoint is provided to obtain the server time.

```
POST /auth/gettime
```

## Get the server authKey

If you want to obtain the authKey, the server exposes the following endpoint.

```
POST /auth/getauthkey
```
This returns the encrypted authKey using AES-CBC. Decrypt it with the same key configured as `auth_crypt_key` in the server configuration.

**Note:** `auth_crypt_key` in the nps configuration must be 16 bytes long.
- Decryption key length: 128 bits
- IV equals the key
- Padding: pkcs5padding
- Encoding of the encrypted string: hexadecimal

## Detailed documentation
- **[Details](webapi.md)** (thanks @avengexyz)
