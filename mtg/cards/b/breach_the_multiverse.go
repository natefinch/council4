package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreachTheMultiverse is the card definition for Breach the Multiverse.
//
// Type: Sorcery
// Cost: {5}{B}{B}
//
// Oracle text:
//
//	Each player mills ten cards. For each player, choose a creature or planeswalker card in that player's graveyard. Put those cards onto the battlefield under your control. Then each creature you control becomes a Phyrexian in addition to its other types.
var BreachTheMultiverse = newBreachTheMultiverse

func newBreachTheMultiverse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Breach the Multiverse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount:      game.Fixed(10),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.ChooseCardFromEachGraveyard{
							Chooser:   game.ControllerReference(),
							Players:   game.AllPlayersReference(),
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}},
							LinkedKey: game.LinkedKey("mass-reanimation-chosen"),
						},
					},
					{
						Primitive: game.ReanimateLinkedCards{
							Controller: game.ControllerReference(),
							LinkedKey:  game.LinkedKey("mass-reanimation-chosen"),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:       game.LayerType,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									AddSubtypes: []types.Sub{types.Sub("Phyrexian")},
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player mills ten cards. For each player, choose a creature or planeswalker card in that player's graveyard. Put those cards onto the battlefield under your control. Then each creature you control becomes a Phyrexian in addition to its other types.
		`,
		},
	}
}
