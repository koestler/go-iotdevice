package ftpServer

import "sync"

type FileList struct {
	sync.RWMutex
	items map[string]*VirtualFile
}

type FileListItem struct {
	Key string
	Value *VirtualFile
}

func NewFileList() FileList {
	return FileList{
		items: make(map[string]*VirtualFile),
	}
}

func (fl *FileList) Set(key string, file *VirtualFile) {
	fl.Lock()
	fl.items[key] = file
	fl.Unlock()
}

func (fl *FileList) Unset(key string) {
	fl.Lock()
	delete(fl.items, key)
	fl.Unlock()
}

func (fl *FileList) Get (key string) (file *VirtualFile, ok bool) {
	fl.RLock()
	defer fl.RUnlock()
	file, ok = fl.items[key]
	return
}

func (fl *FileList) Length() int  {
	fl.RLock()
	defer fl.RUnlock()
	return len(fl.items)
}

func (fl *FileList) Iter() <-chan FileListItem {
	c := make(chan FileListItem)

	go func() {
		fl.RLock()
		defer fl.RUnlock()

		for key, file := range fl.items {
			c <- FileListItem{key, file}
		}

		close(c)
	}()

	return c
}