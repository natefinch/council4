package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var MoggSalvage = cardset.Entry{
	OracleID:     "0923ad85-2ea4-4463-8f71-b7d0a162b486",
	Name:         "Mogg Salvage",
	New:          newMoggSalvage,
	Verification: "mtg/rules/mogg_salvage_test.go: TestAuthoredMoggSalvageScenario",
}

func newMoggSalvage() *game.CardDef {
	const oracleText = "If an opponent controls an Island and you control a Mountain, you may cast this spell without paying its mana cost.\nDestroy target artifact."
	const effectText = "Destroy target artifact."

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Mogg Salvage",
			OracleText: oracleText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Red},
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Mountain},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Island},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: effectText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
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
