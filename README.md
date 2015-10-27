# Go Plugins [![GoDoc](https://godoc.org/github.com/myodc/go-plugins?status.svg)](https://godoc.org/github.com/myodc/go-plugins)

A repository for go-micro plugins.

## Usage

Plugins can be added to go-micro in the following ways

```go
import (
	"github.com/myodc/go-micro/cmd"
	"github.com/myodc/go-plugins/registry/kubernetes"
)

func main() {
	cmd.Registries["kubernetes"] = kubernetes.NewRegistry
	cmd.Init()
}
```

OR

```go
import (
	"github.com/myodc/go-plugins/registry/kubernetes"
)

func main() {
	c kubernetes.NewRegistry([]string{}) // default to using env vars for master API
}
```
