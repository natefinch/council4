package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

var Mundungu = cardset.Entry{
	OracleID: "3fb4161d-a5b9-411d-942c-2fdb039e7bde", Name: "Mundungu", New: newMundungu,
	Verification: "mtg/rules/activated_utility_family_test.go: TestMundunguTargetsOnlySpells, TestMundunguCurrentControllerPaysCombinedCostAndClone, TestMundunguFailedPaymentCounters, TestMundunguUncounterableAndDepartedTargets; printing d6e320ca-848b-4743-93f1-ec04ef1ce402; CR 2026-09-25 117.1b, 118.1-118.3c, 602.2-602.2b, 608.2b, 701.6a-b",
}

var SlyInstigator = cardset.Entry{
	OracleID: "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98", Name: "Sly Instigator", New: newSlyInstigator,
	Verification: "mtg/rules/activated_utility_family_test.go: TestSlyInstigatorSourceDepartureAndCopy, TestSlyInstigatorTargetRevalidation, TestSlyInstigatorDurationUsesAbilityController; printing c9ed183a-a78a-4934-9317-9f16b29e69a2; Scryfall rulings 2021-04-16 and 2023-06-30; CR 2026-09-25 602.2-602.2b, 608.2b, 611.2a-c, 701.15a-d",
}

var HalfElfMonk = cardset.Entry{
	OracleID: "00d11662-9d10-4915-b80c-c1ad54b4e71b", Name: "Half-Elf Monk", New: newHalfElfMonk,
	Verification: "mtg/rules/activated_utility_family_test.go: TestHalfElfMonkVigilance, TestHalfElfMonkActivationCostsAndTiming, TestHalfElfMonkTargetLoss; printing f10f1687-fb20-4ad2-bad7-52e4470b35e4; CR 2026-09-25 302.6, 509.1a-b, 602.2-602.2b, 608.2b",
}

func newMundungu() *game.CardDef {
	const text = "{T}: Counter target spell unless its controller pays {1} and 1 life."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Mundungu", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U, cost.B}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Wizard},
			Colors: []color.Color{color.Blue, color.Black},
			Power:  opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield, AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
					}},
					Sequence: []game.Instruction{
						{
							Primitive: game.Pay{Payment: game.ResolutionPayment{
								Prompt:          "Pay {1} and 1 life?",
								Payer:           opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
								AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}},
							}},
							PublishResult: "mundungu-paid",
						},
						{
							Primitive:  game.CounterObject{Object: game.TargetStackObjectReference(0)},
							ResultGate: opt.Val(game.InstructionResultGate{Key: "mundungu-paid", Succeeded: game.TriFalse}),
						},
					},
				}.Ability(),
			}},
		},
	}
}

func newSlyInstigator() *game.CardDef {
	const text = "{U}, {T}: Until your next turn, target creature an opponent controls can't be blocked. Goad that creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sly Instigator", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Wizard},
			Colors: []color.Color{color.Blue},
			Power:  opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield, ManaCost: opt.Val(cost.Mana{cost.U}),
				AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					}},
					Sequence: []game.Instruction{
						{
							Primitive: game.ApplyRule{
								Object:      opt.Val(game.TargetPermanentReference(0)),
								RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBeBlocked}},
								Duration:    game.DurationUntilYourNextTurn,
							},
						},
						{Primitive: game.Goad{Object: game.TargetPermanentReference(0)}},
					},
				}.Ability(),
			}},
		},
	}
}

func newHalfElfMonk() *game.CardDef {
	const text = "Vigilance\nStunning Strike — {1}{W}, {T}: Tap target creature."
	abilityText := "Stunning Strike — {1}{W}, {T}: Tap target creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Half-Elf Monk", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Elf, types.Monk},
			Colors: []color.Color{color.White},
			Power:  opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{game.VigilanceStaticBody},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: abilityText, ZoneOfFunction: zone.Battlefield,
				ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}), AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}},
					Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
				}.Ability(),
			}},
		},
	}
}
