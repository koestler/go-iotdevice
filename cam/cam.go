package cam

import "log"

type cam struct {
	Name string
}

func camUpdatePicture(vf *VirtualFile) {
	log.Printf("cam: camUpdatePicture deviceName=%v filePath=%v", vf.deviceName, vf.filePath)
}
