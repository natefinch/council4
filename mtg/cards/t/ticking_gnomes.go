package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TickingGnomes is the card definition for Ticking Gnomes.
//
// Type: Artifact Creature — Gnome
// Cost: {3}
//
// Oracle text:
//
//	Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	Sacrifice this creature: It deals 1 damage to any target.
var TickingGnomes = newTickingGnomes

func newTickingGnomes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ticking Gnomes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gnome},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this creature: It deals 1 damage to any target.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			Sacrifice this creature: It deals 1 damage to any target.
		`,
		},
	}
}
