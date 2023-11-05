build:
	go build
run: build
	./home-backend-public
clean:
	git clean -X