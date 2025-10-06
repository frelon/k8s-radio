package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"

	rtlsdr "github.com/frelon/k8s-radio/device-plugin/rtl-sdr"
)

type RadioDeviceLister struct {
	ResUpdateChan chan dpm.PluginNameList
	Heartbeat     chan bool
}

func (l *RadioDeviceLister) GetResourceNamespace() string {
	return "frelon.se"
}

func (l *RadioDeviceLister) Discover(pluginListCh chan dpm.PluginNameList) {
	for {
		select {
		case newResourcesList := <-l.ResUpdateChan:
			pluginListCh <- newResourcesList
		case <-pluginListCh:
			return
		}
	}
}

func (l *RadioDeviceLister) NewPlugin(resourceLastName string) dpm.PluginInterface {
	if resourceLastName == rtlsdr.ResourceName {
		return &rtlsdr.Plugin{
			Heartbeat: l.Heartbeat,
			RtlSdrs:   make(map[string]*rtlsdr.RtlSdrDev),
		}
	}

	glog.Errorf("Unknown resource name: '%s'", resourceLastName)
	return nil
}

func main() {
	flag.Parse()

	glog.Info("Starting radio device plugin")

	l := RadioDeviceLister{
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
		l.ResUpdateChan <- []string{rtlsdr.ResourceName}
	}()

	manager.Run()
}
