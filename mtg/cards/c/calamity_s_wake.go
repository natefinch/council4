package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CalamitySWake is the card definition for Calamity's Wake.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Exile all graveyards. Players can't cast noncreature spells this turn. Exile Calamity's Wake.
var CalamitySWake = newCalamitySWake

func newCalamitySWake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Calamity's Wake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							PlayerGroup: game.AllPlayersReference(),
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:               game.RuleEffectCantCastSpells,
									Duration:           game.DurationPermanent,
									ExcludedSpellTypes: []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile all graveyards. Players can't cast noncreature spells this turn. Exile Calamity's Wake.
		`,
		},
	}
}
