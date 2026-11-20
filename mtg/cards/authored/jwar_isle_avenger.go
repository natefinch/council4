package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var JwarIsleAvenger = cardset.Entry{
	OracleID:     "2db8fc61-13af-4227-83f9-9f426ab978d3",
	Name:         "Jwar Isle Avenger",
	New:          newJwarIsleAvenger,
	Verification: "mtg/rules/jwar_isle_avenger_surge_test.go: TestJwarIsleAvengerSurgeCostAfterCasterCast",
}

func newJwarIsleAvenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Jwar Isle Avenger",
			ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sphinx},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: "Surge {2}{U} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nFlying",
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {2}{U}",
				ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.U}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
		},
	}
}
