package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TappingAtTheWindow is the card definition for Tapping at the Window.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Look at the top three cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest into your graveyard.
//	Flashback {2}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var TappingAtTheWindow = newTappingAtTheWindow()

func newTappingAtTheWindow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tapping at the Window",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(2), cost.G}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:   game.ControllerReference(),
							Look:     game.Fixed(3),
							Take:     game.Fixed(1),
							Filter:   opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							TakeUpTo: true,
							Reveal:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top three cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest into your graveyard.
			Flashback {2}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
