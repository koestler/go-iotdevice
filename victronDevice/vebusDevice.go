package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mk2driver"
	"github.com/tarm/serial"
	"golang.org/x/exp/maps"
	"io"
	"log"
	"time"
)

const unavailableInterval = 2 * time.Second

func runVebus(ctx context.Context, c *DeviceStruct, output dataflow.Fillable) (err error, immediateError bool) {
	log.Printf("device[%s]: start vebus source", c.Name())

	// open mk2 device
	mk2, err := getMk2Device(c.victronConfig.Device(), c.Config().LogComDebug())
	if err != nil {
		return err, true
	}
	defer mk2.Close()

	// setup registers
	registers := maps.Values(RegisterListMultiplus)
	registers = dataflow.FilterRegisters(registers, c.Config().Filter())
	c.RegisterDb().AddStruct(registers...)

	unavailableTimer := time.NewTicker(time.Hour)
	defer unavailableTimer.Stop()
	defer func() {
		c.SetAvailable(false)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-unavailableTimer.C:
			c.SetAvailable(false)
		case info := <-mk2.C():
			if c.Config().LogDebug() {
				log.Printf("device[%s]: recevied frame: %#v", c.Name(), info)
			}

			log.Printf("device[%s]: recevied errors: %v", c.Name(), info.Errors)

			if !info.Valid {
				return fmt.Errorf("invalid frame received"), false
			}

			c.SetAvailable(true)
			unavailableTimer.Reset(unavailableInterval)

			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["BatteryVoltage"],
				info.BatVoltage,
			))
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["BatteryCurrent"],
				info.BatCurrent,
			))

			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["InputVoltage"],
				info.InVoltage,
			))
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["InputCurrent"],
				info.InCurrent,
			))
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["InputFrequency"],
				info.InFrequency,
			))

			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["OutputVoltage"],
				info.OutVoltage,
			))
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["OutputCurrent"],
				info.OutCurrent,
			))
			output.Fill(dataflow.NewNumericRegisterValue(
				c.Name(),
				RegisterListMultiplus["OutputFrequency"],
				info.OutFrequency,
			))

			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["Mains"],
				ledStateToOnOffEnum(info.LEDs[mk2driver.LedMain]),
			))
			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["ChargerMode"],
				ledStateToChargerModeEnum(info.LEDs),
			))

			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["Inverter"],
				ledStateToOnOffEnum(info.LEDs[mk2driver.LedInverter]),
			))
			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["Overload"],
				ledStateToFaultEnum(info.LEDs[mk2driver.LedOverload]),
			))
			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["LowBattery"],
				ledStateToFaultEnum(info.LEDs[mk2driver.LedLowBattery]),
			))
			output.Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				RegisterListMultiplus["Temperature"],
				ledStateToFaultEnum(info.LEDs[mk2driver.LedTemperature]),
			))

			output.Fill(dataflow.NewTextRegisterValue(
				c.Name(),
				RegisterListMultiplus["Version"],
				fmt.Sprintf("0x%x", info.Version),
			))
		}
	}

}

func getMk2Device(dev string, logDebug bool) (mk2driver.Mk2, error) {
	var p io.ReadWriteCloser
	var err error

	serialConfig := &serial.Config{Name: dev, Baud: 2400}
	p, err = serial.OpenPort(serialConfig)
	if err != nil {
		return nil, err
	}

	mk2, err := mk2driver.NewMk2Connection(p, logDebug)
	if err != nil {
		return nil, err
	}

	return mk2, nil
}
