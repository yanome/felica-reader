package felica

import (
	"fmt"
	"strings"
	"time"
)

const (
	EDY_SYSTEM               = 0xFE00
	EDY_NUMBER_SERVICE       = 0x110B
	EDY_NUMER_BLOCKS         = 2
	EDY_BALANCE_SERVICE      = 0x1317
	EDY_BALANCE_BLOCKS       = 1
	EDY_TRANSACTIONS_SERVICE = 0x170F
	EDY_TRANSACTIONS_BLOCKS  = 6
)

var edyTransactionTypes = map[uint8]string{
	0x02: "Recharge",
	0x04: "Recharge",
	0x20: "Purchase",
}

func edyNumber(b []block, s *strings.Builder) {
	s.WriteString(fmt.Sprintf(
		"\nEdy number: %04d-%04d-%04d-%04d",
		bcd16(b[0x00][0x02], b[0x00][0x03]),
		bcd16(b[0x00][0x04], b[0x00][0x05]),
		bcd16(b[0x00][0x06], b[0x00][0x07]),
		bcd16(b[0x00][0x08], b[0x00][0x09]),
	))
}

func edyBalance(b []block, s *strings.Builder) {
	s.WriteString(fmt.Sprintf("\nEdy balance: %d", number32(b[0x00][0x03], b[0x00][0x02], b[0x00][0x01], b[0x00][0x00])))
}

func edyTransactions(b []block, s *strings.Builder) {
	startTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	s.WriteString("\nEdy transactions")
	s.WriteString("\nType     Number Date       Time     Amount Balance")
	for idx := range b {
		days := int(b[idx][0x04])<<7 + int(b[idx][0x05])>>1
		seconds := time.Duration(b[idx][0x05]&0x01)<<16 + time.Duration(b[idx][0x06])<<8 + time.Duration(b[idx][0x07])
		s.WriteString(fmt.Sprintf(
			"\n%- 8s  % 5d %s %s  % 5d   % 5d",
			edyTransactionTypes[b[idx][0x00]],
			number16(b[idx][0x02], b[idx][0x03]),
			startTime.AddDate(0, 0, days).Format(time.DateOnly),
			startTime.Add(time.Second*seconds).Format(time.TimeOnly),
			number32(b[idx][0x08], b[idx][0x09], b[idx][0x0a], b[idx][0x0b]),
			number32(b[idx][0x0c], b[idx][0x0d], b[idx][0x0e], b[idx][0x0f]),
		))
	}
}
