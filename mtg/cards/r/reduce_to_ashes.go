package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReduceToAshes is the card definition for Reduce to Ashes.
//
// Type: Sorcery
// Cost: {4}{R}
//
// Oracle text:
//
//	Reduce to Ashes deals 5 damage to target creature. If that creature would die this turn, exile it instead.
var ReduceToAshes = newReduceToAshes()

func newReduceToAshes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Reduce to Ashes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.CreateReplacement{
							Replacement: &game.ReplacementEffect{
								MatchEvent:    game.EventZoneChanged,
								MatchFromZone: true,
								FromZone:      zone.Battlefield,
								MatchToZone:   true,
								ToZone:        zone.Graveyard,
								ReplaceToZone: zone.Exile,
							},
							Object:   game.TargetPermanentReference(0),
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reduce to Ashes deals 5 damage to target creature. If that creature would die this turn, exile it instead.
		`,
		},
	}
}
