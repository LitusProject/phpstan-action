FROM golang:1.22.0-alpine AS golang

RUN apk add --no-cache \
  git

WORKDIR /go/src
COPY . .

ENV CGO_ENABLED=0
RUN go get -d -v ./... && \
  go install -v ./...

FROM ghcr.io/litusproject/phpstan:latest AS phpstan

COPY --from=golang /go/bin/phpstan-action /usr/bin/

WORKDIR /
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
