package cardgen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerCreateTokenSpell lowers vanilla creature-token creation: the controller,
// a referenced object's controller, or a single targeted player ("Target
// opponent creates ...") creates a fixed-power/toughness creature token with one
// or two subtypes, up to two colors (or colorless), an optional leading
// artifact/enchantment permanent type, an optional single creature keyword, an
// optional tapped entry, an optional attacking entry ("... token that's tapped
// and attacking"), and an optional explicit Oracle name ("... creature token
// named <Name>"). The token count may be a fixed number, the spell's variable X,
// or a recognized rules-derived dynamic count ("for each <X>", "number of …
// equal to <X>", "where X is <X>"). Richer token shapes (a blocking entry,
// quoted abilities, multiple keywords, modifiers) and unrepresentable dynamic
// counts fail closed pending follow-up work under the token-creation epic.
func lowerCreateTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerCreateTokenSpellLinked(ctx, "")
}

// lowerCreateTokenSpellLinked lowers a token-creation effect, optionally
// publishing the created token(s) under publishLinked so a following clause
// ("That token gains <keyword> until end of turn.") can reference them. A blank
// publishLinked leaves the token unpublished, the ordinary standalone form. The
// copy- and choice-token variants do not thread the link key; callers that need a
// published token must restrict themselves to the synthesized token forms.
func lowerCreateTokenSpellLinked(ctx contentCtx, publishLinked game.LinkedKey) (game.AbilityContent, *shared.Diagnostic) {
	// lowerCreateTokenSpellLinked is reached only through lowerImmediateSingleEffectSpell's
	// EffectCreate arm (the len==1 gate at lower_spell.go:297) or via contextForEffect
	// (lower_remap.go), which narrows content to exactly one effect; a different count is a
	// dispatch bug, not an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateTokenSpellLinked: reached with %d effects; the EffectCreate dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.TokenCopyOfTarget {
		return lowerCreateCopyTokenSpell(ctx)
	}
	if effect.TokenCopyOfReference {
		return lowerCreateCopyTokenReferenceSpell(ctx)
	}
	if effect.TokenCopyOfAttached {
		return lowerCreateCopyTokenAttachedSpell(ctx)
	}
	if effect.TokenCopyOfTriggeringSet {
		return lowerCreateCopyTokenTriggeringSetSpell(ctx)
	}
	if effect.TokenCopyOfForEach {
		return lowerCreateCopyTokenForEachSpell(ctx)
	}
	controllerRecipient := effect.Context == parser.EffectContextController
	referencedRecipient := effect.Context == parser.EffectContextReferencedObjectController
	targetRecipient := effect.Context == parser.EffectContextTarget
	extraKeywords, keywordsOK := tokenContentKeywords(ctx.content)
	if group, ok := createTokenRecipientGroup(effect.Context); ok {
		return lowerCreateTokenGroupRecipient(ctx, &effect, group, publishLinked, extraKeywords, keywordsOK)
	}
	expectedTargets := 0
	if targetRecipient {
		expectedTargets = 1
	}
	// Every caller guarantees the sole effect is an EffectCreate: lower_spell.go:1820
	// dispatches here only for `case compiler.EffectCreate`, and the sequence callers
	// (lower_create_token_counter.go:32, lower_create_token_attach.go:32,
	// lower_spell_sequence.go:1083, lower_instead_token_count.go:36) each guard
	// createEffect.Kind == EffectCreate before threading it through contextForEffect. A
	// different kind is a dispatch bug, not an unsupported card.
	if effect.Kind != compiler.EffectCreate {
		panic(fmt.Sprintf("lowerCreateTokenSpellLinked: reached with effect kind %v; every caller guarantees EffectCreate", effect.Kind))
	}
	if !effect.Exact ||
		(!controllerRecipient && !referencedRecipient && !targetRecipient) ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		!createTokenDurationOK(effect.Duration) ||
		len(ctx.content.Targets) != expectedTargets ||
		len(ctx.content.Conditions) != 0 ||
		!keywordsOK ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	var recipient opt.V[game.PlayerReference]
	var targets []game.TargetSpec
	var amountObject game.ObjectReference
	amountReferencesObject := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount ||
		effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower
	switch {
	case controllerRecipient:
		if amountReferencesObject {
			if len(ctx.content.References) != 1 {
				return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
			}
			object, ok := lowerObjectReference(ctx.content.References[0],
				referenceLoweringContext{AllowSource: true, AllowEvent: true})
			if !ok {
				return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
			}
			amountObject = object
		} else if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
	case referencedRecipient:
		if len(ctx.content.References) != 1 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		recipient = opt.Val(game.ObjectControllerReference(object))
	case targetRecipient:
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		spec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		targets = []game.TargetSpec{spec}
		recipient = opt.Val(game.TargetPlayerReference(0))
	default:
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	if effect.TokenChoice {
		return lowerCreateNamedTokenChoiceSpell(ctx, &effect, recipient, targets)
	}
	if len(effect.AdditionalTokens) > 0 {
		return lowerMultiTokenCreate(ctx, &effect, recipient, controllerRecipient, publishLinked, extraKeywords)
	}
	def, ok := synthesizeCreatureTokenDef(&effect, extraKeywords)
	if !ok && len(extraKeywords) == 0 {
		def, ok = synthesizeNamedArtifactTokenDef(&effect)
	}
	if !ok && len(extraKeywords) == 0 {
		def, ok = synthesizePredefinedTokenDef(&effect)
	}
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, dynamicSize, ok := createTokenAmountAndSize(ctx, &effect, amountObject)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:         amount,
				Source:         game.TokenDef(def),
				Recipient:      recipient,
				EntryTapped:    effect.Selector.Tapped,
				EntryAttacking: effect.Selector.Attacking,
				Power:          dynamicSize,
				Toughness:      dynamicSize,
				PublishLinked:  publishLinked,
			},
		}},
	}.Ability(), nil
}

