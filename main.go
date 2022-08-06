package main

import (
	"fmt"

	"github.com/idertator/eogrec/devices"
	"github.com/idertator/eogrec/models"
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
	err := bitalino.Connect("/dev/cu.BITalino-9D-70", 115200, 1000)

	if err != nil {
		return err
	}

	err = bitalino.Initialize(1, 2)
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

	return bitalino.Close()
}

func TestBitalinoDataFetching() error {
	data := make([]models.Sample, 100)
	fmt.Println("Fetching Data")

	bitalino := devices.Bitalino{}
	err := bitalino.Connect("/dev/cu.BITalino-9D-70", 115200, 1000)

	if err != nil {
		return err
	}

	err = bitalino.Initialize(1, 2)
	if err != nil {
		return err
	}

	err = bitalino.Start()
	if err != nil {
		return err
	}

	fmt.Println("Here")

	err = bitalino.Read(data, 100)
	if err != nil {
		return err
	}

	err = bitalino.Stop()
	if err != nil {
		return err
	}

	for idx, sample := range data {
		fmt.Printf("%d -> %d (%d, %d)", idx, sample.Index, sample.Horizontal, sample.Vertical)
	}

	return bitalino.Close()
}

func main() {
	PrintPorts()

	// err := PrintBitalinoInfo()
	// if err != nil {
	// 	panic(err)
	// }

	err := TestBitalinoDataFetching()
	if err != nil {
		panic(err)
	}
}
