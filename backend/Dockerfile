FROM golang:1.12 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /dist/cbbpoll cmd/cbbpoll/main.go

WORKDIR /dist

# Gather static libs we are linking against.  cgo is currently required for
# sqlite support.
RUN ldd cbbpoll | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/

# final stage
# At some point we might find we need enough "extra" stuff to warrant using golang:alpine or
# gcr.io/distroless/base over scratch for the final image base, but for now scratch works.
FROM scratch
COPY --from=builder /dist /

# Need ssl certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# copy keys
COPY jwtRS256.key .
COPY jwtRS256.key.pub .

EXPOSE 8000
ENTRYPOINT ["/cbbpoll"]