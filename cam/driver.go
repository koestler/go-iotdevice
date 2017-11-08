// Package sample is a sample server driver
package cam

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"log"

	"github.com/fclairamb/ftpserver/server"
	"strings"
)

// MainDriver defines a very basic ftpserver driver
type MainDriver struct {
}

// ClientDriver defines a very basic client driver
type ClientDriver struct {
	directories map[string]bool
	files       map[string]virtualFile
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

	drv := &MainDriver{}

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
	log.Printf("ftpcam-diver: WelcomeUser cc.ID=%v", cc.ID())

	cc.SetDebug(true)
	return fmt.Sprintf(
		"Welcome on go-ve-sensor ftpserver, your ID is %d, your IP:port is %s", cc.ID(), cc.RemoteAddr(),
	), nil
}

// AuthUser authenticates the user and selects an handling driver
func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {
	log.Printf("ftpcam-diver: AuthUser cc.ID=%v", cc.ID())

	if user == "bad" || pass == "bad" {
		return nil, errors.New("bad username or password")
	}

	return &ClientDriver{
		directories: make(map[string]bool),
		files:       make(map[string]virtualFile),
	}, nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {
	log.Printf("ftpcam-diver: UserLeft")
}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	log.Printf("ftpcam-diver: ChangeDirectory cc.ID=%v, directory=%v", cc.ID(), directory)

	if directory == "/debug" {
		cc.SetDebug(!cc.Debug())
		return nil
	}
	return nil
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	log.Printf("ftpcam-diver: MakeDirectory, cc.ID=%v, directory=%v", cc.ID(), directory)
	driver.directories[directory] = true
	return nil;
}

func (driver *ClientDriver) LogContent() {
	log.Printf("directories=%v", driver.directories)
	log.Printf("files=%v", driver.files)
}

// ListFiles lists the files of a directory
func (driver *ClientDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {
	log.Printf("ftpcam-diver: ListFiles cc.Path=%v", cc.Path())

	driver.LogContent()

	path := cc.Path()
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	files := make([]os.FileInfo, 0)
	for directory, _ := range driver.directories {
		if !strings.HasPrefix(directory, path) {
			continue
		}

		reminder := directory[len(path):]
		if len(reminder) < 1 || strings.Contains(reminder, "/") {
			// subdir -> ignore
			continue
		}

		files = append(files,
			virtualFileInfo{
				name: reminder,
				mode: os.FileMode(0666) | os.ModeDir,
				size: 4096,
			},
		)
	}

	return files, nil

	/*
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
*/
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	log.Printf("ftpcam-diver: OpenFile cc.ID=%v path=%v flag=%v", cc.ID(), path, flag)

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			delete(driver.files, path);
		}
	}

	if (flag & os.O_CREATE) != 0 {
		driver.files[path] = virtualFile{
			make([]byte, 0),
			0,
			0,
		}
	}

	file, ok := driver.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &file, nil
}

// GetFileInfo gets some info around a file or a directory
func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	log.Printf("ftpcam-diver: GetFileInfo cc.ID=%v path=%v", cc.ID(), path)

	if _, ok := driver.directories[path]; !ok {
		return nil, os.ErrNotExist
	}

	return &virtualFileInfo{
		name: path,
		mode: os.FileMode(0666) | os.ModeDir,
		size: 4096,
	}, nil
}

// CanAllocate gives the approval to allocate some data
func (driver *ClientDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	log.Printf("ftpcam-diver: CanAllocate cc.ID=%v size=%v", cc.ID(), size)
	return true, nil
}

// ChmodFile changes the attributes of the file
func (driver *ClientDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	log.Printf("ftpcam-diver: ChmodFile cc.ID=%v path=%v, mode=%v", cc.ID(), path, mode)
	return os.ErrPermission
}

// DeleteFile deletes a file or a directory
func (driver *ClientDriver) DeleteFile(cc server.ClientContext, path string) error {
	log.Printf("ftpcam-diver: DeleteFile cc.ID=%v path=%v", cc.ID(), path)
	return os.ErrPermission
}

// RenameFile renames a file or a directory
func (driver *ClientDriver) RenameFile(cc server.ClientContext, from, to string) error {
	log.Printf("ftpcam-diver: RenameFile cc.ID=%v from=%v to=%v", cc.ID(), from, to)
	return os.ErrPermission
}

// The virtual file is an example of how you can implement a purely virtual file
type virtualFile struct {
	content     []byte // Content of the file
	readOffset  int    // Reading offset
	writeOffset int    // Reading offset
}

func (f *virtualFile) Close() error {
	log.Printf("ftpcam-driver: virtualFile.Close f=%v", f)
	return nil
}

func (f *virtualFile) Read(buffer []byte) (int, error) {
	log.Printf("ftpcam-driver: virtualFile.Read f=%v", f)

	n := copy(buffer, f.content[f.readOffset:])
	f.readOffset += n
	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func (f *virtualFile) Seek(n int64, w int) (int64, error) {
	log.Printf("ftpcam-driver: virtualFile.Seek f=%v", f)

	return 0, nil
}

func (f *virtualFile) Write(buffer []byte) (int, error) {
	log.Printf("ftpcam-driver: virtualFile.Write f=%v", f)

	n := copy(f.content[f.writeOffset:], buffer)
	f.writeOffset += n

	return n, nil
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
