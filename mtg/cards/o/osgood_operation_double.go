package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OsgoodOperationDouble is the card definition for Osgood, Operation Double.
//
// Type: Legendary Creature — Human Alien Shapeshifter
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	When you cast this spell, create a token that's a copy of it, except it isn't legendary.
//	{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact.
//	Paradox — Whenever you cast a spell from anywhere other than your hand, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
var OsgoodOperationDouble = newOsgoodOperationDouble()

func newOsgoodOperationDouble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Osgood, Operation Double",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Alien, types.Shapeshifter},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateArtifact,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:          game.TokenCopySourceObject,
										Object:          game.EventStackObjectReference(),
										SetNotLegendary: true,
									}),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							ExcludeFromZone: true,
							FromZone:        zone.Hand,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Investigate{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When you cast this spell, create a token that's a copy of it, except it isn't legendary.
			{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact.
			Paradox — Whenever you cast a spell from anywhere other than your hand, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
		`,
		},
	}
}
