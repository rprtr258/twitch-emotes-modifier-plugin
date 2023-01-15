# bump dependencies
@bump:
	go get -u ./...
	go mod tidy

# run sample
@run:
	go run main.go 60aed440997b4b396ed9ec39 6129ea8afd97806f9d734a76 'peepoHappy,snowTime,^'
	go run cmd/dump/dump.go out.webp
