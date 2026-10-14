package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Humility has two parts of one static effect across layers 6 and 7b.
var Humility = cardset.Entry{
	OracleID: "ed7bdb3e-5c51-4547-9266-76a791e0b2b0", Name: "Humility", New: newHumility,
	Verification: "mtg/rules/humility_opalescence_test.go; mtg/rules/challenge_followup_test.go: TestHumilityRemovesThaliaTaxAndRhonasRestrictions; TestHumilityAllowsLaterExternallyGrantedRuleAbility; TestHumilityPrintedHastePreservesControlHistoryForAttackTapUntap; CR 302.6, 613.1/4/6–8/11 (2026-09-25)",
}

// Opalescence preserves its layer-4 recipient set through layer 7b.
var Opalescence = cardset.Entry{
	OracleID: "59489b46-9d02-4f3c-bcd0-884e7605e9a5", Name: "Opalescence", New: newOpalescence,
	Verification: "mtg/rules/humility_opalescence_test.go; mtg/rules/challenge_followup_test.go: TestOpalescenceBrokenSkyManaValueAndDoubleFacedTokenCopies; TestCopiedOpalescenceUsesNewAbilityTimestampAndPhysicalExclusion; Scryfall rulings 2006-02-01, 2009-10-01; CR 202.3b, 613.7a, 707.8a, 712.8e (2026-09-25)",
}

func newHumility() *game.CardDef {
	const text = "All creatures lose all abilities and have base power and toughness 1/1."
	group := game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Humility", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W, cost.W}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.White}, OracleText: text,
			StaticAbilities: []game.StaticAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield,
				ContinuousEffects: []game.ContinuousEffect{
					{Layer: game.LayerAbility, Group: group, RemoveAllAbilities: true},
					{Layer: game.LayerPowerToughnessSet, Group: group, SetPower: opt.Val(game.PT{Value: 1}), SetToughness: opt.Val(game.PT{Value: 1})},
				},
			}},
		},
	}
}

func newOpalescence() *game.CardDef {
	const text = "Each other non-Aura enchantment is a creature in addition to its other types and has base power and base toughness each equal to its mana value."
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Enchantment}, ExcludedSubtype: types.Aura, ExcludeSource: true,
	})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Opalescence", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W, cost.W}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.White}, OracleText: text,
			StaticAbilities: []game.StaticAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield,
				ContinuousEffects: []game.ContinuousEffect{
					{Layer: game.LayerType, Group: group, AddTypes: []types.Card{types.Creature}},
					{Layer: game.LayerPowerToughnessSet, Group: group, SetPowerToManaValue: true, SetToughnessToManaValue: true},
				},
			}},
		},
	}
}
