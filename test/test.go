package main

import (
	"fmt"
	"go-hotplugin/hotplugin"
	"time"
)

func main() {
	options := hotplugin.ManagerOptions{
		Dir:    "./",
		Suffix: ".so",
	}
	go hotplugin.StartManager(options)
	time.Sleep(5 * time.Second)
	result := hotplugin.Call("testplugin", "Test", "my world")
	fmt.Println(result...)
	select {}
}
