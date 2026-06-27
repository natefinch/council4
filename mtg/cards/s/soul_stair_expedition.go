package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SoulStairExpedition is the card definition for Soul Stair Expedition.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: Return up to two target creature cards from your graveyard to your hand.
var SoulStairExpedition = newSoulStairExpedition()

func newSoulStairExpedition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Soul Stair Expedition",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three quest counters from this enchantment and sacrifice it: Return up to two target creature cards from your graveyard to your hand.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three quest counters from this enchantment",
							Amount:      3,
							CounterKind: counter.Quest,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "sacrifice it",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target creature cards from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
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
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Quest,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Landfall — Whenever a land you control enters, you may put a quest counter on this enchantment.
			Remove three quest counters from this enchantment and sacrifice it: Return up to two target creature cards from your graveyard to your hand.
		`,
		},
	}
}
