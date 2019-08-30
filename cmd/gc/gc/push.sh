CGO_ENABLED=0 go build
docker build -t wlhtorresowen/fujitsu-coredump-gc:v0.1 .
docker push wlhtorresowen/fujitsu-coredump-gc:v0.1
