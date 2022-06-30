name = smsToEmail


.PHONY: build docker runforever stop
build:
	go build -o $(name) ./main.go

docker:
	echo "TODO"

runforever: build
	mkdir -p ./logs
	nohup ./$(name) > ./logs/info.log 2>./logs/error.log &

stop:
	killall $(name)