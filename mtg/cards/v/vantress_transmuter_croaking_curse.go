package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VantressTransmuter is the card definition for Vantress Transmuter // Croaking Curse.
//
// Type: Creature — Human Wizard // Sorcery — Adventure
// Cost: {3}{U} // {1}{U}
// Face: Croaking Curse — Sorcery — Adventure ({1}{U})
//
// Oracle text:
//
//	Croaking Curse
//	Tap target creature. Create a Cursed Role token attached to it. (Enchanted creature is 1/1.)
var VantressTransmuter = newVantressTransmuter

func newVantressTransmuter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vantress Transmuter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Croaking Curse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:          game.Fixed(1),
							Source:          game.TokenDef(vantressTransmuterToken),
							EntryAttachedTo: opt.Val(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature. Create a Cursed Role token attached to it. (Enchanted creature is 1/1.)
		`,
		}),
	}
}

var vantressTransmuterToken = newVantressTransmuterToken()

func newVantressTransmuterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Cursed Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetPower:     opt.Val(game.PT{Value: 1}),
							SetToughness: opt.Val(game.PT{Value: 1}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has base power and toughness 1/1.
		`,
		},
	}
}