// lowerMultiTokenCreate lowers a multi-token create effect ("Create a 1/1 green
// Snake creature token, a 2/2 green Wolf creature token, and a 3/3 green
// Elephant creature token.") to one Mode whose sequence creates each token in
// source order. The effect's own token fields describe the first token; each
// AdditionalTokens entry describes one of the rest. Every token must be a fixed
// power/toughness creature token the single-token path already synthesizes, the
// recipient must be the controller, and the clause must not be a linked or
// keyword-content create; any other shape fails closed. Each token enters once
// (the per-token "a"/"an" article), so every instruction creates a single token.
func lowerMultiTokenCreate(ctx contentCtx, effect *compiler.CompiledEffect, recipient opt.V[game.PlayerReference], controllerRecipient bool, publishLinked game.LinkedKey, extraKeywords []parser.KeywordKind) (game.AbilityContent, *shared.Diagnostic) {
	if !controllerRecipient ||
		publishLinked != "" ||
		len(extraKeywords) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	specs := make([]*compiler.CompiledEffect, 0, 1+len(effect.AdditionalTokens))
	specs = append(specs, effect)
	for i := range effect.AdditionalTokens {
		specs = append(specs, &effect.AdditionalTokens[i])
	}
	sequence := make([]game.Instruction, 0, len(specs))
	seen := make(map[string]*game.CardDef, len(specs))
	for _, spec := range specs {
		if spec.TokenPTVariableX || spec.TokenChoice || spec.TokenGrantedAbility != nil {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		def, ok := synthesizeCreatureTokenDef(spec, spec.TokenKeywords)
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		// tokenDefKey now covers every identity-bearing field a synthesized token
		// carries (including Supertypes and StaticAbilities), so two tokens that
		// differ only in their abilities get distinct keys and distinct vars
		// (Wurmcoil Engine's deathtouch/lifelink Wurms). This guard remains a
		// defensive net: if two tokens ever share a key yet are not identical,
		// fail closed so a multi-token card never renders a wrong token. Tokens
		// that share a key and are fully identical reuse one var correctly.
		if prior, ok := seen[tokenDefKey(def)]; ok && !reflect.DeepEqual(prior, def) {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		seen[tokenDefKey(def)] = def
		sequence = append(sequence, game.Instruction{
			Primitive: game.CreateToken{
				Amount:         game.Fixed(1),
				Source:         game.TokenDef(def),
				Recipient:      recipient,
				EntryTapped:    spec.Selector.Tapped,
				EntryAttacking: spec.Selector.Attacking,
			},
		})
	}
	return game.Mode{Sequence: sequence}.Ability(), nil
}

// tokenPTDynamicQuantity maps a variable "X/X" token's bound dynamic-amount kind
// onto a runtime quantity the create handler evaluates once at creation. It
// fails closed for kinds with no token-sizing representation.
func tokenPTDynamicQuantity(kind parser.EffectDynamicAmountKind) (game.Quantity, bool) {
	switch kind {
	case parser.EffectDynamicAmountLifeGainedThisTurn:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn}), true
	default:
		return game.Quantity{}, false
	}
}

// createTokenAmountAndSize resolves a recognized create-token effect's token
// count together with the optional dynamic power/toughness of a variable "X/X"
// token. A fixed-power/toughness token carries no dynamic size, so its count
// comes straight from createTokenAmount and the size quantity stays empty. A
// variable "X/X" token threads through variableTokenSize, which separates the
// printed-X size from the number of tokens created. A single printed X drives
// both the token's power and toughness, so the one returned size is shared by
// both.
func createTokenAmountAndSize(ctx contentCtx, effect *compiler.CompiledEffect, amountObject game.ObjectReference) (game.Quantity, opt.V[game.Quantity], bool) {
	if !effect.TokenPTVariableX {
		amount, ok := createTokenAmount(ctx, effect, amountObject)
		return amount, opt.V[game.Quantity]{}, ok
	}
	size, count, ok := variableTokenSize(ctx, effect, amountObject)
	if !ok {
		return game.Quantity{}, opt.V[game.Quantity]{}, false
	}
	return count, opt.Val(size), true
}

// variableTokenSize resolves a variable "X/X" token's printed size and the
// number of such tokens created. A single printed X drives both the token's
// power and toughness, so the returned size is shared by both.
//
// Three shapes are recognized. A clause that binds X to a cost amount ("...
// where X is the life paid as this entered") carries that binding in
// TokenPTDynamic, and its count is the ordinary article/number amount. A "create
// an X/X ... where X is <dynamic>" clause routes the size dynamic onto the
// effect's amount during parsing (the trailing "where X is" clause overrides the
// singular article), so the amount is the size and exactly one token is created.
// A fixed-count clause ("create two X/X ... tokens") whose X has no separate
// definition takes its size from the spell's own X. Every other shape, including
// a variable-count clause with no size binding, fails closed.
func variableTokenSize(ctx contentCtx, effect *compiler.CompiledEffect, amountObject game.ObjectReference) (size game.Quantity, count game.Quantity, ok bool) {
	if effect.TokenPTDynamic != parser.EffectDynamicAmountNone {
		quantity, ok := tokenPTDynamicQuantity(effect.TokenPTDynamic)
		if !ok {
			return game.Quantity{}, game.Quantity{}, false
		}
		amount, ok := createTokenAmount(ctx, effect, amountObject)
		if !ok {
			return game.Quantity{}, game.Quantity{}, false
		}
		return quantity, amount, true
	}
	if effect.Amount.DynamicForm == compiler.DynamicAmountWhereX &&
		effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		quantity, ok := createTokenAmount(ctx, effect, amountObject)
		if !ok {
			return game.Quantity{}, game.Quantity{}, false
		}
		return quantity, game.Fixed(1), true
	}
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return game.Quantity{}, game.Quantity{}, false
		}
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), game.Fixed(effect.Amount.Value), true
	}
	return game.Quantity{}, game.Quantity{}, false
}

