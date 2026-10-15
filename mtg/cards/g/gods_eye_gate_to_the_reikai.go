package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GodsEyeGateToTheReikai is the card definition for Gods' Eye, Gate to the Reikai.
//
// Type: Legendary Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	When Gods' Eye is put into a graveyard from the battlefield, create a 1/1 colorless Spirit creature token.
var GodsEyeGateToTheReikai = newGodsEyeGateToTheReikai

func newGodsEyeGateToTheReikai() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Gods' Eye, Gate to the Reikai",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Gods' Eye is put into a graveyard from the battlefield, create a 1/1 colorless Spirit creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchFromZone:   true,
							FromZone:        zone.Battlefield,
							MatchToZone:     true,
							ToZone:          zone.Graveyard,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(godsEyeGateToTheReikaiToken),
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
			{T}: Add {C}.
			When Gods' Eye is put into a graveyard from the battlefield, create a 1/1 colorless Spirit creature token.
		`,
		},
	}
}

var godsEyeGateToTheReikaiToken = newGodsEyeGateToTheReikaiToken()

func newGodsEyeGateToTheReikaiToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
