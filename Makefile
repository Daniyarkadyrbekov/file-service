.PHONY: build-server-for-linux
build-server-for-linux:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/server ./server/main.go

.PHONY: build-client-for-windows
build-client-for-windows:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/client ./client/*.go

.PHONY: deploy-server-bin
deploy-server-bin:
	scp ./bin/server root@206.81.22.60:/root/umeford/


