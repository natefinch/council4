package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Panharmonicon is the card definition for Panharmonicon.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
var Panharmonicon = newPanharmonicon

func newPanharmonicon() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Panharmonicon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectAdditionalTriggerForEnteringPermanent,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Artifact, types.Creature},
						},
					},
				},
			},
			OracleText: `
			If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
		`,
		},
	}
}
