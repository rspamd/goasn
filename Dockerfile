FROM golang:1.23-alpine AS builder

RUN apk add --no-cache patch gcc musl-dev

WORKDIR /build
COPY . .

RUN go mod vendor
RUN patch -u -p1 -i gobgp.patch
RUN CGO_ENABLED=0 go build

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/goasn /usr/local/bin/goasn

RUN mkdir -p /goasn

WORKDIR /goasn

ENTRYPOINT ["goasn"]

CMD ["--debug","--download-bgp","--download-asn","--file-v4","/goasn/v4.zone","--file-v6","/goasn/v6.zone"]

VOLUME ["/root/.cache/goasn", "/goasn"]
