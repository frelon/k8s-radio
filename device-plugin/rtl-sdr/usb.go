package rtlsdr

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"strconv"
)

type UsbDevice struct {
	Serial    string
	VendorID  string
	ProductID string
	Bus       int
	Dev       int
}

// supportedProducts is a map of supported products per vendor ID.
var supportedProducts = map[string][]string{
	// 0bda = RealTek, 2838 = RTL2838UHIDIR
	"0bda": {"2838"},
}

func (d UsbDevice) DevicePath() string {
	return fmt.Sprintf("/dev/bus/usb/%03d/%03d", d.Bus, d.Dev)
}

func ListUsbDevices(fsys fs.FS) ([]*UsbDevice, error) {
	const devicesPath = "sys/bus/usb/devices"

	entries, err := fs.ReadDir(fsys, devicesPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s': %w", devicesPath, err)
	}

	devices := []*UsbDevice{}
	for i := range entries {
		// Check if it is a directory or symlink
		if !entries[i].IsDir() && entries[i].Type()&fs.ModeSymlink == 0 {
			slog.Info("Skipping", slog.String("name", entries[i].Name()))
			continue
		}

		path := filepath.Join(devicesPath, entries[i].Name())
		dev, err := ReadUsbDevice(fsys, path)
		if err != nil {
			slog.Debug("failed reading USB device", slog.String("path", path), slog.Any("error", err))
			continue
		}

		if !isSupportedProduct(dev.VendorID, dev.ProductID) {
			continue
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

func isSupportedProduct(vendorID, productID string) bool {
	if prods, ok := supportedProducts[vendorID]; ok {
		return slices.Contains(prods, productID)
	}

	return false
}

func ReadUsbDevice(fsys fs.FS, dir string) (*UsbDevice, error) {
	vendorID, err := fs.ReadFile(fsys, filepath.Join(dir, "idVendor"))
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s/idVendor': %w", dir, err)
	}

	productID, err := fs.ReadFile(fsys, filepath.Join(dir, "idProduct"))
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s/idProduct': %w", dir, err)
	}

	busnum, err := fs.ReadFile(fsys, filepath.Join(dir, "busnum"))
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s/busnum': %w", dir, err)
	}

	bus, err := strconv.Atoi(string(bytes.TrimSuffix(busnum, []byte("\n"))))
	if err != nil {
		return nil, fmt.Errorf("failed converting '%s' to int: %w", busnum, err)
	}

	devnum, err := fs.ReadFile(fsys, filepath.Join(dir, "devnum"))
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s/devnum': %w", dir, err)
	}

	dev, err := strconv.Atoi(string(bytes.TrimSuffix(devnum, []byte("\n"))))
	if err != nil {
		return nil, fmt.Errorf("failed converting '%s' to int: %w", devnum, err)
	}

	serial, err := fs.ReadFile(fsys, filepath.Join(dir, "serial"))
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s/serial': %w", dir, err)
	}

	return &UsbDevice{
		VendorID:  string(bytes.TrimSuffix(vendorID, []byte("\n"))),
		ProductID: string(bytes.TrimSuffix(productID, []byte("\n"))),
		Bus:       bus,
		Dev:       dev,
		Serial:    string(bytes.TrimSuffix(serial, []byte("\n"))),
	}, nil
}
