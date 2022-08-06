package devices

import (
	"github.com/idertator/eogrec/models"
	"go.bug.st/serial"
)

type SerialDevice struct {
	Port              string
	Baudrate          uint32
	SamplingRate      uint16
	HorizontalChannel byte
	VerticalChannel   byte
	Serial            serial.Port
}

type Recordable interface {
	Name() string
	Version() (string, error)
	Battery() (uint8, error)

	AvailableSampleRates() []uint16
	AvailableChannels() []uint8

	Initialize(horizontalChannel byte, verticalChannel byte) error
	Start() error
	Stop() error
	Read(samples []models.Sample, n uint32) error
	Close() error

	setSampleRate(rate uint16) error
	setChannels(channels []uint8) error
}

func (s *SerialDevice) Connect(port string, baudrate uint32, rate uint16) error {
	s.Port = port
	s.Baudrate = baudrate
	s.SamplingRate = rate

	mode := &serial.Mode{
		BaudRate: int(baudrate),
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	serialPort, err := serial.Open(port, mode)
	if err != nil {
		return err
	}

	s.Serial = serialPort

	return nil
}

func (s *SerialDevice) Send(data []byte) (uint32, error) {
	count, err := s.Serial.Write(data)
	if err != nil {
		return 0, err
	}
	return uint32(count), nil
}

func (s *SerialDevice) Recv(data []byte) (uint32, error) {
	count, err := s.Serial.Read(data)
	if err != nil {
		return 0, err
	}
	return uint32(count), nil
}

func (s *SerialDevice) RecvN(data []byte, n uint32) (uint32, error) {
	buff := make([]byte, 1)
	var idx uint32 = 0

	for idx < n {
		_, err := s.Serial.Read(buff)
		if err != nil {
			return 0, err
		}

		data[idx] = buff[0]
		idx += 1
	}

	return idx, nil
}

func (s *SerialDevice) RecvUntil(data []byte, until byte) (uint32, error) {
	buff := make([]byte, 1)
	var idx uint32 = 0

	for {
		_, err := s.Serial.Read(buff)
		if err != nil {
			return 0, err
		}

		data[idx] = buff[0]
		idx += 1

		if buff[0] == until {
			break
		}
	}
	return idx, nil
}

func PortList() ([]string, error) {
	ports, err := serial.GetPortsList()

	return ports, err
}
