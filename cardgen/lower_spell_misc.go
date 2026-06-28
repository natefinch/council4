package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerFixedLifeSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
	groupPrimitiveFactory func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
	case triggeringEventQuantityKind(effect.Amount.DynamicKind):
		// "you gain/lose that much life" reads its anaphor from the enclosing
		// trigger. The parser pins every "that much life" phrase to one
		// life-change kind without knowing which event fired, so the lowering
		// resolves it on whichever event actually fired — the damage dealt in a
		// damage trigger ("Whenever this creature deals damage, you gain that
		// much life."), the life gained or lost in a life-change trigger, and so
		// on — keeping the parser text-blind.
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, effect.Amount)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		if effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower ||
			effect.Amount.DynamicKind == compiler.DynamicAmountSourceToughness {
			dynamic, ok := lifeSourcePowerAmount(ctx, effect)
			if !ok {
				return game.AbilityContent{}, contentDiagnostic(
					ctx,
					"unsupported life spell",
					"the executable source backend supports only exact supported life changes",
				)
			}
			amount = game.Dynamic(dynamic)
			break
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		sourceCounterReferences := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount &&
			singleSelfReference(ctx.content.References)
		if !ok ||
			len(ctx.content.References) != 0 && !sourceCounterReferences {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case len(ctx.content.References) != 0:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact supported life changes",
		)
	default:
	}
	if !effect.Exact ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	if len(ctx.content.Targets) == 0 {
		switch effect.Context {
		case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.OpponentsReference()),
				}},
			}.Ability(), nil
		case parser.EffectContextEachPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.AllPlayersReference()),
				}},
			}.Ability(), nil
		}
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ctx.content.Targets) == 0 &&
		effect.Context == parser.EffectContextController:
	case len(ctx.content.Targets) == 0 &&
		effect.Context == parser.EffectContextDefendingPlayer:
		playerRef = game.DefendingPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(effect.Context == parser.EffectContextEventPlayer &&
			ctx.content.References[0].Kind == compiler.ReferencePronoun &&
			ctx.content.References[0].Pronoun == compiler.ReferencePronounThey ||
			effect.Context == parser.EffectContextReferencedPlayer &&
				ctx.content.References[0].Kind == compiler.ReferenceThatPlayer &&
				ctx.content.References[0].Binding != compiler.ReferenceBindingTarget):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 1 &&
		effect.Context == parser.EffectContextReferencedObjectController:
		ref, ok := referencedControllerPlayerRef(ctx)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		effect.Context == parser.EffectContextReferencedPlayer &&
		hasThatPlayerTargetReference(ctx.content.References):
		ref, ok := referencedThatPlayerRef(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}

		targets = []game.TargetSpec{targetSpec}
		playerRef = game.TargetPlayerReference(0)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: primitiveFactory(amount, playerRef),
		}},
	}.Ability(), nil
}

// lifeSourcePowerAmount lowers "gain/lose life equal to its power/toughness" by
// binding the amount to the power or toughness of the object its referent names —
// the source permanent in a static or non-trigger context, or the triggering
// permanent's last-known characteristic in a leaves/dies trigger where "its"
// binds to the event ("When this creature dies, you gain life equal to its
// toughness.", Angelic Chorus). Every reference in the ability must bind to a
// source or event object so the form stays exact and fails closed on any foreign
// referent.
func lifeSourcePowerAmount(ctx contentCtx, effect compiler.CompiledEffect) (game.DynamicAmount, bool) {
	object, ok := referencedSourceOrEventPowerObject(effect.Amount, ctx.content.References)
	if !ok {
		return game.DynamicAmount{}, false
	}
	return lowerDynamicAmount(effect.Amount, object)
}

// referencedSourceOrEventPowerObject returns the object whose power or toughness
// feeds a source-power amount, found by matching the amount's referent span. It
// requires every reference to lower as a source or event object so trigger
// subjects ("this creature") and the amount referent ("its") are both accounted
// for.
func referencedSourceOrEventPowerObject(amount compiler.CompiledAmount, references []compiler.CompiledReference) (game.ObjectReference, bool) {
	var object game.ObjectReference
	found := false
	for i := range references {
		resolved, ok := lowerObjectReference(references[i], referenceLoweringContext{
			AllowSource: true,
			AllowEvent:  true,
		})
		if !ok {
			return game.ObjectReference{}, false
		}
		if references[i].Span == amount.ReferenceSpan {
			object = resolved
			found = true
		}
	}
	return object, found
}

func lowerFixedDestroySpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	preventRegeneration := ctx.content.Effects[0].PreventRegeneration
	if group, ok := exactMassDestroyGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Destroy{
						Group:               group,
						PreventRegeneration: preventRegeneration,
					},
				},
			},
		}.Ability(), nil
	}
	if content, ok := lowerMultiTargetPermanentSpell(ctx, func(object game.ObjectReference) game.Primitive {
		return game.Destroy{Object: object, PreventRegeneration: preventRegeneration}
	}); ok {
		return content, nil
	}
	if content, ok := lowerMultiDistinctTargetPermanentSpell(ctx, func(object game.ObjectReference) game.Primitive {
		return game.Destroy{Object: object, PreventRegeneration: preventRegeneration}
	}); ok {
		return content, nil
	}
	colorGate, hasColorGate := targetColorGateSelection(ctx.content.Conditions)
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 && !hasColorGate ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	instruction := game.Instruction{
		Primitive: game.Destroy{
			Object:              game.TargetPermanentReference(0),
			PreventRegeneration: preventRegeneration,
		},
	}
	if hasColorGate {
		instruction.Condition = opt.Val(targetColorEffectCondition(
			game.TargetPermanentReference(0),
			colorGate,
			ctx.content.Conditions[0].Text,
		))
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{instruction},
	}.Ability(), nil
}

func lowerFixedExileSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerSourceSpellExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerTargetedGraveyardExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerControllerGraveyardChoiceExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerLinkedAnyGraveyardChoiceExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerPlayerGraveyardExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerTargetPlayerHandExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerAllGraveyardExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourceAttachedExile(ctx); ok {
		return content, nil
	}
	if group, ok := exactMassExileGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Exile{Group: group},
			}},
		}.Ability(), nil
	}
	if content, ok := lowerMultiTargetExileSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerMultiDistinctTargetPermanentSpell(ctx, func(object game.ObjectReference) game.Primitive {
		return game.Exile{Object: object}
	}); ok {
		return content, nil
	}
	return lowerFixedPermanentTargetSpell(ctx, "Exile", func(object game.ObjectReference) game.Primitive {
		return game.Exile{Object: object}
	})
}

func lowerSourceSpellExile(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilitySpell ||
		len(ctx.content.Effects) != 1 ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.References) != 1 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingSource, 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Exile{SourceSpell: true},
	}}}.Ability(), true
}

