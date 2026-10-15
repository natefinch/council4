package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IncendiarySabotage is the card definition for Incendiary Sabotage.
//
// Type: Instant
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact.
//	Incendiary Sabotage deals 3 damage to each creature.
var IncendiarySabotage = newIncendiarySabotage

func newIncendiarySabotage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Incendiary Sabotage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice an artifact.
			Incendiary Sabotage deals 3 damage to each creature.
		`,
		},
	}
}
