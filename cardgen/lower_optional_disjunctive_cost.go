package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// disjunctiveCostResultKeyA and disjunctiveCostResultKeyB wire the two
// mutually-exclusive optional cost clauses of a "you may sacrifice X or discard
// Y. If you do, REWARD" body to their gated reward copies. Clause A is offered
// first; clause B is offered only when clause A is declined, so at most one
// resolves and the reward fires once.
const (
	disjunctiveCostResultKeyA = game.ResultKey("disjunctive-cost-a")
	disjunctiveCostResultKeyB = game.ResultKey("disjunctive-cost-b")
)

// lowerOptionalDisjunctiveSacrificeDiscard lowers the resolving-optional
// disjunctive sacrifice-or-discard cost body —
//
//	"You may sacrifice <X> or discard <Y>. If you do, <REWARD>."
//
// (K'un-Lun Warrior, Crypt Lurker, Vision of Love, Reckless Detective, Contract
// Hero), where the controller may pay one of two alternative self-costs and, on
// paying either, receives a reward.
//
// The parser compiles the disjunction as two sibling cost effects — a leading
// optional cost effect and a second cost effect joined by EffectConnectionOr —
// followed by one or more and-joined reward effects, all gated on a single
// resolving "if you do" condition (ConditionPredicatePriorInstructionAccepted).
//
// Because the two costs are mutually exclusive (the controller pays at most one),
// the runtime models the choice as two sequential optional instructions: cost A
// is offered first and publishes its result; cost B is offered only when cost A
// was not accepted, gated on cost A's result. The reward instructions are then
// emitted twice — once gated on cost A succeeding and once on cost B succeeding —
// so exactly one copy fires. This duplication is sound only for non-targeted
// rewards (the supported family draws cards and/or pumps the source), so the
// reward clauses must lower to target-free single-instruction modes; any reward
// that owns a target, publishes a result, or is itself gated leaves the body
// unsupported rather than lowered to a wrong shape.
func lowerOptionalDisjunctiveSacrificeDiscard(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if !disjunctiveSacrificeDiscardShape(ctx) {
		return game.AbilityContent{}, false
	}
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)

	costA, ok := lowerDisjunctiveCostClause(cardName, ctx, clauseSyntaxes, 0, true)
	if !ok {
		return game.AbilityContent{}, false
	}
	costB, ok := lowerDisjunctiveCostClause(cardName, ctx, clauseSyntaxes, 1, true)
	if !ok {
		return game.AbilityContent{}, false
	}

	var reward []game.Instruction
	for i := 2; i < len(ctx.content.Effects); i++ {
		instr, ok := lowerDisjunctiveCostClause(cardName, ctx, clauseSyntaxes, i, false)
		if !ok {
			return game.AbilityContent{}, false
		}
		reward = append(reward, instr)
	}

	costA.Optional = true
	costA.PublishResult = disjunctiveCostResultKeyA
	costB.Optional = true
	costB.PublishResult = disjunctiveCostResultKeyB
	costB.ResultGate = opt.Val(game.InstructionResultGate{
		Key:      disjunctiveCostResultKeyA,
		Accepted: game.TriFalse,
	})

	sequence := []game.Instruction{costA, costB}
	sequence = appendGatedReward(sequence, reward, disjunctiveCostResultKeyA)
	sequence = appendGatedReward(sequence, reward, disjunctiveCostResultKeyB)

	return game.Mode{Sequence: sequence}.Ability(), true
}

// appendGatedReward appends a copy of every reward instruction gated on the
// given cost key having succeeded, so the reward fires exactly when that cost was
// paid.
func appendGatedReward(
	sequence []game.Instruction,
	reward []game.Instruction,
	key game.ResultKey,
) []game.Instruction {
	for _, instr := range reward {
		gated := instr
		gated.ResultGate = opt.Val(game.InstructionResultGate{
			Key:       key,
			Succeeded: game.TriTrue,
		})
		sequence = append(sequence, gated)
	}
	return sequence
}

// disjunctiveSacrificeDiscardShape reports whether the ability content is the
// "you may sacrifice X or discard Y. If you do, REWARD" disjunctive cost shape
// lowerOptionalDisjunctiveSacrificeDiscard handles: a leading optional
// controller sacrifice or discard cost, a second cost of the other kind joined
// by EffectConnectionOr, one or more trailing reward effects, and exactly one
// resolving "if you do" condition gating the reward. It fails closed for any
// other shape (modes, keywords, content-level targets, extra conditions, a
// negated or delayed cost, or a reward that is itself an alternative).
func disjunctiveSacrificeDiscardShape(ctx contentCtx) bool {
	content := ctx.content
	if ctx.optional ||
		len(content.Modes) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Effects) < 3 {
		return false
	}
	if !disjunctiveResolvingCondition(content.Conditions) {
		return false
	}
	costA := content.Effects[0]
	costB := content.Effects[1]
	if !isDisjunctiveCostEffect(costA) ||
		!costA.Optional ||
		costA.Connection != parser.EffectConnectionNone {
		return false
	}
	if !isDisjunctiveCostEffect(costB) ||
		costB.Optional ||
		costB.Connection != parser.EffectConnectionOr {
		return false
	}
	// The two costs must be the two distinct kinds (sacrifice and discard), not
	// the same kind twice, so the disjunction is the supported sacrifice-or-
	// discard choice.
	if costA.Kind == costB.Kind {
		return false
	}
	for i := 2; i < len(content.Effects); i++ {
		reward := content.Effects[i]
		if reward.Optional ||
			reward.Negated ||
			reward.DelayedTiming != 0 ||
			reward.Connection == parser.EffectConnectionOr {
			return false
		}
	}
	return true
}

