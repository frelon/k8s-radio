package rtlsdr

import (
	"context"
	"io/fs"
	"log/slog"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	ResourceName = "rtl-sdr"
)

type Plugin struct {
	devices   map[string]*UsbDevice
	heartbeat chan bool
	fsys      fs.FS
}

func NewPlugin(heartbeat chan bool, fsys fs.FS) *Plugin {
	return &Plugin{
		heartbeat: heartbeat,
		fsys:      fsys,
		devices:   make(map[string]*UsbDevice),
	}
}

func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *Plugin) UpdateDevices() ([]*pluginapi.Device, error) {
	connectedDevs, err := ListUsbDevices(p.fsys)
	if err != nil {
		slog.Info("Error listing devices", slog.Any("error", err))
		return nil, err
	}

	slog.Info("Found devices", "len", len(connectedDevs))

	connectedDevsBySerial := map[string]*UsbDevice{}
	for i := range connectedDevs {
		connectedDevsBySerial[connectedDevs[i].Serial] = connectedDevs[i]
		p.devices[connectedDevs[i].Serial] = connectedDevs[i]
	}

	pdevs := make([]*pluginapi.Device, len(p.devices))
	i := 0
	for _, rtl := range p.devices {
		pdevs[i] = &pluginapi.Device{
			ID:     rtl.Serial,
			Health: pluginapi.Unhealthy,
		}

		if _, ok := connectedDevsBySerial[rtl.Serial]; ok {
			pdevs[i].Health = pluginapi.Healthy
		}

		i++
	}

	return pdevs, nil
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs, err := p.UpdateDevices()
	if err != nil {
		slog.Error("Error listing devices", slog.Any("error", err))
	}

	err = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		slog.Error("Error sending initial response", slog.Any("error", err))
	}

	slog.Info("Waiting for updates...")

	for range p.heartbeat {
		devs, err = p.UpdateDevices()
		if err != nil {
			slog.Error("Error reading devices", slog.Any("error", err))
			continue
		}

		slog.Info("Devices updated", slog.Int("len", len(devs)))

		err = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		if err != nil {
			slog.Error("Error sending response", slog.Any("error", err))
			continue
		}
	}

	return nil
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
			slog.Info("Allocating device", slog.String("ID", id))

			dev.HostPath = p.devices[id].DevicePath()
			dev.ContainerPath = p.devices[id].DevicePath()
		}

		response.ContainerResponses = append(response.ContainerResponses, &car)
	}

	return &response, nil
}
