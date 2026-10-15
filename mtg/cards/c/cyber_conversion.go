package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CyberConversion is the card definition for Cyber Conversion.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Turn target creature face down. It's a 2/2 Cyberman artifact creature.
var CyberConversion = newCyberConversion

func newCyberConversion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cyber Conversion",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.TurnFaceDown{
							Object: game.TargetPermanentReference(0),
							Characteristics: opt.Val(game.FaceDownCharacteristics{
								Types:    []types.Card{types.Artifact, types.Creature},
								Subtypes: []types.Sub{types.Cyberman},
								Power: game.PT{
									Value: 2,
								},
								Toughness: game.PT{
									Value: 2,
								},
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Turn target creature face down. It's a 2/2 Cyberman artifact creature.
		`,
		},
	}
}
