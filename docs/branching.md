# branching

## Updating dependencies

```text
Please help me update @go.mod. Start with the first require block, which holds direct dependencies. For each dependency, run `go get -u`.

In this first pass though, please skip the following:
* k8s packages
* operator apis
* dependencies with comments
* dependencies mentioned in the replace block

When you are done with these updates, please run `go mod tidy && go mod vendor`.
```

## Validating functionality

```bash
go test -v -tags=integration -run TestConfigmapCreate ./integration/
```
