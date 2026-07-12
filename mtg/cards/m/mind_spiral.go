package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindSpiral is the card definition for Mind Spiral.
//
// Type: Sorcery
// Cost: {4}{U}
//
// Oracle text:
//
//	Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
//	Target player draws three cards. If the gift was promised, tap target creature an opponent controls and put a stun counter on it. (If a permanent with a stun counter would become untapped, remove one from it instead.)
var MindSpiral = newMindSpiral

func newMindSpiral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mind Spiral",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.GiftKeyword{Delivery: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.CreateToken{
										Amount:      game.Fixed(1),
										Source:      game.TokenDef(mindSpiralToken),
										Recipient:   opt.Val(game.GiftRecipientReference()),
										EntryTapped: true,
									},
								},
							},
						}.Ability()},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature an opponent controls",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
						Gate:       game.TargetGateGiftPromised,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Object:      game.TargetPermanentReference(1),
							CounterKind: counter.Stun,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
			Target player draws three cards. If the gift was promised, tap target creature an opponent controls and put a stun counter on it. (If a permanent with a stun counter would become untapped, remove one from it instead.)
		`,
		},
	}
}

var mindSpiralToken = newMindSpiralToken()

func newMindSpiralToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fish",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
