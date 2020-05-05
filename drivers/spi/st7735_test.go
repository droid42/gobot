package spi

import (
	"gobot.io/x/gobot/gobottest"
	"log"
	"testing"
)

type TestAdaptor struct {
}

func (t TestAdaptor) DigitalWrite(s string, b byte) (err error) {
	// FIXME
	return nil
}

func (t TestAdaptor) GetSpiConnection(busNum, chip, mode, bits int, maxSpeed int64) (device Connection, err error) {
	c := &TestConnector{}
	spiCon, _ := c.GetSpiConnection(0, 0, 0, 0, 0)
	return spiCon, nil
}

func (t TestAdaptor) GetSpiDefaultBus() int {
	panic("implement me")
}

func (t TestAdaptor) GetSpiDefaultChip() int {
	panic("implement me")
}

func (t TestAdaptor) GetSpiDefaultMode() int {
	panic("implement me")
}

func (t TestAdaptor) GetSpiDefaultBits() int {
	panic("implement me")
}

func (t TestAdaptor) GetSpiDefaultMaxSpeed() int64 {
	panic("implement me")
}

func TestNewST7735Driver(t *testing.T) {

	d, err := NewST7735Driver(TestAdaptor{})
	if err != nil {
		log.Fatal(err)
	}

	d.Configure(ST7735Config{
		Width:        128,
		Height:       128,
		Rotation:     0,
		Model:        0,
		RowOffset:    0,
		ColumnOffset: 0,
	})

	gobottest.Assert(t, int16(128), d.height)
	gobottest.Assert(t, int16(128), d.width)
}
