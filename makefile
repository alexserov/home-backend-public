build:
	go build
run: build
	./home-backend-public
clean:
	git clean -fX
refresh:
	sudo service home-backend stop
	git pull
	go build
	sudo service home-backend start
logs:
	sudo journalctl -u home-backend.service -f -n 100
rl:
	make refresh
	make logs