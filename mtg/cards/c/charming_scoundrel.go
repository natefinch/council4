package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CharmingScoundrel is the card definition for Charming Scoundrel.
//
// Type: Creature — Human Rogue
// Cost: {1}{R}
//
// Oracle text:
//
//	Haste
//	When this creature enters, choose one —
//	• Discard a card, then draw a card.
//	• Create a Treasure token.
//	• Create a Wicked Role token attached to target creature you control.
var CharmingScoundrel = newCharmingScoundrel

func newCharmingScoundrel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Charming Scoundrel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
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
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Discard a card, then draw a card.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Treasure token.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(charmingScoundrelToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Wicked Role token attached to target creature you control.",
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
										Primitive: game.CreateToken{
											Amount:          game.Fixed(1),
											Source:          game.TokenDef(charmingScoundrelToken2),
											EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Haste
			When this creature enters, choose one —
			• Discard a card, then draw a card.
			• Create a Treasure token.
			• Create a Wicked Role token attached to target creature you control.
		`,
		},
	}
}

var charmingScoundrelToken = newCharmingScoundrelToken()

func newCharmingScoundrelToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

var charmingScoundrelToken2 = newCharmingScoundrelToken2()

func newCharmingScoundrelToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Wicked Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
							MatchToZone:   true,
							ToZone:        zone.Graveyard,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1.
			When this Aura is put into a graveyard from the battlefield, each opponent loses 1 life.
		`,
		},
	}
}
