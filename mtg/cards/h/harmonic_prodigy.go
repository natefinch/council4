package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarmonicProdigy is the card definition for Harmonic Prodigy.
//
// Type: Creature — Human Wizard
// Cost: {1}{R}
//
// Oracle text:
//
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	If a triggered ability of a Shaman or another Wizard you control triggers, that ability triggers an additional time.
var HarmonicProdigy = newHarmonicProdigy()

func newHarmonicProdigy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Harmonic Prodigy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
							AffectedSelection: game.Selection{AnyOf: []game.Selection{game.Selection{SubtypesAny: []types.Sub{types.Sub("Shaman")}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Wizard")}, ExcludeSource: true}}},
						},
					},
				},
			},
			OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			If a triggered ability of a Shaman or another Wizard you control triggers, that ability triggers an additional time.
		`,
		},
	}
}
