FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN ./make.sh

FROM scratch

LABEL org.opencontainers.image.vendor="酸柠檬猹Char/SourLemonJuice"
LABEL org.opencontainers.image.authors="SourLemonJuice233@outlook.com"
LABEL org.opencontainers.image.title="IPAPI-agent"
LABEL org.opencontainers.image.url="https://github.com/SourLemonJuice/ipapi-agent"
LABEL org.opencontainers.image.licenses=""

WORKDIR /

# ipapi-agent required time zone info to work
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip .
COPY --from=build /src/ipapi-agent .

ENV ZONEINFO="/zoneinfo.zip"

ENTRYPOINT ["/ipapi-agent"]
