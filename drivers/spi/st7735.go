package spi

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"image/color"
	"time"

	"errors"
)

// Registers
const (
	NOP        = 0x00
	SWRESET    = 0x01
	RDDID      = 0x04
	RDDST      = 0x09
	SLPIN      = 0x10
	SLPOUT     = 0x11
	PTLON      = 0x12
	NORON      = 0x13
	INVOFF     = 0x20
	INVON      = 0x21
	DISPOFF    = 0x28
	DISPON     = 0x29
	CASET      = 0x2A
	RASET      = 0x2B
	RAMWR      = 0x2C
	RAMRD      = 0x2E
	PTLAR      = 0x30
	COLMOD     = 0x3A
	MADCTL     = 0x36
	MADCTL_MY  = 0x80
	MADCTL_MX  = 0x40
	MADCTL_MV  = 0x20
	MADCTL_ML  = 0x10
	MADCTL_RGB = 0x00
	MADCTL_BGR = 0x08
	MADCTL_MH  = 0x04
	RDID1      = 0xDA
	RDID2      = 0xDB
	RDID3      = 0xDC
	RDID4      = 0xDD
	FRMCTR1    = 0xB1
	FRMCTR2    = 0xB2
	FRMCTR3    = 0xB3
	INVCTR     = 0xB4
	DISSET5    = 0xB6
	PWCTR1     = 0xC0
	PWCTR2     = 0xC1
	PWCTR3     = 0xC2
	PWCTR4     = 0xC3
	PWCTR5     = 0xC4
	VMCTR1     = 0xC5
	PWCTR6     = 0xFC
	GMCTRP1    = 0xE0
	GMCTRN1    = 0xE1
	VSCRDEF    = 0x33
	VSCRSADD   = 0x37

	GREENTAB   Model = 0
	MINI80x160 Model = 1

	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)

type Model uint8
type Rotation uint8

// Config is the configuration for the display
type ST7735Config struct {
	Width        int16
	Height       int16
	Rotation     Rotation
	Model        Model
	RowOffset    int16
	ColumnOffset int16
}

// APA102Driver is a driver for the APA102 programmable RGB LEDs.
type ST7735Driver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	gobot.Commander

	bus      Connection
	dcPin    *Pin
	resetPin *Pin
	csPin    *Pin
	blPin    *Pin

	width        int16
	height       int16
	columnOffset int16
	rowOffset    int16
	rotation     Rotation
	batchLength  int16
	model        Model
	isBGR        bool
	batchData    []uint8
}

func (d *ST7735Driver) Name() string {
	return d.name
}

func (d *ST7735Driver) SetName(s string) {
	d.name = s
}

func (d *ST7735Driver) Start() error {
	// FIXME
	return nil
}

func (d *ST7735Driver) Halt() error {
	// FIXME
	return nil
}

func (d *ST7735Driver) Connection() gobot.Connection {
	// FIXME
	return nil
}

type Pin struct {
	pin string
	a   gpio.DigitalWriter
}

func NewPin(a gpio.DigitalWriter, pin string) *Pin {
	return &Pin{
		pin: pin,
		a:   a,
	}
}

func (p *Pin) DigitalWrite(b byte) {
	p.a.DigitalWrite(p.pin, b)
}

func (p *Pin) High() {
	p.DigitalWrite(1)
}

func (p *Pin) Low() {
	p.DigitalWrite(0)
}

func (p *Pin) Set(b bool) {
	if b {
		p.High()
	} else {
		p.Low()
	}
}

type STAdaptor interface {
	gpio.DigitalWriter
	Connector
}

//NewST7735Driver
func NewST7735Driver(a STAdaptor, options ...func(Config)) (*ST7735Driver, error) {

	// FIXME with: bus, chip

	c, err := a.GetSpiConnection(0, 0, 0, 8, 500000)
	if err != nil {
		return nil, err
	}

	d := &ST7735Driver{
		name:      gobot.DefaultName("ST7735"),
		connector: a,
		Config:    NewConfig(),

		bus:      c,
		dcPin:    NewPin(a, "13"),
		resetPin: NewPin(a, "11"),
		csPin:    NewPin(a, "24"),
		blPin:    NewPin(a, "15"),

		isBGR:     false,
		batchData: nil,
	}
	for _, option := range options {
		option(d)
	}
	return d, nil
}

