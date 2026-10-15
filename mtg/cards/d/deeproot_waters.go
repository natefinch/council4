package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeeprootWaters is the card definition for Deeproot Waters.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Whenever you cast a Merfolk spell, create a 1/1 blue Merfolk creature token with hexproof. (A creature with hexproof can't be the target of spells or abilities your opponents control.)
var DeeprootWaters = newDeeprootWaters

func newDeeprootWaters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Deeproot Waters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a Merfolk spell, create a 1/1 blue Merfolk creature token with hexproof. (A creature with hexproof can't be the target of spells or abilities your opponents control.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{SubtypesAny: []types.Sub{types.Merfolk}},
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
											Source: game.TokenDef(deeprootWatersToken),
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
			Whenever you cast a Merfolk spell, create a 1/1 blue Merfolk creature token with hexproof. (A creature with hexproof can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}

var deeprootWatersToken = newDeeprootWatersToken()

func newDeeprootWatersToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Merfolk",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
		},
	}
}
