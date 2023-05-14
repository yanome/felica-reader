package felica

import (
	"fmt"
	"strings"
)

const (
	WAON_SYSTEM               = 0xFE00
	WAON_NUMBER_SERVICE       = 0x684F
	WAON_NUMBER_BLOCKS        = 2
	WAON_BALANCE_SERVICE      = 0x6817
	WAON_BALANCE_BLOCKS       = 1
	WAON_TRANSACTIONS_SERVICE = 0x680B
	WAON_TRANSACTIONS_BLOCKS  = 9
)

var waonTransactionTypes = map[uint8]string{
	0x04: "Purchase",
	0x0c: "Recharge",
	0x10: "Recharge",
}

func waonNumber(b []block, s *strings.Builder) {
	s.WriteString(fmt.Sprintf(
		"\nWaon number: %04d-%04d-%04d-%04d",
		bcd16(b[0x00][0x00], b[0x00][0x01]),
		bcd16(b[0x00][0x02], b[0x00][0x03]),
		bcd16(b[0x00][0x04], b[0x00][0x05]),
		bcd16(b[0x00][0x06], b[0x00][0x07]),
	))
}

func waonBalance(b []block, s *strings.Builder) {
	s.WriteString(fmt.Sprintf("\nWaon balance: %d", number32(0, b[0x00][0x02], b[0x00][0x01], b[0x00][0x00])))
}

func waonTransactions(b []block, s *strings.Builder) {
	s.WriteString("\nWaon transactions")
	s.WriteString("\nNumber Type     Date     Time  Balance Amount")
	for idx := 5; idx >= 1; idx -= 2 {
		year := 5 + b[idx][0x02]>>3
		month := (b[idx][0x02]&0x07)<<1 + (b[idx][0x03]&0x80)>>7
		day := (b[idx][0x03] & 0x7f) >> 2
		hour := (b[idx][0x03]&0x03)<<3 + (b[idx][0x04]&0xe0)>>5
		minute := (b[idx][0x04]&0x1f)<<1 + (b[idx][0x05]&0x80)>>7
		balance := uint32(b[idx][0x05]&0x7f)<<11 + uint32(b[idx][0x06])<<3 + uint32(b[idx][0x07]&0xe0)>>5
		amount := uint32(b[idx][0x07]&0x1f)<<13 + uint32(b[idx][0x08])<<5 + uint32(b[idx][0x09]&0xf8)>>3
		if amount == 0 {
			amount = uint32(b[idx][0x09]&0x07)<<14 + uint32(b[idx][0x0a])<<6 + uint32(b[idx][0x0b]&0xfc)>>2
		}
		s.WriteString(fmt.Sprintf(
			"\n % 5d %- 8s %02d-%02d-%02d %02d:%02d   % 5d  % 5d",
			number16(b[idx-1][0x0d], b[idx-1][0x0e]),
			waonTransactionTypes[b[idx][0x01]],
			year, month, day,
			hour, minute,
			balance,
			amount,
		))

	}
}
