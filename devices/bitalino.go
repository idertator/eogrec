package devices

import (
	"errors"
	"math"
	"strings"

	"github.com/idertator/eogrec/models"
)

const BITALINO_BATTERY_MIN_VALUE uint16 = 30
const BITALINO_BATTERY_MAX_VALUE uint16 = 650

type Bitalino struct {
	SerialDevice

	version   string
	recording bool
	counter   uint32
}

type BitalinoStatus struct {
	I1               byte
	I2               byte
	O1               byte
	O2               byte
	Battery          uint16
	BatteryThreshold uint8
	A1               uint16
	A2               uint16
	A3               uint16
	A4               uint16
	A5               uint16
	A6               uint16
}

// StartSection - Recordable

func (b *Bitalino) Name() string {
	return "BITalino"
}

func (b *Bitalino) Version() (string, error) {
	if b.recording == false {
		b.Send([]byte{0x07})

		buff := make([]byte, 32)

		_, err := b.RecvUntil(buff, []byte("\n")[0])

		if err != nil {
			return "", err
		}

		version := strings.Split(string(buff[:]), "_")[1]

		return version, nil
	}
	return "", errors.New("Cannot check version while recording")
}

func (b *Bitalino) Battery() (uint8, error) {
	if b.recording == false {
		status, err := b.Status()
		if err != nil {
			return 0, err
		}
		if status.Battery > BITALINO_BATTERY_MAX_VALUE {
			status.Battery = BITALINO_BATTERY_MAX_VALUE
		}
		bat := status.Battery - BITALINO_BATTERY_MIN_VALUE
		max := BITALINO_BATTERY_MAX_VALUE - BITALINO_BATTERY_MIN_VALUE
		percent := math.Round((float64(bat) / float64(max)) * 100.0)
		return uint8(percent), nil
	}
	return 0, errors.New("Cannot check battery while recording")
}

func (b *Bitalino) AvailableSampleRates() []uint16 {
	return []uint16{1, 10, 100, 1000}
}

func (b *Bitalino) AvailableChannels() []uint8 {
	return []uint8{1, 2, 3, 4, 5, 6}
}

func (b *Bitalino) setSampleRate(rate uint16) error {
	if b.recording == false {
		var CMDs = map[uint16][]byte{
			1:    {0x03},
			10:   {0x43},
			100:  {0x83},
			1000: {0xC3},
		}
		cmd := CMDs[rate]

		sent, err := b.Send(cmd)
		if err != nil {
			return err
		}

		if sent != 1 {
			return errors.New("The amount of data sent is wrong")
		}

		return nil
	}
	return errors.New("Cannot set sample rate while recording")
}

func (b *Bitalino) setChannels(channels []uint8) error {
	if b.recording == false {
		if len(channels) == 0 {
			return errors.New("You should specify at least one channel")
		}

		if len(channels) > 4 {
			return errors.New("The maximum amount of channels supported is 4")
		}

		var cmd uint8 = 1
		for _, channel := range channels {
			cmd = cmd | (1 << (2 + channel))
		}

		sent, err := b.Send([]byte{cmd})

		if err != nil {
			return err
		}

		if sent != 1 {
			return errors.New("The amount of data sent is wrong")
		}

		return nil
	}
	return errors.New("Cannot set channels rate while recording")
}

func (b *Bitalino) setBatteryThreshold(threshold uint8) error {
	if threshold > 63 {
		return errors.New("Battery threshold must be in the range of 0 to 63")
	}

	cmd := threshold << 2

	_, err := b.Send([]byte{cmd})
	if err != nil {
		return err
	}

	return nil
}

func (b *Bitalino) Initialize(horizontalChannel byte, verticalChannel byte) error {
	if b.recording == false {
		b.HorizontalChannel = horizontalChannel
		b.VerticalChannel = verticalChannel

		version, err := b.Version()

		if err != nil {
			return err
		}

		b.version = version

		if err := b.setBatteryThreshold(uint8(BITALINO_BATTERY_MIN_VALUE)); err != nil {
			return err
		}

		return nil
	}
	return errors.New("Cannot initialize if already recording")
}

func (b *Bitalino) Start() error {
	if b.recording == false {
		b.counter = 0
		if err := b.setSampleRate(b.SamplingRate); err != nil {
			return err
		}

		if err := b.setChannels([]byte{b.HorizontalChannel, b.VerticalChannel}); err != nil {
			return err
		}

		b.recording = true

		return nil
	}
	return errors.New("Already recording")
}
func (b *Bitalino) Stop() error {
	if b.recording == true {
		_, err := b.Send([]byte{0})
		return err
	} else {
		if b.version != "v5.2" {
			_, err := b.Send([]byte{255})
			return err
		}
	}
	return errors.New("Already stopped")
}

