package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WindbriskHeights is the card definition for Windbrisk Heights.
//
// Type: Land
//
// Oracle text:
//
//	Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
//	This land enters tapped.
//	{T}: Add {W}.
//	{W}, {T}: You may play the exiled card without paying its mana cost if you attacked with three or more creatures this turn.
var WindbriskHeights = newWindbriskHeights()

func newWindbriskHeights() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Windbrisk Heights",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{W}, {T}: You may play the exiled card without paying its mana cost if you attacked with three or more creatures this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayHideawayCard{},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
											Event:      game.EventAttackerDeclared,
											Controller: game.TriggerControllerYou,
										}, Window: game.EventHistoryCurrentTurn, MinCount: 3}),
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.HideawayTriggeredAbility(4),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
			This land enters tapped.
			{T}: Add {W}.
			{W}, {T}: You may play the exiled card without paying its mana cost if you attacked with three or more creatures this turn.
		`,
		},
	}
}
