package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SisterOfSilence is the card definition for Sister of Silence.
//
// Type: Creature — Human Knight
// Cost: {4}{U}
//
// Oracle text:
//
//	Flash
//	Psychic Abomination — When this creature enters, counter target instant spell, sorcery spell, activated ability, or triggered ability.
var SisterOfSilence = newSisterOfSilence

func newSisterOfSilence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sister of Silence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant spell, sorcery spell, activated ability, or triggered ability",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Psychic Abomination — When this creature enters, counter target instant spell, sorcery spell, activated ability, or triggered ability.
		`,
		},
	}
}
