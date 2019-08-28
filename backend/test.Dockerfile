FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c ./internal/app -o app.test

FROM golang:alpine as test
COPY --from=builder /app/app.test /app/

ENTRYPOINT ["/app/app.test"]