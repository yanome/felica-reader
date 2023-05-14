package felica

import (
	"strings"
)

type decoderKey struct {
	systemId  uint16
	serviceId uint16
	blocks    int
}

type decoder func(b []block, s *strings.Builder)

var decoders = map[decoderKey]decoder{
	{
		systemId:  SUICA_SYSTEM,
		serviceId: SUICA_BALANCE_SERVICE,
		blocks:    SUICA_BALANCE_BLOCKS,
	}: suicaBalance,
	{
		systemId:  SUICA_SYSTEM,
		serviceId: SUICA_TRANSACTIONS_SERVICE,
		blocks:    SUICA_TRANSACTIONS_BLOCKS,
	}: suicaTransactions,
	{
		systemId:  SUICA_SYSTEM,
		serviceId: SUICA_GATE_LOG_SERVICE,
		blocks:    SUICA_GATE_LOG_BLOCKS,
	}: suicaGateLog,
	{
		systemId:  EDY_SYSTEM,
		serviceId: EDY_NUMBER_SERVICE,
		blocks:    EDY_NUMER_BLOCKS,
	}: edyNumber,
	{
		systemId:  EDY_SYSTEM,
		serviceId: EDY_BALANCE_SERVICE,
		blocks:    EDY_BALANCE_BLOCKS,
	}: edyBalance,
	{
		systemId:  EDY_SYSTEM,
		serviceId: EDY_TRANSACTIONS_SERVICE,
		blocks:    EDY_TRANSACTIONS_BLOCKS,
	}: edyTransactions,
	{
		systemId:  WAON_SYSTEM,
		serviceId: WAON_NUMBER_SERVICE,
		blocks:    WAON_NUMBER_BLOCKS,
	}: waonNumber,
	{
		systemId:  WAON_SYSTEM,
		serviceId: WAON_BALANCE_SERVICE,
		blocks:    WAON_BALANCE_BLOCKS,
	}: waonBalance,
	{
		systemId:  WAON_SYSTEM,
		serviceId: WAON_TRANSACTIONS_SERVICE,
		blocks:    WAON_TRANSACTIONS_BLOCKS,
	}: waonTransactions,
	{
		systemId:  JAL_SYSTEM,
		serviceId: JAL_MEMBER_SERVICE,
		blocks:    JAL_MEMBER_BLOCKS,
	}: jalMember,
}

func bcd8(b0 uint8) uint8 {
	return 10*(b0>>4) + b0&0x0f
}

func bcd16(b1, b0 uint8) uint16 {
	return uint16(bcd8(b1))*100 + uint16(bcd8(b0))
}

func number16(b1, b0 uint8) uint16 {
	return uint16(b1)<<8 + uint16(b0)
}

func number32(b3, b2, b1, b0 uint8) uint32 {
	return uint32(number16(b3, b2))<<16 + uint32(number16(b1, b0))
}
