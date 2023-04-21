package usb

import (
	"fmt"
)

const (
	CMD_POLL     = 0x02
	CMD_SYSTEMS  = 0x04
	CMD_SERVICES = 0x06
	CMD_BLOCKS   = 0x08

	STATUS_OK       = 0x00
	STATUS_CONTINUE = 0x01
)

func (e *Endpoints) command(commandCode uint8, data ...uint8) ([]uint8, error) {
	command := append([]uint8{commandCode}, data...)
	n, err := e.out.Write(command)
	if err != nil {
		return nil, fmt.Errorf("error sending command: %s", err)
	}
	if n != len(command) {
		return nil, fmt.Errorf("error sending data, sent %d bytes, expected %d", n, len(command))
	}
	return e.response(commandCode)
}

func (e *Endpoints) response(commandCode uint8) ([]uint8, error) {
	response := make([]uint8, e.in.Desc.MaxPacketSize)
	n, err := e.in.Read(response)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("zero bytes received")
	}
	if response[0] != commandCode+1 {
		return nil, fmt.Errorf("unexpected command code %02x, expected %02x", response[0], commandCode+1)
	}
	for {
		switch response[1] {
		case STATUS_OK:
			return response[2:], nil
		case STATUS_CONTINUE:
			next, err := e.response(commandCode)
			if err != nil {
				return nil, err
			}
			return append(response[2:], next...), nil
		default:
			return nil, fmt.Errorf("command error %02x", response[1])
		}
	}
}

func (e *Endpoints) Poll() ([]uint8, error) {
	response, err := e.command(CMD_POLL)
	if err != nil {
		return nil, err
	}
	if response[0] == 0 {
		return nil, nil
	}
	return response[1:], nil
}

func (e *Endpoints) Systems() ([]uint16, error) {
	response, err := e.command(CMD_SYSTEMS)
	if err != nil {
		return nil, err
	}
	systems := []uint16{}
	for idx := 0; idx < int(response[0]); idx++ {
		id := uint16(response[2*idx+1])<<8 + uint16(response[2*idx+2])
		systems = append(systems, id)
	}
	return systems, nil
}

func (e *Endpoints) Services(system int) ([]uint16, error) {
	response, err := e.command(CMD_SERVICES, uint8(system))
	if err != nil {
		return nil, err
	}
	services := []uint16{}
	for idx := 0; idx < int(response[0]); idx++ {
		id := uint16(response[2*idx+1])<<8 + uint16(response[2*idx+2])
		services = append(services, id)
	}
	return services, nil
}

func (e *Endpoints) Blocks(service int) (int, []uint8, error) {
	response, err := e.command(CMD_BLOCKS, uint8(service))
	if err != nil {
		return 0, nil, err
	}
	return int(response[0]), response[1:], nil
}
