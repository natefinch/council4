package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FerociousWerefox is the card definition for Ferocious Werefox // Guard Change.
//
// Type: Creature — Elf Fox Warrior // Instant — Adventure
// Cost: {3}{G} // {1}{G}
// Face: Guard Change — Instant — Adventure ({1}{G})
//
// Oracle text:
//
//	Trample
var FerociousWerefox = newFerociousWerefox

func newFerociousWerefox() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ferocious Werefox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Fox, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Guard Change",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:          game.Fixed(1),
							Source:          game.TokenDef(ferociousWerefoxToken),
							EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a Monster Role token attached to target creature you control. (Enchanted creature gets +1/+1 and has trample.)
		`,
		}),
	}
}

var ferociousWerefoxToken = newFerociousWerefoxToken()

func newFerociousWerefoxToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Monster Role",
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
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has trample.
		`,
		},
	}
}
