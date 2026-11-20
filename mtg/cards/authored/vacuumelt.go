package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const vacuumeltOracleID = "eb421a1b-147f-422b-bb8d-483a182a4096"

const vacuumeltOracleText = "Replicate {2}{U} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)\nReturn target creature to its owner's hand."

var Vacuumelt = cardset.Entry{
	OracleID: vacuumeltOracleID,
	Name:     "Vacuumelt",
	New:      newVacuumelt,
	Verification: "mtg/rules/vacuumelt_test.go: " +
		"TestVacuumeltReplicateReturnsTargetToOwnerHand",
}

func newVacuumelt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Vacuumelt",
			OracleText: vacuumeltOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.O(2), cost.U}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: "Return target creature to its owner's hand.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.MovePermanent{
						Object:      game.TargetPermanentReference(0),
						Destination: zone.Hand,
					},
				}},
			}.Ability()),
		},
	}
}
