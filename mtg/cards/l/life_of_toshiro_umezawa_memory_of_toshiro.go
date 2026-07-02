package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LifeOfToshiroUmezawa is the card definition for Life of Toshiro Umezawa // Memory of Toshiro.
//
// Type: Enchantment — Saga // Enchantment Creature — Human Samurai
// Face: Memory of Toshiro — Enchantment Creature — Human Samurai
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I, II — Choose one —
//	• Target creature gets +2/+2 until end of turn.
//	• Target creature gets -1/-1 until end of turn.
//	• You gain 2 life.
//	III — Exile this Saga, then return it to the battlefield transformed under your control.
var LifeOfToshiroUmezawa = newLifeOfToshiroUmezawa()

func newLifeOfToshiroUmezawa() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Life of Toshiro Umezawa",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text: `I, II — Choose one —
		• Target creature gets +2/+2 until end of turn.
		• Target creature gets -1/-1 until end of turn.
		• You gain 2 life.`,
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Target creature gets +2/+2 until end of turn.",
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
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Target creature gets -1/-1 until end of turn.",
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
										Primitive: game.ModifyPT{
											Object:         game.TargetPermanentReference(0),
											PowerDelta:     game.Fixed(-1),
											ToughnessDelta: game.Fixed(-1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "You gain 2 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Exile this Saga, then return it to the battlefield transformed under your control.",
					Chapters: []int{3},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.SourcePermanentReference(),
									ExileLinkedKey: game.LinkedKey("self-blink"),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source:    game.LinkedBattlefieldSource(game.LinkedKey("self-blink")),
									Recipient: opt.Val(game.ControllerReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter.)
			I, II — Choose one —
			• Target creature gets +2/+2 until end of turn.
			• Target creature gets -1/-1 until end of turn.
			• You gain 2 life.
			III — Exile this Saga, then return it to the battlefield transformed under your control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Memory of Toshiro",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Samurai},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 1 life",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastInstantOrSorcerySpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Pay 1 life: Add {B}. Spend this mana only to cast an instant or sorcery spell.
		`,
		}),
	}
}
