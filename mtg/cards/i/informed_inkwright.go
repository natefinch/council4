package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InformedInkwright is the card definition for Informed Inkwright.
//
// Type: Creature — Human Wizard
// Cost: {1}{W}
//
// Oracle text:
//
//	Vigilance
//	Repartee — Whenever you cast an instant or sorcery spell that targets a creature, create a 1/1 white and black Inkling creature token with flying.
var InformedInkwright = newInformedInkwright

func newInformedInkwright() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Informed Inkwright",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Repartee — Whenever you cast an instant or sorcery spell that targets a creature, create a 1/1 white and black Inkling creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventSpellCast,
							Controller:         game.TriggerControllerYou,
							CardSelection:      game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							StackObjectKind:    game.StackSpell,
							SpellTargetAllow:   game.TargetAllowPermanent,
							SpellTargetPattern: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(informedInkwrightToken),
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
			OracleText: `
			Vigilance
			Repartee — Whenever you cast an instant or sorcery spell that targets a creature, create a 1/1 white and black Inkling creature token with flying.
		`,
		},
	}
}

var informedInkwrightToken = newInformedInkwrightToken()

func newInformedInkwrightToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Inkling",
			Colors:    []color.Color{color.White, color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Inkling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
