# Go Plugins [![GoDoc](https://godoc.org/github.com/micro/go-plugins?status.svg)](https://godoc.org/github.com/micro/go-plugins) [![Travis CI](https://travis-ci.org/micro/go-plugins.svg?branch=master)](https://travis-ci.org/micro/go-plugins)

A repository for go-micro plugins.

## Usage

Plugins can be added to go-micro in the following ways

```go
import (
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-plugins/registry/kubernetes"
)

func main() {
	cmd.Registries["kubernetes"] = kubernetes.NewRegistry
	cmd.Init()
}
```

OR

```go
import (
	"github.com/micro/go-plugins/registry/kubernetes"
)

func main() {
	c kubernetes.NewRegistry([]string{}) // default to using env vars for master API
}
```
