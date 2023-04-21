package usb

import (
	"fmt"

	"github.com/google/gousb"
)

const VID = 0x16C0
const PID = 0x0486

type Endpoints struct {
	in  *gousb.InEndpoint
	out *gousb.OutEndpoint
}

type releaseFunc func()

func Init() (*Endpoints, releaseFunc, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()
	dev, err := ctx.OpenDeviceWithVIDPID(VID, PID)
	if err != nil {
		return nil, nil, fmt.Errorf("error listing devices: %s", err)
	}
	if dev == nil {
		return nil, nil, fmt.Errorf("device %04x:%04x not found", VID, PID)
	}
	defer dev.Close()
	if err = dev.SetAutoDetach(true); err != nil {
		return nil, nil, fmt.Errorf("error enabling autodetach: %s", err)
	}
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		return nil, nil, fmt.Errorf("error claiming interface: %s", err)
	}
	endpoints, err := getEndpoints(intf)
	if err != nil {
		done()
		return nil, nil, err
	}
	return endpoints, done, nil
}

func getEndpoints(i *gousb.Interface) (*Endpoints, error) {
	var err error
	endpoints := &Endpoints{}
	for _, epDesc := range i.Setting.Endpoints {
		switch epDesc.Direction {
		case gousb.EndpointDirectionIn:
			if endpoints.in, err = i.InEndpoint(epDesc.Number); err != nil {
				return nil, fmt.Errorf("error preparing IN endpoint (%s): %s", epDesc.Address, err)
			}
		case gousb.EndpointDirectionOut:
			if endpoints.out, err = i.OutEndpoint(epDesc.Number); err != nil {
				return nil, fmt.Errorf("error preparing OUT endpoint (%s): %s", epDesc.Address, err)
			}
		}
	}
	if endpoints.in == nil || endpoints.out == nil {
		return nil, fmt.Errorf("error preparing endpoints")
	}
	return endpoints, nil
}
