package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VolleyOfBoulders is the card definition for Volley of Boulders.
//
// Type: Sorcery
// Cost: {8}{R}
//
// Oracle text:
//
//	Volley of Boulders deals 6 damage divided as you choose among any number of targets.
//	Flashback {R}{R}{R}{R}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var VolleyOfBoulders = newVolleyOfBoulders

func newVolleyOfBoulders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Volley of Boulders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {R}{R}{R}{R}{R}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.R,
								cost.R,
								cost.R,
								cost.R,
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
						MaxTargets: 6,
						Constraint: "any number of targets",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(6),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Volley of Boulders deals 6 damage divided as you choose among any number of targets.
			Flashback {R}{R}{R}{R}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
