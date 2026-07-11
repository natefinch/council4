package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnuffOut is the card definition for Snuff Out.
//
// Type: Instant
// Cost: {3}{B}
//
// Oracle text:
//
//	If you control a Swamp, you may pay 4 life rather than pay this spell's mana cost.
//	Destroy target nonblack creature. It can't be regenerated.
var SnuffOut = newSnuffOut

func newSnuffOut() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Snuff Out",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Pay 4 life",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "pay 4 life",
							Amount: 4,
						},
					},
					Condition:        cost.AlternativeConditionControlsPermanentSubtype,
					ConditionSubtype: types.Swamp,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonblack creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.Black}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a Swamp, you may pay 4 life rather than pay this spell's mana cost.
			Destroy target nonblack creature. It can't be regenerated.
		`,
		},
	}
}
