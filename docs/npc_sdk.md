# npc SDK reference

```
Start the client in command-line mode.
Since v0.26.10, this function blocks until the client exits. The caller is responsible for handling reconnection.
p0 -> server address
p1 -> vkey
p2 -> connection type (tcp or udp)
p3 -> proxy URL

extern GoInt StartClientByVerifyKey(char* p0, char* p1, char* p2, char* p3);

Returns the status of the currently started client: 1 for online, 0 for offline.
extern GoInt GetClientStatus();

Close the client.
extern void CloseClient();

Return the client version.
extern char* Version();

Return logs, updated in real time.
extern char* Logs();
```
