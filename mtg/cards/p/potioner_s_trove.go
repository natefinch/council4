package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PotionerSTrove is the card definition for Potioner's Trove.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add one mana of any color.
//	{T}: You gain 2 life. Activate only if you've cast an instant or sorcery spell this turn.
var PotionerSTrove = newPotionerSTrove

func newPotionerSTrove() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Potioner's Trove",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: You gain 2 life. Activate only if you've cast an instant or sorcery spell this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						}, Window: game.EventHistoryCurrentTurn}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add one mana of any color.
			{T}: You gain 2 life. Activate only if you've cast an instant or sorcery spell this turn.
		`,
		},
	}
}
