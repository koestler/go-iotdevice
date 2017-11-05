// Package sample is a sample server driver
package cam

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/fclairamb/ftpserver/server"
	"github.com/go-kit/kit/log"
)

// MainDriver defines a very basic ftpserver driver
type MainDriver struct {
	Logger  log.Logger // Logger
	BaseDir string     // Base directory from which to serve file
}

// ClientDriver defines a very basic client driver
type ClientDriver struct {
	BaseDir string // Base directory from which to server file
}

// NewSampleDriver creates a sample driver
func NewDriver(dir string) (*MainDriver, error) {
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "ftpserver")
		if err != nil {
			return nil, fmt.Errorf("could not find a temporary dir, err: %v", err)
		}
	}

	drv := &MainDriver{
		Logger:  log.NewNopLogger(),
		BaseDir: dir,
	}

	return drv, nil
}

// GetSettings returns some general settings around the server setup
func (driver *MainDriver) GetSettings() *server.Settings {

	var config server.Settings

	config.ListenHost = "0.0.0.0"
	config.ListenPort = 2121
	config.PublicHost = "::1"
	config.MaxConnections = 32
	config.DataPortRange = &server.PortRange{Start: 2122, End: 2200}
	config.DisableMLSD = true
	config.NonStandardActiveDataPort = false

	return &config
}

// GetTLSConfig returns a TLS Certificate to use
func (driver *MainDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("tls not supported")
}

// WelcomeUser is called to send the very first welcome message
func (driver *MainDriver) WelcomeUser(cc server.ClientContext) (string, error) {
	cc.SetDebug(true)
	// This will remain the official name for now
	return fmt.Sprintf(
		"Welcome on go-ve-sensor ftpserver, you're on dir %s, your ID is %d, your IP:port is %s",
		driver.BaseDir, cc.ID(), cc.RemoteAddr(),
	), nil
}

// AuthUser authenticates the user and selects an handling driver
func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {
	if user == "bad" || pass == "bad" {
		return nil, errors.New("bad username or password")
	}

	return &ClientDriver{BaseDir: driver.BaseDir}, nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {

}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	if directory == "/debug" {
		cc.SetDebug(!cc.Debug())
		return nil
	} else if directory == "/virtual" {
		return nil
	}
	_, err := os.Stat(driver.BaseDir + directory)
	return err
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.BaseDir+directory, 0777)
}

// ListFiles lists the files of a directory
func (driver *ClientDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {

	if cc.Path() == "/virtual" {
		files := make([]os.FileInfo, 0)
		files = append(files,
			virtualFileInfo{
				name: "localpath.txt",
				mode: os.FileMode(0666),
				size: 1024,
			},
			virtualFileInfo{
				name: "file2.txt",
				mode: os.FileMode(0666),
				size: 2048,
			},
		)
		return files, nil
	}

	path := driver.BaseDir + cc.Path()

	files, err := ioutil.ReadDir(path)

	// We add a virtual dir
	if cc.Path() == "/" && err == nil {
		files = append(files, virtualFileInfo{
			name: "virtual",
			mode: os.FileMode(0666) | os.ModeDir,
			size: 4096,
		})
	}

	return files, err
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {

	if path == "/virtual/localpath.txt" {
		return &virtualFile{content: []byte(driver.BaseDir)}, nil
	}

	path = driver.BaseDir + path

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			os.Remove(path)
		}
	}

	return os.OpenFile(path, flag, 0666)
}

// GetFileInfo gets some info around a file or a directory
func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	path = driver.BaseDir + path

	return os.Stat(path)
}

// CanAllocate gives the approval to allocate some data
func (driver *ClientDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

// ChmodFile changes the attributes of the file
func (driver *ClientDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	path = driver.BaseDir + path

	return os.Chmod(path, mode)
}

// DeleteFile deletes a file or a directory
func (driver *ClientDriver) DeleteFile(cc server.ClientContext, path string) error {
	path = driver.BaseDir + path

	return os.Remove(path)
}

// RenameFile renames a file or a directory
func (driver *ClientDriver) RenameFile(cc server.ClientContext, from, to string) error {
	from = driver.BaseDir + from
	to = driver.BaseDir + to

	return os.Rename(from, to)
}

// The virtual file is an example of how you can implement a purely virtual file
type virtualFile struct {
	content    []byte // Content of the file
	readOffset int    // Reading offset
}

func (f *virtualFile) Close() error {
	return nil
}

func (f *virtualFile) Read(buffer []byte) (int, error) {
	n := copy(buffer, f.content[f.readOffset:])
	f.readOffset += n
	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func (f *virtualFile) Seek(n int64, w int) (int64, error) {
	return 0, nil
}

func (f *virtualFile) Write(buffer []byte) (int, error) {
	return 0, nil
}

type virtualFileInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (f virtualFileInfo) Name() string {
	return f.name
}

func (f virtualFileInfo) Size() int64 {
	return f.size
}

func (f virtualFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f virtualFileInfo) IsDir() bool {
	return f.mode.IsDir()
}

func (f virtualFileInfo) ModTime() time.Time {
	return time.Now().UTC()
}

func (f virtualFileInfo) Sys() interface{} {
	return nil
}

func externalIP() (string, error) {
	// If you need to take a bet, amazon is about as reliable & sustainable a service as you can get
	rsp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
