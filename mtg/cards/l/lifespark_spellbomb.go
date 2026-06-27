package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LifesparkSpellbomb is the card definition for Lifespark Spellbomb.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	{G}, Sacrifice this artifact: Until end of turn, target land becomes a 3/3 creature that's still a land.
//	{1}, Sacrifice this artifact: Draw a card.
var LifesparkSpellbomb = newLifesparkSpellbomb()

func newLifesparkSpellbomb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lifespark Spellbomb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{G}, Sacrifice this artifact: Until end of turn, target land becomes a 3/3 creature that's still a land.",
					ManaCost: opt.Val(cost.Mana{cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:    game.LayerType,
											AddTypes: []types.Card{types.Creature},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 3}),
											SetToughness: opt.Val(game.PT{Value: 3}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice this artifact: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{G}, Sacrifice this artifact: Until end of turn, target land becomes a 3/3 creature that's still a land.
			{1}, Sacrifice this artifact: Draw a card.
		`,
		},
	}
}