// lowerSourceAttachedExile lowers the attached-recipient exile form "Exile
// enchanted creature." (Aura) or "Exile equipped creature." (Equipment) into a
// single Exile instruction acting on the source attached-permanent reference,
// the permanent the source Aura or Equipment is attached to. It mirrors the
// attached-recipient regenerate path: a single exact controller effect with no
// target, reference, condition, mode, keyword, or optional offer, and fails
// closed for every other shape.
func lowerSourceAttachedExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.ExileAttached ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	object := game.SourceAttachedPermanentReference()
	if len(object.Validate()) != 0 {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Exile{Object: object},
	}}}.Ability(), true
}

// isExactSourceSpellShuffleIntoLibrary reports whether effect is the exact
// "Shuffle <this card> into its owner's library." resolution tail naming the
// resolving spell itself. The parser validated the precise wording, so this
// keys off the exact EffectShuffle clause whose destination is the library and
// which names the source spell.
func isExactSourceSpellShuffleIntoLibrary(effect *compiler.CompiledEffect) bool {
	if !effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationNone ||
		effect.Kind != compiler.EffectShuffle ||
		effect.ToZone != zone.Library {
		return false
	}
	return referencesContainKind(effect.References, compiler.ReferenceSelfName) ||
		referencesContainKind(effect.References, compiler.ReferenceThisObject)
}

// lowerSourceSpellShuffleIntoLibrary lowers the exact "Shuffle <this card> into
// its owner's library." resolution tail (Green Sun's Zenith, the Beacon cycle)
// to a single source-spell shuffle-into-library instruction. The shuffled
// object is the resolving spell itself, so the instruction carries no referent.
func lowerSourceSpellShuffleIntoLibrary(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilitySpell ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		!isExactSourceSpellShuffleIntoLibrary(&ctx.content.Effects[0]) {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ShuffleSpellIntoLibrary{},
	}}}.Ability(), true
}

// isExactControllerGraveyardShuffleIntoLibrary recognizes the exact "Shuffle
// your graveyard into your library." resolution clause: a controller-scoped
// shuffle whose graveyard source and library destination are typed and which
// carries no targets or referents.
func isExactControllerGraveyardShuffleIntoLibrary(effect *compiler.CompiledEffect) bool {
	return effect.Exact &&
		!effect.Negated &&
		effect.Duration == compiler.DurationNone &&
		effect.Kind == compiler.EffectShuffle &&
		effect.Context == parser.EffectContextController &&
		effect.FromZone == zone.Graveyard &&
		effect.ToZone == zone.Library &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

// lowerControllerGraveyardShuffleIntoLibrary lowers "Shuffle your graveyard into
// your library." (The Mending of Dominaria chapter III) to a single
// shuffle-graveyard-into-library instruction targeting the controller.
func lowerControllerGraveyardShuffleIntoLibrary(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		!isExactControllerGraveyardShuffleIntoLibrary(&ctx.content.Effects[0]) {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ShuffleGraveyardIntoLibrary{Player: game.ControllerReference()},
	}}}.Ability(), true
}

func lowerPlayerRuleEffect(ctx contentCtx, kind game.RuleEffectKind) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	keywordsValid := len(ctx.content.Keywords) == 0
	if kind == game.RuleEffectPlayerProtection {
		keywordsValid = len(ctx.content.Keywords) == 1 &&
			ctx.content.Keywords[0].Kind == parser.KeywordProtection &&
			ctx.content.Keywords[0].ProtectionKnown &&
			ctx.content.Keywords[0].Protection.Everything &&
			len(ctx.content.Keywords[0].Protection.FromColors) == 0 &&
			len(ctx.content.Keywords[0].Protection.FromTypes) == 0 &&
			len(ctx.content.Keywords[0].Protection.FromSubtypes) == 0 &&
			!ctx.content.Keywords[0].Protection.Multicolored &&
			!ctx.content.Keywords[0].Protection.Monocolored &&
			!ctx.content.Keywords[0].Protection.EachColor
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilYourNextTurn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!keywordsValid {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported player rule effect",
			"the player-scoped rule effect did not match its exact supported form",
		)
	}
	ruleEffect := game.RuleEffect{
		Kind:           kind,
		AffectedPlayer: game.PlayerYou,
	}
	if kind == game.RuleEffectPlayerProtection {
		ruleEffect.Protection = ctx.content.Keywords[0].Protection
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{ruleEffect},
			Duration:    game.DurationUntilYourNextTurn,
		},
	}}}.Ability(), nil
}

// lowerAdditionalLandPlays lowers the controller-scoped additional-land-play
// grant ("Play an additional land this turn.", "You may play two additional
// lands this turn.") to an ApplyRule that raises the controller's land-play
// allowance for the rest of the turn. The "may" is a permission, not a resolving
// choice, so it is modeled as an unconditional allowance the player need not use.
// It fails closed for any target, reference, condition, keyword, or mode.
func lowerAdditionalLandPlays(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Duration != compiler.DurationThisTurn ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported additional land play effect",
			"the executable source backend supports only the exact controller-scoped additional-land-play grant this turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:                game.RuleEffectAdditionalLandPlays,
				AffectedPlayer:      game.PlayerYou,
				AdditionalLandPlays: effect.Amount.Value,
			}},
			Duration: game.DurationThisTurn,
		},
	}}}.Ability(), nil
}

// lowerAdditionalCombatPhase lowers the extra-phase-insertion effect "After this
// [main] phase, there is an additional combat phase[ followed by an additional
// main phase]." (Aggravated Assault, Aurelia the Warleader, World at War) to an
// AddExtraPhases primitive that queues the extra phases onto the current turn.
// It reads only the typed AdditionalCombatPhase / AdditionalMainPhase flags and
// fails closed for any target, reference, condition, keyword, mode, amount,
// duration, or negation so only the bare extra-phase clause lowers.
func lowerAdditionalCombatPhase(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		!effect.AdditionalCombatPhase ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationNone ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported additional combat phase effect",
			"the executable source backend supports only the exact additional combat phase insertion this turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.AddExtraPhases{
			Combat: effect.AdditionalCombatPhase,
			Main:   effect.AdditionalMainPhase,
		},
	}}}.Ability(), nil
}

// lowerCastAsThoughFlash lowers the controller-scoped timing permission "You may
// cast spells this turn as though they had flash." to an ApplyRule that lets the
// controller cast spells at instant speed for the rest of the turn. Like
// lowerAdditionalLandPlays the "may" is a permission, not a resolving choice, so
// it is modeled as an unconditional turn-scoped allowance. The parser's exact
// nine-word match fixes the wording, so the inherent "flash" keyword and
// "they"/"spells" references in the same sentence are expected; only targets,
// conditions, modes, an amount, a negation, or a non-controller scope fail
// closed.
func lowerCastAsThoughFlash(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationThisTurn ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported cast-as-though-flash effect",
			"the executable source backend supports only the exact controller-scoped cast-spells-as-though-flash grant this turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer: game.PlayerYou,
			}},
			Duration: game.DurationThisTurn,
		},
	}}}.Ability(), nil
}

