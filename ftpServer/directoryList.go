package ftpServer

import "sync"

type DirectoryList struct {
	sync.RWMutex
	items map[string]bool
}

func NewDirectoryList() DirectoryList {
	return DirectoryList{
		items: make(map[string]bool),
	}
}

func (dl *DirectoryList) Set(key string) {
	dl.Lock()
	dl.items[key] = true
	dl.Unlock()
}

func (dl *DirectoryList) Unset(key string) {
	dl.Lock()
	delete(dl.items, key)
	dl.Unlock()
}

func (dl *DirectoryList) Isset(key string) (isset bool) {
	dl.RLock()
	_, isset = dl.items[key]
	dl.Unlock()
	return
}

func (dl *DirectoryList) Iter() <-chan string {
	c := make(chan string)

	go func() {
		dl.RLock()
		defer dl.RUnlock()

		for key, _ := range dl.items {
			c <- key
		}

		close(c)
	}()

	return c
}
