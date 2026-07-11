package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MineCollapse is the card definition for Mine Collapse.
//
// Type: Instant
// Cost: {3}{R}
//
// Oracle text:
//
//	If it's your turn, you may sacrifice a Mountain rather than pay this spell's mana cost.
//	Mine Collapse deals 5 damage to target creature or planeswalker.
var MineCollapse = newMineCollapse

func newMineCollapse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mine Collapse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a Mountain",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "sacrifice a Mountain",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Mountain},
						},
					},
					Condition: cost.AlternativeConditionYourTurn,
				},
			},
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
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If it's your turn, you may sacrifice a Mountain rather than pay this spell's mana cost.
			Mine Collapse deals 5 damage to target creature or planeswalker.
		`,
		},
	}
}