// lowerPlayFromLibraryTop lowers the controller-scoped, turn-scoped grant "until
// end of turn, you may look at the top card of your library any time and you may
// play cards from the top of your library." (Gwenom, Remorseless) to an ApplyRule
// that grants the controller, until end of turn, the private top-card visibility
// plus permission to play lands and cast spells from the top of their library.
// "Play cards" covers both playing lands and casting nonland spells, so the grant
// emits the land-play and spell-cast permissions together. When the
// PlayFromTopPayLife rider is present, spells cast this way pay life equal to
// their mana value instead of their mana cost. The leading "you may" permissions
// are unconditional allowances (like lowerCastAsThoughFlash). Targets, references,
// conditions, keywords, modes, a negation, an amount, a non-until-end-of-turn
// duration, or a non-controller scope fail closed.
func lowerPlayFromLibraryTop(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported play-from-library-top effect",
			"the executable source backend supports only the exact controller-scoped until-end-of-turn look-at-and-play-from-library-top grant",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{
				{
					Kind:           game.RuleEffectLookAtTopCardAnyTime,
					AffectedPlayer: game.PlayerYou,
				},
				{
					Kind:           game.RuleEffectPlayLandsFromZone,
					AffectedPlayer: game.PlayerYou,
					CastFromZone:   zone.Library,
					PermanentTypes: []types.Card{types.Land},
					TopCardOnly:    true,
				},
				{
					Kind:                    game.RuleEffectCastSpellsFromZone,
					AffectedPlayer:          game.PlayerYou,
					CastFromZone:            zone.Library,
					TopCardOnly:             true,
					PayLifeEqualToManaValue: effect.PlayFromTopPayLife,
				},
			},
			Duration: game.DurationUntilEndOfTurn,
		},
	}}}.Ability(), nil
}

// lowerNoMaximumHandSize lowers the controller-scoped, rest-of-game continuous
// effect "You have no maximum hand size for the rest of the game." (Sea Gate
// Restoration) to an ApplyRule that removes the controller's maximum hand size
// permanently. It reuses the continuous RuleEffectNoMaximumHandSize rule effect
// (shared with the permanent static "You have no maximum hand size." form on
// Reliquary Tower and similar) with a permanent duration, since "for the rest of
// the game" never expires. Targets, references, conditions, keywords, modes, a
// negation, an amount, or a non-controller scope fail closed.
func lowerNoMaximumHandSize(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported no-maximum-hand-size effect",
			"the executable source backend supports only the exact controller-scoped rest-of-game no-maximum-hand-size effect",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectNoMaximumHandSize,
				AffectedPlayer: game.PlayerYou,
			}},
			Duration: game.DurationPermanent,
		},
	}}}.Ability(), nil
}

// lowerCantCastSpells lowers the one-shot, turn-scoped player cast prohibition
// "<players> can't cast spells this turn." (Silence) to an ApplyRule that
// forbids the affected players from casting spells for the rest of the turn. The
// affected players are the controller's opponents ("your opponents", "each
// opponent") or every player ("players", CantCastSpellsAllPlayers). It reuses
// the continuous RuleEffectCantCastSpells rule effect with a this-turn duration.
// Targets, references, conditions, modes, a negation, an amount, or a
// non-controller scope fail closed.
func lowerCantCastSpells(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationThisTurn ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported cant-cast-spells effect",
			"the executable source backend supports only the exact opponents-or-all-players cast-spells prohibition this turn",
		)
	}
	affected := game.PlayerOpponent
	if effect.CantCastSpellsAllPlayers {
		affected = game.PlayerAny
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantCastSpells,
				AffectedPlayer:     affected,
				SpellTypes:         append([]types.Card(nil), effect.CantCastSpellsRequiredTypes...),
				ExcludedSpellTypes: append([]types.Card(nil), effect.CantCastSpellsExcludedTypes...),
			}},
			Duration: game.DurationThisTurn,
		},
	}}}.Ability(), nil
}

// lowerSpellsCantBeCountered lowers the controller-scoped, turn-scoped resolving
// buff "The next spell you cast this turn can't be countered." (Mistrise
// Village) and the all-spells form "Spells you cast this turn can't be
// countered." (Domri, Anarch of Bolas) to an ApplyRule that makes the
// controller's spells uncounterable for the rest of the turn. The
// next-spell-only variant sets AppliesToNextSpellOnly so the buff is consumed by
// the single next spell the controller casts. Targets, references, conditions,
// modes, a negation, an amount, or a non-controller scope fail closed.
func lowerSpellsCantBeCountered(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Duration != compiler.DurationThisTurn ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported spells-cant-be-countered effect",
			"the executable source backend supports only the exact controller-scoped spells-cant-be-countered buff this turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:                   game.RuleEffectCantBeCountered,
				AffectedController:     game.ControllerYou,
				AppliesToNextSpellOnly: effect.SpellsCantBeCounteredNextOnly,
			}},
			Duration: game.DurationThisTurn,
		},
	}}}.Ability(), nil
}

// lowerGroupMustAttack lowers the one-shot forced-attack effect "<group> attack
// this turn if able." (Bident of Thassa: "Creatures your opponents control
// attack this turn if able.") and its duration-scoped variant "Until your next
// turn, <group> attack each combat if able." (The Akroan War chapter II) to an
// ApplyRule that forces the affected creatures to attack for the recognized
// duration. The affected creature group is read from the parser-recognized
// StaticSubject and mapped to a controller relation; the rule reuses the
// continuous RuleEffectMustAttack rule effect. Targets, references, conditions,
// modes, a negation, an amount, an unsupported duration, or an unsupported
// group subject fail closed.
func lowerGroupMustAttack(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	var controller game.ControllerRelation
	switch effect.StaticSubject {
	case compiler.StaticSubjectControlledCreatures:
		controller = game.ControllerYou
	case compiler.StaticSubjectOpponentControlledCreatures:
		controller = game.ControllerOpponent
	case compiler.StaticSubjectAllCreatures:
		controller = game.ControllerAny
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported forced-attack effect",
			"the executable source backend supports only the exact you/opponents/all creatures forced-attack effect this turn or until your next turn",
		)
	}
	var duration game.EffectDuration
	switch effect.Duration {
	case compiler.DurationThisTurn:
		duration = game.DurationThisTurn
	case compiler.DurationUntilYourNextTurn:
		duration = game.DurationUntilYourNextTurn
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported forced-attack effect",
			"the executable source backend supports only the exact you/opponents/all creatures forced-attack effect this turn or until your next turn",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported forced-attack effect",
			"the executable source backend supports only the exact you/opponents/all creatures forced-attack effect this turn or until your next turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectMustAttack,
				AffectedController: controller,
				PermanentTypes:     []types.Card{types.Creature},
			}},
			Duration: duration,
		},
	}}}.Ability(), nil
}

