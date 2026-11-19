package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var BurnTrail = cardset.Entry{
	OracleID: "566b4dcc-b7d7-4443-b2d0-954f1fc33a11",
	Name:     "Burn Trail",
	New:      newBurnTrail,
	Verification: "mtg/rules/burn_trail_test.go: " +
		"TestBurnTrailDealsDamageWithConspire; CR 2026-11-06 105.2, 115.4, " +
		"120.1, 120.2b, 120.3a, 601.2a-i, 603.2-3, 608.2b, 702.78a-b, 707.10/.10c",
}

func newBurnTrail() *game.CardDef {
	const text = "Burn Trail deals 3 damage to any target."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Burn Trail",
			OracleText: text + "\nConspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{
						types.Creature, types.Planeswalker, types.Battle,
					}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Damage{
						Amount:    game.Fixed(3),
						Recipient: game.AnyTargetDamageRecipient(0),
					},
				}},
			}.Ability()),
		},
	}
}
