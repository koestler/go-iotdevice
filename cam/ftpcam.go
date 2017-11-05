package cam

import (
	"image"
	//"image/jpeg"
	"io/ioutil"
	"log"
	"github.com/koestler/go-ve-sensor/config"
	"regexp"
	"github.com/fsnotify/fsnotify"
)

type FtpCam struct {
	Name          string
	DirectoryName string
	CurrentImage  *image.NRGBA
}

func FtpCamStart(config config.CamConfig) {

	if config.Type != "ftp" {
		log.Printf("ftpcam: unsupported camera type. config=%v", config)
		return
	}

	cam := &FtpCam{
		Name:          config.Name,
		DirectoryName: config.DirectoryName,
	}

	//setupWatcher(cam.DirectoryName, cam.CurrentImage)

	getNewestFile(cam.DirectoryName)

}

func getNewestFile(directoryName string) {

	// get files sorted by name
	fileInfos, err := ioutil.ReadDir(directoryName)

	if err != nil {
		log.Print("ftpCam: error listing files: %v", err)
		return
	}

	fileNameRegex, _ := regexp.Compile("\\.(jpg|JPG)$")

	//fileNames := make(string, 0, 20)

	for _, f := range fileInfos {
		if fileNameRegex.Match([]byte(f.Name())) {
			log.Printf("filename: %s", f.Name())
		}
	}

}

func setupWatcher(directoryName string, curr chan string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	//done := make(chan bool)
	go func() {
		//current := ""

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					getNewestFile(directoryName)

					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(directoryName)
	if err != nil {
		log.Fatal(err)
	}
	//<-done
}
