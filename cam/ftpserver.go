package cam

import (
	"github.com/fclairamb/ftpserver/server"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"os"
	"os/signal"
	"syscall"
	"log"
)

var ftpServer *server.FtpServer

func Run() {
	// Setting up the logger
	logger := kitlog.With(
		kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stdout)),
		"ts", kitlog.DefaultTimestampUTC,
		"caller", kitlog.DefaultCaller,
	)

	// Loading the driver
	listenHost := "0.0.0.0"
	listenPort := 2121

	driver, err := NewDriver(listenHost, listenPort)

	if err != nil {
		level.Error(logger).Log("msg", "ftpserver: Could not load the driver", "err", err)
		return
	}

	// Instantiating the server by passing our driver implementation
	ftpServer = server.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	//ftpServer.Logger = log.With(logger, "component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe

	go func() {
		log.Printf("ftpserver: listening on %v:%v", listenHost, listenPort)
		if err := ftpServer.ListenAndServe(); err != nil {
			level.Error(logger).Log("msg", "ftpserver: Problem listening", "err", err)
		}
	}()
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
