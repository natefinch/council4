package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UneshCriosphinxSovereign is the card definition for Unesh, Criosphinx Sovereign.
//
// Type: Legendary Creature — Sphinx
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Sphinx spells you cast cost {2} less to cast.
//	Whenever Unesh or another Sphinx you control enters, reveal the top four cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
var UneshCriosphinxSovereign = newUneshCriosphinxSovereign()

func newUneshCriosphinxSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Unesh, Criosphinx Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sphinx},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Sub("Sphinx")}},
								GenericReduction: 2,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Sphinx")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PileSplit{
									Player:            game.ControllerReference(),
									Amount:            game.Fixed(4),
									SeparatorOpponent: true,
									Kept:              zone.Hand,
									Other:             zone.Graveyard,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Sphinx spells you cast cost {2} less to cast.
			Whenever Unesh or another Sphinx you control enters, reveal the top four cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
		`,
		},
	}
}
