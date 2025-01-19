package victronBleDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-victron/vebleapi"
	"github.com/koestler/go-victron/vedirectapi"
	"github.com/koestler/go-victron/velog"
	"github.com/koestler/go-victron/veregister"
	"log"
	"time"
)

type Config interface {
	AnnouncedName() string
	EncryptionKey() string
}

type DeviceStruct struct {
	device.State
	victronBleConfig Config

	bleAdapter *vebleapi.Adapter
	model      string
}

func NewDevice(
	deviceConfig device.Config,
	victronBleConfig Config,
	bleAdapter *vebleapi.Adapter,
	stateStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		victronBleConfig: victronBleConfig,
		bleAdapter:       bleAdapter,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	log.Printf("device[%s]: start Victron Bluetooth Low Energy source", c.Name())

	if c.Config().LogComDebug() {

	}

	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	c.model = api.Product.String()
	log.Printf("device[%s]: source: connect to %s", c.Name(), c.model)

	// filter registers by skip list
	api.Registers.FilterRegister(func() func(r veregister.Register) bool {
		rf := dataflow.RegisterFilter(c.Config().Filter())
		return func(r veregister.Register) bool {
			return rf(r)
		}
	}())
	addToRegisterDb(c.RegisterDb(), api.Registers)

	nonStaticRegisters := api.Registers
	nonStaticRegisters.FilterRegister(func(r veregister.Register) bool {
		return !r.Static()
	})

	deviceName := c.Name()
	valueHandler := vedirectapi.ValueHandler{
		Number: func(v vedirectapi.NumberRegisterValue) {
			output.Fill(dataflow.NewNumericRegisterValue(
				deviceName,
				Register{v},
				v.Value(),
			))
		},
		Text: func(v vedirectapi.TextRegisterValue) {
			output.Fill(dataflow.NewTextRegisterValue(
				deviceName,
				Register{v},
				v.Value(),
			))
		},
		Enum: func(v vedirectapi.EnumRegisterValue) {
			output.Fill(dataflow.NewEnumRegisterValue(
				deviceName,
				Register{v},
				v.Value().Idx(),
			))
		},
		FieldList: func(v vedirectapi.FieldListValue) {
			output.Fill(dataflow.NewTextRegisterValue(
				deviceName,
				Register{v},
				v.CommaString(),
			))
		},
	}

	var lastFetch time.Time
	fetch := func(regs veregister.RegisterList) (took time.Duration, err error) {
		// log fetching intervals
		if c.Config().LogDebug() {
			log.Printf("device[%s]: start fetching, since(lastFetch)=%.3fs", deviceName, time.Since(lastFetch).Seconds())
			lastFetch = time.Now()
		}

		start := time.Now()

		// execute a ping before fetching to make sure the device is reachable
		// also this makes orientation in the io log easier
		if e := api.Vd.Ping(); e != nil {
			err = fmt.Errorf("ping failed: %w", e)
			return
		}

		if e := api.StreamRegisterList(ctx, regs, valueHandler); e != nil {
			err = fmt.Errorf("fetching failed: %w", e)
			return
		}

		took = time.Since(start)

		if c.Config().LogDebug() {
			log.Printf("device[%s]: %d registers fetched, took=%.3fs", deviceName, regs.Len(), took.Seconds())
		}
		return
	}

	// fetch all registers
	if _, err := fetch(api.Registers); err != nil {
		return err, true
	}

	pollInterval := c.victronConfig.PollInterval()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			// run fetch whenever the ticker ticks
			// but when fetching took longer than pollInterval, fetch again immediately
			for {
				if took, err := fetch(nonStaticRegisters); err != nil {
					if errors.Is(err, vedirectapi.ErrCtxDone) {
						// do not return an error when the context is done
						err = nil
					}
					return err, false
				} else if took < pollInterval {
					break
				} else {
					// if there is an unused tick, consume it
					select {
					case <-ticker.C:
					default:
					}

					// reset ticker to pollInterval in case the next run is fast enough
					ticker.Reset(pollInterval)
				}
			}
		}
	}
}

func (c *DeviceStruct) getBleAdapterLogger() velog.Logger {
	if c.Config().LogComDebug() {
		return velog.DefaultLogger{
			Prefix: fmt.Sprintf("ble[%s]: ", c.Name()),
		}
	}

	return velog.NoOppLogger{}
}

func (c *DeviceStruct) Model() string {
	return c.model
}
