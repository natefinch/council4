package m

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

// MosswortBridge is the card definition for Mosswort Bridge.
//
// Type: Land
//
// Oracle text:
//
//	Hideaway 4 (When this land enters, look at the top four cards of your library, exile one face down, then put the rest on the bottom in a random order.)
//	This land enters tapped.
//	{T}: Add {G}.
//	{G}, {T}: You may play the exiled card without paying its mana cost if creatures you control have total power 10 or greater.
var MosswortBridge = newMosswortBridge()

func newMosswortBridge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:  "Mosswort Bridge",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{G}, {T}: You may play the exiled card without paying its mana cost if creatures you control have total power 10 or greater.",
					ManaCost:        opt.Val(cost.Mana{cost.G}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayHideawayCard{},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
											TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 10}),
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
				game.TapManaAbility(mana.G),
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
			{T}: Add {G}.
			{G}, {T}: You may play the exiled card without paying its mana cost if creatures you control have total power 10 or greater.
		`,
		},
	}
}
