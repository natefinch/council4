package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// tibaltsTrickeryMillCountKey names the random mill-count choice that Tibalt's
// Trickery publishes from its "Choose 1, 2, or 3 at random." prelude and that the
// mill instruction consumes through DynamicAmountChosenNumber. The choose and the
// mill must share this one key so the milled amount is exactly the chosen number.
const tibaltsTrickeryMillCountKey = game.ChoiceKey("tibalts-trickery-mill-count")

const (
	tibaltsTrickeryExiledKey = game.LinkedKey("tibalts-trickery-exiled")
	tibaltsTrickeryFoundKey  = game.ResultKey("tibalts-trickery-found")
)

// lowerTibaltsTrickerySequence lowers Tibalt's Trickery's closed six-effect
// [Counter, Mill, Exile, Exile, Cast, Put] shape into the counter, random-number
// choose, dynamic mill, and different-name-nonland IterativeLibraryProcess
// instructions. The parser marks every effect with TibaltsTrickery, records the
// random mill range and prelude span on the head Counter effect, and credits the
// "Choose 1, 2, or 3 at random." prelude; this text-blind lowerer reads only
// those typed fields and the compiled counter target. The mill, exile, cast, and
// put all resolve in the countered spell's controller context, referenced as the
// controller of target stack object 0. The iterative process publishes its
// outputs for separate free-cast and random-bottom instructions. Any shape
// mismatch, extra target, keyword, mode, or uncovered reference fails closed.
func lowerTibaltsTrickerySequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) != 6 ||
		effects[0].Kind != compiler.EffectCounter ||
		effects[1].Kind != compiler.EffectMill ||
		effects[2].Kind != compiler.EffectExile ||
		effects[3].Kind != compiler.EffectExile ||
		effects[4].Kind != compiler.EffectCast ||
		effects[5].Kind != compiler.EffectPut {
		return game.AbilityContent{}, false
	}
	for i := range effects {
		if !effects[i].TibaltsTrickery {
			return game.AbilityContent{}, false
		}
	}
	head := &effects[0]
	if head.TibaltRandomMillMin < 1 || head.TibaltRandomMillMax < head.TibaltRandomMillMin {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if !isExactMandatoryCounterEffect(head, target) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := counterTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !tibaltsTrickerySpansCovered(ctx, effects, head.TibaltPreludeSpan) {
		return game.AbilityContent{}, false
	}
	targetRef := game.TargetStackObjectReference(0)
	controller := game.ObjectControllerReference(targetRef)
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.CounterObject{Object: targetRef}},
			{Primitive: game.Choose{
				Choice: game.ResolutionChoice{
					Kind:      game.ResolutionChoiceNumber,
					MinNumber: head.TibaltRandomMillMin,
					MaxNumber: head.TibaltRandomMillMax,
					AtRandom:  true,
				},
				PublishChoice: tibaltsTrickeryMillCountKey,
			}},
			{Primitive: game.Mill{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountChosenNumber,
					ResultKey: game.ResultKey(tibaltsTrickeryMillCountKey),
				}),
				Player: controller,
			}},
			{Primitive: game.IterativeLibraryProcess{
				Player:            controller,
				Stop:              game.IterativeLibraryStopDifferentNameNonland,
				DifferentNameFrom: targetRef,
				PublishLinked:     tibaltsTrickeryExiledKey,
			}, PublishResult: tibaltsTrickeryFoundKey},
			{
				Primitive: game.CastForFree{
					Player: controller,
					Card: game.CardReference{
						Kind:   game.CardReferenceLinked,
						LinkID: string(tibaltsTrickeryExiledKey),
					},
					Zone: zone.Exile,
				},
				Optional:      true,
				OptionalActor: opt.Val(controller),
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       tibaltsTrickeryFoundKey,
					Succeeded: game.TriTrue,
				}),
			},
			{Primitive: game.PutLinkedExiledCardsInLibrary{
				LinkedKey:   tibaltsTrickeryExiledKey,
				Bottom:      true,
				RandomOrder: true,
			}},
		},
	}.Ability(), true
}

// tibaltsTrickerySpansCovered reports whether every content reference and
// condition falls within one of the six folded effect spans or the credited
// "Choose 1, 2, or 3 at random." prelude span, so no reference or gating
// condition needing its own instruction is silently dropped by the folded
// lowering.
func tibaltsTrickerySpansCovered(ctx contentCtx, effects []compiler.CompiledEffect, prelude shared.Span) bool {
	spans := make([]shared.Span, 0, len(effects)+1)
	for i := range effects {
		spans = append(spans, effects[i].Span)
	}
	spans = append(spans, prelude)
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, spans) {
			return false
		}
	}
	for ci := range ctx.content.Conditions {
		if !spanCovered(ctx.content.Conditions[ci].Span, spans) {
			return false
		}
	}
	return true
}
