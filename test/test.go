package main

import (
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
}
