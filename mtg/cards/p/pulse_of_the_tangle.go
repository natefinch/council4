package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PulseOfTheTangle is the card definition for Pulse of the Tangle.
//
// Type: Sorcery
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, return Pulse of the Tangle to its owner's hand.
var PulseOfTheTangle = newPulseOfTheTangle

func newPulseOfTheTangle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pulse of the Tangle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(pulseOfTheTangleToken),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Hand,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlComparison: opt.Val(game.ControlCountComparison{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									Left:      game.ControlPlayerAnyOpponent,
									Right:     game.ControlPlayerController,
									Op:        compare.GreaterThan,
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, return Pulse of the Tangle to its owner's hand.
		`,
		},
	}
}

var pulseOfTheTangleToken = newPulseOfTheTangleToken()

func newPulseOfTheTangleToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
