package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrackdownConstruct is the card definition for Crackdown Construct.
//
// Type: Artifact Creature — Construct
// Cost: {4}
//
// Oracle text:
//
//	Whenever you activate an ability of an artifact or creature that isn't a mana ability, this creature gets +1/+1 until end of turn.
var CrackdownConstruct = newCrackdownConstruct

func newCrackdownConstruct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crackdown Construct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you activate an ability of an artifact or creature that isn't a mana ability, this creature gets +1/+1 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventAbilityActivated,
							Player:             game.TriggerPlayerYou,
							SubjectSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
							StackObjectKind:    game.StackSpell,
							ExcludeManaAbility: true,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
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
			Whenever you activate an ability of an artifact or creature that isn't a mana ability, this creature gets +1/+1 until end of turn.
		`,
		},
	}
}
