FROM golang:1.27-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go/plataforma/go.mod go/plataforma/go.sum ./
RUN go mod download

COPY go/plataforma/ ./

ARG TARGET=api
RUN mkdir -p /out && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${TARGET}

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget \
    && adduser -D -H app

COPY --from=build /out/app /usr/local/bin/app
COPY --from=build /src/scripts/node-loop.sh /usr/local/bin/node-loop.sh

USER app
ENTRYPOINT ["/usr/local/bin/app"]