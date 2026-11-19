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
	quirionDryadOracleID = "f42b6510-7322-4733-85ea-7ae6bd6ca8fc"
	quirionDryadText     = "Whenever you cast a spell that's white, blue, black, or red, put a +1/+1 counter on this creature."
)

var QuirionDryad = cardset.Entry{
	OracleID: quirionDryadOracleID,
	Name:     "Quirion Dryad",
	Verification: "mtg/rules/quirion_dryad_test.go: " +
		"TestQuirionDryadCastColorTrigger",
	New: newQuirionDryad,
}

func newQuirionDryad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Quirion Dryad",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dryad},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: quirionDryadText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: quirionDryadText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventSpellCast,
						Player: game.TriggerPlayerYou,
						CardSelection: game.Selection{
							ColorsAny: []color.Color{color.White, color.Blue, color.Black, color.Red},
						},
					},
				},
				Content: game.Mode{
					Text: quirionDryadText,
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(1),
						CounterKind: counter.PlusOnePlusOne,
					}}},
				}.Ability(),
			}},
		},
	}
}
