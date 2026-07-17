package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheWitchSVanity is the card definition for The Witch's Vanity.
//
// Type: Enchantment — Saga
// Cost: {1}{B}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Destroy target creature an opponent controls with mana value 2 or less.
//	II — Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	III — Create a Wicked Role token attached to target creature you control.
var TheWitchSVanity = newTheWitchSVanity

func newTheWitchSVanity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "The Witch's Vanity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Destroy target creature an opponent controls with mana value 2 or less.",
					Chapters: []int{1},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls with mana value 2 or less",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "II — Create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
					Chapters: []int{2},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(theWitchSVanityToken),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "III — Create a Wicked Role token attached to target creature you control.",
					Chapters: []int{3},
					Content: game.Mode{
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
									Source:          game.TokenDef(theWitchSVanityToken2),
									EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — Destroy target creature an opponent controls with mana value 2 or less.
			II — Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			III — Create a Wicked Role token attached to target creature you control.
		`,
		},
	}
}

var theWitchSVanityToken = newTheWitchSVanityToken()

func newTheWitchSVanityToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

var theWitchSVanityToken2 = newTheWitchSVanityToken2()

func newTheWitchSVanityToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Wicked Role",
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
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
							MatchToZone:   true,
							ToZone:        zone.Graveyard,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1.
			When this Aura is put into a graveyard from the battlefield, each opponent loses 1 life.
		`,
		},
	}
}
