package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CraftworkCrusher is the card definition for Craftwork Crusher.
//
// Type: Artifact Creature — Boar Construct
// Cost: {3}{R}{R}{G}{G}
//
// Oracle text:
//
//	Trample
//	When this creature enters, choose two —
//	• This creature deals 4 damage to target creature or planeswalker.
//	• Create a 2/2 colorless Wizard Soldier creature token named Cadet.
//	• Draw a card.
var CraftworkCrusher = newCraftworkCrusher

func newCraftworkCrusher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Craftwork Crusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Boar, types.Construct},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, choose two —\n• This creature deals 4 damage to target creature or planeswalker.\n• Create a 2/2 colorless Wizard Soldier creature token named Cadet.\n• Draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "This creature deals 4 damage to target creature or planeswalker.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature or planeswalker",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(4),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a 2/2 colorless Wizard Soldier creature token named Cadet.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(craftworkCrusherToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Draw a card.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 2,
						MaxModes: 2,
					},
				},
			},
			OracleText: `
			Trample
			When this creature enters, choose two —
			• This creature deals 4 damage to target creature or planeswalker.
			• Create a 2/2 colorless Wizard Soldier creature token named Cadet.
			• Draw a card.
		`,
		},
	}
}

var craftworkCrusherToken = newCraftworkCrusherToken()

func newCraftworkCrusherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cadet",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
