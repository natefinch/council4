package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Six is the card definition for Six.
//
// Type: Legendary Creature — Treefolk
// Cost: {2}{G}
//
// Oracle text:
//
//	Reach
//	Whenever Six attacks, mill three cards. You may put a land card from among them into your hand.
//	During your turn, nonland permanent cards in your graveyard have retrace. (You may cast permanent cards from your graveyard by discarding a land card in addition to paying their other costs.)
var Six = newSix

func newSix() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Six",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Treefolk},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectGrantGraveyardCardKeyword,
							AffectedPlayer:                 game.PlayerYou,
							CardSelection:                  game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle}, ExcludedTypes: []types.Card{types.Land}},
							GrantedKeyword:                 game.Retrace,
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:        game.Fixed(3),
									Player:        game.ControllerReference(),
									PublishLinked: game.LinkedKey("milled-cards"),
								},
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Graveyard,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Hand,
									},
									Riders: game.ChooseRiders{
										FromLinked: game.LinkedKey("milled-cards"),
									},
									Prompt: "Choose a card to return to your hand",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			Whenever Six attacks, mill three cards. You may put a land card from among them into your hand.
			During your turn, nonland permanent cards in your graveyard have retrace. (You may cast permanent cards from your graveyard by discarding a land card in addition to paying their other costs.)
		`,
		},
	}
}
