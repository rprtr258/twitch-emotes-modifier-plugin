# bump dependencies
@bump:
	go get -u ./...
	go mod tidy

# run sample
@run:
	go run main.go 'peepoHappy,snowTime,^'
