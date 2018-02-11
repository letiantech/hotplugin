// Copyright (c) 2017 letian0805@gmail.com
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package hotplugin

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type ManagerOptions struct {
	Dir    string
	Suffix string
}

type Manager interface {
	Run() error
	IsRunning() bool
	GetPlugin(name string) (*Plugin, error)
	GetPluginWithVersion(name string, version uint64) (*Plugin, error)
	GetFunc(module, function string) (f func(...interface{}) []interface{}, err error)
	Call(module, function string, args ...interface{}) []interface{}
	OnLoaded(p *Plugin)
	OnReloaded(p *Plugin)
	OnUnloaded(p *Plugin)
	OnError(p *Plugin, err *PluginError)
}

type manager struct {
	running bool
	options ManagerOptions
	watcher *fsnotify.Watcher
	cache   map[string]*Plugin
	loaded  map[string]map[uint64]*Plugin
}

func NewManager(options ManagerOptions) (Manager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print("error: ", err)
		return nil, err
	}
	m := &manager{
		options: options,
		watcher: watcher,
		running: false,
		cache:   make(map[string]*Plugin),
		loaded:  make(map[string]map[uint64]*Plugin),
	}
	return m, nil
}

func (m *manager) pluginPath(path string) string {
	if !strings.Contains(path, m.options.Dir) {
		dir := []byte(m.options.Dir)
		if dir[len(dir)-1] != '/' {
			path = m.options.Dir + "/" + path
		} else {
			path = m.options.Dir + path
		}
	}
	return path
}

func (m *manager) OnReloaded(p *Plugin) {

}

func (m *manager) OnError(p *Plugin, err *PluginError) {

}

func (m *manager) OnUnloaded(p1 *Plugin) {
	name := p1.Name()
	version := p1.Version()
	log.Print(name, " loaded")
	if mp, ok := m.loaded[name]; ok {
		delete(mp, version)
		if len(mp) == 0 {
			delete(m.loaded, name)
		}
	}
}

func (m *manager) OnLoaded(p1 *Plugin) {
	name := p1.Name()
	version := p1.Version()
	log.Print(name, " loaded")
	if mp, ok := m.loaded[name]; ok {
		mp[version] = p1
	} else {
		mp := make(map[uint64]*Plugin)
		mp[version] = p1
		m.loaded[name] = mp
	}
}

func (m *manager) loadAll() error {
	if m.running {
		return nil
	}
	f, e := os.Open(m.options.Dir)
	if e != nil {
		return e
	}
	for {
		d, e := f.Readdir(100)
		if e != nil {
			if e == io.EOF {
				f.Seek(0, 0)
				break
			}
			return e
		}
		for i := 0; i < len(d); i++ {
			path := m.pluginPath(d[i].Name())
			if !m.isPlugin(path) {
				continue
			}
			log.Println(path)
			p := NewPlugin(path, m)
			m.cache[path] = p
			p.Load()
		}
	}
	return nil
}

func (m *manager) isPlugin(path string) bool {
	return strings.HasSuffix(path, m.options.Suffix)
}

func (m *manager) Run() error {
	if m.running {
		return nil
	}
	e := m.loadAll()
	if e != nil {
		log.Println(e.Error())
		return e
	}
	e = m.watcher.Add(m.options.Dir)
	if e != nil {
		log.Println(e.Error())
		return e
	}
	c := make(chan int, 1)
	go func() {
		m.running = true
		c <- 1
		for {
			select {
			case e := <-m.watcher.Events:
				path := m.pluginPath(e.Name)
				log.Print(e)
				if !m.isPlugin(path) {
					continue
				}
				p, ok := m.cache[path]
				if !ok || p == nil {
					log.Println(path)
					p = NewPlugin(path, m)
					m.cache[path] = p
				}
				if e.Op&fsnotify.Write == fsnotify.Write {
					p.Reload()
					continue
				}
				if e.Op&fsnotify.Create == fsnotify.Create {
					p.Load()
					continue
				}
				if e.Op&fsnotify.Remove == fsnotify.Remove {
					p.Unload()
					continue
				}
			case err := <-m.watcher.Errors:
				log.Print(err)
				continue
			}
		}
	}()
	<-c
	close(c)
	return nil
}

func (m *manager) GetPlugin(name string) (*Plugin, error) {
	mp, ok := m.loaded[name]
	if !ok {
		return nil, errors.New("not found")
	}
	var latestVersion uint64 = 0
	var latestPlugin *Plugin = nil
	for v, p := range mp {
		if latestVersion < v {
			latestVersion = v
			latestPlugin = p
		}
	}
	return latestPlugin, nil
}

func (m *manager) GetPluginWithVersion(name string, version uint64) (*Plugin, error) {
	mp, ok := m.loaded[name]
	if !ok {
		return nil, errors.New("not found")
	}
	p, ok := mp[version]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *manager) GetFunc(module, function string) (f func(...interface{}) []interface{}, err error) {
	var p *Plugin = nil
	p, err = m.GetPlugin(module)
	if err != nil {
		return
	}
	if p == nil {
		err = errors.New("no plugin " + module)
		return
	}

	return p.GetFunc(function)
}

func (m *manager) Call(module, function string, args ...interface{}) []interface{} {
	f, err := m.GetFunc(module, function)
	if err != nil {
		return []interface{}{err}
	}

	return f(args...)
}

func (m *manager) IsRunning() bool {
	return m.running
}
