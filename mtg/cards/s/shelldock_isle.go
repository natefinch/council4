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

// ShelldockIsle is the card definition for Shelldock Isle.
//
// Type: Land
//
// Oracle text:
//
//	Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
//	This land enters tapped.
//	{T}: Add {U}.
//	{U}, {T}: You may play the exiled card without paying its mana cost if a library has twenty or fewer cards in it.
var ShelldockIsle = newShelldockIsle

func newShelldockIsle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Shelldock Isle",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{U}, {T}: You may play the exiled card without paying its mana cost if a library has twenty or fewer cards in it.",
					ManaCost:        opt.Val(cost.Mana{cost.U}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayHideawayCard{},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateMinPlayerLibrarySize, Op: compare.LessOrEqual, Value: 20}},
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
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
			{T}: Add {U}.
			{U}, {T}: You may play the exiled card without paying its mana cost if a library has twenty or fewer cards in it.
		`,
		},
	}
}
