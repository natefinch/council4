package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KederektLeviathan is the card definition for Kederekt Leviathan.
//
// Type: Creature — Leviathan
// Cost: {6}{U}{U}
//
// Oracle text:
//
//	When this creature enters, return all other nonland permanents to their owners' hands.
//	Unearth {6}{U} ({6}{U}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
var KederektLeviathan = newKederektLeviathan

func newKederektLeviathan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Kederekt Leviathan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.UnearthActivatedAbility(cost.Mana{cost.O(6), cost.U}),
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
								Primitive: game.Bounce{
									Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludeSource: true}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, return all other nonland permanents to their owners' hands.
			Unearth {6}{U} ({6}{U}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
		`,
		},
	}
}
