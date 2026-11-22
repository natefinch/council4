package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var GoblinFreerunner = cardset.Entry{
	OracleID:     "372094f8-86dc-4c78-a5d0-444ffbf02630",
	Name:         "Goblin Freerunner",
	New:          newGoblinFreerunner,
	Verification: "mtg/rules/goblin_freerunner_surge_test.go: TestGoblinFreerunnerSurgeCastHasMenace",
}

func newGoblinFreerunner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Goblin Freerunner",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Goblin, types.Warrior, types.Ally},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: "Surge {1}{R} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nMenace (This creature can't be blocked except by two or more creatures.)",
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {1}{R}",
				ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.R}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			StaticAbilities: []game.StaticAbility{game.MenaceStaticBody},
		},
	}
}
