package l

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

// LeylineOfAbundance is the card definition for Leyline of Abundance.
//
// Type: Enchantment
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	Whenever you tap a creature for mana, add an additional {G}.
//	{6}{G}{G}: Put a +1/+1 counter on each creature you control.
var LeylineOfAbundance = newLeylineOfAbundance

func newLeylineOfAbundance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Leyline of Abundance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{6}{G}{G}: Put a +1/+1 counter on each creature you control.",
					ManaCost:       opt.Val(cost.Mana{cost.O(6), cost.G, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Controller:           game.TriggerControllerYou,
							RequireTappedForMana: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			Whenever you tap a creature for mana, add an additional {G}.
			{6}{G}{G}: Put a +1/+1 counter on each creature you control.
		`,
		},
	}
}
