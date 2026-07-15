package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TibaltSTrickery is the card definition for Tibalt's Trickery.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, then exiles cards from the top of their library until they exile a nonland card with a different name than that spell. They may cast that card without paying its mana cost. Then they put the exiled cards on the bottom of their library in a random order.
var TibaltSTrickery = newTibaltSTrickery

func newTibaltSTrickery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tibalt's Trickery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
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
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:      game.ResolutionChoiceNumber,
								MinNumber: 1,
								MaxNumber: 3,
								AtRandom:  true,
							},
							PublishChoice: game.ChoiceKey("tibalts-trickery-mill-count"),
						},
					},
					{
						Primitive: game.Mill{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountChosenNumber,
								ResultKey: game.ResultKey("tibalts-trickery-mill-count"),
							}),
							Player: game.ObjectControllerReference(game.TargetStackObjectReference(0)),
						},
					},
					{
						Primitive: game.IterativeLibraryProcess{
							Player:            game.ObjectControllerReference(game.TargetStackObjectReference(0)),
							Stop:              game.IterativeLibraryStopDifferentNameNonland,
							DifferentNameFrom: game.TargetStackObjectReference(0),
							PublishLinked:     game.LinkedKey("tibalts-trickery-exiled"),
						},
						PublishResult: game.ResultKey("tibalts-trickery-found"),
					},
					{
						Primitive: game.CastForFree{
							Player: game.ObjectControllerReference(game.TargetStackObjectReference(0)),
							Zone:   zone.Exile,
							Card:   game.CardReference{Kind: game.CardReferenceLinked, LinkID: "tibalts-trickery-exiled"},
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "tibalts-trickery-found",
							Succeeded: game.TriTrue,
						}),
						Optional:      true,
						OptionalActor: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
					},
					{
						Primitive: game.PutLinkedExiledCardsInLibrary{
							LinkedKey:   game.LinkedKey("tibalts-trickery-exiled"),
							Bottom:      true,
							RandomOrder: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. Choose 1, 2, or 3 at random. Its controller mills that many cards, then exiles cards from the top of their library until they exile a nonland card with a different name than that spell. They may cast that card without paying its mana cost. Then they put the exiled cards on the bottom of their library in a random order.
		`,
		},
	}
}
