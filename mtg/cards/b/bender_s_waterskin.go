package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BenderSWaterskin is the card definition for Bender's Waterskin.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Untap this artifact during each other player's untap step.
//	{T}: Add one mana of any color.
var BenderSWaterskin = newBenderSWaterskin

func newBenderSWaterskin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bender's Waterskin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Untap this artifact during each other player's untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectUntapDuringOtherPlayersUntapStep,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Untap this artifact during each other player's untap step.
			{T}: Add one mana of any color.
		`,
		},
	}
}
