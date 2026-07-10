package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CoralColony is the card definition for Coral Colony.
//
// Type: Creature — Wall
// Cost: {1}{U}
//
// Oracle text:
//
//	Defender
//	{1}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender. (To mill a card, a player puts the top card of their library into their graveyard.)
var CoralColony = newCoralColony

func newCoralColony() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Coral Colony",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender. (To mill a card, a player puts the top card of their library into their graveyard.)",
					ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.U}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: game.Defender}),
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			{1}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender. (To mill a card, a player puts the top card of their library into their graveyard.)
		`,
		},
	}
}
