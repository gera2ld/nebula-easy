FROM node AS builder

WORKDIR /app
RUN git clone https://github.com/gera2ld/nebula-web.git /app && npm i pnpm -g && pnpm i && pnpm build

FROM gera2ld/nebula as nebula

FROM denoland/deno:alpine

RUN mkdir -p /app/data
COPY src /app/src
COPY --from=builder /app/build /app/dist
COPY --from=nebula /usr/local/bin/nebula* /usr/local/bin
WORKDIR /app
RUN deno check src/main.ts
CMD ["deno", "run", "-A", "src/main.ts"]
