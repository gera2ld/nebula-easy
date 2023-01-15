# NebulaEasy

This is a dashboard for [nebula](https://github.com/slackhq/nebula).

## Usage

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
