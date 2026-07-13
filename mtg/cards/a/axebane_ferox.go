package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AxebaneFerox is the card definition for Axebane Ferox.
//
// Type: Creature — Beast
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Deathtouch, haste
//	Ward—Collect evidence 4. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player exiles cards with total mana value 4 or greater from their graveyard.)
var AxebaneFerox = newAxebaneFerox

func newAxebaneFerox() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Axebane Ferox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.HasteStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalCollectEvidence,
						Text:   "Collect evidence 4",
						Amount: 4,
						Source: zone.Graveyard,
					},
				}),
			},
			OracleText: `
			Deathtouch, haste
			Ward—Collect evidence 4. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player exiles cards with total mana value 4 or greater from their graveyard.)
		`,
		},
	}
}
