# sockfwd

Forward data between sockets.

i.e. sockfwd can export unix socket via tcp port, and can export tcp socket which listen 127.0.0.1 via 0.0.0.0 host.

## Usage

```
Usage:
  sockfwd [flags]

Flags:
  -d, --destination string   destination address, data will send to this address.
  -s, --source string        source address, which will accept connection at this address.
  -q, --quiet                quiet mode.
```

## Example

Export local docker to network:`./sockfwd -s tcp://127.0.0.1:8090 -d unix:///var/run/docker.sock`.

Export `127.0.0.1:8080` to `0.0.0.0:8090`: `./sockfwd -s tcp://127.0.0.1:8090 -d unix://127.0.0.1:8090`.
