# vault2cfg

Library provides configuration binding to vault secrets.
Supports only string values and anything that can be fmt.Sprintf-ed.

## Installation

```bash
go get github.com/hasansino/vault2cfg/v2
```

## Example

```go
package main

import (
	"log"
	"context"

	"github.com/hasansino/vault2cfg"
	"github.com/hashicorp/vault-client-go"
)

type Config struct {
	SecretA string `vault:"SECRET_A"`
	SecretB string `vault:"SECRET_B"`
}

func main() {
	client, err := vault.New(
		vault.WithAddress("http://localhost:8200"),
	)
	if err != nil {
		log.Fatalf("failed to initialise vault client: %v", err)
	}

	err = client.SetToken("qwerty")
	if err != nil {
		log.Fatalf("failed to authenticate in vault: %v", err)
	}

	data, err := client.Secrets.KvV2Read(
		context.Background(), "service/path", vault.WithMountPath("secret"),
	)
	if err != nil {
		log.Fatalf("failed to read vault secrets: %v", err)
	}

	cfg := new(Config)

	if err := vault2cfg.Bind(cfg, data.Data, vault2cfg.WithTagName("vault")); err != nil {
		log.Fatalf("failed to bind vault secrets: %v", err)
	}
}
```
