FROM golang:1.12-alpine AS builder

RUN apk update && \
    apk add --no-cache bash ca-certificates git gcc g++ libc-dev postgresql-dev && \
    update-ca-certificates && \
    adduser -D -g '' appuser && \
    mkdir -p /src
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s -extldflags "-static"' -tags netgo -a -installsuffix cgo -o /app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app /app

EXPOSE 8080
USER appuser

CMD ["/app"]
