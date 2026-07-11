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

// TyvarJubilantBrawler is the card definition for Tyvar, Jubilant Brawler.
//
// Type: Legendary Planeswalker — Tyvar
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	You may activate abilities of creatures you control as though those creatures had haste.
//	+1: Untap up to one target creature.
//	−2: Mill three cards, then you may return a creature card with mana value 2 or less from your graveyard to the battlefield.
var TyvarJubilantBrawler = newTyvarJubilantBrawler

func newTyvarJubilantBrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Tyvar, Jubilant Brawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Tyvar},
			Loyalty:    opt.Val(3),
			StaticAbilities: []game.StaticAbility{
				game.ActivateAbilitiesAsThoughHasteStaticBody,
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Graveyard,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to return to the battlefield",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You may activate abilities of creatures you control as though those creatures had haste.
			+1: Untap up to one target creature.
			−2: Mill three cards, then you may return a creature card with mana value 2 or less from your graveyard to the battlefield.
		`,
		},
	}
}