// lowerCreateNamedTokenChoiceSpell lowers an N-way (N >= 2) choice among
// predefined artifact tokens ("Create a X token or a Y token." and "Create your
// choice of a X token, a Y token, or a Z token.") to a choose-one modal ability:
// one mode per predefined artifact-token alternative, each creating a single
// token for the shared recipient. Every alternative must be a predefined
// artifact token the runtime already models. The target-recipient form is not
// lowered here because modal content cannot carry per-mode targets; it fails
// closed. Any non-predefined alternative, color, keyword, or count other than
// one also fails closed.
func lowerCreateNamedTokenChoiceSpell(ctx contentCtx, effect *compiler.CompiledEffect, recipient opt.V[game.PlayerReference], targets []game.TargetSpec) (game.AbilityContent, *shared.Diagnostic) {
	subtypes := effect.Selector.SubtypesAny()
	if len(targets) != 0 ||
		len(subtypes) < 2 ||
		len(effect.Selector.ColorsAny()) != 0 ||
		effect.Selector.Keyword != parser.KeywordUnknown ||
		effect.Selector.Tapped ||
		effect.TokenPTKnown {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	modes := make([]game.Mode, 0, len(subtypes))
	for _, sub := range subtypes {
		def, ok := namedArtifactTokenDef(sub)
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		modes = append(modes, game.Mode{
			Text: "Create a " + string(sub) + " token.",
			Sequence: []game.Instruction{{
				Primitive: game.CreateToken{
					Amount:    amount,
					Source:    game.TokenDef(def),
					Recipient: recipient,
				},
			}},
		})
	}
	return game.AbilityContent{
		Modes:    modes,
		MinModes: 1,
		MaxModes: 1,
	}, nil
}

// createTokenDurationOK reports whether a recognized exact create-token effect's
// compiled duration is acceptable. Creating a token is instantaneous and the
// token persists, so a create-token clause never carries its own duration. A
// recognized exact create reconstructs to "Create <spec>." with no duration
// words, which proves any non-None duration is spurious metadata that leaked
// from a sibling clause in the same sentence — for example the "this turn" of an
// intervening "if you attacked this turn" trigger condition. Only that
// turn-scoped leak is tolerated; an "until end of turn"/"until your next turn"
// value cannot leak from such a clause (its "until" wording would break create
// exactness) and so stays fail-closed.
func createTokenDurationOK(duration compiler.DurationKind) bool {
	return duration == compiler.DurationNone || duration == compiler.DurationThisTurn
}

// createTokenAmount resolves a recognized create-token effect's token count. A
// fixed literal lowers to that count; the spell's variable X lowers to the
// runtime X amount; and every recognized rules-derived count ("for each <X>",
// "equal to <X>", "where X is <X>") lowers through the shared dynamic-amount
// lowerer. A "the number of <kind> counters on it" count reads the source
// permanent's counters (last-known information once it has died), as for
// Chasm Skulker's death-triggered Squid tokens. Source-power counts and any
// unrepresented dynamic kind fail closed.
func createTokenAmount(ctx contentCtx, effect *compiler.CompiledEffect, amountObject game.ObjectReference) (game.Quantity, bool) {
	switch {
	case effect.Amount.Known:
		if effect.Amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(effect.Amount.Value), true
	case effect.Amount.VariableX:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	case effect.Amount.DynamicKind == compiler.DynamicAmountTriggeringCombatDamage:
		// The parser pins a created "that many" count to the combat-damage
		// triggering-event kind without seeing which event actually fired. The
		// generic resolver reads the enclosing trigger's measured quantity —
		// combat damage dealt, life gained or lost, counters added, or cards
		// drawn or discarded — so "Whenever you put one or more -1/-1 counters on
		// a creature, create that many ... tokens." and the discard- and
		// damage-triggered forms all resolve to their own event's count, while a
		// trigger that publishes no such quantity stays closed.
		dynamic, ok := lowerTriggeringEventQuantity(ctx, effect.Amount)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		object := game.ObjectReference{}
		if effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount ||
			effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			object = amountObject
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, object)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}

// lowerCreateCopyTokenSpell lowers "Create a token that's a copy of <target>[,
// except <it> isn't legendary][. That token gains <keyword>.]" to a CreateToken
// whose source copies the lone target object, applying any copy modifiers. The
// runtime already supports object-copy tokens (TokenCopySourceObject); only a
// single fixed target, controller recipient, and supported copy modifiers are
// accepted here.
func lowerCreateCopyTokenSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's TokenCopyOfTarget branch, which
	// runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateCopyTokenSpell: reached with %d effects; the copy-token dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 1 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, game.TargetPermanentReference(0))
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, &effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:      amount,
				Source:      game.TokenCopyOf(spec),
				EntryTapped: effect.TokenCopyEntersTapped,
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenReferenceSpell lowers "Create a token that's a copy of
// <reference>[, except <it> isn't legendary][. That token gains <keyword>.]"
// (e.g. "... a copy of this creature") to a CreateToken whose source copies the
// object named by the effect's leading explicit reference. The leading reference
// binds the source permanent; any trailing references are the "except" / "that
// token" pronouns of the recognized copy modifiers. Only a controller recipient
// with supported copy modifiers is accepted here.
func lowerCreateCopyTokenReferenceSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's TokenCopyOfReference branch, which
	// runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateCopyTokenReferenceSpell: reached with %d effects; the copy-token dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) > 1 ||
		len(ctx.content.References) < 1 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References[1:]) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	// "Choose target <permanent>. Create a token that's a copy of it." (Yenna,
	// Redtooth Regent) chooses the copied permanent with a separate targeting
	// sentence, so the "it" copy source binds an ability-level target. Emit that
	// target's spec and require the leading reference to bind it; the copy of an
	// inline "target <permanent>" instead goes through lowerCreateCopyTokenSpell.
	var targets []game.TargetSpec
	if len(ctx.content.Targets) == 1 {
		if !referencesBindTo(ctx.content.References[:1], compiler.ReferenceBindingTarget, 0) {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
		}
		targets = []game.TargetSpec{targetSpec}
	}
	object, ok := lowerObjectReference(
		ctx.content.References[0],
		referenceLoweringContext{AllowSource: true, AllowTarget: true, AllowEvent: true},
	)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, object)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, &effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:      amount,
				Source:      game.TokenCopyOf(spec),
				EntryTapped: effect.TokenCopyEntersTapped,
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenAttachedSpell lowers "Create a token that's a copy of
// equipped/enchanted creature[, except <it> isn't legendary][. That token gains
// <keyword>.]" to a CreateToken whose source copies the permanent the source
// Equipment or Aura is attached to. The runtime resolves the attached permanent
// at resolution; only a controller recipient with supported copy modifiers is
// accepted here.
func lowerCreateCopyTokenAttachedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's TokenCopyOfAttached branch, which
	// runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateCopyTokenAttachedSpell: reached with %d effects; the copy-token dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyModifiers(&effect, game.SourceAttachedPermanentReference())
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, &effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:      amount,
				Source:      game.TokenCopyOf(spec),
				EntryTapped: effect.TokenCopyEntersTapped,
			},
		}},
	}.Ability(), nil
}

