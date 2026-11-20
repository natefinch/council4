package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var Pyromatics = cardset.Entry{
	OracleID: "9c01e2a6-61e7-455c-b09f-8c0de2266782",
	Name:     "Pyromatics",
	New:      newPyromatics,
	Verification: "mtg/rules/pyromatics_test.go: TestPyromaticsReplicateDealsDamage; " +
		"CR 115.4, 120.1, 120.2b, 120.3a, 601.2a-i, 603.2-4, 608.2b, " +
		"702.56a-b, 707.10/.10c",
}

func newPyromatics() *game.CardDef {
	const text = "Pyromatics deals 1 damage to any target."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Pyromatics",
			OracleText: "Replicate {1}{R} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)\n" + text,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.O(1), cost.R}},
				},
			}},
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
						Amount:    game.Fixed(1),
						Recipient: game.AnyTargetDamageRecipient(0),
					},
				}},
			}.Ability()),
		},
	}
}