// lowerDirectedTwoPlayerMustAttack lowers The Brothers' War chapter II directed
// forced-attack effect "Choose two target players. Until your next turn, each
// creature they control attacks the other chosen player each combat if able." to
// two distinct player target slots plus an ApplyRule that creates a reciprocal
// pair of directed RuleEffectMustAttack effects: each chosen player's creatures
// must attack the other chosen player (or a planeswalker or battle they control)
// each combat if able, until the source controller's next turn. Any other
// duration, target shape, mode, condition, keyword, or unexpected reference fails
// closed.
func lowerDirectedTwoPlayerMustAttack(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported forced-attack effect",
			"the executable source backend supports only the exact two-target-player directed forced-attack effect until your next turn",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Amount.Known ||
		effect.Context != parser.EffectContextController ||
		effect.Duration != compiler.DurationUntilYourNextTurn ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		!directedMustAttackReferencesSupported(ctx.content.References) {
		return unsupported()
	}
	target := ctx.content.Targets[0]
	if target.Selector.Kind != compiler.SelectorPlayer ||
		target.Selector.Controller != compiler.ControllerAny ||
		target.Selector.Another ||
		target.Selector.Other ||
		target.Cardinality.Min != 2 ||
		target.Cardinality.Max != 2 {
		return unsupported()
	}
	playerSpec := func(distinct bool) game.TargetSpec {
		return game.TargetSpec{
			MinTargets:               1,
			MaxTargets:               1,
			Constraint:               "player",
			Allow:                    game.TargetAllowPlayer,
			DistinctFromPriorTargets: distinct,
		}
	}
	directedEffect := func(affected, required int) game.RuleEffect {
		return game.RuleEffect{
			Kind:                    game.RuleEffectMustAttack,
			PermanentTypes:          []types.Card{types.Creature},
			AffectedPlayerRef:       game.TargetPlayerReference(affected),
			RequiredAttackTargetRef: game.TargetPlayerReference(required),
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{playerSpec(false), playerSpec(true)},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyRule{
				RuleEffects: []game.RuleEffect{
					directedEffect(0, 1),
					directedEffect(1, 0),
				},
				Duration: game.DurationUntilYourNextTurn,
			},
		}},
	}.Ability(), nil
}

// directedMustAttackReferencesSupported reports whether the directed
// forced-attack effect's references are limited to the inherent "they" pronoun
// that names the two chosen players. Any other reference fails closed.
func directedMustAttackReferencesSupported(references []compiler.CompiledReference) bool {
	for i := range references {
		switch references[i].Pronoun {
		case compiler.ReferencePronounThey, compiler.ReferencePronounTheir:
		default:
			return false
		}
	}
	return true
}

// cost modifier "[<type filter>] spells <caster> cast cost {N} more/less to
// cast" scoped by a recognized finite duration ("Artifact spells you cast this
// turn cost {1} less to cast.", Armor Wars chapter II; "Until your next turn,
// spells your opponents cast cost {1} more to cast.", Tax Collector) to an
// ApplyRule that creates a RuleEffectCostModifier rule effect for that lifetime.
// The caster phrase selects the affected-player relation; an optional single
// card-type filter narrows the modifier to that spell type, or a single excluded
// card-type filter narrows it to spells that lack that type ("Noncreature spells
// ...", Elspeth Conquers Death chapter II). A required and an excluded filter
// never combine. Targets, references, conditions, modes, keywords, a negation, or
// an unsupported duration fail closed.
func lowerSpellCostModifier(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.SpellCostModifierAmount <= 0 ||
		len(effect.SpellCostModifierExcludedTypes) > 1 ||
		len(effect.SpellCostModifierRequiredTypes) > 1 ||
		(len(effect.SpellCostModifierExcludedTypes) != 0 && len(effect.SpellCostModifierRequiredTypes) != 0) ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedSpellCostModifierDiagnostic(ctx)
	}
	duration, ok := resolvingSpellCostModifierDuration(effect.Duration)
	if !ok {
		return game.AbilityContent{}, unsupportedSpellCostModifierDiagnostic(ctx)
	}
	affected, ok := spellCostModifierAffectedPlayer(effect.SpellCostModifierCaster)
	if !ok {
		return game.AbilityContent{}, unsupportedSpellCostModifierDiagnostic(ctx)
	}
	modifier := game.CostModifier{Kind: game.CostModifierSpell}
	if effect.SpellCostModifierIncrease {
		modifier.GenericIncrease = effect.SpellCostModifierAmount
	} else {
		modifier.GenericReduction = effect.SpellCostModifierAmount
	}
	if len(effect.SpellCostModifierRequiredTypes) == 1 {
		modifier.CardSelection.RequiredTypes = []types.Card{effect.SpellCostModifierRequiredTypes[0]}
	}
	if len(effect.SpellCostModifierExcludedTypes) == 1 {
		modifier.CardSelection.ExcludedTypes = []types.Card{effect.SpellCostModifierExcludedTypes[0]}
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedPlayer: affected,
				CostModifier:   modifier,
			}},
			Duration: duration,
		},
	}}}.Ability(), nil
}

// resolvingSpellCostModifierDuration maps the supported finite durations of a
// resolving spell cost modifier to their runtime effect durations. A permanent
// or otherwise unsupported duration fails closed: a resolving cost modifier is
// always temporary.
func resolvingSpellCostModifierDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationThisTurn:
		return game.DurationThisTurn, true
	case compiler.DurationUntilEndOfTurn:
		return game.DurationUntilEndOfTurn, true
	case compiler.DurationUntilYourNextTurn:
		return game.DurationUntilYourNextTurn, true
	case compiler.DurationUntilEndOfYourNextTurn:
		return game.DurationUntilEndOfYourNextTurn, true
	default:
		return game.DurationPermanent, false
	}
}

// spellCostModifierAffectedPlayer maps a resolving spell cost modifier's caster
// phrase to the rule effect's affected-player relation: the controller's spells
// ("you cast"), the controller's opponents' spells ("your opponents cast"), or
// every player's spells (an absent caster phrase).
func spellCostModifierAffectedPlayer(caster parser.SpellCostCasterKind) (game.PlayerRelation, bool) {
	switch caster {
	case parser.SpellCostCasterController:
		return game.PlayerYou, true
	case parser.SpellCostCasterOpponents:
		return game.PlayerOpponent, true
	case parser.SpellCostCasterAll:
		return game.PlayerAny, true
	default:
		return game.PlayerAny, false
	}
}

func unsupportedSpellCostModifierDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported spell cost modifier",
		"the executable source backend supports only a duration-bounded resolving spell cost modifier with at most one required card-type filter",
	)
}

func lowerPlayerRuleOrPhaseEffect(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectLifeTotalCantChange:
		content, diagnostic := lowerPlayerRuleEffect(ctx, game.RuleEffectLifeTotalCantChange)
		return content, diagnostic, true
	case compiler.EffectProtectionFromEverything:
		content, diagnostic := lowerPlayerRuleEffect(ctx, game.RuleEffectPlayerProtection)
		return content, diagnostic, true
	case compiler.EffectAdditionalLandPlays:
		content, diagnostic := lowerAdditionalLandPlays(ctx)
		return content, diagnostic, true
	case compiler.EffectCastAsThoughFlash:
		content, diagnostic := lowerCastAsThoughFlash(ctx)
		return content, diagnostic, true
	case compiler.EffectPlayFromLibraryTop:
		content, diagnostic := lowerPlayFromLibraryTop(ctx)
		return content, diagnostic, true
	case compiler.EffectAdditionalCombatPhase:
		content, diagnostic := lowerAdditionalCombatPhase(ctx)
		return content, diagnostic, true
	case compiler.EffectNoMaximumHandSize:
		content, diagnostic := lowerNoMaximumHandSize(ctx)
		return content, diagnostic, true
	case compiler.EffectCantCastSpells:
		content, diagnostic := lowerCantCastSpells(ctx)
		return content, diagnostic, true
	case compiler.EffectSpellCostModifier:
		content, diagnostic := lowerSpellCostModifier(ctx)
		return content, diagnostic, true
	case compiler.EffectSpellsCantBeCountered:
		content, diagnostic := lowerSpellsCantBeCountered(ctx)
		return content, diagnostic, true
	case compiler.EffectMustAttack:
		if ctx.content.Effects[0].Context == parser.EffectContextTarget {
			content, diagnostic := lowerTargetMustAttack(ctx)
			return content, diagnostic, true
		}
		content, diagnostic := lowerGroupMustAttack(ctx)
		return content, diagnostic, true
	case compiler.EffectDirectedMustAttack:
		content, diagnostic := lowerDirectedTwoPlayerMustAttack(ctx)
		return content, diagnostic, true
	case compiler.EffectAttackTax:
		content, diagnostic := lowerAttackTaxSpell(ctx)
		return content, diagnostic, true
	case compiler.EffectPhaseOut:
		content, diagnostic := lowerPhaseOutSpell(ctx)
		return content, diagnostic, true
	default:
		return game.AbilityContent{}, nil, false
	}
}

// lowerPlayerGraveyardExile lowers the whole-graveyard exile "Exile target
// player's graveyard." (and its "target opponent's graveyard." variant) to a
// single target-player TargetSpec plus the player-zone group form of MoveCard,
// which the runtime resolves by moving every card in the chosen player's
// graveyard to exile at once. It reuses the typed GraveyardZoneExile owner
// relation the parser recognized rather than reconstructing wording here. It
// fails closed for any extra clause, condition, mode, keyword, or reference so
// the riders, modal siblings, and "that player's graveyard"/"all graveyards"
// forms stay unsupported.
func lowerPlayerGraveyardExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated {
		return game.AbilityContent{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: ctx.content.Targets[0].Text,
		Allow:      game.TargetAllowPlayer,
	}
	switch effect.GraveyardZoneExile {
	case parser.GraveyardZoneExileTargetPlayer:
	case parser.GraveyardZoneExileTargetOpponent:
		spec.Selection = opt.Val(game.Selection{Player: game.PlayerOpponent})
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Player:      game.TargetPlayerReference(0),
				FromZone:    zone.Graveyard,
				Destination: zone.Exile,
			},
		}},
	}.Ability(), true
}

// lowerAllGraveyardExile lowers the non-targeted whole-graveyard wipe "Exile all
// graveyards." (and the synonymous "Exile each player's graveyard.") to the
// player-group form of MoveCard, which the runtime resolves by moving every card
// in every player's graveyard to exile at once. It carries no target and reuses
// the typed GraveyardZoneExileAll relation the parser recognized. It fails closed
// for any extra clause, target, condition, mode, keyword, or reference so riders
// and modal siblings stay unsupported.
func lowerAllGraveyardExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated ||
		effect.GraveyardZoneExile != parser.GraveyardZoneExileAll {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				PlayerGroup: game.AllPlayersReference(),
				FromZone:    zone.Graveyard,
				Destination: zone.Exile,
			},
		}},
	}.Ability(), true
}

// lowerTargetedGraveyardExile lowers "Exile target card from a graveyard." (and
// its "from your graveyard"/"from an opponent's graveyard", typed-noun, subtype,
// mana-value, and "up to N" variants) to one MoveCard per target slot that moves
// the chosen graveyard card to exile. Exiling a graveyard card is a plain
// zone change, so it reuses the same graveyard-card target spec the graveyard
// return and put paths build (cardInZoneTargetSpec) and the runtime MoveCard
// primitive, which the rules engine resolves by removing the targeted card from
// its graveyard and adding it to exile. It returns ok=false for any non-graveyard
// exile so the mass, multi-target, and single-permanent exile paths are untouched.
func lowerTargetedGraveyardExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Negated ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		ctx.content.Effects[0].FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i},
			FromZone:    zone.Graveyard,
			Destination: zone.Exile,
		}})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// lowerControllerGraveyardChoiceExile lowers the non-target "exile a <filter>
// card from your graveyard" wording, where the exiled card is chosen from the
// controller's own graveyard at resolution rather than targeted (Masked Vandal,
// the Imoen cycle, Aphemia, Forgotten Harvest, ...). The targeted form ("exile
// target ... card from your graveyard") lowers through lowerTargetedGraveyardExile
// instead. It produces one game.ChooseFromZone instruction whose Filter
// carries the same card filter the targeted and search paths reconstruct, so an
// enclosing "you may X. If you do, Y" wrapper marks that single instruction
// Optional and gates Y on the player having exiled a card. It is card-name-blind
// and fails closed (ok=false) on any shape it does not fully model — a reference
// or target, a non-graveyard source, a non-"your" controller scope, a selector
// qualifier it cannot express, or a non-fixed amount — so an unmodeled wording
// falls through to the generic exile path's diagnostic rather than lowering to a
// silently-wrong instruction.
func lowerControllerGraveyardChoiceExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
		selector.Controller != compiler.ControllerYou ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ExileFromGraveyardChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(effect.Amount.Value),
			false,
			"",
		),
	}}}.Ability(), true
}

