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

var defaultManager Manager

func StartManager(options ManagerOptions) error {
	if defaultManager != nil && defaultManager.IsRunning() {
		return nil
	}
	var err error
	defaultManager, err = NewManager(options)
	if err != nil {
		return err
	}
	return defaultManager.Run()
}

func Call(module, function string, args ...interface{}) []interface{} {
	f, err := GetFunc(module, function)
	if err != nil {
		return []interface{}{err}
	}

	return f(args...)
}

func GetPlugin(name string) (*Plugin, error) {
	return defaultManager.GetPlugin(name)
}

func GetPluginWithVersion(name string, version uint64) (*Plugin, error) {
	return defaultManager.GetPluginWithVersion(name, version)
}

func GetFunc(module, function string) (f func(...interface{}) []interface{}, err error) {
	return defaultManager.GetFunc(module, function)
}
