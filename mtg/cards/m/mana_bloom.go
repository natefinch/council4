package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaBloom is the card definition for Mana Bloom.
//
// Type: Enchantment
// Cost: {X}{G}
//
// Oracle text:
//
//	This enchantment enters with X charge counters on it.
//	Remove a charge counter from this enchantment: Add one mana of any color. Activate only once each turn.
//	At the beginning of your upkeep, if this enchantment has no charge counters on it, return it to its owner's hand.
var ManaBloom = newManaBloom

func newManaBloom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mana Bloom",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this enchantment",
							Amount:      1,
							CounterKind: counter.Charge,
						},
					},
					Timing: game.OncePerTurn,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if this enchantment has no charge counters on it",
						InterveningCondition: opt.Val(game.Condition{
							Negate:        true,
							Object:        opt.Val(game.SourcePermanentReference()),
							ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.Charge}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This enchantment enters with X charge counters on it.", game.CounterPlacement{Kind: counter.Charge, AmountFromX: true}),
			},
			OracleText: `
			This enchantment enters with X charge counters on it.
			Remove a charge counter from this enchantment: Add one mana of any color. Activate only once each turn.
			At the beginning of your upkeep, if this enchantment has no charge counters on it, return it to its owner's hand.
		`,
		},
	}
}
