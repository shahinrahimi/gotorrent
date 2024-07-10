build: 
	@go build -o ./bin/gotorrent
run: build
	@./bin/gotorrent