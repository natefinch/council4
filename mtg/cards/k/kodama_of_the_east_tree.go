package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KodamaOfTheEastTree is the card definition for Kodama of the East Tree.
//
// Type: Legendary Creature — Spirit
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Reach
//	Whenever another permanent you control enters, if it wasn't put onto the battlefield with this ability, you may put a permanent card with equal or lesser mana value from your hand onto the battlefield.
//	Partner (You can have two commanders if both have partner.)
var KodamaOfTheEastTree = newKodamaOfTheEastTree

func newKodamaOfTheEastTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kodama of the East Tree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.PartnerStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:       game.EventPermanentEnteredBattlefield,
							Controller:  game.TriggerControllerYou,
							ExcludeSelf: true,
						},
						InterveningIf: "if it wasn't put onto the battlefield with this ability",
						InterveningIfEventPermanentWasNotPutByThisAbilitySource: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ManaValueLessOrEqualEventPermanent: true},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			Whenever another permanent you control enters, if it wasn't put onto the battlefield with this ability, you may put a permanent card with equal or lesser mana value from your hand onto the battlefield.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
