package hotplugin

import (
	"errors"
)

var manager *Manager = nil

func StartManager(options ManagerOptions) error {
	if manager != nil && manager.IsRunning() {
		return nil
	}
	var err error
	manager, err = NewManager(options)
	if err != nil {
		return err
	}
	return manager.Run()
}

func Call(module, function string, args ...interface{}) []interface{} {
	f, err := GetFunc(module, function)
	if err != nil {
		return []interface{}{err}
	}

	return f(args...)
}

func GetPlugin(name string) (*Plugin, error) {
	if manager == nil || !manager.IsRunning() {
		return nil, errors.New("not running")
	}
	return manager.GetPlugin(name)
}

func GetPluginWithVersion(name string, version uint64) (*Plugin, error) {
	if manager == nil || !manager.IsRunning() {
		return nil, errors.New("not running")
	}
	return manager.GetPluginWithVersion(name, version)
}

func GetFunc(module, function string) (f func(...interface{}) []interface{}, err error) {
	var p *Plugin = nil
	p, err = GetPlugin(module)
	if err != nil {
		return
	}
	if p == nil {
		err = errors.New("no plugin " + module)
		return
	}

	return p.GetFunc(function)
}
