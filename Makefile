all:
	go run  ./cmd/tinear/main.go
debug:
	go build -gcflags="all=-N -l" ./cmd/tinear && ./tinear
