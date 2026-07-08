package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NicoMinoruRunaway is the card definition for Nico Minoru, Runaway.
//
// Type: Legendary Creature — Human Warlock Hero
// Cost: {3}{R}
//
// Oracle text:
//
//	Whenever you cast a spell from anywhere other than your hand, Nico Minoru deals 2 damage to each opponent.
//	{2}{R}, {T}, Discard a card: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.
var NicoMinoruRunaway = newNicoMinoruRunaway

func newNicoMinoruRunaway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Nico Minoru, Runaway",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warlock, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{R}, {T}, Discard a card: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ExileLibraryUntilNonlandCast{
									Player: game.ControllerReference(),
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
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							ExcludeFromZone: true,
							FromZone:        zone.Hand,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(2),
									Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast a spell from anywhere other than your hand, Nico Minoru deals 2 damage to each opponent.
			{2}{R}, {T}, Discard a card: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.
		`,
		},
	}
}
