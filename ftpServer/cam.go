package ftpServer

import "log"

type cam struct {
	Name string
}

func camUpdatePicture(vf *VirtualFile) {
	log.Printf("ftpServer: camUpdatePicture deviceName=%v filePath=%v", vf.deviceName, vf.filePath)
}
