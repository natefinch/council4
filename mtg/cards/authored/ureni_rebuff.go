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

var UrenisRebuff = cardset.Entry{
	OracleID: "c0b7e989-8194-484c-93aa-a44df0764a77",
	Name:     "Ureni's Rebuff",
	New:      newUrenisRebuff,
	Verification: "mtg/rules/ureni_rebuff_test.go: " +
		"TestUrenisRebuffHarmonizeReturnsTargetCreatureToOwnersHand; printing " +
		"722716df-9cea-40a7-924b-c28497e227e6; Scryfall rulings retrieved " +
		"2026-11-21: 0; CR 2026-11-06 115.1a, 118.7a, 307.1, 400.3, " +
		"601.2b,c,f-h, 608.2b, 702.180a-b",
}

func newUrenisRebuff() *game.CardDef {
	const spellText = "Return target creature to its owner's hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Ureni's Rebuff",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue},
			OracleText: spellText + "\nHarmonize {5}{U} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)",
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(5), cost.U}),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.MovePermanent{
						Object:      game.TargetPermanentReference(0),
						Destination: zone.Hand,
					},
				}},
			}.Ability()),
		},
	}
}
