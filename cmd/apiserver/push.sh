CGO_ENABLED=0 go build
#apiserver-boot build executables --generate=false
docker build -t wlhtorresowen/fujitsu-coredump:v0.2 .
docker push wlhtorresowen/fujitsu-coredump:v0.2
