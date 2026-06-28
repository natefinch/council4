package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SilversmoteGhoul is the card definition for Silversmote Ghoul.
//
// Type: Creature — Zombie Vampire
// Cost: {2}{B}
//
// Oracle text:
//
//	At the beginning of your end step, if you gained 3 or more life this turn, return this card from your graveyard to the battlefield tapped.
//	{1}{B}, Sacrifice this creature: Draw a card.
var SilversmoteGhoul = newSilversmoteGhoul()

func newSilversmoteGhoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Silversmote Ghoul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Vampire},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{B}, Sacrifice this creature: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if you gained 3 or more life this turn",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGainedLifeThisTurn, Op: compare.GreaterOrEqual, Value: 3}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
									EntryTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, if you gained 3 or more life this turn, return this card from your graveyard to the battlefield tapped.
			{1}{B}, Sacrifice this creature: Draw a card.
		`,
		},
	}
}
