FROM docker.io/library/golang:1.26.1 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/file-server ./cmd/hello

FROM scratch
COPY --from=builder /usr/local/bin/file-server /usr/local/bin/file-server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/file-server"]
