# Usage
**Tip: when using the web mode, the server binary must run from the project root, otherwise the configuration file cannot be loaded.**

## Web management

Open the web UI at `<public ip>:<web port>` (default `8080`). The default password is `123`.

The web UI itself contains detailed instructions.

## Reload the server configuration
On Linux / macOS:
```shell
 sudo nps reload
```
On Windows:
```shell
 nps.exe reload
```
**Note:** only some settings can be reloaded at runtime, e.g. `allow_user_login`, `auth_crypt_key`, `auth_key`, `web_username`, `web_password`. More will be supported in the future.


## Stop or restart the server
On Linux / macOS:
```shell
 sudo nps stop|restart
```
On Windows:
```shell
 nps.exe stop|restart
```
## Update the server
Stop the server first with `sudo nps stop` or `nps.exe stop`, then:

On Linux:
```shell
 sudo nps-update update
```
On Windows:
```shell
 nps-update.exe update
```

Once the update finishes, run `sudo nps start` or `nps.exe start` again to complete the upgrade.

If the update is unsuccessful, download the release archive manually and overwrite the existing nps binary and web directory.

Note: after `nps install`, the binary is no longer in its original location. Use `whereis nps` to locate the actual path before replacing the binary.
