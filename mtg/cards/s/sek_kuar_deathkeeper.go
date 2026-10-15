package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SekKuarDeathkeeper is the card definition for Sek'Kuar, Deathkeeper.
//
// Type: Legendary Creature — Orc Shaman
// Cost: {2}{B}{R}{G}
//
// Oracle text:
//
//	Whenever another nontoken creature you control dies, create a 3/1 black and red Graveborn creature token with haste.
var SekKuarDeathkeeper = newSekKuarDeathkeeper

func newSekKuarDeathkeeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Sek'Kuar, Deathkeeper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Orc, types.Shaman},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another nontoken creature you control dies, create a 3/1 black and red Graveborn creature token with haste.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(sekKuarDeathkeeperToken),
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
			Whenever another nontoken creature you control dies, create a 3/1 black and red Graveborn creature token with haste.
		`,
		},
	}
}

var sekKuarDeathkeeperToken = newSekKuarDeathkeeperToken()

func newSekKuarDeathkeeperToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Graveborn",
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Graveborn},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
