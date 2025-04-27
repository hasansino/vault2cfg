# vault2cfg

Library provides configuration binding to vault secrets.

## Installation

```bash
~ $ go get github.com/hasansino/cfg2env
```

```go
package main

import (
	"log"

	"github.com/hasansino/vault2cfg"
	"github.com/hasansino/libvault"
)

type Config struct {
	SecretA string `vault:"SECRET_A"`
	SecretB string `vault:"SECRET_B"`
}

func main() {
	vcl, err := libvault.New("localhost", nil)
	if err != nil {
		log.Fatal(err)
	}
	err = vcl.K8Auth("role", "serviceaccount", "mountpath", )
	if err != nil {
		log.Fatal(err)
	}
	data, err := vcl.Retrieve("secret/path")
	if err != nil {
		log.Fatal(err)
	}
	cfg := new(Config)
	vault2cfg.Bind(cfg, data)
}
```
