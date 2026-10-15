package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JundCharm is the card definition for Jund Charm.
//
// Type: Instant
// Cost: {B}{R}{G}
//
// Oracle text:
//
//	Choose one —
//	• Exile target player's graveyard.
//	• Jund Charm deals 2 damage to each creature.
//	• Put two +1/+1 counters on target creature.
var JundCharm = newJundCharm

func newJundCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Jund Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target player's graveyard.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player's graveyard",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Player:      game.TargetPlayerReference(0),
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Jund Charm deals 2 damage to each creature.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(2),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
								},
							},
						},
					},
					game.Mode{
						Text: "Put two +1/+1 counters on target creature.",
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
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Exile target player's graveyard.
			• Jund Charm deals 2 damage to each creature.
			• Put two +1/+1 counters on target creature.
		`,
		},
	}
}
