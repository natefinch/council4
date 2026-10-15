package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KyokiSanitySEclipse is the card definition for Kyoki, Sanity's Eclipse.
//
// Type: Legendary Creature — Demon Spirit
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Whenever you cast a Spirit or Arcane spell, target opponent exiles a card from their hand.
var KyokiSanitySEclipse = newKyokiSanitySEclipse

func newKyokiSanitySEclipse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Kyoki, Sanity's Eclipse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon, types.Spirit},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a Spirit or Arcane spell, target opponent exiles a card from their hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{SubtypesAny: []types.Sub{types.Spirit, types.Arcane}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.TargetPlayerReference(0),
											SourceZone: zone.Hand,
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Exile,
											},
											Riders: game.ChooseRiders{
												PublishObjectScoped: true,
											},
											Prompt: "Choose a card to exile",
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
			Whenever you cast a Spirit or Arcane spell, target opponent exiles a card from their hand.
		`,
		},
	}
}
