all:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/apiserver ./cmd/apiserver
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/gc ./cmd/gc
