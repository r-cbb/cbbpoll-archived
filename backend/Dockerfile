FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cbbpoll cmd/cbbpoll/main.go

# TODO: Move testing into a makefile and out of image creation.  Later on can
# figure out a good way to test in docker as well.
# RUN go test ./...

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