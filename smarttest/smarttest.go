package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
)

type Device struct {
	Device struct {
		Name string `json:"name"`
	} `json:"device"`
}

type ScanResult struct {
	Devices []Device `json:"devices"`
}

// GetStorageDevices scans the system for all devices using smartctl --scan -j
func GetStorageDevices() ([]string, error) {
	cmd := exec.Command("smartctl", "--scan", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running smartctl --scan: %v", err)
	}

	var scanResult ScanResult
	err = json.Unmarshal(output, &scanResult)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON from smartctl: %v", err)
	}

	var devices []string
	for _, device := range scanResult.Devices {
		devices = append(devices, device.Device.Name)
	}
	return devices, nil
}

// RunSmartTest initiates a SMART test (short or long) on the given device
func RunSmartTest(device, testType string) error {
	fmt.Printf("Running %s SMART test on %s...\n", testType, device)
	cmd := exec.Command("smartctl", "-t", testType, "-j", device)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running smartctl on %s: %v", device, err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		return fmt.Errorf("error parsing JSON output from smartctl: %v", err)
	}

	fmt.Printf("SMART test (%s) initiated on %s:\n", testType, device)
	fmt.Println(string(output)) // Prints the full JSON output
	return nil
}

func main() {
	// Define a command-line flag for selecting the test type (short or long)
	testType := flag.String("test", "short", "SMART test type to run (short or long)")
	flag.Parse()

	// Validate the input for test type
	if *testType != "short" && *testType != "long" {
		log.Fatalf("Invalid test type: %s. Please use 'short' or 'long'.", *testType)
	}

	// Get the list of detected devices
	devices, err := GetStorageDevices()
	if err != nil {
		log.Fatalf("Failed to detect storage devices: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No storage devices found.")
		return
	}

	// Run the specified SMART test (short or long) on each detected device
	for _, device := range devices {
		err := RunSmartTest(device, *testType)
		if err != nil {
			log.Printf("Failed to run SMART test on %s: %v", device, err)
		}
	}
}
