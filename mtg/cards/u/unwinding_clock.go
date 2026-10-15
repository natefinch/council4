package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnwindingClock is the card definition for Unwinding Clock.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Untap all artifacts you control during each other player's untap step.
var UnwindingClock = newUnwindingClock

func newUnwindingClock() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Unwinding Clock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Untap all artifacts you control during each other player's untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Artifact},
						},
					},
				},
			},
			OracleText: `
			Untap all artifacts you control during each other player's untap step.
		`,
		},
	}
}
