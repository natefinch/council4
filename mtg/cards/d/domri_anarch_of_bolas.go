package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DomriAnarchOfBolas is the card definition for Domri, Anarch of Bolas.
//
// Type: types.Legendary Planeswalker — Domri
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Creatures you control get +1/+0.
//	+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.
//	−2: Target creature you control fights target creature you don't control.
var DomriAnarchOfBolas = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name: "Domri, Anarch of Bolas",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.R,
			cost.G,
		}),
		Colors:     []color.Color{color.Green, color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Planeswalker},
		Subtypes:   []types.Sub{"Domri"},
		Loyalty:    opt.Val(3),
		OracleText: `
			Creatures you control get +1/+0.
			+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.
			−2: Target creature you control fights target creature you don't control.
		`,
		StaticAbilities: []game.StaticAbilityBody{
			{
				Text: `
					Creatures you control get +1/+0.
				`,
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:      game.LayerPowerToughnessModify,
						Selector:   game.EffectSelectorCreaturesYouControl,
						PowerDelta: 1,
					},
				},
			},
		},
		LoyaltyAbilities: []game.LoyaltyAbilityBody{
			{
				Text: `
					+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.
				`,
				LoyaltyCost: 1,
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Choose{
								Choice: game.ResolutionChoice{
									Kind:   game.ResolutionChoiceMana,
									Prompt: "Choose {R} or {G}",
									Colors: []mana.Color{
										mana.R,
										mana.G,
									},
								},
								PublishChoice: game.ChoiceKey("domri-color"),
							},
						},
						{
							Primitive: game.AddMana{
								Amount:     game.Fixed(1),
								ChoiceFrom: game.ChoiceKey("domri-color"),
							},
						},
						{
							Primitive: game.ApplyRule{
								TargetIndex: game.TargetIndexController,
								RuleEffects: []game.RuleEffect{
									{
										Kind:               game.RuleEffectCantBeCountered,
										AffectedController: game.ControllerYou,
										SpellTypes: []types.Card{
											types.Creature,
										},
									},
								},
								Duration: game.DurationThisTurn,
							},
						},
					},
				},
			},
			{
				Text: `
					−2: Target creature you control fights target creature you don't control.
				`,
				LoyaltyCost: -2,
				Content: game.PlainAbilityContent{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerYou,
							},
						},
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you don't control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerNotYou,
							},
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.Fight{},
						},
					},
				},
			},
		},
	},
}
