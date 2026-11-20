package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const gigadrowseReplicateText = "Replicate {U} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)"
const gigadrowseTapText = "Tap target permanent."

var Gigadrowse = cardset.Entry{
	OracleID: "9cc64a91-82d2-4be1-812f-531863e99801",
	Name:     "Gigadrowse",
	New:      newGigadrowse,
	Verification: "mtg/rules/gigadrowse_test.go: " +
		"TestGigadrowseReplicateTapsTargetPermanent",
}

func newGigadrowse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Gigadrowse",
			OracleText: gigadrowseReplicateText + "\n" + gigadrowseTapText,
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{{
				Text: gigadrowseReplicateText,
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.U}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: gigadrowseTapText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target permanent",
					Allow:      game.TargetAllowPermanent,
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Tap{Object: game.TargetPermanentReference(0)},
				}},
			}.Ability()),
		},
	}
}
