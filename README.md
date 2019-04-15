# MAC to IP service

## Quickstart

1. Check environment variables defined in docker-compose
2. Run `docker-compose up`
3. The service is accessible on http://localhost:8081

## API

The following endpoints expect a request body of the form `{ "mac": "11:22:33:44:55:66" }`

### POST `/ip`
returns JSON `{ "ip": "1.2.3.4" }`

### POST `/ipxe`
returns iPXE string:

```
    #!ipxe
    chain http://drp-hostname/0.0.0.1.ipxe
```
