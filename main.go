package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type device struct {
	macAddress string
	name       string
}

type deviceStatus struct {
	device
	batteryStatus int
	connected     bool
}

var statusLine = regexp.MustCompile(`\t(.+): (.+)`)
var getBattery = regexp.MustCompile(`0x([[:xdigit:]]+)`)

func init() {
	log.SetFlags(0)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bluetoothctl", "devices")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not get a list of devices: %s", err.Error())
	}
	devices := getDevices(&out)
	status := make([]*deviceStatus, len(devices))
	var wg sync.WaitGroup
	for i, d := range devices {
		wg.Add(1)
		go func(i int, d *device) {
			defer wg.Done()
			status[i] = getDeviceStatus(d)
		}(i, d)
	}
	wg.Wait()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"MAC", "Name", "Connected", "Battery",
	})
	for _, s := range status {
		t.AppendRow(table.Row{
			s.macAddress, s.name, s.connected, s.batteryStatus,
		})
	}
	t.Render()

}

func getDeviceStatus(d *device) *deviceStatus {
	status := &deviceStatus{
		device: *d,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bluetoothctl", "info", d.macAddress)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not status for %s (%s)", d.name, d.macAddress)
	}
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if statusLine.MatchString(line) {
			parts := statusLine.FindStringSubmatch(line)
			switch parts[1] {
			case "Connected":
				if parts[2] == "yes" {
					status.connected = true
				}

			case "Battery Percentage":
				status.batteryStatus = getBatteryStatus(parts[2])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Could not parse status for %s (%s)", d.name, d.macAddress)
	}

	return status
}

func getBatteryStatus(value string) int {
	if getBattery.MatchString(value) {
		hex := getBattery.FindStringSubmatch(value)
		battery, err := strconv.ParseInt(hex[1], 16, 32)
		if err != nil {
			log.Fatalf("Could not parse battery status: %s - %s", hex[0], err.Error())
		}
		return int(battery)
	}
	return 0
}

func getDevices(reader io.Reader) []*device {
	scanner := bufio.NewScanner(reader)
	devices := make([]*device, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			log.Fatalf("Invalid device line: %s", line)
		}
		devices = append(devices, &device{
			macAddress: parts[1],
			name:       parts[2],
		})
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Could not parse device list: %s", err.Error())
	}
	return devices
}
