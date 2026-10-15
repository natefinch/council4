package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarshalingCry is the card definition for Marshaling Cry.
//
// Type: Sorcery
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Creatures you control get +1/+1 and gain vigilance until end of turn.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//	Flashback {3}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var MarshalingCry = newMarshalingCry

func newMarshalingCry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Marshaling Cry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.W,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     1,
									ToughnessDelta: 1,
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Vigilance},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +1/+1 and gain vigilance until end of turn.
			Cycling {2} ({2}, Discard this card: Draw a card.)
			Flashback {3}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
