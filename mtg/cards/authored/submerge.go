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

const submergeOracleID = "99427ebe-c00d-4206-84ca-9764f6e952c6"

var Submerge = cardset.Entry{
	OracleID: submergeOracleID,
	Name:     "Submerge",
	New:      newSubmerge,
	Verification: "mtg/rules/submerge_test.go: " +
		"TestAuthoredSubmergeConditionalAlternativeCostPutsTargetOnOwnerLibrary",
}

func newSubmerge() *game.CardDef {
	const effectText = "Put target creature on top of its owner's library."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Submerge",
			OracleText: "If an opponent controls a Forest and you control an Island, you may cast this spell without paying its mana cost.\n" + effectText,
			ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Island},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Forest},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: effectText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.MovePermanent{
						Object:      game.TargetPermanentReference(0),
						Destination: zone.Library,
					},
				}},
			}.Ability()),
		},
	}
}
