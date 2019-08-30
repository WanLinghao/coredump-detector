#CGO_ENABLED=0 go build
#apiserver-boot build executables --generate=false
docker build -t  wlhtorresowen/fujitsu-coredump:v0.1 bin
#docker push  wlhtorresowen/fujitsu-coredump:v0.3
