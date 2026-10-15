package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BattleAtTheHelvault is the card definition for Battle at the Helvault.
//
// Type: Enchantment — Saga
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I, II — For each player, exile up to one target non-Saga, nonland permanent that player controls until this Saga leaves the battlefield.
//	III — Create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible.
var BattleAtTheHelvault = newBattleAtTheHelvault

func newBattleAtTheHelvault() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Battle at the Helvault",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this permanent leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchFromZone:   true,
							FromZone:        zone.Battlefield,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves")),
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
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — For each player, exile up to one target non-Saga, nonland permanent that player controls until this Saga leaves the battlefield.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ForEachPlayer{
											Scope:     game.AllPlayersReference(),
											Chooser:   game.ControllerReference(),
											Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludedSubtype: types.Saga},
											Removal:   game.DistributiveRemovalExile,
											LinkedKey: game.LinkedKey("exile-until-leaves"),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(battleAtTheHelvaultToken),
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I, II — For each player, exile up to one target non-Saga, nonland permanent that player controls until this Saga leaves the battlefield.
			III — Create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible.
		`,
		},
	}
}

var battleAtTheHelvaultToken = newBattleAtTheHelvaultToken()

func newBattleAtTheHelvaultToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Avacyn",
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.IndestructibleStaticBody,
			},
		},
	}
}
