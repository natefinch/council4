package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IlluminatorVirtuoso is the card definition for Illuminator Virtuoso.
//
// Type: Creature — Human Rogue
// Cost: {1}{W}
//
// Oracle text:
//
//	Double strike
//	Whenever this creature becomes the target of a spell you control, it connives. (Draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on this creature.)
var IlluminatorVirtuoso = newIlluminatorVirtuoso()

func newIlluminatorVirtuoso() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Illuminator Virtuoso",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventObjectBecameTarget,
							Source:               game.TriggerSourceSelf,
							CauseController:      game.TriggerControllerYou,
							MatchStackObjectKind: true,
							StackObjectKind:      game.StackSpell,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Connive{
									Object: game.EventPermanentReference(),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Double strike
			Whenever this creature becomes the target of a spell you control, it connives. (Draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on this creature.)
		`,
		},
	}
}
