package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnshakableTail is the card definition for Unshakable Tail.
//
// Type: Creature — Zombie Detective
// Cost: {2}{B}
//
// Oracle text:
//
//	When this creature enters and at the beginning of your upkeep, surveil 1.
//	Whenever one or more creature cards are put into your graveyard from your library, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
//	{2}, Sacrifice a Clue: Return this card from your graveyard to your hand.
var UnshakableTail = newUnshakableTail

func newUnshakableTail() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unshakable Tail",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Detective},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, Sacrifice a Clue: Return this card from your graveyard to your hand.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Clue",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Sub("Clue")},
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
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
								Primitive: game.Surveil{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Surveil{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							MatchFromZone:    true,
							FromZone:         zone.Library,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							OneOrMore:        true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Investigate{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters and at the beginning of your upkeep, surveil 1.
			Whenever one or more creature cards are put into your graveyard from your library, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
			{2}, Sacrifice a Clue: Return this card from your graveyard to your hand.
		`,
		},
	}
}
