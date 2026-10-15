package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DamningVerdict is the card definition for Damning Verdict.
//
// Type: Sorcery
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Destroy all creatures with no counters on them.
var DamningVerdict = newDamningVerdict

func newDamningVerdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Damning Verdict",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchNoCounters: true}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures with no counters on them.
		`,
		},
	}
}