// lowerCreateCopyTokenTriggeringSetSpell lowers "Create a token that's a copy of
// one of them[, except <it> isn't legendary]." (Twilight Diviner) to a
// CreateToken whose source is a controller-chosen member of the resolving
// ability's triggering event batch. The "them"/"they" pronoun is a benign
// reference naming that batch; only a controller recipient with supported copy
// modifiers (legendary drop, tapped entry) is accepted here.
func lowerCreateCopyTokenTriggeringSetSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's TokenCopyOfTriggeringSet branch,
	// which runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateCopyTokenTriggeringSetSpell: reached with %d effects; the copy-token dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyTriggeringSetModifiers(&effect)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	amount, ok := createTokenAmount(ctx, &effect, game.ObjectReference{})
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:      amount,
				Source:      game.TokenCopyOf(spec),
				EntryTapped: effect.TokenCopyEntersTapped,
			},
		}},
	}.Ability(), nil
}

// tokenCopyTriggeringSetModifiers builds the runtime copy spec for a "copy of one
// of them" create over the controller-chosen triggering-batch member, applying
// the "except <it> isn't legendary" supertype drop and any folded "that token
// gains <keyword>" rider keywords. It fails closed when a granted keyword has no
// reusable runtime static form.
func tokenCopyTriggeringSetModifiers(effect *compiler.CompiledEffect) (game.TokenCopySpec, bool) {
	spec := game.TokenCopySpec{
		Source:          game.TokenCopySourceChosenFromTriggerBatch,
		SetNotLegendary: effect.TokenCopyDropLegendary,
	}
	for _, kind := range effect.TokenCopyGrantKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return game.TokenCopySpec{}, false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	return spec, true
}

// lowerCreateCopyTokenForEachSpell lowers a per-each copy-token create whose
// copy source is each member of a controlled battlefield group ("For each token
// you control, create a token that's a copy of that permanent." — Second
// Harvest) to a CreateToken whose source iterates the group, copying each
// matched permanent in turn. The "that permanent" reference is a benign
// per-iteration pronoun the runtime resolves member-by-member; only a controller
// recipient over a controlled group with supported copy modifiers is accepted.
func lowerCreateCopyTokenForEachSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	// Reached only from lowerCreateTokenSpellLinked's TokenCopyOfForEach branch, which
	// runs on its single-effect content, so a different count is a dispatch bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCreateCopyTokenForEachSpell: reached with %d effects; the copy-token dispatch is single-effect", len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextController ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		!tokenCopyAuxiliaryReferencesOK(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != len(effect.TokenCopyGrantKeywords) ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	selection, ok := copyForEachGroupSelection(effect.TokenCopyForEachGroup)
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	spec, ok := tokenCopyForEachModifiers(&effect, game.BattlefieldGroup(selection))
	if !ok {
		return game.AbilityContent{}, unsupportedTokenCreationDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:      game.Fixed(1),
				Source:      game.TokenCopyOf(spec),
				EntryTapped: effect.TokenCopyEntersTapped,
			},
		}},
	}.Ability(), nil
}

