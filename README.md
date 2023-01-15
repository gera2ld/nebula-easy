# NebulaEasy

This is a dashboard for [nebula](https://github.com/slackhq/nebula).

Note: It only helps you to manage devices and generate the configuration and certificates. But you still have to deploy the files on your devices by yourself.

## Usage

### Dashboard

Create a `docker-compose.yml`:

```yml
version: '3'

services:
  nebula-easy:
    image: gera2ld/nebula-easy:latest
    ports:
      - '4000:4000'
    volumes:
      - './data:/app/data'
```

Then visit `http://localhost:4000` and manage your networks and hosts. All data (including CA) will be stored in `./data/db.json`.

Note:

- The dashboard only helps you to generates configuration and certificates, the whole process can be done offline.
- The dashboard should not be exposed to the public because it contains the root CA. One choice is to deploy it on your personal laptop.

### Lighthouse / Host

After creating a network and devices, download the configuration for each device and copy the the corresponding device.

Assuming `./config` is the directory containing the configuration and certificates, we can deploy nebula using a `docker-compose.yml` with the following content:

```yml
version: '3'

services:
  nebula:
    image: gera2ld/nebula:latest
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun
    ports:
      - 4242:4242/udp
    volumes:
      - ./config:/etc/nebula
    command: nebula -config /etc/nebula/config.yml
```

## Links

- Source code of the web pages: https://github.com/gera2ld/nebula-web
