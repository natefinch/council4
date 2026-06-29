package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PathToTheWorldTree is the card definition for Path to the World Tree.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	When this enchantment enters, search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	{2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw two cards. Target opponent loses 2 life. This enchantment deals 2 damage to up to one target creature. You create a 2/2 green Bear creature token.
var PathToTheWorldTree = newPathToTheWorldTree()

func newPathToTheWorldTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Path to the World Tree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw two cards. Target opponent loses 2 life. This enchantment deals 2 damage to up to one target creature. You create a 2/2 green Bear creature token.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W, cost.U, cost.B, cost.R, cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this enchantment",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.AnyTargetDamageRecipient(1),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(pathToTheWorldTreeToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Hand,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										Reveal:      true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, search your library for a basic land card, reveal it, put it into your hand, then shuffle.
			{2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw two cards. Target opponent loses 2 life. This enchantment deals 2 damage to up to one target creature. You create a 2/2 green Bear creature token.
		`,
		},
	}
}

var pathToTheWorldTreeToken = newPathToTheWorldTreeToken()

func newPathToTheWorldTreeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bear",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bear},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
