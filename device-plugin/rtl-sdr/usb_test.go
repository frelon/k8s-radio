package rtlsdr

import (
	"testing"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUsb(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "USB Suite")
}

var _ = Describe("USB", func() {
	It("reads devices from /sys/bus/usb", func(ctx SpecContext) {
		fsys := fstest.MapFS{
			"sys/bus/usb/devices/1-2/idVendor": {
				Data: []byte("0bda\n"),
			},
			"sys/bus/usb/devices/1-2/idProduct": {
				Data: []byte("2838\n"),
			},
			"sys/bus/usb/devices/1-2/busnum": {
				Data: []byte("2\n"),
			},
			"sys/bus/usb/devices/1-2/devnum": {
				Data: []byte("8\n"),
			},
			"sys/bus/usb/devices/1-2/serial": {
				Data: []byte("00000001\n"),
			},
		}

		b, err := ListUsbDevices(fsys)
		Expect(err).ToNot(HaveOccurred())
		Expect(b).ToNot(BeEmpty())
		Expect(len(b)).To(Equal(1))
		Expect(b[0].VendorID).To(Equal("0bda"))
		Expect(b[0].ProductID).To(Equal("2838"))
		Expect(b[0].Bus).To(Equal(2))
		Expect(b[0].Dev).To(Equal(8))
		Expect(b[0].Serial).To(Equal("00000001"))
	})
})
