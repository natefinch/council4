package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WisdomOfAges is the card definition for Wisdom of Ages.
//
// Type: Sorcery
// Cost: {4}{U}{U}{U}
//
// Oracle text:
//
//	Return all instant and sorcery cards from your graveyard to your hand. You have no maximum hand size for the rest of the game.
//	Exile Wisdom of Ages.
var WisdomOfAges = newWisdomOfAges

func newWisdomOfAges() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wisdom of Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou},
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectNoMaximumHandSize,
									Duration:       game.DurationPermanent,
									AffectedPlayer: game.PlayerYou,
								},
							},
							Duration: game.DurationPermanent,
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
			Return all instant and sorcery cards from your graveyard to your hand. You have no maximum hand size for the rest of the game.
			Exile Wisdom of Ages.
		`,
		},
	}
}
