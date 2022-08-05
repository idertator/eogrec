package main

import (
	"fmt"

	"github.com/idertator/eogrec/devices"
)

func PrintPorts() {
	ports, err := devices.PortList()

	if err != nil {
		panic(err)
	}

	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}
}

func PrintBitalinoInfo() error {
	bitalino := devices.Bitalino{}
	// err := bitalino.Connect("/dev/cu.BITalino-9D-70", 115200, 1000, []uint8{1, 2})
	err := bitalino.Connect("/dev/cu.BITalino-6A-36", 115200, 1000, []uint8{1, 2})

	if err != nil {
		return err
	}

	err = bitalino.Initialize()
	if err != nil {
		return err
	}

	version, err := bitalino.Version()
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("Version: %s", version)

	battery, err := bitalino.Battery()
	if err != nil {
		return err
	}

	fmt.Printf("Battery: %d%%\n", battery)
	return nil
}

func main() {
	PrintPorts()

	err := PrintBitalinoInfo()
	if err != nil {
		panic(err)
	}
}
