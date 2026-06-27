package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ManifestationSage is the card definition for Manifestation Sage.
//
// Type: Creature — Human Wizard
// Cost: {G/U}{G/U}{G/U}{G/U}
//
// Oracle text:
//
//	When this creature enters, create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of cards in your hand.
var ManifestationSage = newManifestationSage()

func newManifestationSage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Manifestation Sage",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
								Primitive: game.CreateToken{
									Amount:        game.Fixed(1),
									Source:        game.TokenDef(manifestationSageToken),
									PublishLinked: game.LinkedKey("created-token"),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									Object:      game.LinkedObjectReference("created-token"),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of cards in your hand.
		`,
		},
	}
}

var manifestationSageToken = newManifestationSageToken()

func newManifestationSageToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fractal",
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fractal},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
