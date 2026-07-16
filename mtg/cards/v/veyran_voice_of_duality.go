package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VeyranVoiceOfDuality is the card definition for Veyran, Voice of Duality.
//
// Type: Legendary Creature — Efreet Wizard
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Magecraft — Whenever you cast or copy an instant or sorcery spell, Veyran gets +1/+1 until end of turn.
//	If you casting or copying an instant or sorcery spell causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
var VeyranVoiceOfDuality = newVeyranVoiceOfDuality

func newVeyranVoiceOfDuality() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Veyran, Voice of Duality",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Efreet, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                                 game.RuleEffectAdditionalTriggerForControlledPermanent,
							TriggerCauseCastOrCopyInstantSorcery: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:          game.EventSpellCast,
							Controller:     game.TriggerControllerYou,
							MatchSpellCopy: true,
							CardSelection:  game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Magecraft — Whenever you cast or copy an instant or sorcery spell, Veyran gets +1/+1 until end of turn.
			If you casting or copying an instant or sorcery spell causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
		`,
		},
	}
}
