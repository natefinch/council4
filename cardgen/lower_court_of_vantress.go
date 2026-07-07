package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCourtOfVantressUpkeep lowers Court of Vantress's upkeep trigger: "At the
// beginning of your upkeep, choose up to one other target enchantment or
// artifact. If you're the monarch, you may create a token that's a copy of it.
// If you're not the monarch, you may have this enchantment become a copy of it,
// except it has this ability."
//
// The compiled body is two optional controller-context effects that share one
// up-to-one "other" artifact-or-enchantment target and are gated by mutually
// exclusive monarch designations: a monarch-gated token copy of the chosen
// permanent, and a not-monarch-gated become-a-copy of that same permanent whose
// copy keeps Court of Vantress's own upkeep ability. Lowering emits one shared
// target spec and two "you may" instructions over that target — a CreateToken
// copy gated on ControllerIsMonarch and a BecomeCopy (RetainsThisAbility) gated
// on the negated designation. It fails closed for any shape it does not fully
// model, so the generic optional/ordered routes never silently drop a branch or
// a designation gate. Choosing no target leaves both instructions with no
// referent, so each resolves as a harmless no-op.
func lowerCourtOfVantressUpkeep(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional || !courtOfVantressUpkeepContent(ctx.content) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	create := ctx.content.Effects[0]
	tokenSpec, ok := tokenCopyModifiers(&create, game.TargetPermanentReference(0))
	if !ok {
		return game.AbilityContent{}, false
	}
	isMonarch := opt.Val(game.EffectCondition{
		Condition: opt.Val(game.Condition{ControllerIsMonarch: true}),
	})
	notMonarch := opt.Val(game.EffectCondition{
		Condition: opt.Val(game.Condition{ControllerIsMonarch: true, Negate: true}),
	})
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.CreateToken{
					Amount: game.Fixed(1),
					Source: game.TokenCopyOf(tokenSpec),
				},
				Condition: isMonarch,
				Optional:  true,
			},
			{
				Primitive: game.BecomeCopy{
					Object:             game.TargetPermanentReference(0),
					RetainsThisAbility: true,
				},
				Condition: notMonarch,
				Optional:  true,
			},
		},
	}.Ability(), true
}

// courtOfVantressUpkeepContent reports whether content is exactly Court of
// Vantress's upkeep body: one up-to-one "other" artifact-or-enchantment target,
// the positive and negated controller-monarch designations, and the two
// optional controller-context copy branches (a token copy of the chosen
// permanent gated by "you're the monarch", and a become-a-copy that retains this
// ability gated by "you're not the monarch"), each designation falling inside
// its own branch sentence. lowerCourtOfVantressUpkeep consumes this whole body,
// including the "choose ... target" sentence; the triggered-ability span
// accounting keys on this predicate to also cover that sentence's leading
// "choose" verb, which the generic trigger body does not yet account for. The
// shape is unique to Court of Vantress, so no other card's lowering is affected.
func courtOfVantressUpkeepContent(content compiler.AbilityContent) bool {
	if len(content.Effects) != 2 ||
		len(content.Targets) != 1 ||
		len(content.Conditions) != 2 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return false
	}
	target := content.Targets[0]
	if target.Cardinality.Min != 0 || target.Cardinality.Max != 1 || !target.Selector.Other {
		return false
	}
	if !vantressMonarchCondition(content.Conditions[0], false) ||
		!vantressMonarchCondition(content.Conditions[1], true) {
		return false
	}
	create := content.Effects[0]
	become := content.Effects[1]
	if !vantressTokenCopyBranch(create) || !vantressBecomeCopyBranch(become) {
		return false
	}
	// Each branch is gated by the designation printed inside its own sentence, so
	// the monarch gate must fall within the token-copy effect and the not-monarch
	// gate within the become-a-copy effect.
	return spanContains(create.Span, content.Conditions[0].Span) &&
		spanContains(become.Span, content.Conditions[1].Span)
}

// vantressMonarchCondition reports whether the condition is the controller
// monarch designation with the requested negation (positive "If you're the
// monarch" or negated "If you're not the monarch").
func vantressMonarchCondition(condition compiler.CompiledCondition, negated bool) bool {
	return condition.Predicate == compiler.ConditionPredicateControllerIsMonarch &&
		condition.Negated == negated &&
		!condition.Resolving
}

// vantressTokenCopyBranch reports whether the effect is Court of Vantress's
// optional, controller-context "create a token that's a copy of it" — the
// monarch branch that copies the chosen target permanent (carried as the
// effect's target-bound "it" reference).
func vantressTokenCopyBranch(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCreate &&
		effect.Optional &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		effect.TokenCopyOfReference &&
		!effect.TokenCopyOfTarget &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 1 &&
		effect.References[0].Binding == compiler.ReferenceBindingTarget
}

// vantressBecomeCopyBranch reports whether the effect is Court of Vantress's
// optional, controller-context "have this enchantment become a copy of it,
// except it has this ability" — the not-monarch branch whose copy retains Court
// of Vantress's own upkeep ability.
func vantressBecomeCopyBranch(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectBecomeCopy &&
		effect.Optional &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		effect.BecomeCopyRetainsThisAbility &&
		!effect.BecomeCopyUntilEndOfTurn &&
		len(effect.Targets) == 0
}
