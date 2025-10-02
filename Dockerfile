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

FROM scratch

LABEL org.opencontainers.image.vendor="酸柠檬猹Char/SourLemonJuice"
LABEL org.opencontainers.image.authors="SourLemonJuice233@outlook.com"
LABEL org.opencontainers.image.title="IPAPI-agent"
LABEL org.opencontainers.image.url="https://github.com/SourLemonJuice/ipapi-agent"
LABEL org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /

# ipapi-agent required time zone info to work
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip ./
COPY --from=build /usr/src/app/ipapi-agent ./

ENV ZONEINFO="/zoneinfo.zip"

ENTRYPOINT ["/ipapi-agent"]
