package cam

import (
	"image"
	//"image/jpeg"
	"io/ioutil"
	"time"
	"log"
)

type FtpCam struct {
	directoryName string
	currentImage  *image.NRGBA
}

func FtpCamStart(directoryName string) {

	/*
	cam := &FtpCam{
		directoryName: directoryName,
	}
	*/

	go func() {
		for _ = range time.Tick(500 * time.Millisecond) {



		}
	}()


}

func getNewestFile(directoryName string) {

	fileInfos, err := ioutil.ReadDir(directoryName)

	if err != nil {
		log.Print("ftpCam: error listing files: %v", err)
		return
	}

	log.Printf("ftpCam: fileInfos=%v", fileInfos)

}