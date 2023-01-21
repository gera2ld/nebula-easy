FROM node AS node-builder

WORKDIR /app
COPY nebula-web /app
RUN npm i pnpm -g && pnpm i && pnpm build

FROM golang:alpine AS go-builder

WORKDIR /app
COPY go.mod main.go /app
COPY --from=node-builder /app/build /app/dist
RUN go build -ldflags '-s -w' -o /usr/local/bin/nebula-easy

FROM gera2ld/nebula as nebula

FROM alpine

COPY --from=go-builder /usr/local/bin/nebula-easy /usr/local/bin
COPY --from=nebula /usr/local/bin/nebula* /usr/local/bin
WORKDIR /app
CMD ["nebula-easy"]
