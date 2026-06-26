package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PathOfAnnihilation is the card definition for Path of Annihilation.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	When this enchantment enters, create two 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
//	Eldrazi you control have "{T}: Add one mana of any color."
//	Whenever you cast a creature spell with mana value 7 or greater, you gain 4 life.
var PathOfAnnihilation = newPathOfAnnihilation()

func newPathOfAnnihilation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Path of Annihilation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Eldrazi")}}),
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
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
									Amount: game.Fixed(2),
									Source: game.TokenDef(pathOfAnnihilationToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 7})},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(4),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			When this enchantment enters, create two 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
			Eldrazi you control have "{T}: Add one mana of any color."
			Whenever you cast a creature spell with mana value 7 or greater, you gain 4 life.
		`,
		},
	}
}

var pathOfAnnihilationToken = newPathOfAnnihilationToken()

func newPathOfAnnihilationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
