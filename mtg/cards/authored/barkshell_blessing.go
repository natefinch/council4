package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var BarkshellBlessing = cardset.Entry{
	OracleID: "baf3d4d9-0590-4d94-8f4e-1e726194b920",
	Name:     "Barkshell Blessing",
	New:      newBarkshellBlessing,
	Verification: "mtg/rules/barkshell_blessing_test.go: " +
		"TestBarkshellBlessingPumpsTargetWithConspire; CR 2026-11-06 " +
		"202.2d, 115.1a, 601.2a-i, 603.2-3, 608.2b, 611.2a, 613.4c, " +
		"702.78a-b, 707.10/10c",
}

func newBarkshellBlessing() *game.CardDef {
	const text = "Target creature gets +2/+2 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.White),
		CardFace: game.CardFace{
			Name:       "Barkshell Blessing",
			OracleText: text + "\nConspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)",
			ManaCost:   opt.Val(cost.Mana{cost.HybridMana(mana.G, mana.W)}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green, color.White},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ModifyPT{
						Object:         game.TargetPermanentReference(0),
						PowerDelta:     game.Fixed(2),
						ToughnessDelta: game.Fixed(2),
						Duration:       game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability()),
		},
	}
}
