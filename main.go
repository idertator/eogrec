package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"github.com/idertator/eogrec/devices"
	"github.com/idertator/eogrec/formats"
	"github.com/idertator/eogrec/gui"
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
	err := bitalino.Connect("/dev/rfcomm0", 115200, 1000)

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

	record, err := formats.CreateRecord("./test.dat")
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

	for i := 0; i < 100; i++ {
		err = bitalino.Read(data, 100)
		if err != nil {
			return err
		}

		err = record.AddSamples(data, 100)
		if err != nil {
			return err
		}
	}

	err = bitalino.Stop()
	if err != nil {
		return err
	}

	err = record.Close()
	if err != nil {
		return err
	}

	return bitalino.Close()
}

func PrintDataFile(filename string) error {
	samples, err := formats.ReadRecord(filename)
	if err != nil {
		return err
	}

	for _, sample := range samples {
		fmt.Printf("%d %d %d\n", sample.Index, sample.Horizontal, sample.Vertical)
	}

	return nil
}

func main() {
	// PrintBitalinoInfo()
	app := app.New()

	mainWindow := gui.CreateMainWindow(app)
	mainWindow.Show()

	app.Run()
}
