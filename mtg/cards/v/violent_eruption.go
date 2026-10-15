package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ViolentEruption is the card definition for Violent Eruption.
//
// Type: Instant
// Cost: {1}{R}{R}{R}
//
// Oracle text:
//
//	Violent Eruption deals 4 damage divided as you choose among any number of targets.
//	Madness {1}{R}{R} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var ViolentEruption = newViolentEruption

func newViolentEruption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Violent Eruption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Madness {1}{R}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.R,
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 4,
						Constraint: "any number of targets",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Violent Eruption deals 4 damage divided as you choose among any number of targets.
			Madness {1}{R}{R} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
