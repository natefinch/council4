package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const shatteringSpreeReplicateText = "Replicate {R} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)"
const shatteringSpreeDestroyText = "Destroy target artifact."

var ShatteringSpree = cardset.Entry{
	OracleID: "fe725b27-49ef-41bc-91f9-9cd9871ad863",
	Name:     "Shattering Spree",
	New:      newShatteringSpree,
	Verification: "mtg/rules/shattering_spree_test.go: TestAuthoredShatteringSpreeScenario; " +
		"pinned Oracle printing d6dcff21-5900-43c4-a38b-cdc19c704ce4; " +
		"Scryfall WotC rulings 2024-01-12; CR 2026-11-06 601.2b/c/f-i, " +
		"603.2-3, 113.7a, 608.2b, 701.8a-c, 702.56a-b, 707.10",
}

func newShatteringSpree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Shattering Spree",
			OracleText: shatteringSpreeReplicateText + "\n" + shatteringSpreeDestroyText,
			ManaCost:   opt.Val(cost.Mana{cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{{
				Text: shatteringSpreeReplicateText,
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.R}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: shatteringSpreeDestroyText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability()),
		},
	}
}