// copyForEachGroupSelection lowers the controlled battlefield group iterated by a
// per-each copy-token create to a runtime Selection. It requires a "you control"
// controller scope and reuses massGroupSelection for the type/color/keyword
// filters. A bare "token you control" carries no card type, so its token filter
// is the constraint: synthesize a permanent kind so massGroupSelection accepts
// it, then restore the token/nontoken filter the shared lowering does not model.
func copyForEachGroupSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Controller != compiler.ControllerYou {
		return game.Selection{}, false
	}
	adjusted := selector
	if selector.Kind == compiler.SelectorUnknown && (selector.TokenOnly || selector.NonToken) {
		adjusted.Kind = compiler.SelectorPermanent
	}
	return SelectionForSelectorMasked(adjusted, copyForEachGroupSelectionMask)
}

// copyForEachGroupSelectionMask honors the per-object token qualifier a
// copy-for-each iteration can scope to ("for each token you control") while
// dropping the historic, excluded-subtype, and source-relative-power dimensions
// the iterated battlefield group never carries.
var copyForEachGroupSelectionMask = SelectionMask{}.Ignoring(
	DimHistoric,
	DimExcludedSubtype,
	DimPowerVsSource,
).Rejecting(
	DimRequiredName,
)

// tokenCopyForEachModifiers builds the runtime copy spec for a per-each copy
// over the iterated group, applying the "except <it> isn't legendary" supertype
// drop and any folded "that token gains <keyword>" rider keywords. It fails
// closed when a granted keyword has no reusable runtime static form.
func tokenCopyForEachModifiers(effect *compiler.CompiledEffect, group game.GroupReference) (game.TokenCopySpec, bool) {
	spec := game.TokenCopySpec{
		Source:          game.TokenCopySourceEachInGroup,
		Group:           game.GroupRef(group),
		SetNotLegendary: effect.TokenCopyDropLegendary,
	}
	for _, kind := range effect.TokenCopyGrantKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return game.TokenCopySpec{}, false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	return spec, true
}

// tokenCopyModifiers builds the runtime copy spec for a copy-token effect over
// the given source object, applying the "except <it> isn't legendary" supertype
// drop and any folded "that token gains <keyword>" rider keywords. It fails
// closed when a granted keyword has no reusable runtime static form.
func tokenCopyModifiers(effect *compiler.CompiledEffect, object game.ObjectReference) (game.TokenCopySpec, bool) {
	spec := game.TokenCopySpec{
		Source:          game.TokenCopySourceObject,
		Object:          object,
		SetNotLegendary: effect.TokenCopyDropLegendary,
	}
	for _, kind := range effect.TokenCopyGrantKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return game.TokenCopySpec{}, false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	if !applyCopyTokenOverride(&spec, effect) {
		return game.TokenCopySpec{}, false
	}
	return spec, true
}

// tokenCopyAuxiliaryReferencesOK reports whether every reference is a benign
// pronoun introduced by a copy modifier ("except it isn't legendary", "that
// token gains ..."), so the copy-token lowering can tolerate them without
// treating them as additional copy sources.
func tokenCopyAuxiliaryReferencesOK(references []compiler.CompiledReference) bool {
	for i := range references {
		switch references[i].Kind {
		case compiler.ReferencePronoun, compiler.ReferenceThatObject:
		default:
			return false
		}
	}
	return true
}

// synthesizeCreatureTokenDef builds a token CardDef from a recognized create
// effect: a creature with one or two subtypes, up to two colors (or colorless),
// an optional leading artifact/enchantment permanent type, a fixed
// power/toughness, and zero or more creature keywords. The leading "with
// <keyword>" selector keyword is carried on the effect selector; any additional
// conjoined keywords ("... and reach") arrive in extraKeywords. Each keyword
// becomes one static ability in Oracle order. The token's name is its explicit
// Oracle name when one is printed ("... token named <Name>"); otherwise it is
// the joined subtypes, matching paper tokens.
func synthesizeCreatureTokenDef(effect *compiler.CompiledEffect, extraKeywords []parser.KeywordKind) (*game.CardDef, bool) {
	if !effect.TokenPTKnown && !effect.TokenPTVariableX {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) < 1 || len(subtypes) > 2 {
		return nil, false
	}
	colors := effect.Selector.ColorsAny()
	if len(colors) > 5 {
		return nil, false
	}
	supertypes := effect.Selector.Supertypes()
	for _, supertype := range supertypes {
		if supertype != types.Legendary {
			return nil, false
		}
	}
	cardTypes, ok := creatureTokenCardTypes(effect.Selector)
	if !ok {
		return nil, false
	}
	names := make([]string, 0, len(subtypes))
	for _, sub := range subtypes {
		names = append(names, string(sub))
	}
	name := strings.Join(names, " ")
	if effect.TokenName != "" {
		name = effect.TokenName
	}
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:       name,
			Colors:     slices.Clone(colors),
			Types:      cardTypes,
			Subtypes:   slices.Clone(subtypes),
			Supertypes: slices.Clone(supertypes),
		},
	}
	// A fixed-power/toughness token carries its printed power and toughness on
	// the definition. A variable "X/X" token leaves them unset here; the create
	// instruction sizes the token from a dynamic amount at creation time.
	if effect.TokenPTKnown {
		def.Power = opt.Val(game.PT{Value: effect.TokenPower})
		def.Toughness = opt.Val(game.PT{Value: effect.TokenToughness})
	}
	keywords := make([]parser.KeywordKind, 0, 1+len(extraKeywords))
	if effect.Selector.Keyword != parser.KeywordUnknown {
		keywords = append(keywords, effect.Selector.Keyword)
	}
	keywords = append(keywords, extraKeywords...)
	for _, keyword := range keywords {
		if keyword == parser.KeywordToxic {
			body, ok := toxicTokenStaticBody(effect.TokenToxic)
			if !ok {
				return nil, false
			}
			def.StaticAbilities = append(def.StaticAbilities, body)
			continue
		}
		if keyword == parser.KeywordDecayed {
			def.StaticAbilities = append(def.StaticAbilities, game.CantBlockStaticBody)
			def.TriggeredAbilities = append(def.TriggeredAbilities, decayedSacrificeTrigger())
			continue
		}
		static, ok := keywordStaticBodies[keyword]
		if !ok {
			return nil, false
		}
		def.StaticAbilities = append(def.StaticAbilities, static.Body)
	}
	if effect.TokenGrantedAbility != nil {
		if !attachTokenGrantedAbility(def, effect.TokenGrantedAbility) {
			return nil, false
		}
	}
	return def, true
}

