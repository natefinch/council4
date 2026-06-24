package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// LevelBand is a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band (CR 711.4). Low
// is the band's first level; High is its last level, or 0 for the open-ended
// final band ("LEVEL lo+"). Power and Toughness hold the band's printed base
// power/toughness when HasPowerToughness is true; a band with no printed P/T
// (a non-creature leveler band) leaves them unset.
type LevelBand struct {
	Low               int
	High              int
	Power             int
	Toughness         int
	HasPowerToughness bool
	Span              shared.Span
}

// recognizeLevelUpAbility reads a leveler card's "Level up {cost}" activated
// ability line (CR 711.2), returning the parsed mana cost. The line is the two
// words "Level up", one or more mana symbols, then an optional fully
// parenthesized reminder. Any other shape returns false so the line lowers
// through the ordinary paths and fails closed if unsupported.
func recognizeLevelUpAbility(tokens []shared.Token) (cost.Mana, bool) {
	if len(tokens) < 3 {
		return nil, false
	}
	if !equalWord(tokens[0], "Level") || !equalWord(tokens[1], "up") {
		return nil, false
	}
	manaCost, end, ok := parseKeywordManaCost(tokens, 2)
	if !ok {
		return nil, false
	}
	// Only a fully parenthesized reminder may follow the cost.
	if end < len(tokens) && tokens[end].Kind != shared.LeftParen {
		return nil, false
	}
	return manaCost, true
}

// recognizeLevelBandHeader reads a leveler card's "LEVEL lo-hi" or "LEVEL lo+"
// band header (CR 711.4), returning the band's level range. A closed band
// ("LEVEL lo-hi") yields Low=lo, High=hi; an open band ("LEVEL lo+") yields
// High=0. Any other shape returns false.
func recognizeLevelBandHeader(tokens []shared.Token) (low, high int, ok bool) {
	if len(tokens) < 2 || len(tokens) > 4 {
		return 0, 0, false
	}
	if tokens[0].Kind != shared.Word || tokens[0].Text != "LEVEL" {
		return 0, 0, false
	}
	if tokens[1].Kind != shared.Integer {
		return 0, 0, false
	}
	low, err := strconv.Atoi(tokens[1].Text)
	if err != nil || low < 1 {
		return 0, 0, false
	}
	switch len(tokens) {
	case 3:
		// "LEVEL lo+"
		if tokens[2].Kind != shared.Plus {
			return 0, 0, false
		}
		return low, 0, true
	case 4:
		// "LEVEL lo-hi"
		if tokens[2].Kind != shared.Minus || tokens[3].Kind != shared.Integer {
			return 0, 0, false
		}
		high, err := strconv.Atoi(tokens[3].Text)
		if err != nil || high < low {
			return 0, 0, false
		}
		return low, high, true
	default:
		return 0, 0, false
	}
}

// recognizeLevelBandPowerToughness reads a leveler band's printed base
// power/toughness line ("N/M"), returning the two values. Any other shape
// returns false.
func recognizeLevelBandPowerToughness(tokens []shared.Token) (power, toughness int, ok bool) {
	if len(tokens) != 3 {
		return 0, 0, false
	}
	if tokens[0].Kind != shared.Integer || tokens[1].Kind != shared.Slash || tokens[2].Kind != shared.Integer {
		return 0, 0, false
	}
	power, err := strconv.Atoi(tokens[0].Text)
	if err != nil {
		return 0, 0, false
	}
	toughness, err = strconv.Atoi(tokens[2].Text)
	if err != nil {
		return 0, 0, false
	}
	return power, toughness, true
}

// parseLevelBand builds an AbilityLevelBand from the band header at lines[i] and
// consumes the immediately following printed P/T line, if any. It returns the
// band ability, the index of the next unconsumed line, and whether lines[i] is a
// band header. The band header and its P/T line lose their resolving body; the
// abilities printed below belong to the band until the next header.
func parseLevelBand(source string, lines [][]shared.Token, i int) (Ability, int, bool) {
	low, high, ok := recognizeLevelBandHeader(lines[i])
	if !ok {
		return Ability{}, i, false
	}
	headerSpan := shared.SpanOf(lines[i])
	band := &LevelBand{Low: low, High: high, Span: headerSpan}
	endSpan := headerSpan
	next := i + 1
	if next < len(lines) {
		if power, toughness, ptOK := recognizeLevelBandPowerToughness(lines[next]); ptOK {
			band.Power = power
			band.Toughness = toughness
			band.HasPowerToughness = true
			endSpan = shared.SpanOf(lines[next])
			next++
		}
	}
	span := shared.Span{Start: headerSpan.Start, End: endSpan.End}
	ability := Ability{
		Kind:      AbilityLevelBand,
		Span:      span,
		Text:      shared.SliceSpan(source, span),
		Tokens:    cloneTokens(lines[i]),
		LevelBand: band,
	}
	return ability, next, true
}
