package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	fireSagesOracleID        = "08d1940a-ddc6-4b6d-b6c9-9d2aa470b906"
	fireSagesFirebendingText = "Firebending 1 (Whenever this creature attacks, add {R}. " +
		"This mana lasts until end of combat.)"
	fireSagesCounterText = "{1}{R}{R}: Put a +1/+1 counter on this creature."
)

var FireSages = cardset.Entry{
	OracleID: fireSagesOracleID,
	Name:     "Fire Sages",
	New:      newFireSages,
	Verification: "mtg/rules/fire_sages_test.go: " +
		"TestFireSagesAttackSpendsFirebendingManaForCounter",
}

func newFireSages() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Fire Sages",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Cleric},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: fireSagesFirebendingText + "\n" + fireSagesCounterText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: fireSagesFirebendingText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventAttackerDeclared,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:                  game.Fixed(1),
						ManaColor:               mana.R,
						PersistUntilEndOfCombat: true,
					},
				}}}.Ability(),
			}},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           fireSagesCounterText,
				ZoneOfFunction: zone.Battlefield,
				ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.R, cost.R}),
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(1),
						CounterKind: counter.PlusOnePlusOne,
					},
				}}}.Ability(),
			}},
		},
	}
}
