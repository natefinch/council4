package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	caldaiaStrongarmOracleID   = "e0e6938b-7abf-4f69-a3d6-82a58912bff4"
	caldaiaStrongarmEnterText  = "When this creature enters, put two +1/+1 counters on target creature."
	caldaiaStrongarmOracleText = `When this creature enters, put two +1/+1 counters on target creature.
Blitz {3}{G} (If you cast this spell for its blitz cost, it gains haste and "When this creature dies, draw a card." Sacrifice it at the beginning of the next end step.)`
)

var CaldaiaStrongarm = cardset.Entry{
	OracleID: caldaiaStrongarmOracleID,
	Name:     "Caldaia Strongarm",
	New:      newCaldaiaStrongarm,
	Verification: "mtg/rules/caldaia_strongarm_test.go: " +
		"TestCaldaiaStrongarmBlitzCountersTargetCreature",
}

func newCaldaiaStrongarm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:             "Caldaia Strongarm",
			ManaCost:         opt.Val(cost.Mana{cost.O(4), cost.G}),
			AlternativeCosts: []cost.Alternative{cost.BlitzAlternative(cost.Mana{cost.O(3), cost.G})},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Warrior},
			Colors:           []color.Color{color.Green},
			Power:            opt.Val(game.PT{Value: 2}),
			Toughness:        opt.Val(game.PT{Value: 3}),
			OracleText:       caldaiaStrongarmOracleText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: caldaiaStrongarmEnterText,
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{
					Text: caldaiaStrongarmEnterText,
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}},
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Object:      game.TargetPermanentReference(0),
						Amount:      game.Fixed(2),
						CounterKind: counter.PlusOnePlusOne,
					}}},
				}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{game.BlitzStaticAbility()},
		},
	}
}
