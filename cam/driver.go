// Package sample is a sample server driver
package cam

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"log"

	"github.com/fclairamb/ftpserver/server"
	"strings"
)

type VirtualFileSystem struct {
	directories map[string]bool
	files       map[string]*VirtualFile
}

// MainDriver defines a very basic ftpserver driver
type MainDriver struct {
	vfs VirtualFileSystem
}

// ClientDriver defines a very basic client driver
type ClientDriver struct {
	vfs VirtualFileSystem
}

// NewSampleDriver creates a sample driver
func NewDriver() (*MainDriver, error) {
	// create new virtual in-memory filesystem
	driver := &MainDriver{
		vfs: VirtualFileSystem{
			directories: make(map[string]bool),
			files:       make(map[string]*VirtualFile),
		},
	}

	return driver, nil
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

	return &ClientDriver{vfs: driver.vfs}, nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {
	log.Printf("ftpcam-diver: UserLeft")
}

// ChangeDirectory changes the current working directory
func (driver *ClientDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	log.Printf("ftpcam-diver: ChangeDirectory cc.ID=%v, directory=%v", cc.ID(), directory)

	// create directories on the fly
	driver.vfs.directories[directory] = true
	return nil
}

// MakeDirectory creates a directory
func (driver *ClientDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	log.Printf("ftpcam-diver: MakeDirectory, cc.ID=%v, directory=%v", cc.ID(), directory)
	driver.vfs.directories[directory] = true
	return nil;
}

func (driver *ClientDriver) LogContent() {
	log.Printf("directories=%v", driver.vfs.directories)
	log.Printf("files=%v", driver.vfs.files)
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

	for directory, _ := range driver.vfs.directories {
		if !strings.HasPrefix(directory, path) {
			continue
		}

		reminder := directory[len(path):]
		if len(reminder) < 1 || strings.Contains(reminder, "/") {
			// subdir -> ignore
			continue
		}

		files = append(files,
			VirtualFileInfo{
				name: reminder,
				mode: os.FileMode(0666) | os.ModeDir,
				size: 4096,
			},
		)
	}

	for fileName, file := range driver.vfs.files {
		if !strings.HasPrefix(fileName, path) {
			continue
		}

		reminder := fileName[len(path):]
		if len(reminder) < 1 || strings.Contains(reminder, "/") {
			// subdir -> ignore
			continue
		}

		files = append(files, file.getFileInfo(reminder))
	}

	return files, nil

}

func (vf *VirtualFile) getFileInfo(name string) VirtualFileInfo {
	return VirtualFileInfo{
		name: name,
		mode: os.FileMode(0666),
		size: vf.Size(),
	}
}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *ClientDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	log.Printf("ftpcam-diver: OpenFile cc.ID=%v path=%v flag=%v", cc.ID(), path, flag)

	// If we are writing and we are not in append mode, we should remove the file
	if (flag & os.O_WRONLY) != 0 {
		flag |= os.O_CREATE
		if (flag & os.O_APPEND) == 0 {
			delete(driver.vfs.files, path);
		}
	}

	if (flag & os.O_CREATE) != 0 {
		driver.vfs.files[path] = &VirtualFile{}
	}

	file, ok := driver.vfs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return file, nil
}

// GetFileInfo gets some info around a file or a directory
func (driver *ClientDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	log.Printf("ftpcam-diver: GetFileInfo cc.ID=%v path=%v", cc.ID(), path)

	if file, ok := driver.vfs.files[path]; ok {
		return file.getFileInfo(path), nil
	} else if _, ok := driver.vfs.directories[path]; !ok {
		return &VirtualFileInfo{
			name: path,
			mode: os.FileMode(0666) | os.ModeDir,
			size: 4096,
		}, nil
	}

	return nil, os.ErrNotExist
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
