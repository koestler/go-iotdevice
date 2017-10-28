package cam

import (
	"image"
	//"image/jpeg"
	"io/ioutil"
	"time"
	"log"
	"github.com/koestler/go-ve-sensor/config"
	"regexp"
)

type FtpCam struct {
	Name string
	DirectoryName string
	CurrentImage  *image.NRGBA
}

func FtpCamStart(config config.CamConfig) {

	if config.Type != "ftp" {
		log.Printf("ftpcam: unsupported camera type. config=%v", config)
		return
	}

	cam := &FtpCam{
		Name : config.Name,
		DirectoryName:config.DirectoryName,
	}

	go func() {
		for _ = range time.Tick(500 * time.Millisecond) {
			getNewestFile(cam.DirectoryName)
		}
	}()


}

func getNewestFile(directoryName string) {

	// get files sorted by name
	fileInfos, err := ioutil.ReadDir(directoryName)

	if err != nil {
		log.Print("ftpCam: error listing files: %v", err)
		return
	}

	fileNameRegex, _ := regexp.Compile("\\.(jpg|JPG)$")

	fileNames = make(string, 0, 20)

	for _, f := range fileInfos {
		if fileNameRegex.Match(f.Name()) {

		}
		log.Printf("filename: %s", f.Name())
	}

	log.Printf("ftpCam: fileInfos=%v", fileInfos)

}