package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArchaeomancerSMap is the card definition for Archaeomancer's Map.
//
// Type: Artifact
// Cost: {2}{W}
//
// Oracle text:
//
//	When this artifact enters, search your library for up to two basic Plains cards, reveal them, put them into your hand, then shuffle.
//	Whenever a land an opponent controls enters, if that player controls more lands than you, you may put a land card from your hand onto the battlefield.
var ArchaeomancerSMap = newArchaeomancerSMap

func newArchaeomancerSMap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Archaeomancer's Map",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this artifact enters, search your library for up to two basic Plains cards, reveal them, put them into your hand, then shuffle.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Plains}},
												Reveal:      true,
											},
											Amount: game.Fixed(2),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever a land an opponent controls enters, if that player controls more lands than you, you may put a land card from your hand onto the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerOpponent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf: "if that player controls more lands than you",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if that player controls more lands than you",
							ControlComparison: opt.Val(game.ControlCountComparison{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
								Left:      game.ControlPlayerTriggeringPlayer,
								Op:        compare.GreaterThan,
							}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Prompt: "Choose a card to put onto the battlefield",
										},
										Optional: true,
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
			When this artifact enters, search your library for up to two basic Plains cards, reveal them, put them into your hand, then shuffle.
			Whenever a land an opponent controls enters, if that player controls more lands than you, you may put a land card from your hand onto the battlefield.
		`,
		},
	}
}
