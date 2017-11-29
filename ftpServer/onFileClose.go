package ftpServer

import (
	"github.com/koestler/go-ve-sensor/storage"
)

func onFileClose(vf *VirtualFile) {
	//log.Printf("ftpServer: onFileClose device.name=%v filePath=%v", vf.device.Name, vf.filePath)

	picture := storage.Picture{
		Created: vf.modified,
		Path:    vf.filePath,
	}
	storage.PictureDb.SetPicture(vf.device, &picture)
}
