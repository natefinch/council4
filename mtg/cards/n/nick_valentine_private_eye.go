package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NickValentinePrivateEye is the card definition for Nick Valentine, Private Eye.
//
// Type: Legendary Artifact Creature — Synth Detective
// Cost: {2}{U}
//
// Oracle text:
//
//	Nick Valentine can't be blocked except by artifact creatures.
//	Whenever Nick Valentine or another artifact creature you control dies, you may investigate. (To investigate, create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
var NickValentinePrivateEye = newNickValentinePrivateEye

func newNickValentinePrivateEye() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Nick Valentine, Private Eye",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Synth, types.Detective},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedExceptBy,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionArtifact,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentDied,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
						},
					},
					Optional: true,
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
			Nick Valentine can't be blocked except by artifact creatures.
			Whenever Nick Valentine or another artifact creature you control dies, you may investigate. (To investigate, create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
		`,
		},
	}
}
