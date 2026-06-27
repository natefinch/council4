package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WorldBreaker is the card definition for World Breaker.
//
// Type: Creature — Eldrazi
// Cost: {6}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	When you cast this spell, exile target artifact, enchantment, or land.
//	Reach
//	{2}{C}, Sacrifice a land: Return this card from your graveyard to your hand. ({C} represents colorless mana.)
var WorldBreaker = newWorldBreaker()

func newWorldBreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "World Breaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.ReachStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{C}, Sacrifice a land: Return this card from your graveyard to your hand. ({C} represents colorless mana.)",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.C}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a land",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact, enchantment, or land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment, types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			When you cast this spell, exile target artifact, enchantment, or land.
			Reach
			{2}{C}, Sacrifice a land: Return this card from your graveyard to your hand. ({C} represents colorless mana.)
		`,
		},
	}
}
