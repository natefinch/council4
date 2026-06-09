package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FanaticOfRhonas is the card definition for Fanatic of Rhonas.
//
// Type: Creature — Snake Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add {G}.
//	Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.
//	Eternalize {2}{G}{G}
var FanaticOfRhonas = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fanatic of Rhonas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			OracleText: `
				{T}: Add {G}.
				Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.
				Eternalize {2}{G}{G} ({2}{G}{G}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Snake Druid with no mana cost. Eternalize only as a sorcery.)
			`,
		},
	}

	card.ManaAbilities = append(card.ManaAbilities, game.TapManaAbility(mana.G))

	card.ManaAbilities = append(card.ManaAbilities,
		game.ManaAbility{
			Text: `
				Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.
			`,
			AdditionalCosts: cost.Tap,
			ActivationCondition: opt.Val(game.Condition{
				Text: "you control a creature with power 4 or greater",
				ControllerControls: game.PermanentFilter{
					Types: []types.Card{
						types.Creature,
					},
					Power: opt.Val(compare.Int{
						Op:    compare.GreaterOrEqual,
						Value: 4,
					}),
				},
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(4),
							ManaColor: mana.G,
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.EternalizeActivatedBody(
			cost.Mana{cost.O(2), cost.G, cost.G},
			types.Snake, types.Druid,
		),
	)
	return card
}()
