package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const memorySluiceOracleID = "a30b659e-ae2b-4736-9b4f-a9805905bf75"

var memorySluiceSpell = game.ProgramReference{
	ID:      memorySluiceOracleID + "/spell/0",
	Version: "1",
}

var MemorySluice = cardset.Entry{
	OracleID: memorySluiceOracleID,
	Name:     "Memory Sluice",
	New:      newMemorySluice,
	Programs: []cardprogram.Program{{
		Reference: memorySluiceSpell,
		Resolve:   resolveMemorySluice,
	}},
	Verification: "mtg/rules/memory_sluice_test.go: " +
		"TestMemorySluiceConspireMillsDifferentTargetPlayers; CR 2026-11-06 " +
		"105.2, 601.2b/c/f-i, 608.2b/c/n, 701.17a-b, 702.78a, 707.10/.10c",
}

func newMemorySluice() *game.CardDef {
	const text = "Target player mills four cards."
	const conspire = "Conspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:       "Memory Sluice",
			OracleText: text + "\n" + conspire,
			ManaCost:   opt.Val(cost.Mana{cost.HybridMana(mana.U, mana.B)}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue, color.Black},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player",
					Allow:      game.TargetAllowPlayer,
				}},
				Program: memorySluiceSpell,
			}.Ability()),
		},
	}
}

func resolveMemorySluice(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MoveTopOfLibrary{
		Player:      game.TargetPlayerReference(0),
		Amount:      game.Fixed(4),
		Destination: zone.Graveyard,
	})
	return err
}
