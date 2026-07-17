package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BesottedKnight is the card definition for Besotted Knight // Betroth the Beast.
//
// Type: Creature — Human Knight // Sorcery — Adventure
// Cost: {3}{W} // {W}
// Face: Betroth the Beast — Sorcery — Adventure ({W})
//
// Oracle text:
//
//	Betroth the Beast
//	Create a Royal Role token attached to target creature you control. (Enchanted creature gets +1/+1 and has ward {1}.)
var BesottedKnight = newBesottedKnight

func newBesottedKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Besotted Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Betroth the Beast",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Sorcery},
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
							Source:          game.TokenDef(besottedKnightToken),
							EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a Royal Role token attached to target creature you control. (Enchanted creature gets +1/+1 and has ward {1}.)
		`,
		}),
	}
}

var besottedKnightToken = newBesottedKnightToken()

func newBesottedKnightToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Royal Role",
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
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
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
			Enchanted creature gets +1/+1 and has ward {1}.
			(Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
		`,
		},
	}
}
