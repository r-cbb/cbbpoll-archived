checks: test vet

check-swagger:
	which swagger || (go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
	swagger generate spec -o docs/swagger.json --scan-models
	docker-compose build swagger

test:
	go test ./... -coverprofile=coverage.out

vet:
	go vet ./...

clean:
	rm coverage.out
	rm cmd/cbbpoll/cbbpoll

docker:
	docker-compose build cbbpoll

run:
	docker-compose up cbbpoll