// Configure initializes the display with default configuration
func (d *ST7735Driver) Configure(cfg ST7735Config) {
	d.model = cfg.Model
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		if d.model == MINI80x160 {
			d.width = 80
		} else {
			d.width = 128
		}
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 160
	}
	d.rotation = cfg.Rotation
	d.rowOffset = cfg.RowOffset
	d.columnOffset = cfg.ColumnOffset

	d.batchLength = d.width
	if d.height > d.width {
		d.batchLength = d.height
	}
	d.batchLength += d.batchLength & 1
	d.batchData = make([]uint8, d.batchLength*2)

	// reset the device
	d.resetPin.High()
	time.Sleep(5 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(20 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(150 * time.Millisecond)

	// Common initialization
	d.Command(SWRESET)
	time.Sleep(150 * time.Millisecond)
	d.Command(SLPOUT)
	time.Sleep(500 * time.Millisecond)
	d.Command(FRMCTR1)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(FRMCTR2)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(FRMCTR3)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(INVCTR)
	d.Data(0x07)
	d.Command(PWCTR1)
	d.Data(0xA2)
	d.Data(0x02)
	d.Data(0x84)
	d.Command(PWCTR2)
	d.Data(0xC5)
	d.Command(PWCTR3)
	d.Data(0x0A)
	d.Data(0x00)
	d.Command(PWCTR4)
	d.Data(0x8A)
	d.Data(0x2A)
	d.Command(PWCTR5)
	d.Data(0x8A)
	d.Data(0xEE)
	d.Command(VMCTR1)
	d.Data(0x0E)
	d.Command(COLMOD)
	d.Data(0x05)

	if d.model == GREENTAB {
		d.InvertColors(false)
	} else if d.model == MINI80x160 {
		d.isBGR = true
		d.InvertColors(true)
	}

	// common color adjustment
	d.Command(GMCTRP1)
	d.Data(0x02)
	d.Data(0x1C)
	d.Data(0x07)
	d.Data(0x12)
	d.Data(0x37)
	d.Data(0x32)
	d.Data(0x29)
	d.Data(0x2D)
	d.Data(0x29)
	d.Data(0x25)
	d.Data(0x2B)
	d.Data(0x39)
	d.Data(0x00)
	d.Data(0x01)
	d.Data(0x03)
	d.Data(0x10)
	d.Command(GMCTRN1)
	d.Data(0x03)
	d.Data(0x1D)
	d.Data(0x07)
	d.Data(0x06)
	d.Data(0x2E)
	d.Data(0x2C)
	d.Data(0x29)
	d.Data(0x2D)
	d.Data(0x2E)
	d.Data(0x2E)
	d.Data(0x37)
	d.Data(0x3F)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x02)
	d.Data(0x10)

	d.Command(NORON)
	time.Sleep(10 * time.Millisecond)
	d.Command(DISPON)
	time.Sleep(500 * time.Millisecond)

	if cfg.Model == MINI80x160 {
		d.Command(MADCTL)
		d.Data(0xC0)
	}

	d.SetRotation(d.rotation)

	d.blPin.High()
}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *ST7735Driver) Display() error {
	return nil
}

// SetPixel sets a pixel in the screen
func (d *ST7735Driver) SetPixel(x int16, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *ST7735Driver) setWindow(x, y, w, h int16) {
	if d.rotation == NO_ROTATION || d.rotation == ROTATION_180 {
		x += d.columnOffset
		y += d.rowOffset
	} else {
		x += d.rowOffset
		y += d.columnOffset
	}
	d.Tx([]uint8{CASET}, true)
	d.Tx([]uint8{uint8(x >> 8), uint8(x), uint8((x + w - 1) >> 8), uint8(x + w - 1)}, false)
	d.Tx([]uint8{RASET}, true)
	d.Tx([]uint8{uint8(y >> 8), uint8(y), uint8((y + h - 1) >> 8), uint8(y + h - 1)}, false)
	d.Command(RAMWR)
}

