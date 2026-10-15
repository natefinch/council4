package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReclusiveTaxidermist is the card definition for Reclusive Taxidermist.
//
// Type: Creature — Human Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	This creature gets +3/+2 as long as there are four or more creature cards in your graveyard.
//	{T}: Add one mana of any color.
var ReclusiveTaxidermist = newReclusiveTaxidermist

func newReclusiveTaxidermist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Reclusive Taxidermist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +3/+2 as long as there are four or more creature cards in your graveyard.",
					Condition: opt.Val(game.Condition{
						Text:              "as long as there are four or more creature cards in your graveyard",
						SourceCounterKind: counter.PlusOnePlusOne,
						ControllerGraveyardCardOfTypeCountAtLeast: 4,
						ControllerGraveyardCountCardType:          types.Creature,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			This creature gets +3/+2 as long as there are four or more creature cards in your graveyard.
			{T}: Add one mana of any color.
		`,
		},
	}
}
