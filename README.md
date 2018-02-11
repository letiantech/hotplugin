# go-hotplugin
golang plugin framework for hot update, go version >= 1.8

# usage
1. get go-hotplugin
```bash
go get github.com/letian0805/go-hotplugin
```
2. write a plugin with Load, Unload and other functions like this
```go
//testplugin.go
package main

import (
    "fmt"
)

var plugin_name = "testplugin"

var plugin_version uint64 = 0x00010000

func Load() (name string, version uint64, err error) {
    fmt.Printf("loading test plugin\n")
    return plugin_name, plugin_version, nil
}

func Unload() error {
    fmt.Printf("unload %s, version: 0x%x\n", plugin_name, plugin_version)
    return nil
}

func Test(data string) string {
    return "hello " + data
}
```

3. build your plugin
```bash
go build -buildmode ./testplugin.go
```

4. save your testplugin.so to /path/of/plugin/dir

5. write main.go like this
```go
//main.go
package main

import (
	"fmt"
	"hotplugin"
)

func main() {
	options := hotplugin.ManagerOptions{
		Dir:    "./",
		Suffix: ".so",
	}
	hotplugin.StartManager(options)
	result := hotplugin.Call("testplugin", "Test", "my world")
	fmt.Println(result...)
}

```
