package hotplugin

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"strings"
)

type ManagerOptions struct {
	Dir    string
	Suffix string
}

type Manager struct {
	running bool
	options ManagerOptions
	watcher *fsnotify.Watcher
	cache   map[string]*Plugin
	loaded  map[string]map[uint64]*Plugin
}

func NewManager(options ManagerOptions) (*Manager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print("error: ", err)
		return nil, err
	}
	m := &Manager{
		options: options,
		watcher: watcher,
		running: false,
		cache:   make(map[string]*Plugin),
		loaded:  make(map[string]map[uint64]*Plugin),
	}
	return m, nil
}

func (m *Manager) pluginPath(path string) string {
	if !strings.Contains(path, m.options.Dir) {
		path = m.options.Dir + "/" + path
	}
	return path
}

func (m *Manager) pluginOptions() PluginOptions {
	return PluginOptions{
		OnLoaded: func(p1 *Plugin) {
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
		},
		OnUnloaded: func(p1 *Plugin) {
			name := p1.Name()
			version := p1.Version()
			log.Print(name, " loaded")
			if mp, ok := m.loaded[name]; ok {
				delete(mp, version)
				if len(mp) == 0 {
					delete(m.loaded, name)
				}
			}
		},
	}
}

func (m *Manager) LoadAll() error {
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
			if !m.IsPlugin(path) {
				continue
			}
			p := NewPlugin(path, m.pluginOptions())
			m.cache[path] = p
			p.Load()
		}
	}
	return nil
}

func (m *Manager) IsPlugin(path string) bool {
	return strings.HasSuffix(path, m.options.Suffix)
}

func (m *Manager) Run() error {
	if m.running {
		return nil
	}
	m.LoadAll()
	m.watcher.Add(m.options.Dir)
	m.running = true
	for {
		select {
		case e := <-m.watcher.Events:
			path := m.pluginPath(e.Name)
			log.Print(e)
			if !m.IsPlugin(path) {
				continue
			}
			p, ok := m.cache[path]
			if !ok || p == nil {
				p = NewPlugin(path, m.pluginOptions())
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
			return err
		}
	}
	return nil
}

func (m *Manager) GetPlugin(name string) (*Plugin, error) {
	mp, ok := m.loaded[name]
	if !ok {
		return nil, errors.New("not found")
	}
	var latest_version uint64 = 0
	var latest_plugin *Plugin = nil
	for v, p := range mp {
		if latest_version < v {
			latest_version = v
			latest_plugin = p
		}
	}
	return latest_plugin, nil
}

func (m *Manager) GetPluginWithVersion(name string, version uint64) (*Plugin, error) {
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

func (m *Manager) IsRunning() bool {
	return m.running
}
