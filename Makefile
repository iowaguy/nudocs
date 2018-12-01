TARGET = nudocs

$(TARGET): client server

client:
	go build -v -i -o client/client client/main.go

server:
	go build -v -i

.PHONY: docker
docker:
	GOOS=linux go build -v
	GOOS=linux go build -v -o client/client client/main.go
	docker-compose up

.PHONY: test
test:
	go test github.com/iowaguy/nudocs/common/communication
	go test github.com/iowaguy/nudocs/common/clock

.PHONY: docker-clean
docker-clean:
	docker ps -a | grep nudocs | awk '{print $$1}' | xargs docker stop
	docker ps -a | grep nudocs | awk '{print $$1}' | xargs docker rm
