FROM alpine

ARG TARGETOS
ARG TARGETARCH

RUN apk add curl --no-cache

RUN VERSION=$(curl -fsSI https://github.com/slackhq/nebula/releases/latest | sed -n '/tag/s/.*\/v\(.*\)/\1/p' | tr -d \\r | tr -d \\n) \
  && TARBALL=nebula.tar.gz \
  && URL=https://github.com/slackhq/nebula/releases/download/v${VERSION}/nebula-${TARGETOS}-${TARGETARCH}.tar.gz \
  && echo Download $URL \
  && curl -fsSLo $TARBALL $URL \
  && mkdir /nebula \
  && tar xf $TARBALL -C /nebula \
  && mv /nebula/nebula* /usr/local/bin

WORKDIR /nebula

CMD ["nebula"]
