package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CosmotronicWave is the card definition for Cosmotronic Wave.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Cosmotronic Wave deals 1 damage to each creature your opponents control. Creatures your opponents control can't block this turn.
var CosmotronicWave = newCosmotronicWave

func newCosmotronicWave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cosmotronic Wave",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent})),
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:               game.RuleEffectCantBlock,
									AffectedController: game.ControllerOpponent,
									PermanentTypes:     []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cosmotronic Wave deals 1 damage to each creature your opponents control. Creatures your opponents control can't block this turn.
		`,
		},
	}
}
