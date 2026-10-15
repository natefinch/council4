package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IncendiaryFlow is the card definition for Incendiary Flow.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Incendiary Flow deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
var IncendiaryFlow = newIncendiaryFlow

func newIncendiaryFlow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Incendiary Flow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.CreateReplacement{
							Replacement: &game.ReplacementEffect{
								Duration:                     game.DurationPermanent,
								MatchEvent:                   game.EventZoneChanged,
								MatchFromZone:                true,
								FromZone:                     zone.Battlefield,
								MatchToZone:                  true,
								ToZone:                       zone.Graveyard,
								ReplaceToZone:                zone.Exile,
								AffectedObjectMustBeCreature: true,
							},
							Duration: game.DurationThisTurn,
							Object:   game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Incendiary Flow deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
		`,
		},
	}
}
