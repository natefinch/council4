package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonclawStrike is the card definition for Dragonclaw Strike.
//
// Type: Sorcery
// Cost: {2/G}{2/U}{2/R}
//
// Oracle text:
//
//	Double the power and toughness of target creature you control until end of turn. Then it fights up to one target creature an opponent controls. (Each deals damage equal to its power to the other.)
var DragonclawStrike = newDragonclawStrike()

func newDragonclawStrike() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Dragonclaw Strike",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.G),
				cost.Twobrid(mana.U),
				cost.Twobrid(mana.R),
			}),
			Colors: []color.Color{color.Green, color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature an opponent controls",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerPowerToughnessModify,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Fight{
							Object:        game.TargetPermanentReference(0),
							RelatedObject: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Double the power and toughness of target creature you control until end of turn. Then it fights up to one target creature an opponent controls. (Each deals damage equal to its power to the other.)
		`,
		},
	}
}
