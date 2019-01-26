# ![hive](https://raw.githubusercontent.com/lnsp/hive/master/docs/logo.png)

hive is a simple and fast microservice toolkit built on top of HTTP and JSON. It was created in mind for modern software architectures and infrastructure, simple cross-language interaction and orchestrators like Kubernetes and Docker Swarm.

## Getting started
```bash
$ go get github.com/lnsp/hive
$ hive new
Enter service name: skynet
Enter service path: github.com/cyberdyne/skynet
...
skynet up and ready.
$ hive about github.com/cyberdyne/skynet
{
  "name": "skynet",
  "dnsname": "skynet",
  "version": "1.0.0",
  "protocol": "http",
  "methods": {
    "SpawnRobot": {
...
$ cd $GOPATH/src/github.com/cyberdyne/skynet/runtime
$ docker build -t skynet:latest .
$ docker service create --name takeover --network=backend --replicas=1000 skynet:latest
```