// decayedSacrificeTrigger is the attack-triggered ability the Decayed keyword
// (CR 702.148) grants a created token: "When this token attacks, sacrifice it at
// end of combat." It schedules a delayed end-of-combat sacrifice of the token,
// mirroring the delayed-disposal shape cards like Fog Elemental use. Decayed's
// other half — "This creature can't block." — is added separately as
// game.CantBlockStaticBody.
func decayedSacrificeTrigger() game.TriggeredAbility {
	return game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.CreateDelayedTrigger{
					Trigger: game.DelayedTriggerDef{
						Timing: game.DelayedAtEndOfCombat,
						Content: game.Mode{
							Sequence: []game.Instruction{{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
							}},
						}.Ability(),
					},
				},
			}},
		}.Ability(),
	}
}

// toxicTokenStaticBody builds the typed static-ability body granting a created
// token the parameterized toxic keyword sized from its Oracle rank ("with toxic
// 1" -> ToxicKeyword{Amount: 1}). It mirrors the toxic keyword body normal cards
// carry (lowerKeywordStatic), and fails closed for a non-positive rank.
func toxicTokenStaticBody(amount int) (game.StaticAbility, bool) {
	if amount <= 0 {
		return game.StaticAbility{}, false
	}
	return game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: amount}},
	}, true
}

// attachTokenGrantedAbility compiles and lowers the quoted ability a created
// token enters with ("... token with \"When this token dies, you gain 1
// life.\""), appending the resulting triggered, activated, mana, or static
// ability to the token definition. It mirrors lowerStaticGrantedQuotedAbility's
// recursive compile + lower of an already-parsed quoted body, and fails closed
// when the inner document does not compile to exactly one such lowered ability.
func attachTokenGrantedAbility(def *game.CardDef, granted *parser.StaticGrantedAbilitySyntax) bool {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil {
		return false
	}
	switch {
	case len(lowered.staticAbilities) == 1 &&
		!lowered.triggeredAbility.Exists &&
		!lowered.activatedAbility.Exists &&
		!lowered.manaAbility.Exists:
		// A quoted static ability ("This token can't block.") appends its lowered
		// static body to the token definition. Only a lone static ability with no
		// other lowered ability kind is accepted; any richer inner document falls
		// through and fails closed.
		def.StaticAbilities = append(def.StaticAbilities, lowered.staticAbilities[0].Body)
		return true
	case lowered.triggeredAbility.Exists:
		if abilityContentCreatesToken(lowered.triggeredAbility.Val.Content) {
			return false
		}
		def.TriggeredAbilities = append(def.TriggeredAbilities, lowered.triggeredAbility.Val)
		return true
	case lowered.activatedAbility.Exists:
		if abilityContentCreatesToken(lowered.activatedAbility.Val.Content) {
			return false
		}
		def.ActivatedAbilities = append(def.ActivatedAbilities, lowered.activatedAbility.Val)
		return true
	case lowered.manaAbility.Exists:
		if abilityContentCreatesToken(lowered.manaAbility.Val.Content) {
			return false
		}
		def.ManaAbilities = append(def.ManaAbilities, lowered.manaAbility.Val)
		return true
	default:
		return false
	}
}

// abilityContentCreatesToken reports whether a lowered ability body creates a
// token. A token's granted ability that itself creates a token would require the
// renderer to emit a second, nested token definition from within a token
// definition (Wolf's Quarry's Boar creating a Food token; the Fish/Whale/Kraken
// chain). The token-definition emitter does not synthesize those nested defs, so
// such granted abilities fail closed here rather than producing a token def that
// references an unemitted variable.
func abilityContentCreatesToken(content game.AbilityContent) bool {
	for i := range content.Modes {
		for j := range content.Modes[i].Sequence {
			if _, ok := content.Modes[i].Sequence[j].Primitive.(game.CreateToken); ok {
				return true
			}
		}
	}
	return false
}

// creatureTokenCardTypes returns the card types for a synthesized creature
// token. A bare creature token compiles to [Creature]; an artifact- or
// enchantment-creature token prepends the additional permanent type, matching
// the Oracle "<type> creature" ordering. Any other required-type set fails
// closed.
func creatureTokenCardTypes(selector compiler.CompiledSelector) ([]types.Card, bool) {
	required := selector.RequiredTypesAny()
	if len(required) == 0 {
		return []types.Card{types.Creature}, true
	}
	hasCreature := false
	var extra types.Card
	for _, cardType := range required {
		switch cardType {
		case types.Creature:
			hasCreature = true
		case types.Artifact, types.Enchantment:
			if extra != "" {
				return nil, false
			}
			extra = cardType
		default:
			return nil, false
		}
	}
	if !hasCreature {
		return nil, false
	}
	if extra == "" {
		return []types.Card{types.Creature}, true
	}
	return []types.Card{extra, types.Creature}, true
}

