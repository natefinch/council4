package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ImperialHellkite is the card definition for Imperial Hellkite.
//
// Type: Creature — Dragon
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Flying
//	Morph {6}{R}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
//	When this creature is turned face up, you may search your library for a Dragon card, reveal it, put it into your hand, then shuffle.
var ImperialHellkite = newImperialHellkite

func newImperialHellkite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Imperial Hellkite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Morph {6}{R}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(6),
								cost.R,
								cost.R,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature is turned face up, you may search your library for a Dragon card, reveal it, put it into your hand, then shuffle.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentTurnedFaceUp,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{SubtypesAny: []types.Sub{types.Dragon}},
												Reveal:      true,
											},
											Amount: game.Fixed(1),
										},
										Optional: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying
			Morph {6}{R}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
			When this creature is turned face up, you may search your library for a Dragon card, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
