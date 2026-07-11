package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChampionOfStraySouls is the card definition for Champion of Stray Souls.
//
// Type: Creature — Skeleton Warrior
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature cards from your graveyard to the battlefield.
//	{5}{B}{B}: Put this card from your graveyard on top of your library.
var ChampionOfStraySouls = newChampionOfStraySouls

func newChampionOfStraySouls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Champion of Stray Souls",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature cards from your graveyard to the battlefield.",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice X other creatures",
							AmountFromX:        true,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets:   0,
								MaxTargets:   20,
								Constraint:   "target creature cards from your graveyard",
								Allow:        game.TargetAllowCard,
								TargetZone:   zone.Graveyard,
								Selection:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								CountEqualsX: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 4}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 5}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 6}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 7}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 8}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 9}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 10}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 11}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 12}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 13}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 14}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 15}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 16}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 17}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 18}),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 19}),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{5}{B}{B}: Put this card from your graveyard on top of your library.",
					ManaCost:       opt.Val(cost.Mana{cost.O(5), cost.B, cost.B}),
					ZoneOfFunction: zone.Graveyard,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Library,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature cards from your graveyard to the battlefield.
			{5}{B}{B}: Put this card from your graveyard on top of your library.
		`,
		},
	}
}
