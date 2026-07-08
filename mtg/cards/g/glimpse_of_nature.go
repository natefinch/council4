package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlimpseOfNature is the card definition for Glimpse of Nature.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Whenever you cast a creature spell this turn, draw a card.
var GlimpseOfNature = newGlimpseOfNature

func newGlimpseOfNature() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Glimpse of Nature",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								EventPattern: opt.Val(game.TriggerPattern{
									Event:         game.EventSpellCast,
									Controller:    game.TriggerControllerYou,
									CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}),
								Window: game.DelayedWindowThisTurn,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.Draw{
												Amount: game.Fixed(1),
												Player: game.ControllerReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Whenever you cast a creature spell this turn, draw a card.
		`,
		},
	}
}
