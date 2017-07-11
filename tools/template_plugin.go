package main

import (
	"fmt"
)

var plugin_name = "template_plugin"

var plugin_version uint64 = 0x00010000

func Load() (name string, version uint64, err error) {
	return plugin_name, plugin_version, nil
}

func Unload() error {
	fmt.Printf("unload %s, version: 0x%x\n", plugin_name, plugin_version)
	return nil
}

func Test(data string) string {
	return "hello " + data
}
