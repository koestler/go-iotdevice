package ftpServer

// https://github.com/orcaman/writerseeker/blob/master/writerseeker.go
// bytes/reader

import (
	"io"
	"errors"
	"time"
	"github.com/koestler/go-ve-sensor/deviceDb"
)

type VirtualFile struct {
	buffer        []byte
	writePosition int
	readPosition  int
	device        *deviceDb.Device
	filePath      string
	modified      time.Time
}

type VirtualFileByCreated []*VirtualFile

func (s VirtualFileByCreated) Len() int { return len(s) }
func (s VirtualFileByCreated) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s VirtualFileByCreated) Less(i, j int) bool {
	return s[i].modified.Before(s[j].modified)
}

func (vf *VirtualFile) Write(p []byte) (n int, err error) {
	minCap := vf.writePosition + len(p)
	if minCap > cap(vf.buffer) { // Make sure buffer has enough capacity:
		buf2 := make([]byte, len(vf.buffer), minCap+len(p)) // add some extra
		copy(buf2, vf.buffer)
		vf.buffer = buf2
	}
	if minCap > len(vf.buffer) {
		vf.buffer = vf.buffer[:minCap]
	}
	copy(vf.buffer[vf.writePosition:], p)
	vf.writePosition += len(p)
	return len(p), nil
}

func (vf *VirtualFile) Seek(offset int64, whence int) (int64, error) {
	newPos, offs := 0, int(offset)
	switch whence {
	case io.SeekStart:
		newPos = offs
	case io.SeekCurrent:
		newPos = vf.writePosition + offs
	case io.SeekEnd:
		newPos = len(vf.buffer) + offs
	}
	if newPos < 0 {
		return 0, errors.New("negative result writePosition")
	}
	vf.writePosition = newPos
	return int64(newPos), nil
}

func (vf *VirtualFile) Size() int64 {
	return int64(len(vf.buffer))
}

func (vf *VirtualFile) Read(b []byte) (n int, err error) {
	if vf.readPosition >= len(vf.buffer) {
		return 0, io.EOF
	}
	n = copy(b, vf.buffer[vf.readPosition:])
	vf.readPosition += n
	return
}

func (vf *VirtualFile) Close() error {
	onFileClose(vf)
	return nil
}
