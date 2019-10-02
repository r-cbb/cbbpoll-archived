FROM golang:1.12 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o cbbpoll cmd/cbbpoll/main.go

# final stage
# At some point we might find we need enough "extra" stuff to warrant using golang:alpine or
# gcr.io/distroless/base over scratch for the final image base, but for now scratch works.
FROM scratch
COPY --from=builder /app/cbbpoll /app/

# Need ssl certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# copy keys
COPY jwtRS256.key .
COPY jwtRS256.key.pub .


EXPOSE 8000
ENTRYPOINT ["/app/cbbpoll"]