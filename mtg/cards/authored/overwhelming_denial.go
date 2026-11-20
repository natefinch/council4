package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var OverwhelmingDenial = cardset.Entry{
	OracleID: "45aa8bdb-54da-4069-bd6c-c63396154464",
	Name:     "Overwhelming Denial",
	New:      newOverwhelmingDenial,
	Verification: "mtg/rules/overwhelming_denial_test.go: " +
		"TestOverwhelmingDenialSurgeUncounterableAndCountersTargetSpell",
}

func newOverwhelmingDenial() *game.CardDef {
	const text = "Surge {U}{U} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nThis spell can't be countered.\nCounter target spell."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Overwhelming Denial",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {U}{U}",
				ManaCost:  opt.Val(cost.Mana{cost.U, cost.U}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			StaticAbilities: []game.StaticAbility{game.CantBeCounteredStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Text: "Counter target spell.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowStackObject,
					Predicate: game.TargetPredicate{
						StackObjectKinds: []game.StackObjectKind{game.StackSpell},
					},
				}},
				Sequence: []game.Instruction{{
					Primitive: game.CounterObject{
						Object: game.TargetStackObjectReference(0),
					},
				}},
			}.Ability()),
		},
	}
}
