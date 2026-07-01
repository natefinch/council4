package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerThatMuchLifeBackref lowers a "you gain/lose that much life" clause whose
// amount back-references the life change defined by the immediately preceding
// clause in the same ordered sequence. It backs the else branch of a
// kicked-conditional gain-or-lose pair — Sheoldred's Restoration ("… you gain
// life equal to that card's mana value. Otherwise, you lose that much life.") —
// where the "Otherwise" branch loses the same amount the gain branch would have
// gained. The two branches are made mutually exclusive by the sequence's
// otherwise gate; this lowerer only resolves the back-referenced amount.
//
// The amount referent ("that much") compiles to a triggering-life-change dynamic
// amount with no formula of its own, so it carries the preceding life change's
// exact quantity and the result gate that conditions that quantity (a
// reanimation mana-value read is gated on the move reaching the battlefield).
// Copying both keeps the loss reading the same last-known characteristic the
// gain read. The recipient is the spell's controller ("you"); any other context,
// multiplier, formula, target, keyword, condition, or leftover reference fails
// closed so an unsupported back-reference is never silently mislowered.
func lowerThatMuchLifeBackref(
	ctx contentCtx,
	effectIndex int,
	sequence []game.Instruction,
) (game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := &ctx.content.Effects[0]
	if (effect.Kind != compiler.EffectGain && effect.Kind != compiler.EffectLose) ||
		!effect.LifeObject ||
		!effect.Exact ||
		effect.Negated ||
		ctx.optional ||
		effect.Amount.Known ||
		effect.Amount.DynamicKind != compiler.DynamicAmountTriggeringLifeChange ||
		effect.Amount.DynamicForm != compiler.DynamicAmountFormNone ||
		effect.Amount.Multiplier != 0 ||
		effect.Context != parser.EffectContextController ||
		len(effect.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	prior := sequence[effectIndex-1]
	amount, ok := priorLifeChangeAmount(prior.Primitive)
	if !ok {
		return game.AbilityContent{}, false
	}
	player := game.ControllerReference()
	var primitive game.Primitive
	switch effect.Kind {
	case compiler.EffectGain:
		primitive = game.GainLife{Amount: amount, Player: player}
	case compiler.EffectLose:
		primitive = game.LoseLife{Amount: amount, Player: player}
	default:
		panic(fmt.Sprintf(
			"lowerThatMuchLifeBackref: effect.Kind %v is neither EffectGain nor EffectLose despite the earlier kind guard",
			effect.Kind,
		))
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive:  primitive,
			ResultGate: prior.ResultGate,
		}},
	}.Ability(), true
}

// priorLifeChangeAmount returns the quantity of a preceding GainLife or LoseLife
// instruction, the only primitives a "that much life" back-reference reads. It
// fails closed for every other primitive.
func priorLifeChangeAmount(primitive game.Primitive) (game.Quantity, bool) {
	if gain, ok := primitive.(game.GainLife); ok {
		return gain.Amount, true
	}
	if lose, ok := primitive.(game.LoseLife); ok {
		return lose.Amount, true
	}
	return game.Quantity{}, false
}