// target has a plural ("Exile two target creatures.") or optional ("Exile up to
// two target artifacts.", "Exile up to one target permanent.") cardinality. It
// emits one multi-target spec carrying the chosen MinTargets/MaxTargets range
// and one Exile instruction per slot, each addressing its target index. Slots
// the player declines to fill at announcement leave fewer chosen targets, and
// the runtime Exile primitive no-ops on an unresolved target index, so an "up
// to" exile of N safely exiles only the chosen targets. It returns ok=false for
// the single-target form so that path stays on lowerFixedPermanentTargetSpell.
func lowerMultiTargetExileSpell(ctx contentCtx) (game.AbilityContent, bool) {
	return lowerMultiTargetPermanentSpell(ctx, func(object game.ObjectReference) game.Primitive {
		return game.Exile{Object: object}
	})
}

// lowerMultiDistinctTargetPermanentSpell lowers a single permanent verb (destroy,
// exile) applied to two or more distinct single-permanent targets, each of its
// own type ("Destroy target artifact, target creature, target enchantment, and
// target land." — Decimate, "Destroy target artifact and target creature." —
// shorter heterogeneous forms). Each "target <type>" clause compiles to its own
// {1,1} exact TargetSpec, and the verb emits one primitive per target slot.
// Object references address chosen targets by a flat slot index across all
// specs, so with every spec admitting one slot the slot index equals the spec
// index, letting slot i carry the i-th distinct target. It fails closed for the
// single-target form (handled by the single-target path) and for any optional,
// negated, conditional, keyword, modal, or referenced shape it does not model,
// and for any target permanentTargetSpec cannot express.
func lowerMultiDistinctTargetPermanentSpell(
	ctx contentCtx,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) < 2 ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	specs := make([]game.TargetSpec, 0, len(ctx.content.Targets))
	sequence := make([]game.Instruction, 0, len(ctx.content.Targets))
	for i := range ctx.content.Targets {
		spec, ok := permanentTargetSpec(ctx.content.Targets[i])
		if !ok {
			return game.AbilityContent{}, false
		}
		specs = append(specs, spec)
		sequence = append(sequence, game.Instruction{
			Primitive: primitiveFactory(game.TargetPermanentReference(i)),
		})
	}
	return game.Mode{
		Targets:  specs,
		Sequence: sequence,
	}.Ability(), true
}

// lowerMultiTargetPermanentSpell lowers a single-object permanent verb (exile,
// destroy, tap, untap, regenerate) whose one permanent target has a plural
// ("Destroy two target creatures.") or optional ("Tap up to two target
// creatures.", "Exile up to one target permanent.") cardinality. It emits one
// multi-target spec carrying the chosen MinTargets/MaxTargets range and one
// primitive per slot, each addressing its own target index. Slots the player
// declines to fill at announcement leave fewer chosen targets, and the runtime
// single-object handlers no-op on an unresolved target index, so an "up to"
// spell of N safely affects only the chosen targets. It returns ok=false for
// the single-target form so that path stays on lowerFixedPermanentTargetSpell,
// and for any reference, condition, keyword, or mode it does not consume.
func lowerMultiTargetPermanentSpell(
	ctx contentCtx,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if targetCardinalityIsOne(target) ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	return multiTargetPermanentMode(target, primitiveFactory)
}

// multiTargetPermanentMode builds the one-spec/one-primitive-per-slot multi-
// target mode shared by the multi-target permanent verbs. It returns ok=false
// when the target is not an exact multi-target permanent the executable backend
// can represent.
func multiTargetPermanentMode(
	target compiler.CompiledTarget,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, bool) {
	targetSpec, ok := permanentTargetSpecWithCardinality(target)
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: primitiveFactory(game.TargetPermanentReference(i)),
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

func exactMassDestroyGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx)
}

func exactMassExileGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx)
}

// lowerMassOrSinglePermanentSpell lowers a tap or untap effect that is either an
// exact mass group ("Tap all creatures your opponents control.", "Untap all
// creatures you control.") or an exact single permanent target ("Tap target
// creature."). The mass group reuses exactMassGroup unchanged: the tap/untap
// verbs carry no destination, reference, or possessive suffix, so the bare
// mass-group constraints apply just as they do for destroy and exile. The
// resolved group feeds the group constructor (game.Tap{Group}/game.Untap{Group}),
// which the rules engine resolves by tapping or untapping every permanent the
// group matches; the single target falls through to the shared fixed-target
// path. groupPrimitive and objectPrimitive build the same primitive type with
// its Group or Object field set, respectively.
func lowerMassOrSinglePermanentSpell(
	ctx contentCtx,
	verb string,
	groupPrimitive func(group game.GroupReference) game.Primitive,
	objectPrimitive func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{Primitive: groupPrimitive(group)}},
		}.Ability(), nil
	}
	return lowerFixedPermanentTargetSpell(ctx, verb, objectPrimitive)
}

// lowerPhaseOutSpell lowers the "phases out"/"phase out" family (CR 702.26) into
// a single PhaseOut instruction. It supports three recipients, each reusing the
// existing phase-out runtime:
//
//   - every permanent in an exact mass group ("All permanents you control phase
//     out.", Teferi's Protection's effect shape);
//   - a chosen target permanent ("Target creature phases out.", with any
//     supported target restriction), including multi-target forms; and
//   - the ability's own source ("This creature phases out." / "This permanent
//     phases out." / "<CardName> phases out."), which lowers to the source
//     permanent reference and needs no target.
//
// Any other phase-out shape — a negated or non-controller effect, conditional or
// modal content, or an unrepresentable recipient — fails closed.
func lowerPhaseOutSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{Primitive: game.PhaseOut{Group: group}}},
		}.Ability(), nil
	}
	if len(ctx.content.Targets) > 0 {
		return lowerFixedPermanentTargetSpell(ctx, "Phase out", func(object game.ObjectReference) game.Primitive {
			return game.PhaseOut{Object: object}
		})
	}
	if object, ok := lowerSourcePhaseOutObject(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{Primitive: game.PhaseOut{Object: object}}},
		}.Ability(), nil
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported phase out effect",
		"the executable source backend supports only exact phase out of one target, source, or mass group",
	)
}

// lowerSourcePhaseOutObject resolves the non-target phase-out recipient: the
// ability's own source ("This creature phases out." / "<CardName> phases out.").
// It requires a single exact controller effect with no conditional or modal
// content and a single source reference, and fails closed for every other shape.
func lowerSourcePhaseOutObject(ctx contentCtx) (game.ObjectReference, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.ObjectReference{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated || effect.Optional {
		return game.ObjectReference{}, false
	}
	if len(ctx.content.References) != 1 {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
}

// lowerBoundedUntapSpell lowers the "Untap up to N <permanent filter>" family
// ("Untap up to two lands." — Snap, "Untap up to three lands." — Frantic Search,
// "Untap up to two creatures.", "Untap up to one artifact you control.") into a
// ChooseUpTo untap over a battlefield group. The resolving controller chooses up
// to Maximum distinct permanents matching the selector. It accepts only the
// untargeted "up to N" range (Minimum 0) and a selector massGroupSelection can
// express, failing closed otherwise so targeted or unexpressible forms fall
// through to the mass / single-target untap paths.
func lowerBoundedUntapSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		!effect.Amount.RangeKnown ||
		effect.Amount.Minimum != 0 ||
		effect.Amount.Maximum < 1 ||
		effect.Selector.All {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Untap{
			Group:      game.BattlefieldGroup(selection),
			ChooseUpTo: true,
			Amount:     game.Fixed(effect.Amount.Maximum),
		},
	}}}.Ability(), true
}

func lowerUntapSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerBoundedUntapSpell(ctx); ok {
		return content, nil
	}
	return lowerMassOrSinglePermanentSpell(ctx, "Untap", func(group game.GroupReference) game.Primitive {
		return game.Untap{Group: group}
	}, func(object game.ObjectReference) game.Primitive {
		return game.Untap{Object: object}
	})
}

// exactMassBounceGroup mirrors exactMassGroup for the mass return-to-hand
// "Return all <group> to their owners' hands." and the singular "each" form
// "Return each <group> to its owner's hand." The return wording differs from the
// bare destroy/exile mass clause only by its destination suffix, which the
// compiler records as an ambiguous "their"/"its" possessive pronoun reference. A
// counter-qualifier ("without a +1/+1 counter on it") adds a further "it"/"them"
// pronoun that belongs to the group, not a separate object. Those reference
// pronouns are the only ones tolerated; every other reference (and any target,
// condition, or mode) fails closed.
func exactMassBounceGroup(ctx contentCtx) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		!ctx.content.Effects[0].Selector.All ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		!massBounceReferencesOnly(ctx.content.References, ctx.content.Effects[0].Selector) {
		return game.GroupReference{}, false
	}
	if len(ctx.content.Keywords) != 0 {
		return game.GroupReference{}, false
	}
	selection, ok := massGroupSelection(ctx.content.Effects[0].Selector)
	if !ok {
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

// massBounceReferencesOnly reports whether every reference of a mass return is
// one the group bounce addresses directly rather than through the reference. The
// destination possessive ("their owners' hands", "its owner's hand") is always
// tolerated because a group bounce sends every permanent to its own owner's hand
// regardless of how the compiler bound that possessive: an "Each player" mass
// return leaves it ambiguous, while a triggered "When ~ enters, return all ..."
// binds it to the entering permanent, and neither binding changes the per-owner
// resolution. When the group selector carries a counter qualifier ("without a
// +1/+1 counter on it"), the qualifier's trailing "it"/"them" pronoun is part of
// the selected group and is tolerated too. Every other reference fails closed. A
// reference-free group always passes.
func massBounceReferencesOnly(
	references []compiler.CompiledReference,
	selector compiler.CompiledSelector,
) bool {
	hasCounterQualifier := selector.MatchCounter || selector.MatchNoCounters ||
		selector.MatchAnyCounter || selector.MatchExcludedCounter
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun {
			return false
		}
		switch reference.Pronoun {
		case compiler.ReferencePronounTheir, compiler.ReferencePronounIts:
		case compiler.ReferencePronounIt, compiler.ReferencePronounThem:
			if !hasCounterQualifier {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func exactMassGroup(ctx contentCtx) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		!ctx.content.Effects[0].Selector.All ||
		!massCounterQualifierReferencesOnly(ctx.content.References, ctx.content.Effects[0].Selector) {
		return game.GroupReference{}, false
	}
	selection, ok := massGroupSelection(ctx.content.Effects[0].Selector)
	if !ok {
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

// massCounterQualifierReferencesOnly reports whether every reference is the
// trailing "it"/"them" pronoun that belongs to a counter qualifier ("each
// creature with a +1/+1 counter on it", "all creatures with no counters on
// them"). That pronoun is part of the selected group, not a separate object, so
// when the group selector carries a counter qualifier it is the only reference
// the mass group tolerates; with no counter selector, or any other reference,
// this fails closed. A reference-free group always passes.
func massCounterQualifierReferencesOnly(
	references []compiler.CompiledReference,
	selector compiler.CompiledSelector,
) bool {
	if len(references) == 0 {
		return true
	}
	if !selector.MatchCounter && !selector.MatchNoCounters {
		return false
	}
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			(reference.Pronoun != compiler.ReferencePronounIt &&
				reference.Pronoun != compiler.ReferencePronounThem) {
			return false
		}
	}
	return true
}

// massGroupSelection projects the battlefield-group selector of a mass effect
// ("Destroy all artifacts", "Each creature you control ...") onto a Selection.
// It is the canonical projector restricted to the dimensions a mass-effect
// group can express: the per-object token, historic, and source-relative-power
// qualifiers belong to other contexts and never reach a mass group, so the mask
// drops them. An excluded creature subtype is honored for "non-<subtype>" board
// wipes.
func massGroupSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	return SelectionForSelectorMasked(selector, massGroupSelectionMask)
}

// massGroupSelectionMask drops the canonical dimensions a mass-effect group
// never carries: per-object token state, the historic disjunction, and the
// source-relative power comparison. An excluded creature subtype IS honored: a
// "Destroy all non-Dragon creatures" board wipe carries one excluded subtype the
// group must emit so the runtime spares the named subtype.
var massGroupSelectionMask = SelectionMask{}.Ignoring(
	DimNonToken,
	DimTokenOnly,
	DimHistoric,
	DimPowerVsSource,
).Rejecting(
	DimRequiredName,
)

func massGroupRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true
	case compiler.SelectorBattle:
		return types.Battle, true
	default:
		return "", false
	}
}

// lowerControllerAndReferencedPlayerDraw lowers a "You and that player each draw
// N cards" trigger body: the controller and the triggering event's player ("that
// player") each draw. The amount is either a fixed count ("each draw two cards",
// Black Widow) or the triggering event's quantity ("each draw that many cards"
// after a combat-damage trigger, Diviner Spirit). It is modeled as two parallel
// draw instructions, one for the controller and one for the event player.
func lowerControllerAndReferencedPlayerDraw(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	if effect.Negated || effect.Optional || ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 {
		return unsupported()
	}
	// The recipient "that player" must bind the triggering event's player. Any
	// remaining references are the trigger's own source ("Whenever <self> deals
	// combat damage ..."), which carries no recipient and is tolerated.
	recipientRefs := 0
	for _, reference := range ctx.content.References {
		if reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingEventPlayer {
			recipientRefs++
			continue
		}
		if reference.Binding == compiler.ReferenceBindingSource {
			continue
		}
		return unsupported()
	}
	if recipientRefs != 1 {
		return unsupported()
	}
	amount, ok := controllerAndReferencedPlayerDrawAmount(ctx, effect.Amount)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: amount, Player: game.ControllerReference()}},
			{Primitive: game.Draw{Amount: amount, Player: game.EventPlayerReference()}},
		},
	}.Ability(), nil
}

