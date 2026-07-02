FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cms ./cmd/cms

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/cms /usr/local/bin/cms
USER app
ENTRYPOINT ["/usr/local/bin/cms"]
