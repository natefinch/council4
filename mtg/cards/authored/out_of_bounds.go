package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const outOfBoundsOracleID = "431b2e52-1290-47b1-920d-2c1817553d96"

var OutOfBounds = cardset.Entry{
	OracleID: outOfBoundsOracleID,
	Name:     "Out of Bounds",
	New:      newOutOfBounds,
	Verification: "mtg/rules/out_of_bounds_test.go: TestOutOfBoundsAssistCountersTargetSpell; " +
		"CR 2026-09-25 601.2a-i, 608.2b-c, 701.6a-b, 702.132a",
}

func newOutOfBounds() *game.CardDef {
	const text = "Counter target spell."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Out of Bounds",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			OracleText: "Assist (Another player can pay up to {3} of this spell's cost.)\n" + text,
			StaticAbilities: []game.StaticAbility{
				game.AssistStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowStackObject,
					Predicate: game.TargetPredicate{
						StackObjectKinds: []game.StackObjectKind{game.StackSpell},
					},
				}},
				Sequence: []game.Instruction{{
					Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
				}},
			}.Ability()),
		},
	}
}
