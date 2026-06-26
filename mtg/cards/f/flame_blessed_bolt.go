package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlameBlessedBolt is the card definition for Flame-Blessed Bolt.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Flame-Blessed Bolt deals 2 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
var FlameBlessedBolt = newFlameBlessedBolt()

func newFlameBlessedBolt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flame-Blessed Bolt",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
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
			Flame-Blessed Bolt deals 2 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
		`,
		},
	}
}
