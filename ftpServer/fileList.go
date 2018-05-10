package ftpServer

import "sync"

type FileList struct {
	sync.RWMutex
	items map[string]*VirtualFile
}

type FileListItem struct {
	Path  string
	Value *VirtualFile
}

func NewFileList() *FileList {
	return &FileList{
		items: make(map[string]*VirtualFile),
	}
}

func (fl *FileList) Create(path string, file *VirtualFile) {
	fl.Lock()
	fl.items[path] = file
	fl.Unlock()
}

func (fl *FileList) Delete(path string) {
	fl.Lock()
	delete(fl.items, path)
	fl.Unlock()
}

func (fl *FileList) Get(path string) (file *VirtualFile, ok bool) {
	fl.RLock()
	defer fl.RUnlock()
	file, ok = fl.items[path]
	return
}

func (fl *FileList) Length() int  {
	fl.RLock()
	defer fl.RUnlock()
	return len(fl.items)
}

func (fl *FileList) Iterate() <-chan FileListItem {
	c := make(chan FileListItem)

	go func() {
		fl.RLock()
		defer fl.RUnlock()

		for path, file := range fl.items {
			c <- FileListItem{path, file}
		}

		close(c)
	}()

	return c
}