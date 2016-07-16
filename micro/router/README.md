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

## Routes

Routes are used to config request to match and the response to return. Here's an example.

```json
{
	"api": {
		"routes": [
			{
				"request": {
					"method": "GET",
					"host": "127.0.0.1:10001",
					"path": "/"
				},
				"response": {
					"status_code": 302,
					"header": {
						"location": "http://example.com"
					}
				}
			},
			{
				"request": {
					"method": "POST",
					"host": "127.0.0.1:10001",
					"path": "/foo"
				},
				"response": {
					"status_code": 301,
					"header": {
						"location": "http://foo.bar.com"
					}
				}
			}
		]
	}
}
```
