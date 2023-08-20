package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/koestler/go-iotdevice/config"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

// is set through linker by build.sh
var buildVersion string
var buildTime string

type CmdOptions struct {
	Version    bool           `long:"version" description:"Print the build version and timestamp"`
	Config     flags.Filename `short:"c" long:"config" description:"Config File in yaml format" default:"./config.yaml"`
	CpuProfile flags.Filename `long:"cpuprofile" description:"write cpu profile to <file>"`
	MemProfile flags.Filename `long:"memprofile" description:"write memory profile to <file>"`
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
		fmt.Println("github.com/koestler/go-iotdevice version:", buildVersion)
		fmt.Println("build at:", buildTime)
		os.Exit(ExitSuccess)
	}

	return cmdOptions, parser.Name
}

func getConfig(cmdOptions CmdOptions, cmdName string) *config.Config {
	// read, transform and validate configuration
	cfg, err := config.ReadConfigFile(cmdName, string(cmdOptions.Config))
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

	// call defer statements before os.Exit
	exitCode := func() (exitCode int) {
		if cfg.LogWorkerStart() {
			log.Printf("main: start go-iotdevice version=%s", buildVersion)
		}

		// start cpu profiling if enabled
		if runCpuProfile(string(cmdOptions.CpuProfile)) {
			defer pprof.StopCPUProfile()
		}

		// start storage
		stateStorageLogPrefix := ""
		commandStorageLogPrefix := ""
		if cfg.LogStorageDebug() {
			stateStorageLogPrefix = "stateStorage"
			commandStorageLogPrefix = "commandStorage"
		}

		stateStorage := runStorage(stateStorageLogPrefix)
		defer stateStorage.Shutdown()
		commandStorage := runStorage(commandStorageLogPrefix)
		defer commandStorage.Shutdown()

		// start mqtt clients
		mqttClientPoolInstance := runMqttClient(cfg)
		defer mqttClientPoolInstance.Shutdown()

		// start hass discovery
		hassDiscoveryInstance := runHassDisovery(cfg, stateStorage, mqttClientPoolInstance)
		defer hassDiscoveryInstance.Shutdown()

		// start modbus device handlers
		modbusPoolInstance := runModbus(cfg)
		defer modbusPoolInstance.Shutdown()

		// start devices
		devicePoolInstance := runDevices(cfg, mqttClientPoolInstance, modbusPoolInstance, stateStorage, commandStorage)
		defer devicePoolInstance.Shutdown()

		// start http server
		httpServerInstance := runHttpServer(cfg, devicePoolInstance, stateStorage, commandStorage)
		if httpServerInstance != nil {
			defer httpServerInstance.Shutdown()
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
		exitCode = ExitSuccess

		// write memory profile; after that defer will run the shutdown methods
		writeMemProfile(string(cmdOptions.MemProfile))

		return
	}()

	if cfg.LogWorkerStart() {
		log.Printf("main: stutdown completed; exit %d", exitCode)
	}
	os.Exit(exitCode)
}