func (b *Bitalino) Read(samples []models.Sample, n uint32) error {
	buff := make([]byte, 4)
	var i uint32
	for i = 0; i < n; i++ {
		count, err := b.RecvN(buff, 4)
		if err != nil {
			return err
		}
		if count != 4 {
			return errors.New("Packets must have 4 bytes")
		}

		packet_crc := buff[3] & 0x0F
		buff[3] = buff[3] & 0xF0

		computed_crc := CRC(buff, 4)

		if packet_crc == (computed_crc & 0x0F) {
			samples[i].Index = b.counter
			samples[i].Horizontal = ((uint32(buff[2] & 0x0F)) << 6) | (uint32(buff[1] >> 2))
			samples[i].Vertical = ((uint32(buff[1] & 0x03)) << 8) | uint32(buff[0])
		} else {
			samples[i].Index = 0xFFFFFFFF
			samples[i].Horizontal = 0xFFFFFFFF
			samples[i].Vertical = 0xFFFFFFFF
		}

		b.counter++
	}
	return nil
}

func (b *Bitalino) Close() error {
	if b.recording {
		b.Stop()
	}
	err := b.Serial.Close()
	return err
}

// EndSection

// StartSection - Own Methods

func (b *Bitalino) Status() (BitalinoStatus, error) {
	if b.recording == false {
		b.Send([]byte{0x0B})

		if b.version == "v5.2" {
			buff := make([]byte, 17)

			_, err := b.RecvN(buff, 17)

			if err != nil {
				return BitalinoStatus{}, err
			}

			packet_crc := buff[16] & 0x0F
			buff[16] = buff[16] & 0xF0

			computed_crc := CRC(buff, 17)

			if packet_crc == (computed_crc & 0x0F) {
				return BitalinoStatus{
					A1:               (uint16(buff[1]) << 8) | uint16(buff[0]),
					A2:               (uint16(buff[3]) << 8) | uint16(buff[2]),
					A3:               (uint16(buff[5]) << 8) | uint16(buff[4]),
					A4:               (uint16(buff[7]) << 8) | uint16(buff[6]),
					A5:               (uint16(buff[9]) << 8) | uint16(buff[8]),
					A6:               (uint16(buff[11]) << 8) | uint16(buff[10]),
					Battery:          (uint16(buff[13]) << 8) | uint16(buff[12]),
					BatteryThreshold: buff[14],
					O2:               buff[16] >> 4 & 0x01,
					O1:               buff[16] >> 5 & 0x01,
					I2:               buff[16] >> 6 & 0x01,
					I1:               buff[16] >> 7 & 0x01,
				}, nil
			}
		} else {
			buff := make([]byte, 16)

			_, err := b.RecvN(buff, 16)

			if err != nil {
				return BitalinoStatus{}, err
			}

			packet_crc := buff[15] & 0x0F
			buff[15] = buff[15] & 0xF0

			computed_crc := CRC(buff, 16)

			if packet_crc == (computed_crc & 0x0F) {
				return BitalinoStatus{
					A1:               (uint16(buff[1]) << 8) | uint16(buff[0]),
					A2:               (uint16(buff[3]) << 8) | uint16(buff[2]),
					A3:               (uint16(buff[5]) << 8) | uint16(buff[4]),
					A4:               (uint16(buff[7]) << 8) | uint16(buff[6]),
					A5:               (uint16(buff[9]) << 8) | uint16(buff[8]),
					A6:               (uint16(buff[11]) << 8) | uint16(buff[10]),
					Battery:          (uint16(buff[13]) << 8) | uint16(buff[12]),
					BatteryThreshold: buff[14],
					O2:               buff[15] >> 4 & 0x01,
					O1:               buff[15] >> 5 & 0x01,
					I2:               buff[15] >> 6 & 0x01,
					I1:               buff[15] >> 7 & 0x01,
				}, nil
			}
		}
		return BitalinoStatus{}, errors.New("Packet Checksum Failed")
	}
	return BitalinoStatus{}, errors.New("Cannot check status while recording")
}

// EndSection

// StartSection - Utilities

func CRC(data []byte, n uint32) byte {
	var i uint32
	var crc byte
	crc = 0
	for i = 0; i < n; i++ {
		var bit int8
		for bit = 7; bit >= 0; bit-- {
			crc = crc << 1
			if (crc & 0x10) > 0 {
				crc = crc ^ 0x03
			}
			crc = crc ^ ((data[i] >> bit) & 0x01)
		}
	}
	return crc
}

// EndSection
