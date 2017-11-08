package cam

import (
	"image"
	"github.com/koestler/go-ve-sensor/config"

	"github.com/fclairamb/ftpserver/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"os"
	"os/signal"
	"syscall"
)

type FtpCam struct {
	Name          string
	DirectoryName string
	CurrentImage  *image.NRGBA
}

var (
	ftpServer *server.FtpServer
)

func FtpCamStart(config config.CamConfig) {
	// Setting up the logger
	logger := log.With(
		log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)),
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	// Loading the driver
	driver, err := NewDriver("")

	if err != nil {
		level.Error(logger).Log("msg", "Could not load the driver", "err", err)
		return
	}

	// Instantiating the server by passing our driver implementation
	ftpServer = server.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	//ftpServer.Logger = log.With(logger, "component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if err := ftpServer.ListenAndServe(); err != nil {
		level.Error(logger).Log("msg", "Problem listening", "err", err)
	}
}

func signalHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	for {
		switch <-ch {
		case syscall.SIGTERM:
			ftpServer.Stop()
			break
		}
	}
}
