package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Connive is the card definition for Connive // Concoct.
//
// Type: Sorcery // Sorcery
// Cost: {2}{U/B}{U/B} // {3}{U}{B}
// Face: Concoct — Sorcery ({3}{U}{B})
//
// Oracle text:
//
//	Gain control of target creature with power 2 or less.
var Connive = newConnive

func newConnive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Connive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors: []color.Color{color.Blue, color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with power 2 or less",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:      game.DurationPermanent,
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Gain control of target creature with power 2 or less.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Concoct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Blue, color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to return to the battlefield",
						},
					},
				},
			}.Ability()),
			OracleText: `
			Surveil 3, then return a creature card from your graveyard to the battlefield.
		`,
		}),
	}
}
