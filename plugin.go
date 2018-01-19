package hotplugin

import (
	"errors"
	"fmt"
	"log"
	"plugin"
	"reflect"
	"sync"
	"time"
)

const PLUGIN_TIMEOUT = 100 * time.Millisecond

const (
	PLUGIN_STATUS_NONE = iota
	PLUGIN_STATUS_LOADING
	PLUGIN_STATUS_RELOADING
	PLUGIN_STATUS_UNLOADING
	PLUGIN_STATUS_LOADED
)

const (
	PLUGIN_ERROR_NONE = iota
	PLUGIN_ERROR_LOADFAILED
)

type PluginError struct {
	Type int
	Err  error
}

type Plugin struct {
	options PluginOptions
	name    string
	version uint64
	path    string
	status  int
	plugin  *plugin.Plugin
	timer   *time.Timer
	c       chan int
	lock    *sync.RWMutex
	refs    int
}

type PluginOptions struct {
	OnLoaded   func(p *Plugin)
	OnReloaded func(p *Plugin)
	OnUnloaded func(p *Plugin)
	OnError    func(p *Plugin, err *PluginError)
}

func NewPlugin(path string, options PluginOptions) *Plugin {
	p := &Plugin{
		options: options,
		path:    path,
		status:  PLUGIN_STATUS_NONE,
		c:       make(chan int, 10),
		timer:   time.NewTimer(PLUGIN_TIMEOUT),
		lock:    &sync.RWMutex{},
		refs:    0,
	}
	go p.start()
	return p
}

func (p *Plugin) setStatus(status int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.status = status
}

func (p *Plugin) Status() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.status
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) Version() uint64 {
	return p.version
}

func (p *Plugin) start() error {
	s := PLUGIN_STATUS_NONE
	defer p.unload()
	for {
		select {
		case s = <-p.c:
			if p.Status() == PLUGIN_STATUS_NONE && s == PLUGIN_STATUS_RELOADING {
				s = PLUGIN_STATUS_LOADING
			}
			if p.Status() == PLUGIN_STATUS_LOADED && s == PLUGIN_STATUS_LOADING {
				s = PLUGIN_STATUS_RELOADING
			}
			p.timer.Reset(PLUGIN_TIMEOUT)
		case <-p.timer.C:
			if p.Status() == PLUGIN_STATUS_NONE && s == PLUGIN_STATUS_LOADING {
				e := p.load()
				if e != nil {
					return e
				}
			} else if p.Status() == PLUGIN_STATUS_LOADED && s == PLUGIN_STATUS_RELOADING {
				p.reload()
			} else if s == PLUGIN_STATUS_UNLOADING {
				return nil
			}
		}
	}
	return nil
}

func (p *Plugin) Path() string {
	return p.path
}

func (p *Plugin) Load() error {
	p.c <- PLUGIN_STATUS_LOADING

	return nil
}

func (p *Plugin) load() error {
	path := p.path
	plugin, e := plugin.Open(path)
	if e != nil {
		log.Print("load plugin ", path, " error: ", e)
		return e
	}
	p.plugin = plugin
	f, e := plugin.Lookup("Load")
	if e != nil {
		log.Print("load plugin ", path, " error: ", e)
		return e
	}
	name, version, e := f.(func() (string, uint64, error))()
	p.name = name
	p.version = version
	s := fmt.Sprintf("load plugin: %s, version: 0x%x", name, version)
	if e == nil {
		log.Print(s)
	} else {
		log.Print(s, ", error: ", e)
	}
	p.setStatus(PLUGIN_STATUS_LOADED)
	if p.options.OnLoaded != nil {
		p.options.OnLoaded(p)
	}

	return e
}

func (p *Plugin) Reload() error {
	p.c <- PLUGIN_STATUS_RELOADING
	return nil
}

func (p *Plugin) reload() error {
	name := p.name
	version := p.version
	s := fmt.Sprintf("reload plugin: %s, version: 0x%x", name, version)
	log.Print(s)
	p.setStatus(PLUGIN_STATUS_LOADED)
	return nil
}

func (p *Plugin) Unload() error {
	p.c <- PLUGIN_STATUS_UNLOADING
	return nil
}

func (p *Plugin) unload() error {
	name := p.name
	version := p.version
	s := fmt.Sprintf("unload plugin: %s, version: 0x%x", name, version)
	f, e := p.plugin.Lookup("Unload")
	if e != nil {
		log.Print(s, ", error: ", e)
		return e
	}
	err := f.(func() error)()
	log.Print(s)
	p.setStatus(PLUGIN_STATUS_NONE)
	return err
}

func (p *Plugin) Call(fun string, params ...interface{}) []interface{} {
	f, err := p.GetFunc(fun)
	if err != nil {
		return []interface{}{err}
	}
	return f(params...)
}

func (p *Plugin) GetFunc(fun string) (f func(...interface{}) []interface{}, err error) {
	if p.plugin == nil {
		err = errors.New("plugin not loaded")
		return
	}
	f1, err := p.plugin.Lookup(fun)
	if err != nil {
		return
	}
	return func(params ...interface{}) []interface{} {
		f2 := reflect.ValueOf(f1)
		out := make([]interface{}, f2.Type().NumOut())
		if len(params) != f2.Type().NumIn() {
			err := errors.New("The number of params is not adapted.")
			out[len(out)-1] = err
			return out
		}
		in := make([]reflect.Value, len(params))
		for k, param := range params {
			in[k] = reflect.ValueOf(param)
			if f2.Type().In(k).Name() != in[k].Type().Name() {
				err := fmt.Sprintf("The type of params is not adapted, params[%d] require type %s",
					k, f2.Type().In(k).Name())
				out[len(out)-1] = errors.New(err)
				return out
			}
		}
		for j := 0; j < f2.Type().NumOut(); j++ {
			log.Print(f2.Type().Out(j))
		}
		result := f2.Call(in)
		for i := 0; i < len(result); i++ {
			out[i] = result[i].Interface()
		}

		return out
	}, nil
}
