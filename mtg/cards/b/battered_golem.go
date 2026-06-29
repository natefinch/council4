package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BatteredGolem is the card definition for Battered Golem.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	This creature doesn't untap during your untap step.
//	Whenever an artifact enters, you may untap this creature.
var BatteredGolem = newBatteredGolem()

func newBatteredGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Battered Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							AffectedSource: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature doesn't untap during your untap step.
			Whenever an artifact enters, you may untap this creature.
		`,
		},
	}
}
