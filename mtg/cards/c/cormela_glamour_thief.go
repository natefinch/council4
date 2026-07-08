package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CormelaGlamourThief is the card definition for Cormela, Glamour Thief.
//
// Type: Legendary Creature — Vampire Rogue
// Cost: {1}{U}{B}{R}
//
// Oracle text:
//
//	Haste
//	{1}, {T}: Add {U}{B}{R}. Spend this mana only to cast instant and/or sorcery spells.
//	When Cormela dies, return up to one target instant or sorcery card from your graveyard to your hand.
var CormelaGlamourThief = newCormelaGlamourThief

func newCormelaGlamourThief() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Cormela, Glamour Thief",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastInstantOrSorcerySpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
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
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target instant or sorcery card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Haste
			{1}, {T}: Add {U}{B}{R}. Spend this mana only to cast instant and/or sorcery spells.
			When Cormela dies, return up to one target instant or sorcery card from your graveyard to your hand.
		`,
		},
	}
}
