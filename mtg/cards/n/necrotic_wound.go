package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NecroticWound is the card definition for Necrotic Wound.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Undergrowth — Target creature gets -X/-X until end of turn, where X is the number of creature cards in your graveyard. If that creature would die this turn, exile it instead.
var NecroticWound = newNecroticWound

func newNecroticWound() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Necrotic Wound",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
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
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: -1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}),
							ToughnessDelta: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: -1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateReplacement{
							Replacement: &game.ReplacementEffect{
								MatchEvent:    game.EventZoneChanged,
								MatchFromZone: true,
								FromZone:      zone.Battlefield,
								MatchToZone:   true,
								ToZone:        zone.Graveyard,
								ReplaceToZone: zone.Exile,
							},
							Object:   game.TargetPermanentReference(0),
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Undergrowth — Target creature gets -X/-X until end of turn, where X is the number of creature cards in your graveyard. If that creature would die this turn, exile it instead.
		`,
		},
	}
}
