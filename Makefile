build:
	docker build -t validator-service-img .

start: build
	docker-compose up -d

run-test:
	go run tests/test.go

start-tests:
	make start
	make run-test
	make stop

stop:
	docker-compose down