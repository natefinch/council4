package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var gleefulSabotageProgram = game.ProgramReference{
	ID:      "25fe48be-95c4-4011-8c60-4628fb5ecfcb/spell/0",
	Version: "1",
}

var GleefulSabotage = cardset.Entry{
	OracleID: "25fe48be-95c4-4011-8c60-4628fb5ecfcb", Name: "Gleeful Sabotage",
	New: newGleefulSabotage,
	Programs: []cardprogram.Program{{
		Reference: gleefulSabotageProgram,
		Resolve:   resolveDestroyTarget,
	}},
	Verification: "mtg/rules/gleeful_sabotage_test.go: TestGleefulSabotageConspireDestroysArtifactAndEnchantment; CR 2026-11-06 601.2b,c,f,h,i, 603.2-3, 608.2b, 701.8, 702.78a, 707.10c",
}

func newGleefulSabotage() *game.CardDef {
	const text = "Destroy target artifact or enchantment."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Gleeful Sabotage",
			OracleText: text + "\nConspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Program: gleefulSabotageProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
					}),
				}},
			}.Ability()),
		},
	}
}
