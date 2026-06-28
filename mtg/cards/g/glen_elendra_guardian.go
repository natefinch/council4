package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GlenElendraGuardian is the card definition for Glen Elendra Guardian.
//
// Type: Creature — Faerie Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	Flash
//	Flying
//	This creature enters with a -1/-1 counter on it.
//	{1}{U}, Remove a counter from this creature: Counter target noncreature spell. Its controller draws a card.
var GlenElendraGuardian = newGlenElendraGuardian()

func newGlenElendraGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glen Elendra Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{U}, Remove a counter from this creature: Counter target noncreature spell. Its controller draws a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalRemoveCounter,
							Text:           "Remove a counter from this creature",
							Amount:         1,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target noncreature spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									ExcludedSpellCardTypes: []types.Card{types.Creature},
									StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.TargetStackObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a -1/-1 counter on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 1}),
			},
			OracleText: `
			Flash
			Flying
			This creature enters with a -1/-1 counter on it.
			{1}{U}, Remove a counter from this creature: Counter target noncreature spell. Its controller draws a card.
		`,
		},
	}
}
