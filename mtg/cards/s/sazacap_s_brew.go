package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SazacapSBrew is the card definition for Sazacap's Brew.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
//	As an additional cost to cast this spell, discard a card.
//	Target player draws two cards. If the gift was promised, target creature you control gets +2/+0 until end of turn.
var SazacapSBrew = newSazacapSBrew

func newSazacapSBrew() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Sazacap's Brew",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.GiftKeyword{Delivery: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.CreateToken{
										Amount:      game.Fixed(1),
										Source:      game.TokenDef(sazacapSBrewToken),
										Recipient:   opt.Val(game.GiftRecipientReference()),
										EntryTapped: true,
									},
								},
							},
						}.Ability()},
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalDiscard,
					Text:   "discard a card",
					Amount: 1,
					Source: zone.Hand,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						Gate:       game.TargetGateGiftPromised,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(1),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								GiftPromised: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gift a tapped Fish (You may promise an opponent a gift as you cast this spell. If you do, they create a tapped 1/1 blue Fish creature token before its other effects.)
			As an additional cost to cast this spell, discard a card.
			Target player draws two cards. If the gift was promised, target creature you control gets +2/+0 until end of turn.
		`,
		},
	}
}

var sazacapSBrewToken = newSazacapSBrewToken()

func newSazacapSBrewToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fish",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fish},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
