pull:
	git checkout .
	git pull

build:
	go mod download
	go build -o chatgpt-telegram-bot

install: pull build
	sudo cp chatgpt-telegram-bot /usr/local/bin/chatgpt-telegram-bot