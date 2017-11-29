package ftpServer

import "log"

func camUpdatePicture(vf *VirtualFile) {
	log.Printf("ftpServer: camUpdatePicture deviceName=%v filePath=%v", vf.deviceName, vf.filePath)
}
