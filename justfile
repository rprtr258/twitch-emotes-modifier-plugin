# bump dependencies
@bump:
	go get -u ./...
	go mod tidy

@dump:
	ls *.webp | xargs go run cmd/dump/dump.go

# run trollDespair(static),snowTime(animated) sample
@run-despair-snow:
	go run main.go '603caea243b9e100141caf4f,6129ea8afd97806f9d734a76>over'

# run peepoClap(animated),snowTime(animated) sample
@run-peepo-snow:
	go run main.go '60aed440997b4b396ed9ec39,6129ea8afd97806f9d734a76>over'
