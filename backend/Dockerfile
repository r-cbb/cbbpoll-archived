FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cbbpoll cmd/cbbpoll/main.go

# final stage
FROM scratch
COPY --from=builder /app/cbbpoll /app/

EXPOSE 8000
ENTRYPOINT ["/app/cbbpoll"]