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

// OsgoodOperationDoubleTokendce71fa45e0b4c19807e59a60917fe6b is the card definition for Osgood, Operation Double.
//
// Type: Token Creature — Human Alien Shapeshifter
// Cost: {2}{U}{U}
//
// Oracle text:
//   {T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact.
//   Paradox — Whenever you cast a spell from anywhere other than your hand, investigate.
//   (This token's mana cost is {2}{U}{U}.)

// OsgoodOperationDoubleTokendce71fa45e0b4c19807e59a60917fe6b is the card definition for Osgood, Operation Double.
var OsgoodOperationDoubleTokendce71fa45e0b4c19807e59a60917fe6b = newOsgoodOperationDoubleTokendce71fa45e0b4c19807e59a60917fe6b()

func newOsgoodOperationDoubleTokendce71fa45e0b4c19807e59a60917fe6b() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Osgood, Operation Double",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Alien, types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
			{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact.
			Paradox — Whenever you cast a spell from anywhere other than your hand, investigate.
			(This token's mana cost is {2}{U}{U}.)
		`,
		},
	}
}
