[![Build Status](https://api.travis-ci.org/letiantech/hotplugin.svg)](https://travis-ci.org/letiantech/hotplugin)

# hotplugin
golang plugin framework for hot update, go version >= 1.8

# usage
1. get hotplugin
```bash
go get github.com/letiantech/hotplugin
```
2. write a plugin with Load, Unload and other functions like this
```go
//testplugin.go
package main

import (
	"fmt"
	"log"
)

const (
	pluginName    = "testplugin"
	pluginVersion = 0x00010000
)

func Load(register func(name string, version uint64) error) error {
	err := register(pluginName, pluginVersion)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Println("loading test plugin")
	return nil
}

func Unload() error {
	fmt.Printf("unload %s, version: 0x%x\n", pluginName, pluginVersion)
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
