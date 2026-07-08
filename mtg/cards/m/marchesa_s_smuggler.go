package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MarchesaSSmuggler is the card definition for Marchesa's Smuggler.
//
// Type: Creature — Human Rogue
// Cost: {U}{R}
//
// Oracle text:
//
//	Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
//	{1}{U}{R}: Target creature you control gains haste until end of turn and can't be blocked this turn.
var MarchesaSSmuggler = newMarchesaSSmuggler

func newMarchesaSSmuggler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Marchesa's Smuggler",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{U}{R}: Target creature you control gains haste until end of turn and can't be blocked this turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.U, cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.DethroneTriggeredBody,
			},
			OracleText: `
			Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
			{1}{U}{R}: Target creature you control gains haste until end of turn and can't be blocked this turn.
		`,
		},
	}
}