// synthesizeNamedArtifactTokenDef builds a token CardDef for a predefined
// artifact token with no printed power/toughness (Treasure, Food, Clue, Blood,
// Gold, Lander, Mutagen). These tokens carry a fixed Oracle ability the runtime
// CreateToken/TokenDef model already represents. Any other named token fails
// closed.
func synthesizeNamedArtifactTokenDef(effect *compiler.CompiledEffect) (*game.CardDef, bool) {
	if effect.TokenPTKnown {
		return nil, false
	}
	subtypes := effect.Selector.SubtypesAny()
	if len(subtypes) != 1 ||
		len(effect.Selector.ColorsAny()) != 0 ||
		effect.Selector.Keyword != parser.KeywordUnknown {
		return nil, false
	}
	return namedArtifactTokenDef(subtypes[0])
}

// synthesizePredefinedTokenDef builds a token CardDef for a predefined named
// token whose name is a card name rather than a card subtype (Mutavault). The
// create clause carries only the name, so the token's full definition — its
// types, mana ability, and activated abilities — is fixed here, mirroring the
// printed token's reminder text. Any name with no fixed definition fails closed.
func synthesizePredefinedTokenDef(effect *compiler.CompiledEffect) (*game.CardDef, bool) {
	if effect.TokenPredefinedName == "" || effect.TokenPTKnown {
		return nil, false
	}
	switch effect.TokenPredefinedName {
	case "Mutavault":
		return mutavaultTokenDef(), true
	default:
		return nil, false
	}
}

// mutavaultTokenDef builds the Mutavault token: a colorless land with the
// intrinsic colorless mana ability "{T}: Add {C}." and the self-animation
// ability "{1}: This token becomes a 2/2 creature with all creature types until
// end of turn. It's still a land." The animation adds the creature type and every
// creature subtype, and sets the token to 2/2, for the rest of the turn while
// leaving the land type intact (CR 613: the type layer adds rather than replaces).
func mutavaultTokenDef() *game.CardDef {
	manaAbility := game.TapManaAbility(mana.C)
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:          "Mutavault",
			Types:         []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{manaAbility},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:     "{1}: This token becomes a 2/2 creature with all creature types until end of turn. It's still a land.",
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.SourcePermanentReference()),
						ContinuousEffects: []game.ContinuousEffect{
							{
								Layer:                game.LayerType,
								AddTypes:             []types.Card{types.Creature},
								AddEveryCreatureType: true,
							},
							{
								Layer:        game.LayerPowerToughnessSet,
								SetPower:     opt.Val(game.PT{Value: 2}),
								SetToughness: opt.Val(game.PT{Value: 2}),
							},
						},
						Duration: game.DurationUntilEndOfTurn,
					},
				}}}.Ability(),
			}},
		},
	}
}

// namedArtifactTokenDef returns the synthesized CardDef for a recognized
// predefined artifact token, or false for an unrepresented one.
func namedArtifactTokenDef(sub types.Sub) (*game.CardDef, bool) {
	switch sub {
	case types.Treasure:
		return treasureTokenDef(), true
	case types.Food:
		return foodTokenDef(), true
	case types.Clue:
		return clueTokenDef(), true
	case types.Blood:
		return bloodTokenDef(), true
	case types.Gold:
		return goldTokenDef(), true
	case types.Lander:
		return landerTokenDef(), true
	case types.Mutagen:
		return mutagenTokenDef(), true
	case types.Map:
		return mapTokenDef(), true
	case types.Junk:
		return junkTokenDef(), true
	case types.Powerstone:
		return powerstoneTokenDef(), true
	default:
		return nil, false
	}
}

// sacrificeArtifactCost is the "Sacrifice this artifact" additional cost shared
// by predefined artifact tokens.
func sacrificeArtifactCost() cost.Additional {
	return cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}
}

// artifactTokenDef builds a single-ability colorless artifact token CardDef
// named for its subtype, matching paper predefined tokens.
func artifactTokenDef(sub types.Sub, ability *game.ActivatedAbility) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:               string(sub),
			Types:              []types.Card{types.Artifact},
			Subtypes:           []types.Sub{sub},
			ActivatedAbilities: []game.ActivatedAbility{*ability},
		},
	}
}

// treasureTokenDef builds the Treasure token: tap and sacrifice it to add one
// mana of any color.
func treasureTokenDef() *game.CardDef {
	ability := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	ability.Text = "{T}, Sacrifice this artifact: Add one mana of any color."
	ability.AdditionalCosts = append(slices.Clone(ability.AdditionalCosts), sacrificeArtifactCost())
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:          string(types.Treasure),
			Types:         []types.Card{types.Artifact},
			Subtypes:      []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{ability},
		},
	}
}

