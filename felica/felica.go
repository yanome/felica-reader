package felica

import (
	"fmt"
	"strings"

	"github.com/yanome/felica-reader/usb"
)

const (
	ID_LENGTH        = 8
	BLOCK_SIZE       = 16
	SERVICE_TYPE     = 0x003F
	SERVICE_READABLE = 0x01
)

type block []uint8

func (b block) String() string {
	bytes := []string{}
	for idx := range b {
		bytes = append(bytes, fmt.Sprintf("%02x", b[idx]))
	}
	return strings.Join(bytes, " ")
}

var serviceTypes = map[uint16]string{
	0b001000: "Random RW Auth",
	0b001001: "Random RW",
	0b001010: "Random RO Auth",
	0b001011: "Random RO",
	0b001100: "Cyclic RW Auth",
	0b001101: "Cyclic RW",
	0b001110: "Cyclic RO Auth",
	0b001111: "Cyclic RO",
	0b010000: "Purse Direct Auth",
	0b010001: "Purse Direct",
	0b010010: "Purse Cashback Auth",
	0b010011: "Purse Cashback",
	0b010100: "Purse Decrement Auth",
	0b010101: "Purse Decrement",
	0b010110: "Purse RO Auth",
	0b010111: "Purse RO",
}

type service struct {
	Id     uint16 `json:"ServiceId"`
	Blocks []block
}

func (s service) isReadable() bool {
	return s.Id&SERVICE_READABLE == SERVICE_READABLE
}

func (s service) String() string {
	return fmt.Sprintf("%04x %s", s.Id, serviceTypes[s.Id&SERVICE_TYPE])
}

type system struct {
	Id       uint16 `json:"SystemId"`
	Services []service
}

func (s system) DecodedService(i int) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("System code: %s", s))
	service := s.Services[i]
	serviceNumber := service.Id
	for mask := SERVICE_TYPE; mask > 0; mask >>= 1 {
		serviceNumber >>= 1
	}
	serviceAttribute := service.Id & SERVICE_TYPE
	b.WriteString(fmt.Sprintf("\nService code: %04x, number: %03x, attribute: %02x %s", service.Id, serviceNumber, serviceAttribute, serviceTypes[serviceAttribute]))
	if decoder, found := decoders[decoderKey{
		systemId:  s.Id,
		serviceId: service.Id,
		blocks:    len(service.Blocks),
	}]; found {
		decoder(service.Blocks, &b)
	}
	return b.String()
}

func (s system) RawService(i int) string {
	service := s.Services[i]
	b := strings.Builder{}
	if len(service.Blocks) > 0 {
		b.WriteString("    ")
		for idx := 0; idx < BLOCK_SIZE; idx++ {
			b.WriteString(fmt.Sprintf(" %02x", idx))
		}
		for idx := range service.Blocks {
			b.WriteString(fmt.Sprintf("\n%04x %s", idx, service.Blocks[idx]))
		}
	}
	return b.String()
}

func (s system) String() string {
	return fmt.Sprintf("%04x", s.Id)
}

type Card struct {
	Id      []uint8 `json:"CardId"`
	Systems []system
}

func (c *Card) String() string {
	bytes := []string{}
	for idx := range c.Id {
		bytes = append(bytes, fmt.Sprintf("%02x", c.Id[idx]))
	}
	return strings.Join(bytes, ":")
}

func Read(e *usb.Endpoints) (*Card, error) {
	id, err := e.Poll()
	if err != nil {
		return nil, fmt.Errorf("polling error: %s", err)
	}
	if id == nil {
		return nil, nil
	}
	systems, err := systems(e)
	if err != nil {
		return nil, err
	}
	return &Card{
		Id:      id[:ID_LENGTH],
		Systems: systems,
	}, nil
}

func systems(e *usb.Endpoints) ([]system, error) {
	ids, err := e.Systems()
	if err != nil {
		return nil, fmt.Errorf("error reading card systems: %s", err)
	}
	systems := []system{}
	for idx, id := range ids {
		services, err := services(e, idx)
		if err != nil {
			return nil, err
		}
		systems = append(systems, system{
			Id:       id,
			Services: services,
		})
	}
	return systems, nil
}

func services(e *usb.Endpoints, system int) ([]service, error) {
	ids, err := e.Services(system)
	if err != nil {
		return nil, fmt.Errorf("error reading card services for system %d: %s", system, err)
	}
	services := []service{}
	for idx, id := range ids {
		service := service{
			Id: id,
		}
		if service.isReadable() {
			b, err := blocks(e, system, idx)
			if err != nil {
				return nil, err
			}
			service.Blocks = b
		}
		services = append(services, service)
	}
	return services, nil
}

func blocks(e *usb.Endpoints, system int, service int) ([]block, error) {
	n, data, err := e.Blocks(service)
	if err != nil {
		return nil, fmt.Errorf("error reading card blocks for system %d, service %d: %s", system, service, err)
	}
	blocks := []block{}
	for idx := 0; idx < n; idx++ {
		blocks = append(blocks, data[idx*BLOCK_SIZE:idx*BLOCK_SIZE+BLOCK_SIZE])
	}
	return blocks, nil
}
