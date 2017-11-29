package ftpServer

import "log"

func onFileClose(vf *VirtualFile) {
	log.Printf("ftpServer: onFileClose device.name=%v filePath=%v", vf.device.Name, vf.filePath)

	picture := Picture{
		created: vf.modified,
		path:    vf.filePath,
	}
	PictureStorage.SetPicture(vf.device, &picture)
}