// foodTokenDef builds the Food token: pay two generic mana, tap, and sacrifice
// it to gain 3 life.
func foodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Food, &game.ActivatedAbility{
		Text:            "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// clueTokenDef builds the Clue token: pay two generic mana and sacrifice it to
// draw a card.
func clueTokenDef() *game.CardDef {
	return artifactTokenDef(types.Clue, &game.ActivatedAbility{
		Text:            "{2}, Sacrifice this artifact: Draw a card.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{sacrificeArtifactCost()},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// bloodTokenDef builds the Blood token: pay one generic mana, tap, discard a
// card, and sacrifice it to draw a card.
func bloodTokenDef() *game.CardDef {
	return artifactTokenDef(types.Blood, &game.ActivatedAbility{
		Text:     "{1}, {T}, Discard a card, Sacrifice this artifact: Draw a card.",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{
			cost.T,
			{Kind: cost.AdditionalDiscard, Text: "Discard a card", Amount: 1, Source: zone.Hand},
			sacrificeArtifactCost(),
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	})
}

// mapTokenDef builds the Map token: pay one generic mana, tap, and sacrifice it
// to make a target creature you control explore, at sorcery speed.
func mapTokenDef() *game.CardDef {
	return artifactTokenDef(types.Map, &game.ActivatedAbility{
		Text:            "{1}, {T}, Sacrifice this artifact: Target creature you control explores. Activate only as a sorcery.",
		ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeArtifactCost()},
		ZoneOfFunction:  zone.Battlefield,
		Timing:          game.SorceryOnly,
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature you control",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
					Controller:       game.ControllerYou,
				}),
			}},
			Sequence: []game.Instruction{
				{Primitive: game.Explore{Creature: game.TargetPermanentReference(0)}},
			},
		}.Ability(),
	})
}

// powerstoneTokenDef builds the Powerstone token: tap for one colorless mana that
// can't be spent to cast a nonartifact spell. The mana ability tags its produced
// mana with the artifact-spell spend restriction; the token enters tapped via the
// creating spell's "tapped" modifier.
func powerstoneTokenDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     string(types.Powerstone),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Powerstone},
			ManaAbilities: []game.ManaAbility{{
				Text:            "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
				AdditionalCosts: []cost.Additional{cost.T},
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{
						Amount:    game.Fixed(1),
						ManaColor: mana.C,
						SpendRider: opt.Val(game.ManaSpendRider{
							Condition:   game.ManaSpendCastArtifactSpell,
							Restriction: game.ManaSpendRestrictedToCondition,
						}),
					}},
				}}.Ability(),
			}},
		},
	}
}

// junkTokenDef builds the Junk token: tap and sacrifice it to exile the top card
// of your library and play that card this turn, at sorcery speed.
func junkTokenDef() *game.CardDef {
	return artifactTokenDef(types.Junk, &game.ActivatedAbility{
		Text:            "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.",
		AdditionalCosts: []cost.Additional{cost.T, sacrificeTokenCost()},
		ZoneOfFunction:  zone.Battlefield,
		Timing:          game.SorceryOnly,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.ImpulseExile{
				Player:   game.ControllerReference(),
				Amount:   game.Fixed(1),
				Duration: game.DurationThisTurn,
			}},
		}}.Ability(),
	})
}

// sacrificeTokenCost is the "Sacrifice this token" additional cost carried by
// the predefined tokens (Gold, Lander, Mutagen) whose printed Oracle text names
// the token rather than its artifact type.
func sacrificeTokenCost() cost.Additional {
	return cost.Additional{
		Kind:   cost.AdditionalSacrificeSource,
		Text:   "Sacrifice this token",
		Amount: 1,
	}
}

// goldTokenDef builds the Gold token: sacrifice it to add one mana of any color.
func goldTokenDef() *game.CardDef {
	choice := game.ChoiceKey("oracle-mana-color")
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     string(types.Gold),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Gold},
			ManaAbilities: []game.ManaAbility{{
				Text:            "Sacrifice this token: Add one mana of any color.",
				AdditionalCosts: []cost.Additional{sacrificeTokenCost()},
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
						},
						PublishChoice: choice,
					}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ChoiceFrom: choice}},
				}}.Ability(),
			}},
		},
	}
}

// landerTokenDef builds the Lander token: pay two generic mana, tap, and
// sacrifice it to search your library for a basic land and put it onto the
// battlefield tapped.
func landerTokenDef() *game.CardDef {
	return artifactTokenDef(types.Lander, &game.ActivatedAbility{
		Text:            "{2}, {T}, Sacrifice this token: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeTokenCost()},
		ZoneOfFunction:  zone.Battlefield,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Battlefield,
					Filter: game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Supertypes:    []types.Super{types.Basic},
					},
					EntersTapped: true,
				},
				Amount: game.Fixed(1),
			}},
		}}.Ability(),
	})
}

// mutagenTokenDef builds the Mutagen token: pay one generic mana, tap, and
// sacrifice it to put a +1/+1 counter on target creature, at sorcery speed.
func mutagenTokenDef() *game.CardDef {
	return artifactTokenDef(types.Mutagen, &game.ActivatedAbility{
		Text:            "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
		ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: []cost.Additional{cost.T, sacrificeTokenCost()},
		ZoneOfFunction:  zone.Battlefield,
		Timing:          game.SorceryOnly,
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{
				{Primitive: game.AddCounter{
					Amount:      game.Fixed(1),
					Object:      game.TargetPermanentReference(0),
					CounterKind: counter.PlusOnePlusOne,
				}},
			},
		}.Ability(),
	})
}

// tokenContentKeywords returns the kinds of the conjoined bare keywords in a
// create-token ability ("... and reach" -> [Reach]); these are the rider
// keywords beyond the leading "with <keyword>" selector keyword. It reports
// false when any compiled ability keyword carries a parameter, since a
// parameterized keyword is not a plain creature-token rider and must fail closed.
func tokenContentKeywords(content compiler.AbilityContent) ([]parser.KeywordKind, bool) {
	kinds := make([]parser.KeywordKind, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keyword.ParameterKind != parser.KeywordParameterNone {
			return nil, false
		}
		kinds = append(kinds, keyword.Kind)
	}
	return kinds, true
}

func unsupportedTokenCreationDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported token creation",
		"the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color",
	)
}
