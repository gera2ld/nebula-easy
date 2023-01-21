FROM gera2ld/nebula as nebula

FROM alpine

ARG TARGETOS
ARG TARGETARCH

COPY bin/nebula-easy-${TARGETOS}-${TARGETARCH} /usr/local/bin/nebula-easy
COPY --from=nebula /usr/local/bin/nebula* /usr/local/bin
CMD ["nebula-easy"]
