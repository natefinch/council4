package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LupinflowerVillage is the card definition for Lupinflower Village.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {W}. Spend this mana only to cast a creature spell.
//	{1}{W}, {T}, Sacrifice this land: Look at the top six cards of your library. You may reveal a Bat, Bird, Mouse, or Rabbit card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var LupinflowerVillage = newLupinflowerVillage()

func newLupinflowerVillage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Lupinflower Village",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{W}, {T}, Sacrifice this land: Look at the top six cards of your library. You may reveal a Bat, Bird, Mouse, or Rabbit card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this land",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(6),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Bat"), types.Sub("Bird"), types.Sub("Mouse"), types.Sub("Rabbit")}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.W,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastCreatureSpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}.
			{T}: Add {W}. Spend this mana only to cast a creature spell.
			{1}{W}, {T}, Sacrifice this land: Look at the top six cards of your library. You may reveal a Bat, Bird, Mouse, or Rabbit card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
