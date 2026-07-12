package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

func compileLifeCharacteristicExchangeKind(
	kind parser.LifeCharacteristicExchangeKind,
) LifeCharacteristicExchangeKind {
	switch kind {
	case parser.LifeCharacteristicExchangeSourcePower:
		return LifeCharacteristicExchangeSourcePower
	case parser.LifeCharacteristicExchangeSourceToughness:
		return LifeCharacteristicExchangeSourceToughness
	default:
		return LifeCharacteristicExchangeNone
	}
}
