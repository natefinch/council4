package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MondrakGloryDominus is the card definition for Mondrak, Glory Dominus.
//
// Type: Legendary Creature — Phyrexian Horror
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	If one or more tokens would be created under your control, twice that many of those tokens are created instead.
//	{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures: Put an indestructible counter on Mondrak. ({W/P} can be paid with either {W} or 2 life.)
var MondrakGloryDominus = newMondrakGloryDominus

func newMondrakGloryDominus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mondrak, Glory Dominus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Horror},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures: Put an indestructible counter on Mondrak. ({W/P} can be paid with either {W} or 2 life.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.PhyrexianMana(mana.W),
						cost.PhyrexianMana(mana.W),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice two other artifacts and/or creatures",
							Amount:             2,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							PermanentTypeAlt:   types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Indestructible,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If one or more tokens would be created under your control, twice that many of those tokens are created instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more tokens would be created under your control, twice that many of those tokens are created instead.
			{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures: Put an indestructible counter on Mondrak. ({W/P} can be paid with either {W} or 2 life.)
		`,
		},
	}
}