// isDisjunctiveCostEffect reports whether the effect is a controller sacrifice
// or discard self-cost usable as one side of the disjunction: a non-negated,
// non-delayed sacrifice or discard performed by the spell or ability controller.
func isDisjunctiveCostEffect(effect compiler.CompiledEffect) bool {
	if effect.Kind != compiler.EffectSacrifice &&
		effect.Kind != compiler.EffectDiscard {
		return false
	}
	if effect.Negated || effect.DelayedTiming != 0 {
		return false
	}
	return effect.Context == parser.EffectContextController ||
		effect.Context == parser.EffectContextUnknown
}

// disjunctiveResolvingCondition reports whether the conditions are exactly the
// single resolving "if you do" gate
// (ConditionPredicatePriorInstructionAccepted) that the disjunctive cost body's
// reward depends on. The custom lowering realizes the gate through ResultGate
// duplication, so any other or additional condition leaves the body unsupported.
func disjunctiveResolvingCondition(conditions []compiler.CompiledCondition) bool {
	if len(conditions) != 1 {
		return false
	}
	condition := conditions[0]
	return condition.Predicate == compiler.ConditionPredicatePriorInstructionAccepted &&
		!condition.Negated
}

// lowerDisjunctiveCostClause lowers the single effect at index i of the
// disjunctive cost body to one standalone instruction, reusing the shared
// single-effect lowering path. It strips the effect's optionality and
// connection so the effect lowers as an independent mandatory clause; the caller
// re-applies the optional and gating envelope. It fails closed unless the clause
// lowers to exactly one target-free instruction with no result envelope of its
// own, which the duplication-based reward gating requires.
func lowerDisjunctiveCostClause(
	cardName string,
	ctx contentCtx,
	clauseSyntaxes []parser.Ability,
	i int,
	costClause bool,
) (game.Instruction, bool) {
	effect := ctx.content.Effects[i]
	resolved := effect
	resolved.Optional = false
	resolved.Connection = parser.EffectConnectionNone
	resolved.RequiresOrderedLowering = false
	// The "or <cost>" alternative shares its subject with the leading cost and so
	// compiles with an unset (controller) context; normalize it to the controller
	// so the standalone single-effect lowerer recognizes the self-cost.
	if costClause && resolved.Context == parser.EffectContextUnknown {
		resolved.Context = parser.EffectContextController
	}
	// A non-leading "or <cost>" alternative is parsed in the base-verb form
	// ("or discard a card"), so its exact reconstruction does not match the
	// standalone finite-verb form and the compiler leaves it non-exact and marks
	// it an unrecognized sibling. These are artifacts of the disjunction sibling,
	// not a genuinely unrecognized clause: the standalone single-effect lowerer
	// re-validates the effect kind, fixed amount, selector, and controller
	// context and fails closed otherwise, so clear them for a recognized fixed
	// self-cost (a known sacrifice or discard count).
	if costClause &&
		(resolved.Kind == compiler.EffectSacrifice || resolved.Kind == compiler.EffectDiscard) &&
		resolved.Amount.Known &&
		resolved.Amount.Value >= 1 {
		resolved.Exact = true
		resolved.HasUnrecognizedSibling = false
	}

	clauseCtx := contextForEffect(ctx, &resolved)
	clauseCtx.optional = false
	// The resolving "if you do" gate condition is realized by the caller through
	// ResultGate duplication, not as a per-clause condition, so the single-effect
	// lowerer must see no inherited conditions or modes.
	clauseCtx.content.Conditions = nil
	clauseCtx.content.Modes = nil

	clauseSyntax := clauseSyntaxes[i]
	if clauseSyntax.Span != effect.Span {
		if clauseText := joinedTokenText(clauseSyntax.Tokens); clauseText != "" {
			clauseSyntax.Text = upperFirst(clauseText)
		}
	} else {
		clauseSyntax.Text = effect.Text
	}

	content, diagnostic := lowerContent(cardName, clauseCtx, &clauseSyntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes) != 1 {
		return game.Instruction{}, false
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		return game.Instruction{}, false
	}
	instr := mode.Sequence[0]
	if instr.Optional ||
		instr.PublishResult != "" ||
		instr.ResultGate.Exists {
		return game.Instruction{}, false
	}
	return instr, true
}
