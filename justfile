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

# run reverseTime(diesfromcringe) sample
@run-revt:
	go run main.go '611523959bf574f1fded6d72>revt'

# run reverseX(diesfromcringe) sample
@run-revx:
	go run main.go '611523959bf574f1fded6d72>revx'

# run reverseY(diesfromcringe) sample
@run-revy:
	go run main.go '611523959bf574f1fded6d72>revy'

# run concat(diesfromcringe, reverseTime(diesofcringe)) sample
@run-stackt:
	go run main.go '611523959bf574f1fded6d72,611523959bf574f1fded6d72>revt>stackt'

# run stackx(diesfromcringe, reverseTime(diesofcringe)) sample
@run-stackx:
	go run main.go '611523959bf574f1fded6d72,611523959bf574f1fded6d72>revt>stackx'

# run stacky(diesfromcringe, reverseTime(diesofcringe)) sample
@run-stacky:
	go run main.go '611523959bf574f1fded6d72,611523959bf574f1fded6d72>revt>stacky'
