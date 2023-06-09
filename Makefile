output:
	mkdir output

certificat: output
	openssl genrsa -out output/privateKey.key 2048
	openssl req -new -key output/privateKey.key -out output/server.csr
	openssl x509 -req -days 365 -in output/server.csr -signkey output/privateKey.key -out output/certificate.crt
	rm output/server.csr

build: output
	cp bearer.txt Handler/
	cp bearer.txt Agent/
	go build -o output/Handler -ldflags="-s -w" Handler/handler.go
	go build -o output/RatChatPT_unix -ldflags="-s -w" Agent/ratchatPT.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ HOST=x86_64-w64-mingw32 go build -C Agent -ldflags "-s -w -H=windowsgui -extldflags=-static" -p 4 -v -o ../output/RatChatPT_windows.exe
	rm Handler/bearer.txt Agent/bearer.txt

run_agent:
	docker run -it --rm -v $(shell pwd)/output/:/root/output ubuntu /root/output/ratchatPT
