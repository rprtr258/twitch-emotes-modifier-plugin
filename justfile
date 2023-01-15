# bump dependencies
@bump:
	go get -u ./...
	go mod tidy

@dump:
	ls *.webp | xargs go run cmd/dump/dump.go

# run trollDespair(static),cage(static) sample
@run-despair-cage:
	go run main.go '603caea243b9e100141caf4f,612b180e8560641da7250ce8>over'

# run trollDespair(static),snowTime(animated) sample
@run-despair-snow:
	go run main.go '603caea243b9e100141caf4f,6129ea8afd97806f9d734a76>over'

# run peepoClap(animated),snowTime(animated) sample
@run-peepo-snow:
	go run main.go '60aed440997b4b396ed9ec39,6129ea8afd97806f9d734a76>over'

# run reverse(diesfromcringe) sample
@run-alivefromcringe:
	go run main.go '611523959bf574f1fded6d72>rev'

