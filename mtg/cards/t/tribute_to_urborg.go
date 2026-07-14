package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TributeToUrborg is the card definition for Tribute to Urborg.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Kicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)
//	Target creature gets -2/-2 until end of turn. If this spell was kicked, that creature gets an additional -1/-1 until end of turn for each instant and sorcery card in your graveyard.
var TributeToUrborg = newTributeToUrborg

func newTributeToUrborg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Tribute to Urborg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.U}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
							PowerDelta:     game.Fixed(-2),
							ToughnessDelta: game.Fixed(-2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: -1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							}),
							ToughnessDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: -1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)
			Target creature gets -2/-2 until end of turn. If this spell was kicked, that creature gets an additional -1/-1 until end of turn for each instant and sorcery card in your graveyard.
		`,
		},
	}
}
