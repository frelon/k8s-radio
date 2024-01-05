package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/google/gousb"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Plugin struct {
	RtlSdrs   map[string]*RtlSdrDev
	Heartbeat chan bool
}

type Lister struct {
	ResUpdateChan chan dpm.PluginNameList
	Heartbeat     chan bool
}

func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	rtls, err := ListDevices()
	if err != nil {
		glog.Infof("Error listing devices: %s", err.Error())
		return err
	}

	glog.Infof("Found %d devices", len(rtls))

	p.RtlSdrs = make(map[string]*RtlSdrDev)

	devs := make([]*pluginapi.Device, len(rtls))

	for i := range rtls {
		devs[i] = &pluginapi.Device{
			ID:     rtls[i].SerialNumber,
			Health: pluginapi.Healthy,
		}

		p.RtlSdrs[rtls[i].SerialNumber] = rtls[i]
	}

	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

	glog.Info("Sent ListAndWatchResponse")

	for {
		select {
		case <-p.Heartbeat:
			glog.Info("ListAndWatch heartbeat")
			for i := 0; i < len(rtls); i++ {
				devs[i].Health = pluginapi.Healthy
			}

			s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		}
	}
}

func (p *Plugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (p *Plugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse
	var car pluginapi.ContainerAllocateResponse
	var dev *pluginapi.DeviceSpec

	for _, req := range r.ContainerRequests {
		car = pluginapi.ContainerAllocateResponse{}

		dev = new(pluginapi.DeviceSpec)
		dev.Permissions = "rw"
		car.Devices = append(car.Devices, dev)

		for _, id := range req.DevicesIDs {
			glog.Infof("Allocating device ID: %s", id)

			dev.HostPath = p.RtlSdrs[id].DevicePath()
			dev.ContainerPath = p.RtlSdrs[id].DevicePath()
		}

		response.ContainerResponses = append(response.ContainerResponses, &car)
	}

	return &response, nil
}

func (l *Lister) GetResourceNamespace() string {
	return "frelon.se"
}

func (l *Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	for {
		select {
		case newResourcesList := <-l.ResUpdateChan:
			pluginListCh <- newResourcesList
		case <-pluginListCh:
			return
		}
	}
}

func (l *Lister) NewPlugin(resourceLastName string) dpm.PluginInterface {
	return &Plugin{
		Heartbeat: l.Heartbeat,
	}
}

func main() {
	flag.Parse()

	glog.Info("Starting rtl-sdr device plugin")

	l := Lister{
		Heartbeat:     make(chan bool),
		ResUpdateChan: make(chan dpm.PluginNameList),
	}

	manager := dpm.NewManager(&l)

	go func() {
		pulse := 2
		glog.Infof("Heartbeating every %d seconds", pulse)
		for {
			time.Sleep(time.Second * time.Duration(pulse))
			l.Heartbeat <- true
		}
	}()

	go func() {
		l.ResUpdateChan <- []string{"rtl-sdr"}
	}()

	manager.Run()
}

type RtlSdrDev struct {
	*gousb.Device

	SerialNumber string
}

func (r RtlSdrDev) DevicePath() string {
	return fmt.Sprintf("/dev/bus/usb/%03d/%03d", r.Device.Desc.Bus, r.Device.Desc.Address)
}

func NewRtlSdrDev(dev *gousb.Device) *RtlSdrDev {
	serial, _ := dev.SerialNumber()

	return &RtlSdrDev{
		Device:       dev,
		SerialNumber: serial,
	}
}

func ListDevices() ([]*RtlSdrDev, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == 0x0bda
	})
	if err != nil {
		return nil, err
	}

	devices := make([]*RtlSdrDev, len(devs))

	for i := range devs {
		defer devs[i].Close()

		devices[i] = NewRtlSdrDev(devs[i])
	}

	return devices, err
}
