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

func main() {
	PrintPorts()

	bitalino := devices.Bitalino{}
	// err := bitalino.Connect("/dev/cu.BITalino-9D-70", 115200, 1000, []uint8{1, 2})
	err := bitalino.Connect("/dev/cu.BITalino-6A-36", 115200, 1000, []uint8{1, 2})

	if err != nil {
		panic(err)
	}

	err = bitalino.Initialize()
	if err != nil {
		panic(err)
	}

	version, err := bitalino.Version()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)

	battery, err := bitalino.Battery()
	if err != nil {
		panic(err)
	}

	fmt.Println(battery)
}
