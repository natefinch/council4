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

const aethertowOracleID = "dd0cac88-0a78-408f-8eb0-080f79150180"

var aethertowSpell = game.ProgramReference{
	ID:      aethertowOracleID + "/spell/0",
	Version: "1",
}

var Aethertow = cardset.Entry{
	OracleID: aethertowOracleID,
	Name:     "Aethertow",
	New:      newAethertow,
	Programs: []cardprogram.Program{{
		Reference: aethertowSpell,
		Resolve:   resolveAethertow,
	}},
	Verification: "mtg/rules/aethertow_test.go: TestAethertowConspirePutsCombatCreaturesOnOwnersLibraries; " +
		"CR 2026-11-06 105.2-3, 400.3/400.7, 506.4, 508.1a/f, 509.1a/g, " +
		"601.2b/c/e-i, 603.2, 608.2b-c, 702.78a-b, 707.10/.10c",
}

func newAethertow() *game.CardDef {
	const text = "Put target attacking or blocking creature on top of its owner's library."
	const conspire = "Conspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:       "Aethertow",
			OracleText: text + "\n" + conspire,
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.W, mana.U),
			}),
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.White, color.Blue},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target attacking or blocking creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						CombatState:   game.CombatStateAttackingOrBlocking,
					}),
				}},
				Program: aethertowSpell,
			}.Ability()),
		},
	}
}

func resolveAethertow(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Library,
	})
	return err
}
