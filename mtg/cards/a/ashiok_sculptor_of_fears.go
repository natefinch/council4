package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshiokSculptorOfFears is the card definition for Ashiok, Sculptor of Fears.
//
// Type: Legendary Planeswalker — Ashiok
// Cost: {4}{U}{B}
//
// Oracle text:
//
//	+2: Draw a card. Each player mills two cards.
//	−5: Put target creature card from a graveyard onto the battlefield under your control.
//	−11: Gain control of all creatures target opponent controls.
var AshiokSculptorOfFears = newAshiokSculptorOfFears

func newAshiokSculptorOfFears() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ashiok, Sculptor of Fears",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Ashiok},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(2),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -5,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from a graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									Recipient: opt.Val(game.ControllerReference()),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -11,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:         game.LayerControl,
											NewController: opt.Val(game.Player1),
											Group:         game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+2: Draw a card. Each player mills two cards.
			−5: Put target creature card from a graveyard onto the battlefield under your control.
			−11: Gain control of all creatures target opponent controls.
		`,
		},
	}
}
