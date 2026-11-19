package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	nightClubberOracleID   = "4d2abfe1-f5f8-4ed5-915d-bfe98c6321ee"
	nightClubberEnterText  = "When this creature enters, creatures your opponents control get -1/-1 until end of turn."
	nightClubberOracleText = `When this creature enters, creatures your opponents control get -1/-1 until end of turn.
Blitz {2}{B} (If you cast this spell for its blitz cost, it gains haste and "When this creature dies, draw a card." Sacrifice it at the beginning of the next end step.)`
)

var NightClubber = cardset.Entry{
	OracleID: nightClubberOracleID,
	Name:     "Night Clubber",
	New:      newNightClubber,
	Verification: "mtg/rules/night_clubber_test.go: " +
		"TestNightClubberBlitzEntryGivesOpponentsMinusOneUntilEndOfTurnAndDraws",
}

func newNightClubber() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:             "Night Clubber",
			ManaCost:         opt.Val(cost.Mana{cost.O(1), cost.B, cost.B}),
			AlternativeCosts: []cost.Alternative{cost.BlitzAlternative(cost.Mana{cost.O(2), cost.B})},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Warrior},
			Colors:           []color.Color{color.Black},
			Power:            opt.Val(game.PT{Value: 2}),
			Toughness:        opt.Val(game.PT{Value: 2}),
			OracleText:       nightClubberOracleText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: nightClubberEnterText,
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						ContinuousEffects: []game.ContinuousEffect{{
							Layer: game.LayerPowerToughnessModify,
							Group: game.BattlefieldGroup(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								Controller:    game.ControllerOpponent,
							}),
							PowerDelta:     -1,
							ToughnessDelta: -1,
						}},
						Duration: game.DurationUntilEndOfTurn,
					},
				}}}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{game.BlitzStaticAbility()},
		},
	}
}
