# Router Plugin

The router plugin is a HTTP handler plugin for the Micro API which enables you to define routes via go-platform/config. This is 
dynamic configuration that can then be leveraged via anything that implements the go-platform/config interface e.g file, etcd, consul 
or the config service.

## Usage

Register the plugin before building Micro

```
package main

import (
	"github.com/micro/micro/plugin"
	"github.com/micro/go-plugins/micro/router"
)

func init() {
	plugin.Register(router.NewRouter())
}
```


