TEST_OPTIONS :=

gen:
	go run github.com/99designs/gqlgen generate

test:
	mkdir -p test_output
	go test ./... $(TEST_OPTIONS) -coverprofile=test_output/coverage.out
	go tool cover -html=test_output/coverage.out -o test_output/coverage.html