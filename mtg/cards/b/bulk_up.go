package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BulkUp is the card definition for Bulk Up.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Double target creature's power until end of turn.
//	Flashback {4}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var BulkUp = newBulkUp

func newBulkUp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Bulk Up",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(4), cost.R, cost.R}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature's power",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectPower,
								Multiplier: 1,
								Object:     game.TargetPermanentReference(0),
							}),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Double target creature's power until end of turn.
			Flashback {4}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
