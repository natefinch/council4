package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Nighthaze is the card definition for Nighthaze.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)
//	Draw a card.
var Nighthaze = newNighthaze

func newNighthaze() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nighthaze",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
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
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Layer:    game.LayerAbility,
									AddAbilities: []game.Ability{
										&game.StaticAbility{
											Text: "Swampwalk",
											KeywordAbilities: []game.KeywordAbility{
												game.LandwalkKeyword{
													Subtype: types.Swamp,
												},
											},
										},
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)
			Draw a card.
		`,
		},
	}
}