// controllerAndReferencedPlayerDrawAmount resolves the per-player draw count of a
// "you and that player each draw ..." body. It accepts a fixed count and the
// triggering-event quantity anaphor ("that many cards"), failing closed on every
// other dynamic form.
func controllerAndReferencedPlayerDrawAmount(ctx contentCtx, compiled compiler.CompiledAmount) (game.Quantity, bool) {
	switch {
	case compiled.Known:
		if compiled.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(compiled.Value), true
	case triggeringEventQuantityKind(compiled.DynamicKind):
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, compiled)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}

// lowerControllerAndTargetDraw lowers a "You and target <player> each draw N
// cards" body: the controller and the single player target each draw, modeled as
// two parallel draw instructions sharing the mode's player target.
func lowerControllerAndTargetDraw(ctx contentCtx, amount game.Quantity) (game.AbilityContent, *shared.Diagnostic) {
	target, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: amount, Player: game.ControllerReference()}},
			{Primitive: game.Draw{Amount: amount, Player: game.TargetPlayerReference(0)}},
		},
	}.Ability(), nil
}

func lowerFixedDrawSpell(
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Context == parser.EffectContextControllerAndReferencedPlayer {
		return lowerControllerAndReferencedPlayerDraw(ctx, effect)
	}
	// Allow a single EventPlayer reference for "They draw N card(s)." bodies;
	// reject all other non-zero-reference forms.
	hasEventPlayerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPlayer
	hasReferencedControllerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.Context == parser.EffectContextReferencedObjectController
	hasSourceCounterRef := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount &&
		singleSelfReference(ctx.content.References)
	// "When this creature leaves the battlefield, draw a card for each +1/+1
	// counter on it." (Bloodtracker) counts the +1/+1 counters on the triggering
	// permanent. In a zone-change/dies trigger the "it"/"them" of the counter
	// phrase binds to the event permanent rather than the live source, so the
	// counted amount reads that permanent's last-known counters once it has left
	// the battlefield (CR 603.10, CR 122).
	hasEventCounterRef := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Kind == compiler.ReferencePronoun &&
		(ctx.content.References[0].Pronoun == compiler.ReferencePronounIt ||
			ctx.content.References[0].Pronoun == compiler.ReferencePronounThem) &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent
	// "Draw a card for each creature you control with a +1/+1 counter on it."
	// counts a counter-qualified group; the qualifier's trailing "it"/"them" is
	// part of the counted selection, not a separate recipient, so a single such
	// reference is tolerated rather than rejected.
	hasCountCounterRef := effect.Amount.DynamicKind == compiler.DynamicAmountCount &&
		effect.Amount.Selector().MatchCounter &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Kind == compiler.ReferencePronoun &&
		(ctx.content.References[0].Pronoun == compiler.ReferencePronounIt ||
			ctx.content.References[0].Pronoun == compiler.ReferencePronounThem)
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef && !hasReferencedControllerRef && !hasSourceCounterRef && !hasCountCounterRef && !hasEventCounterRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
	case triggeringEventQuantityKind(effect.Amount.DynamicKind):
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, effect.Amount)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		counterObject := game.SourcePermanentReference()
		if hasEventCounterRef {
			counterObject = game.EventPermanentReference()
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, counterObject)
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	default:
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	if len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 {
		switch effect.Context {
		case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: amount, PlayerGroup: game.OpponentsReference()},
				}},
			}.Ability(), nil
		case parser.EffectContextEachPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: amount, PlayerGroup: game.AllPlayersReference()},
				}},
			}.Ability(), nil
		}
	}
	switch {
	case effect.Context == parser.EffectContextControllerAndTarget &&
		len(ctx.content.Targets) == 1 && effect.Amount.Known:
		return lowerControllerAndTargetDraw(ctx, amount)
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		(effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer) &&
		effect.Amount.Known:
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		!hasEventPlayerRef &&
		effect.Context == parser.EffectContextController:
	case hasReferencedControllerRef && len(ctx.content.Targets) == 1 && effect.Amount.Known:
		ref, ok := referencedControllerPlayerRef(ctx)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported draw spell", "the executable source backend supports only exact fixed card draw")
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		target, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact fixed card draw",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		target.Constraint = "target player"
		targets = []game.TargetSpec{target}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: game.Draw{
					Amount: amount,
					Player: playerRef,
				},
			},
		},
	}.Ability(), nil
}

// referencedControllerPlayerRef resolves the recipient player for an "Its
// controller <effect>" body whose subject is the controller of the inherited
// antecedent target in an ordered sequence. The antecedent target's selector
// drives the object reference kind: a permanent target (including bare subtype
// and compound type leads) yields a permanent reference, a spell on the stack
// yields a stack-object reference (so a counterspell's "its controller" resolves
// the countered spell's controller). It returns false (fail closed) for any
// other shape or antecedent kind. The embedded clause-local target index is
// rebased by the sequence machinery.
func referencedControllerPlayerRef(ctx contentCtx) (game.PlayerReference, bool) {
	if len(ctx.content.Effects) == 0 ||
		ctx.content.Effects[0].Context != parser.EffectContextReferencedObjectController ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingTarget ||
		ctx.content.References[0].Occurrence < 0 ||
		len(ctx.content.Targets) != 1 {
		return game.PlayerReference{}, false
	}
	occ := ctx.content.References[0].Occurrence
	object, ok := inheritedRemovalTargetObjectRef(ctx.content.Targets[0], occ)
	if !ok {
		return game.PlayerReference{}, false
	}
	return game.ObjectControllerReference(object), true
}

// hasThatPlayerTargetReference reports whether the clause carries a "that
// player" reference bound to an inherited antecedent target, the typed shape an
// ordered-sequence life clause takes when its subject ("That player loses N
// life.") is the player established by the preceding clause's target.
func hasThatPlayerTargetReference(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingTarget {
			return true
		}
	}
	return false
}

// referencedThatPlayerRef resolves the recipient player for a "That player
// <gains/loses> N life" body whose subject is the inherited antecedent target of
// an ordered sequence. A player target ("Target opponent ... That player loses N
// life.") denotes that player directly; a permanent target ("Destroy target
// creature an opponent controls. That player loses N life.", including bare
// subtype and compound type leads) denotes its controller. It returns false
// (fail closed) for any other antecedent kind. The single inherited target sits
// at clause-local index 0.
func referencedThatPlayerRef(target compiler.CompiledTarget) (game.PlayerReference, bool) {
	switch target.Selector.Kind {
	case compiler.SelectorPlayer, compiler.SelectorOpponent:
		return game.TargetPlayerReference(0), true
	}
	object, ok := inheritedRemovalTargetObjectRef(target, 0)
	if !ok {
		return game.PlayerReference{}, false
	}
	return game.ObjectControllerReference(object), true
}
