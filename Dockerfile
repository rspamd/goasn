FROM golang:1.23-alpine AS builder

RUN apk add --no-cache patch gcc musl-dev

WORKDIR /build
COPY ./ .

RUN go mod vendor
RUN patch -u -p1 -i gobgp.patch
RUN CGO_ENABLED=0 go build

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN update-ca-certificates

RUN addgroup -g 1001 -S goasn && adduser -u 1001 -S goasn -G goasn

COPY --from=builder /build/goasn /usr/local/bin/goasn

RUN mkdir -p /goasn && chown goasn:goasn /goasn

WORKDIR /goasn

USER goasn

ENTRYPOINT ["goasn"]

CMD ["--cache-dir","/goasn/cache","--download-bgp","--download-asn","--on-update-only","--file-v4","/goasn/zones/asn.rspamd.com_ip4trie","--file-v6","/goasn/zones/asn6.rspamd.com_ip6trie","--zone-tmp-ext","__tmp"]

VOLUME ["/goasn"]
