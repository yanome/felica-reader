package felica

import (
	"fmt"
	"strings"
)

const (
	SUICA_SYSTEM               = 0x0003
	SUICA_BALANCE_SERVICE      = 0x008B
	SUICA_BALANCE_BLOCKS       = 1
	SUICA_TRANSACTIONS_SERVICE = 0x090F
	SUICA_TRANSACTIONS_BLOCKS  = 20
	SUICA_GATE_LOG_SERVICE     = 0x108F
	SUICA_GATE_LOG_BLOCKS      = 3
)

var suicaTransactionTypes = map[uint8]string{
	0x01: "Train fare",
	0x02: "Recharge",
	0x0f: "Bus fare",
	0x14: "Recharge",
	0x46: "Purchase",
}

var suicaTransactionTypeWithTime = map[uint8]bool{
	0x46: true,
}

var suicaTransactionTypeWithStations = map[uint8]bool{
	0x01: true,
}

var suicaGateLogTypes = map[uint8]string{
	0xa0: "Entry",
	0x20: "Exit",
}

func suicaBalance(b []block, s *strings.Builder) {
	s.WriteString(fmt.Sprintf("\nSuica balance: %d", number16(b[0x00][0x0c], b[0x00][0x0b])))
}

func suicaTransactions(b []block, s *strings.Builder) {
	s.WriteString("\nSuica transactions")
	s.WriteString("\nType       Date     Time  Stations        Balance Number")
	for idx := range b {
		transactionType := b[idx][0x01]
		transactionTime := ""
		if suicaTransactionTypeWithTime[transactionType] {
			hour := b[idx][0x06] >> 3
			minute := (b[idx][0x06]&0x07)<<3 + b[idx][0x07]>>5
			transactionTime = fmt.Sprintf("%02d:%02d", hour, minute)
		}
		stations := ""
		if suicaTransactionTypeWithStations[transactionType] {
			stations = fmt.Sprintf(
				"%03d-%03d %03d-%03d",
				b[idx][0x06],
				b[idx][0x07],
				b[idx][0x08],
				b[idx][0x09],
			)
		}
		s.WriteString(fmt.Sprintf(
			"\n%- 10s %s %- 5s %- 15s   % 5d  % 5d",
			suicaTransactionTypes[transactionType],
			suicaDate(b[idx][0x04], b[idx][0x05]),
			transactionTime,
			stations,
			number16(b[idx][0x0b], b[idx][0x0a]),
			number16(b[idx][0x0d], b[idx][0x0e]),
		))
	}
}

func suicaGateLog(b []block, s *strings.Builder) {
	s.WriteString("\nSuica gate log")
	s.WriteString("\nType  Station Date     Time  Amount")
	for idx := range b {
		s.WriteString(fmt.Sprintf(
			"\n%- 5s %03d-%03d %s %02d:%02d  % 5d",
			suicaGateLogTypes[b[idx][0x00]],
			b[idx][0x02], b[idx][0x03],
			suicaDate(b[idx][0x06], b[idx][0x07]),
			bcd8(b[idx][0x08]), bcd8(b[idx][0x09]),
			number16(b[idx][0x0b], b[idx][0x0a]),
		))
	}
}

func suicaDate(b1, b0 uint8) string {
	year := b1 >> 1
	month := (b1&0x01)<<3 + b0>>5
	day := b0 & 0x1f
	return fmt.Sprintf("%02d-%02d-%02d", year, month, day)
}
