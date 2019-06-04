# fujitsu-coredump
 This project aims to implement a general tool to manage core dump functions in kubernetes. It is built by [apiserver-builder-alpha](https://github.com/kubernetes-incubator/apiserver-builder-alpha).
 You could built your own one step by step followed by commits.
### Built

Installation instructions:
 - Download the source code of latest [apiserver-builder-alpha](https://github.com/kubernetes-incubator/apiserver-builder-alpha)
 - Enter `apiserver-builder-alpha` directory, execute `git checkout release-1.12`
 - Execute `make build`
 - Observe archives generated, extract one of them depends on your system
 - Add ./apiserver-builder-alpha/bin to your path export PATH=$PATH:[apiserver-builder-path]/bin
 - Test apiserver-boot by `apiserver-boot -h`
 - Enter `fujitsu-coredump` directory, execute `apiserver-boot build executables --generate=false`
 - Observe `apiserver` and `controller-manager` in ./bin
 - Execute `apiserver-boot build config --name coredump --namespace default --image [your image name]`
 - Observe ./config was generated

### Deploy
Previous section shows how the fujitsu-coredump component was built. And then we should deploy this tool to kubernetes cluster. Before that, you need configure your cluster firstly as [k8s documents](https://kubernetes.io/docs/tasks/access-kubernetes-api/configure-aggregation-layer/#kubernetes-apiserver-client-authentication) said to let aggregation api layer runnable.

After that, build your own docker image:

    $ tee ./bin/Dockerfile <<- 'EOF'
      FROM alpine
      COPY apiserver /
      COPY controller-manager /
      ENTRYPOINT ["/apiserver"]
      EOF

    $ docker build -t [account]/[image name]:[tag] bin

    $ docker push [account]/[image name]:[tag]

Now you have your an image with fujitsu-coredump components, it's time to generate config yaml file:

    $ apiserver-boot build config --name [name of service] --namespace [namespace of the service in] --image [your image name]

    $ kubectl create -f config/apiserver.yaml
