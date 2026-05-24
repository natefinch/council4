package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

type permanentEffectiveValues struct {
	power       int
	powerOK     bool
	toughness   int
	toughnessOK bool
	keywords    map[game.Keyword]bool
}

func effectivePower(g *game.Game, permanent *game.Permanent) int {
	values := effectivePermanentValues(g, permanent)
	if !values.powerOK {
		return 0
	}
	return max(0, values.power)
}

func effectiveToughness(g *game.Game, permanent *game.Permanent) (int, bool) {
	values := effectivePermanentValues(g, permanent)
	return values.toughness, values.toughnessOK
}

func hasKeyword(g *game.Game, permanent *game.Permanent, keyword game.Keyword) bool {
	return effectivePermanentValues(g, permanent).keywords[keyword]
}

func effectivePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := basePermanentValues(g, permanent)
	applyCounterAndTemporaryValues(permanent, &values)
	applyStaticPTValues(g, permanent, &values)
	return values
}

func basePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := permanentEffectiveValues{keywords: make(map[game.Keyword]bool)}
	card := permanentCardDef(g, permanent)
	if card == nil {
		return values
	}
	if card.Power != nil && !card.Power.IsStar {
		values.power = card.Power.Value
		values.powerOK = true
	}
	if card.Toughness != nil && !card.Toughness.IsStar {
		values.toughness = card.Toughness.Value
		values.toughnessOK = true
	}
	for i := range card.Abilities {
		for _, keyword := range card.Abilities[i].Keywords {
			values.keywords[keyword] = true
		}
	}
	for _, keyword := range keywordCounters(permanent) {
		values.keywords[keyword] = true
	}
	return values
}

func applyCounterAndTemporaryValues(permanent *game.Permanent, values *permanentEffectiveValues) {
	if permanent == nil || values == nil {
		return
	}
	counterDelta := powerToughnessCounterDelta(permanent)
	if values.powerOK {
		values.power += counterDelta + permanent.TemporaryPowerModifier
	}
	if values.toughnessOK {
		values.toughness += counterDelta + permanent.TemporaryToughnessModifier
	}
}

func applyStaticPTValues(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) {
	if g == nil || permanent == nil || values == nil {
		return
	}
	for _, source := range g.Battlefield {
		sourceDef := permanentCardDef(g, source)
		if sourceDef == nil {
			continue
		}
		for i := range sourceDef.Abilities {
			ability := &sourceDef.Abilities[i]
			if ability.Kind != game.StaticAbility || !abilityFunctionsOnBattlefield(ability) {
				continue
			}
			for _, effect := range ability.Effects {
				if effect.Type != game.EffectModifyPT || !permanentMatchesSelectorForSource(g, source, source.Controller, permanent, effect.Selector) {
					continue
				}
				if values.powerOK {
					values.power += effect.PowerDelta
				}
				if values.toughnessOK {
					values.toughness += effect.ToughnessDelta
				}
			}
		}
	}
}

func abilityFunctionsOnBattlefield(ability *game.AbilityDef) bool {
	return ability != nil && (ability.ZoneOfFunction == game.ZoneNone || ability.ZoneOfFunction == game.ZoneBattlefield)
}

func powerToughnessCounterDelta(permanent *game.Permanent) int {
	if permanent == nil {
		return 0
	}
	return permanent.Counters.Get(counter.PlusOnePlusOne) - permanent.Counters.Get(counter.MinusOneMinusOne)
}

func keywordCounters(permanent *game.Permanent) []game.Keyword {
	if permanent == nil {
		return nil
	}
	var keywords []game.Keyword
	for _, keyword := range []game.Keyword{
		game.Deathtouch,
		game.FirstStrike,
		game.Flying,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Reach,
		game.Trample,
		game.Vigilance,
	} {
		counterKind, ok := keywordCounterKind(keyword)
		if ok && permanent.Counters.Get(counterKind) > 0 {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}
