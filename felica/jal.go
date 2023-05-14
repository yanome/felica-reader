package felica

import (
	"fmt"
	"strings"
)

const (
	JAL_SYSTEM         = 0xFE00
	JAL_MEMBER_SERVICE = 0x2F4B
	JAL_MEMBER_BLOCKS  = 4
)

func jalMember(b []block, s *strings.Builder) {
	n0 := bcd8(b[0x00][0x0f])
	n1 := bcd16(b[0x00][0x0e], b[0x00][0x0d])
	n2 := bcd16(b[0x00][0x0c], b[0x00][0x0b])
	s.WriteString(fmt.Sprintf("\nJAL number: %02d-%03d-%04d", n0, n1/10, 1000*(n1%10)+n2/10))
	s.WriteString("\nJAL name: ")
	for idx := BLOCK_SIZE - 1; idx >= 0; idx-- {
		s.WriteByte(b[0x01][idx])
	}
}
