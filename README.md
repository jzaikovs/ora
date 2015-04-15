# ora

Oracle database driver for Go uses OCI

For now it's only implements basic `database/sql/driver` interface.

For 64bit go you will need 64bit oracle instant client.

# windows

* you will only need instant client installed
* path to instant client should be in PATH system variables

# linux

* you will only need instant client installed
* `sudo ldconfig -p | grep libclntsh` should return some lines

## usage

```
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/jzaikovs/ora"
)

func main() {
	db, err := sql.Open("ora", "user/password@//localhost:1521/XE")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from dual")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	dummy := ""
	for rows.Next() {
		if err = rows.Scan(&dummy); err != nil {
			panic(err)
		}
		fmt.Println(dummy)
	}
}
```
