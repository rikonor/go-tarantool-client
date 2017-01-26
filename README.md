Go Tarantool Client
---

This is a simple client to query the online [Tarantool](http://sh6.tarantool.org/) avro validation endpoint.

#### Usage

```go
package main

import (
	"log"

	tarantool "github.com/rikonor/go-tarantool-client"
)

func main() {
	schema := `{
	  "type": "record",
	  "name": "User",
	  "fields": [
	    {"name": "username", "type": "array"},
	    {"name": "phone", "type": "long"},
	    {"name": "age", "type": "int"}
	  ]
	}`

	if err := tarantool.Validate(schema); err != nil {
		log.Fatalf("Schema did not pass validation: %s", err)
	}
}
```
