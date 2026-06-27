package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZanarkandAncientMetropolis is the card definition for Zanarkand, Ancient Metropolis // Lasting Fayth.
//
// Type: Land — Town // Sorcery — Adventure
// Cost: {4}{G}{G}
// Face: Lasting Fayth — Sorcery — Adventure ({4}{G}{G})
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G}.
var ZanarkandAncientMetropolis = newZanarkandAncientMetropolis()

func newZanarkandAncientMetropolis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:     "Zanarkand, Ancient Metropolis",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G}.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Lasting Fayth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(zanarkandAncientMetropolisToken),
							PublishLinked: game.LinkedKey("created-token"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							}),
							Object:      game.LinkedObjectReference("created-token"),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 colorless Hero creature token. Put a +1/+1 counter on it for each land you control. (Then exile this card. You may play the land later from exile.)
		`,
		}),
	}
}

var zanarkandAncientMetropolisToken = newZanarkandAncientMetropolisToken()

func newZanarkandAncientMetropolisToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Hero",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hero},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
