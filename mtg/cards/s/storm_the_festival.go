package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StormTheFestival is the card definition for Storm the Festival.
//
// Type: Sorcery
// Cost: {3}{G}{G}{G}
//
// Oracle text:
//
//	Look at the top five cards of your library. You may put up to two permanent cards with mana value 5 or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
//	Flashback {7}{G}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var StormTheFestival = newStormTheFestival()

func newStormTheFestival() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Storm the Festival",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(7), cost.G, cost.G, cost.G}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:      game.ControllerReference(),
							Look:        game.Fixed(5),
							Take:        game.Fixed(2),
							Remainder:   game.DigRemainderLibraryBottom,
							Filter:      opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 5})}),
							TakeUpTo:    true,
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top five cards of your library. You may put up to two permanent cards with mana value 5 or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
			Flashback {7}{G}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
