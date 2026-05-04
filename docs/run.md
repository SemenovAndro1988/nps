# Run
## Server
After downloading the server archive, unzip it and `cd` into the extracted folder.

- Run the install command

On Linux / macOS: ```sudo ./nps install```

On Windows, run cmd as administrator and enter the installation directory: ```nps.exe install```

- Start

On Linux / macOS: ```sudo nps start```

On Windows, run cmd as administrator, enter the program directory: ```nps.exe start```

```After installation, the Windows config file is located at C:\Program Files\nps; on Linux / macOS it is at /etc/nps```

`stop` and `restart` are also available.

**If the server does not start successfully, run `nps(.exe) stop`, then run `nps(.exe)` directly to debug, or check the log** (the Windows log is in the current directory; on Linux / macOS it is at /var/log/nps.log).
- Access the server at server-ip:web-port (default 8080).
- Log in with the default username and password (admin/123). **You must change them before deploying to production.**
- Create a client.

## Client
- Download the client archive and extract it, then `cd` into the extracted directory.
- In the web UI, click the `+` icon in front of the client to copy the startup command.
- Run the startup command. On Linux you can run it directly. On Windows, replace `./npc` with `npc.exe` and run from **cmd**.

If you run it from `powershell`, **wrap the IP address in quotes!**

If you want to register npc as a system service, see [Register as a system service](/use).

## Version check
- Both the server and the client accept the `-version` flag to print the version.
- `nps -version` or `./nps -version`
- `npc -version` or `./npc -version`

## Configuration
- After the client connects, configure the desired penetration services in the web UI.
- See the [Examples](/example).
