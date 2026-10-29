package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/koestler/go-iotdevice/v3/internal/config"
	"github.com/koestler/go-iotdevice/v3/internal/run"
)

// is set through linker by build.sh
var buildVersion string
var buildTime string

type CmdOptions struct {
	Version bool           `long:"version" description:"Print the build version and timestamp"`
	Config  flags.Filename `short:"c" long:"config" description:"Config File in yaml format" default:"./config.yaml"`
	DryRun  bool           `short:"d" long:"dry-run" description:"Read and check the config and exit."`
}

const (
	ExitSuccess         = 0
	ExitDueToCmdOptions = 1
	ExitDueToConfig     = 2
)

func getCmdOptions() (cmdOptions CmdOptions, cmdName string) {
	// parse command line options
	parser := flags.NewParser(&cmdOptions, flags.Default)
	parser.Usage = "[-c <path to yaml config file>]"
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(ExitSuccess)
		} else {
			os.Exit(ExitDueToCmdOptions)
		}
	}

	if cmdOptions.Version {
		if buildVersion == "" {
			fmt.Println("github.com/koestler/go-iotdevice/v3 locally built version")
		} else {
			fmt.Println("github.com/koestler/go-iotdevice/v3 version:", buildVersion)
			fmt.Println("built at:", buildTime)
		}
		os.Exit(ExitSuccess)
	}

	return cmdOptions, parser.Name
}

func getConfig(cmdOptions CmdOptions, cmdName string) *config.Config {
	// read, transform and validate configuration
	cfg, err := config.ReadConfigFile(cmdName, string(cmdOptions.Config), false)
	if len(err) > 0 {
		for _, e := range err {
			log.Printf("config: error: %v", e)
		}
		os.Exit(ExitDueToConfig)
	}

	if cfg.LogConfig() {
		if err := cfg.PrintConfig(); err != nil {
			log.Printf("config: cannot print: %s", err)
		}
	}

	return &cfg
}

func main() {
	// read cmd parameters and configuration file; on error: os.Exit
	cmdOptions, cmdName := getCmdOptions()
	cfg := getConfig(cmdOptions, cmdName)
	if cmdOptions.DryRun {
		os.Exit(ExitSuccess)
	}

	// call defer statements before os.Exit
	func() {
		if cfg.LogWorkerStart() {
			log.Printf("main: start go-iotdevice version=%s", buildVersion)
		}

		// start storage
		stateStorageLogPrefix := ""
		if cfg.LogStateStorageDebug() {
			stateStorageLogPrefix = "stateStorage"
		}
		stateStorage := run.Storage(stateStorageLogPrefix)
		defer stateStorage.Shutdown()

		commandStorageLogPrefix := ""
		if cfg.LogCommandStorageDebug() {
			commandStorageLogPrefix = "commandStorage"
		}
		commandStorage := run.Storage(commandStorageLogPrefix)
		defer commandStorage.Shutdown()

		// start modbus device handlers
		modbusPool := run.Modbus(cfg)
		defer modbusPool.Shutdown()

		// start device pool
		devicePool := run.DevicePool()
		defer devicePool.Shutdown()

		// start non mqtt devices
		run.NonMqttGensetDevices(cfg, devicePool, modbusPool, stateStorage, commandStorage)

		// start mqtt client pool
		mqttClientPool := run.MqttClient(cfg)
		defer mqttClientPool.Shutdown()

		// start mqtt clients
		run.MqttDevices(cfg, devicePool, mqttClientPool, stateStorage, commandStorage)

		// start genset devices
		run.GensetDevices(cfg, devicePool, stateStorage, commandStorage)

		// start mqtt forwarders
		run.MqttForwarders(cfg, devicePool, mqttClientPool, stateStorage, commandStorage)

		// start http server
		httpServer := run.HttpServer(cfg, devicePool, stateStorage, commandStorage, buildVersion)
		if httpServer != nil {
			defer httpServer.Shutdown()
		}

		// setup SIGTERM, SIGINT handlers
		gracefulStop := make(chan os.Signal, 1)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)

		// wait for something to trigger a shutdown
		sig := <-gracefulStop

		if cfg.LogWorkerStart() {
			log.Printf("main: graceful shutdown; caught signal: %+v", sig)
		}
	}()

	if cfg.LogWorkerStart() {
		log.Printf("main: shutdown completed; exit %d", ExitSuccess)
	}
	os.Exit(ExitSuccess)
}
