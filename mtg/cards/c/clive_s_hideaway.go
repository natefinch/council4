package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CliveSHideaway is the card definition for Clive's Hideaway.
//
// Type: Land — Town
//
// Oracle text:
//
//	Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
//	{T}: Add {C}.
//	{2}, {T}: You may play the exiled card without paying its mana cost if you control four or more legendary creatures.
var CliveSHideaway = newCliveSHideaway

func newCliveSHideaway() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Clive's Hideaway",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: You may play the exiled card without paying its mana cost if you control four or more legendary creatures.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayHideawayCard{},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
											MinCount:  4,
										}),
									}),
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.HideawayTriggeredAbility(4),
			},
			OracleText: `
			Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
			{T}: Add {C}.
			{2}, {T}: You may play the exiled card without paying its mana cost if you control four or more legendary creatures.
		`,
		},
	}
}
