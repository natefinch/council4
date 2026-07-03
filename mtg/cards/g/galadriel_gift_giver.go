package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GaladrielGiftGiver is the card definition for Galadriel, Gift-Giver.
//
// Type: Legendary Creature — Elf Noble
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Whenever Galadriel enters or attacks, choose one —
//	• Put a +1/+1 counter on another target creature.
//	• Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	• Create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")
var GaladrielGiftGiver = newGaladrielGiftGiver()

func newGaladrielGiftGiver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Galadriel, Gift-Giver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Noble},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerDeclared,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Put a +1/+1 counter on another target creature.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(galadrielGiftGiverToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(galadrielGiftGiverToken2),
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
			Whenever Galadriel enters or attacks, choose one —
			• Put a +1/+1 counter on another target creature.
			• Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			• Create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")
		`,
		},
	}
}

var galadrielGiftGiverToken = newGaladrielGiftGiverToken()

func newGaladrielGiftGiverToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
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
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

var galadrielGiftGiverToken2 = newGaladrielGiftGiverToken2()

func newGaladrielGiftGiverToken2() *game.CardDef {
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
