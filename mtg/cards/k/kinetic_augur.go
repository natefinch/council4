package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KineticAugur is the card definition for Kinetic Augur.
//
// Type: Creature — Human Shaman
// Cost: {3}{R}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	Kinetic Augur's power is equal to the number of instant and sorcery cards in your graveyard.
//	When this creature enters, discard up to two cards, then draw that many cards.
var KineticAugur = newKineticAugur

func newKineticAugur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kinetic Augur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:       []color.Color{color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Shaman},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerInstantOrSorceryCardsInGraveyard}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.DiscardThenDraw{
									Player: game.ControllerReference(),
									Max:    2,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			Kinetic Augur's power is equal to the number of instant and sorcery cards in your graveyard.
			When this creature enters, discard up to two cards, then draw that many cards.
		`,
		},
	}
}