// SetScrollWindow sets an area to scroll with fixed top and bottom parts of the display
func (d *ST7735Driver) SetScrollArea(topFixedArea, bottomFixedArea int16) {
	d.Command(VSCRDEF)
	d.Tx([]uint8{
		uint8(topFixedArea >> 8), uint8(topFixedArea),
		uint8(d.height - topFixedArea - bottomFixedArea>>8), uint8(d.height - topFixedArea - bottomFixedArea),
		uint8(bottomFixedArea >> 8), uint8(bottomFixedArea)},
		false)
}

// SetScroll sets the vertical scroll address of the display.
func (d *ST7735Driver) SetScroll(line int16) {
	d.Command(VSCRSADD)
	d.Tx([]uint8{uint8(line >> 8), uint8(line)}, false)
}

// SpotScroll returns the display to its normal state
func (d *ST7735Driver) StopScroll() {
	d.Command(NORON)
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *ST7735Driver) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	c1 := uint8(c565 >> 8)
	c2 := uint8(c565)

	for i = 0; i < d.batchLength; i++ {
		d.batchData[i*2] = c1
		d.batchData[i*2+1] = c2
	}
	i = width * height
	for i > 0 {
		if i >= d.batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:i*2], false)
		}
		i -= d.batchLength
	}
	return nil
}

// FillRectangle fills a rectangle at a given coordinates with a buffer
func (d *ST7735Driver) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	k, l := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= l || (y+height) > l {
		return errors.New("rectangle coordinates outside display area")
	}
	k = width * height
	l = int16(len(buffer))
	if k != l {
		return errors.New("buffer length does not match with rectangle size")
	}

	d.setWindow(x, y, width, height)

	offset := int16(0)
	for k > 0 {
		for i := int16(0); i < d.batchLength; i++ {
			if offset+i < l {
				c565 := RGBATo565(buffer[offset+i])
				c1 := uint8(c565 >> 8)
				c2 := uint8(c565)
				d.batchData[i*2] = c1
				d.batchData[i*2+1] = c2
			}
		}
		if k >= d.batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:k*2], false)
		}
		k -= d.batchLength
		offset += d.batchLength
	}
	return nil
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *ST7735Driver) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *ST7735Driver) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	d.FillRectangle(x0, y, x1-x0+1, y, c)
}

// FillScreen fills the screen with a given color
func (d *ST7735Driver) FillScreen(c color.RGBA) {
	if d.rotation == NO_ROTATION || d.rotation == ROTATION_180 {
		d.FillRectangle(0, 0, d.width, d.height, c)
	} else {
		d.FillRectangle(0, 0, d.height, d.width, c)
	}
}

// SetRotation changes the rotation of the device (clock-wise)
func (d *ST7735Driver) SetRotation(rotation Rotation) {
	madctl := uint8(0)
	switch rotation % 4 {
	case 0:
		madctl = MADCTL_MX | MADCTL_MY
		break
	case 1:
		madctl = MADCTL_MY | MADCTL_MV
		break
	case 2:
		break
	case 3:
		madctl = MADCTL_MX | MADCTL_MV
		break
	}
	if d.isBGR {
		madctl |= MADCTL_BGR
	}
	d.Command(MADCTL)
	d.Data(madctl)

}

// Command sends a command to the display
func (d *ST7735Driver) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Command sends a data to the display
func (d *ST7735Driver) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

// Tx sends data to the display
func (d *ST7735Driver) Tx(data []byte, isCommand bool) {
	d.dcPin.Set(!isCommand)
	d.bus.Tx(data, nil)
}

// Size returns the current size of the display.
func (d *ST7735Driver) Size() (w, h int16) {
	if d.rotation == NO_ROTATION || d.rotation == ROTATION_180 {
		return d.width, d.height
	}
	return d.height, d.width
}

// EnableBacklight enables or disables the backlight
func (d *ST7735Driver) EnableBacklight(enable bool) {
	if enable {
		d.blPin.High()
	} else {
		d.blPin.Low()
	}
}

// InverColors inverts the colors of the screen
func (d *ST7735Driver) InvertColors(invert bool) {
	if invert {
		d.Command(INVON)
	} else {
		d.Command(INVOFF)
	}
}

// IsBGR changes the color mode (RGB/BGR)
func (d *ST7735Driver) IsBGR(bgr bool) {
	d.isBGR = bgr
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
