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

func (p *Plugin) UpdateDevices() (error) {
	rtls, err := ListDevices()
	if err != nil {
		glog.Infof("Error listing devices: %s", err.Error())
		return err
	}

	glog.Infof("Found %d devices", len(rtls))

    for _, rtl := range p.RtlSdrs {
        rtl.Connected = false
    }

	for i := range rtls {
		p.RtlSdrs[rtls[i].SerialNumber] = rtls[i]
	}

	return nil
}

func (p *Plugin) GetDevices() []*pluginapi.Device {
	devs := make([]*pluginapi.Device, len(p.RtlSdrs))
    i := 0
    for _, rtl := range p.RtlSdrs {
		devs[i] = &pluginapi.Device{
			ID:     rtl.SerialNumber,
            Health: pluginapi.Unhealthy,
		}

        if rtl.Connected {
            devs[i].Health = pluginapi.Healthy
        }

        i++
    }

    return devs
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {

	err := p.UpdateDevices()
	if err != nil {
		glog.Errorf("Error listing devices: %s", err.Error())
	}

    devs := p.GetDevices()

	err = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		glog.Errorf("Error sending initial response: %s", err.Error())
	}

	glog.Info("Waiting for updates...")

	for {
		select {
		case <-p.Heartbeat:
			err = p.UpdateDevices()
			if err != nil {
				glog.Errorf("Error reading devices: %s", err.Error())
				continue
			}

            devs := p.GetDevices()
			glog.Infof("Devices updated (len %d)", len(devs))

			err = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
			if err != nil {
				glog.Errorf("Error sending response: %s", err.Error())
				continue
			}
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
	    RtlSdrs: make(map[string]*RtlSdrDev),
	}
}

func main() {
	flag.Parse()

	glog.Info("Starting rtl-sdr device plugin")

	l := Lister{
		ResUpdateChan: make(chan dpm.PluginNameList),
		Heartbeat:     make(chan bool),
	}

	pulse := 2
	go func() {
		glog.Infof("Heart beating every %d seconds", pulse)

		for {
			time.Sleep(time.Second * time.Duration(pulse))
			l.Heartbeat <- true
		}
	}()

	manager := dpm.NewManager(&l)

	go func() {
		l.ResUpdateChan <- []string{"rtl-sdr"}
	}()

	manager.Run()
}

type RtlSdrDev struct {
	*gousb.Device

	SerialNumber string
    Connected bool
}

func (r RtlSdrDev) DevicePath() string {
	return fmt.Sprintf("/dev/bus/usb/%03d/%03d", r.Device.Desc.Bus, r.Device.Desc.Address)
}

func NewRtlSdrDev(dev *gousb.Device) *RtlSdrDev {
	serial, _ := dev.SerialNumber()

	return &RtlSdrDev{
		Device:       dev,
		SerialNumber: serial,
        Connected:    true,
	}
}

func ListDevices() ([]*RtlSdrDev, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == 0x0bda
	})
	for i := range devs {
		defer devs[i].Close()
	}

	if err != nil {
		glog.Infof("Err open device (%d): %s", len(devs), err.Error())
		return nil, err
	}

	devices := make([]*RtlSdrDev, len(devs))
	for i := range devs {
		devices[i] = NewRtlSdrDev(devs[i])
	}

	return devices, nil
}
