package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RealmCloakedGiant is the card definition for Realm-Cloaked Giant // Cast Off.
//
// Type: Creature — Giant // Sorcery — Adventure
// Cost: {5}{W}{W} // {3}{W}{W}
// Face: Cast Off — Sorcery — Adventure ({3}{W}{W})
//
// Oracle text:
//
//	Vigilance
var RealmCloakedGiant = newRealmCloakedGiant()

func newRealmCloakedGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Realm-Cloaked Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Cast Off",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Giant")}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all non-Giant creatures. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
