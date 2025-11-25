FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# cross compile is more fast
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0
RUN go build -v -o ipapi-agent .

FROM alpine:3.22

LABEL org.opencontainers.image.vendor="酸柠檬猹Char/SourLemonJuice"
LABEL org.opencontainers.image.authors="SourLemonJuice233@outlook.com"
LABEL org.opencontainers.image.title="IPAPI-agent"
LABEL org.opencontainers.image.url="https://github.com/SourLemonJuice/ipapi-agent"
LABEL org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /

RUN apk add --no-cache tzdata ca-certificates

COPY --from=build /usr/src/app/ipapi-agent ./

ENTRYPOINT ["/ipapi-agent"]
