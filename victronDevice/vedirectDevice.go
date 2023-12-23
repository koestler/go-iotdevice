package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-victron/vedirect"
	"github.com/koestler/go-victron/vedirectapi"
	"github.com/koestler/go-victron/veregisters"
	"log"
	"time"
)

func runVedirect(ctx context.Context, c *DeviceStruct, output dataflow.Fillable) (err error, immediateError bool) {
	log.Printf("device[%s]: start vedirect source", c.Name())

	vedirectConfig := vedirect.Config{}

	if c.Config().LogComDebug() {
		vedirectConfig.DebugLogger = log.New(
			log.Writer(),
			fmt.Sprintf("device[%s]: vedirect: ", c.Name()),
			log.LstdFlags|log.Lmsgprefix,
		)
	}

	if ioLog := c.victronConfig.IoLog(); ioLog != "" {
		if logger, err := vedirectapi.NewFileLogger(ioLog); err != nil {
			log.Printf("device[%s]: cannot log io: %s", c.Name(), err)
		} else {
			defer logger.Close()
			vedirectConfig.IoLogger = logger
		}
	}

	api, err := vedirectapi.NewRegistertApi(c.victronConfig.Device(), vedirectConfig)
	if err != nil {
		return err, true
	}
	defer func() {
		if err := api.Close(); err != nil {
			log.Printf("device[%s]: Close failed: %s", c.Name(), err)
		}
	}()

	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	c.model = api.Product.String()
	log.Printf("device[%s]: source: connect to %s", c.Name(), c.model)

	// filter registers by skip list
	api.Registers.FilterRegister(func() func(r veregisters.Register) bool {
		rf := dataflow.RegisterFilter(c.Config().Filter())
		return func(r veregisters.Register) bool {
			return rf(r)
		}
	}())
	addToRegisterDb(c.RegisterDb(), api.Registers)

	nonStaticRegisters := api.Registers
	nonStaticRegisters.FilterRegister(func(r veregisters.Register) bool {
		return !r.Static()
	})

	valueHandler := vedirectapi.ValueHandler{
		Number: func(v vedirectapi.NumberRegisterValue) {
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				Register{v.RegisterStruct},
				v.Value(),
			))
		},
		Text: func(v vedirectapi.TextRegisterValue) {
			output.Fill(dataflow.NewTextRegisterValue(
				c.Name(),
				Register{v.RegisterStruct},
				v.Value(),
			))
		},
		Enum: func(v vedirectapi.EnumRegisterValue) {
			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				Register{v.RegisterStruct},
				v.Idx(),
			))
		},
	}

	// start polling loop
	regs := api.Registers
	last := time.Now()
	minPollInterval := c.victronConfig.PollInterval()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		default:
		}

		start := time.Now()
		if err := api.StreamRegisterList(regs, valueHandler); err != nil {
			return fmt.Errorf("device[%s]: fetching failed: %s", c.Name(), err), false
		}
		took := time.Since(start)

		if c.Config().LogDebug() {
			log.Printf(
				"device[%s]: %d registers fetched, took=%.3fs, pollInterval=%.3fs",
				c.Name(),
				regs.Len(), took.Seconds(), time.Since(last).Seconds(),
			)
		}
		last = time.Now()

		// fetch static registers only once
		regs = nonStaticRegisters

		// limit to one fetch per poll interval
		if took < minPollInterval {
			time.Sleep(minPollInterval - took)
		}
	}
}
