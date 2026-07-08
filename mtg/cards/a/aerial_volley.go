package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AerialVolley is the card definition for Aerial Volley.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Aerial Volley deals 3 damage divided as you choose among one, two, or three target creatures with flying.
var AerialVolley = newAerialVolley

func newAerialVolley() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Aerial Volley",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 3,
						Constraint: "one, two, or three target creatures with flying",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Keyword: game.Flying}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Aerial Volley deals 3 damage divided as you choose among one, two, or three target creatures with flying.
		`,
		},
	}
}
