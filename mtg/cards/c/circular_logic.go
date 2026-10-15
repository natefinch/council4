package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CircularLogic is the card definition for Circular Logic.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Counter target spell unless its controller pays {1} for each card in your graveyard.
//	Madness {U} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var CircularLogic = newCircularLogic

func newCircularLogic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Circular Logic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Madness {U}",
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{
							Cost: cost.Mana{
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {1} for each card in your graveyard?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost: opt.Val(cost.Mana{
									cost.O(1),
								}),
								ManaCostMultiplier: opt.Val(&game.DynamicAmount{
									Kind:        game.DynamicAmountCountCardsInZone,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
									CardZone:    zone.Graveyard,
									Selection:   &game.Selection{},
								}),
							},
						},
						PublishResult: game.ResultKey("unless-paid"),
					},
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "unless-paid",
							Succeeded: game.TriFalse,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell unless its controller pays {1} for each card in your graveyard.
			Madness {U} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
