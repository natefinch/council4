package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RoccoCabarettiCaterer is the card definition for Rocco, Cabaretti Caterer.
//
// Type: Legendary Creature — Elf Druid
// Cost: {X}{R}{G}{W}
//
// Oracle text:
//
//	When Rocco enters, if you cast it, you may search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle.
var RoccoCabarettiCaterer = newRoccoCabarettiCaterer

func newRoccoCabarettiCaterer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Rocco, Cabaretti Caterer",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Druid},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Rocco enters, if you cast it, you may search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if you cast it",
						InterveningIfEventPermanentWasCastByController: true,
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:        zone.Library,
												Destination:       zone.Battlefield,
												Filter:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
												MaxManaValueFromX: true,
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
			When Rocco enters, if you cast it, you may search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
