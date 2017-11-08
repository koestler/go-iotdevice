package cam

import (
	"os"
	"time"
)

type VirtualFileInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (f VirtualFileInfo) Name() string {
	return f.name
}

func (f VirtualFileInfo) Size() int64 {
	return f.size
}

func (f VirtualFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f VirtualFileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f VirtualFileInfo) ModTime() time.Time {
	return time.Now().UTC()
}

func (f VirtualFileInfo) Sys() interface{} {
	return nil
}
