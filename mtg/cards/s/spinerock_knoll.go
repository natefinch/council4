package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpinerockKnoll is the card definition for Spinerock Knoll.
//
// Type: Land
//
// Oracle text:
//
//	Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
//	This land enters tapped.
//	{T}: Add {R}.
//	{R}, {T}: You may play the exiled card without paying its mana cost if an opponent was dealt 7 or more damage this turn.
var SpinerockKnoll = newSpinerockKnoll

func newSpinerockKnoll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Spinerock Knoll",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{R}, {T}: You may play the exiled card without paying its mana cost if an opponent was dealt 7 or more damage this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.R}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayHideawayCard{},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentDamageTakenThisTurn, Op: compare.GreaterOrEqual, Value: 7}},
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
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
			{T}: Add {R}.
			{R}, {T}: You may play the exiled card without paying its mana cost if an opponent was dealt 7 or more damage this turn.
		`,
		},
	}
}
