package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConduitOfWorlds is the card definition for Conduit of Worlds.
//
// Type: Artifact
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	You may play lands from your graveyard.
//	{T}: Choose target nonland permanent card in your graveyard. If you haven't cast a spell this turn, you may cast that card. If you do, you can't cast additional spells this turn. Activate only as a sorcery.
var ConduitOfWorlds = newConduitOfWorlds

func newConduitOfWorlds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Conduit of Worlds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.PlayLandsFromGraveyardStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Choose target nonland permanent card in your graveyard. If you haven't cast a spell this turn, you may cast that card. If you do, you can't cast additional spells this turn. Activate only as a sorcery.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Timing:          game.SorceryOnly,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent card in your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CastForFree{
									Player:      game.ControllerReference(),
									Zone:        zone.Graveyard,
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									PayManaCost: true,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate: true,
										EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
											Event:      game.EventSpellCast,
											Controller: game.TriggerControllerYou,
										}, Window: game.EventHistoryCurrentTurn}),
									}),
								}),
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:           game.RuleEffectCantCastSpells,
											AffectedPlayer: game.PlayerYou,
										},
									},
									Duration: game.DurationThisTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You may play lands from your graveyard.
			{T}: Choose target nonland permanent card in your graveyard. If you haven't cast a spell this turn, you may cast that card. If you do, you can't cast additional spells this turn. Activate only as a sorcery.
		`,
		},
	}
}
