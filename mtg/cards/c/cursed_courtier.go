package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CursedCourtier is the card definition for Cursed Courtier.
//
// Type: Creature — Human Noble
// Cost: {2}{W}
//
// Oracle text:
//
//	Lifelink
//	When this creature enters, create a Cursed Role token attached to it. (Enchanted creature is 1/1.)
var CursedCourtier = newCursedCourtier

func newCursedCourtier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cursed Courtier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Noble},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(cursedCourtierToken),
									EntryAttachedTo: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Lifelink
			When this creature enters, create a Cursed Role token attached to it. (Enchanted creature is 1/1.)
		`,
		},
	}
}

var cursedCourtierToken = newCursedCourtierToken()

func newCursedCourtierToken() *game.CardDef {
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
