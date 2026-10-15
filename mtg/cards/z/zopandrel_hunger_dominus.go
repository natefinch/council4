package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZopandrelHungerDominus is the card definition for Zopandrel, Hunger Dominus.
//
// Type: Legendary Creature — Phyrexian Horror
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Reach
//	At the beginning of each combat, double the power and toughness of each creature you control until end of turn.
//	{G/P}{G/P}, Sacrifice two other creatures: Put an indestructible counter on Zopandrel. ({G/P} can be paid with either {G} or 2 life.)
var ZopandrelHungerDominus = newZopandrelHungerDominus

func newZopandrelHungerDominus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Zopandrel, Hunger Dominus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Horror},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{G/P}{G/P}, Sacrifice two other creatures: Put an indestructible counter on Zopandrel. ({G/P} can be paid with either {G} or 2 life.)",
					ManaCost: opt.Val(cost.Mana{
						cost.PhyrexianMana(mana.G),
						cost.PhyrexianMana(mana.G),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice two other creatures",
							Amount:             2,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Indestructible,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of each combat, double the power and toughness of each creature you control until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:        game.DurationPermanent,
													Group:           game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:           game.LayerPowerToughnessModify,
													DoublePower:     true,
													DoubleToughness: true,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Reach
			At the beginning of each combat, double the power and toughness of each creature you control until end of turn.
			{G/P}{G/P}, Sacrifice two other creatures: Put an indestructible counter on Zopandrel. ({G/P} can be paid with either {G} or 2 life.)
		`,
		},
	}
}
