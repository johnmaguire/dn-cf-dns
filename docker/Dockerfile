FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-X main.version=${VERSION}" -o /nebula-dns .

FROM alpine:3

RUN apk add --no-cache ca-certificates

COPY --from=builder /nebula-dns /usr/local/bin/nebula-dns
COPY docker/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
