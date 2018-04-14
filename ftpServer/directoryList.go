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

func (dl *DirectoryList) Create(path string) {
	dl.Lock()
	dl.items[path] = true
	dl.Unlock()
}

func (dl *DirectoryList) Delete(path string) {
	dl.Lock()
	delete(dl.items, path)
	dl.Unlock()
}

func (dl *DirectoryList) Exists(path string) (exists bool) {
	dl.RLock()
	_, exists = dl.items[path]
	dl.Unlock()
	return
}

func (dl *DirectoryList) Iterate() <-chan string {
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
