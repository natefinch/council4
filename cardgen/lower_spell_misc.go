package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
	case effect.Amount.DynamicKind == compiler.DynamicAmountTriggeringLifeChange:
		dynamic, ok := lowerEventLifeChangeAmount(ctx, effect.Amount)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		sourceCounterReferences := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount &&
			singleSelfReference(ctx.content.References)
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower ||
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
	if content, ok := lowerPlayerGraveyardExile(ctx); ok {
		return content, nil
	}
	if content, ok := lowerAllGraveyardExile(ctx); ok {
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

// lowerGroupMustAttack lowers the one-shot, turn-scoped forced-attack effect
// "<group> attack this turn if able." (Bident of Thassa: "Creatures your
// opponents control attack this turn if able.") to an ApplyRule that forces the
// affected creatures to attack for the rest of the turn. The affected creature
// group is read from the parser-recognized StaticSubject and mapped to a
// controller relation; the rule reuses the continuous RuleEffectMustAttack rule
// effect with a this-turn duration. Targets, references, conditions, modes, a
// negation, an amount, or an unsupported group subject fail closed.
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
			"the executable source backend supports only the exact you/opponents/all creatures forced-attack effect this turn",
		)
	}
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
			"unsupported forced-attack effect",
			"the executable source backend supports only the exact you/opponents/all creatures forced-attack effect this turn",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectMustAttack,
				AffectedController: controller,
				PermanentTypes:     []types.Card{types.Creature},
			}},
			Duration: game.DurationThisTurn,
		},
	}}}.Ability(), nil
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
	case compiler.EffectSpellsCantBeCountered:
		content, diagnostic := lowerSpellsCantBeCountered(ctx)
		return content, diagnostic, true
	case compiler.EffectMustAttack:
		content, diagnostic := lowerGroupMustAttack(ctx)
		return content, diagnostic, true
	case compiler.EffectPhaseOut:
		content, diagnostic := lowerMassOrSinglePermanentSpell(ctx, "Phase out", func(group game.GroupReference) game.Primitive {
			return game.PhaseOut{Group: group}
		}, func(object game.ObjectReference) game.Primitive {
			return game.PhaseOut{Object: object}
		})
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
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
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

// lowerMultiTargetExileSpell lowers exile abilities whose single permanent
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
// "Return all <group> to their owners' hands." The return wording differs from
// the bare destroy/exile mass clause only by its "to their owners' hands"
// destination suffix, which the compiler records as a single ambiguous "their"
// possessive pronoun reference. The bounced objects come from the group, not
// that reference, so the possessive is the only reference tolerated; every other
// reference (and any target, condition, or mode) fails closed.
func exactMassBounceGroup(ctx contentCtx) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		!ctx.content.Effects[0].Selector.All ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		!bounceDestinationPossessiveReferencesOnly(ctx.content.References) {
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

func exactMassGroup(ctx contentCtx) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		!ctx.content.Effects[0].Selector.All {
		return game.GroupReference{}, false
	}
	selection, ok := massGroupSelection(ctx.content.Effects[0].Selector)
	if !ok {
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func massGroupSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Zone != zone.None ||
		selector.BasicLandType ||
		selector.PlayerOrPlaneswalker ||
		len(selector.SourceTypes()) != 0 ||
		(selector.Tapped && selector.Untapped) {
		return game.Selection{}, false
	}
	selection := game.Selection{
		RequiredTypesAny: append([]types.Card(nil), selector.RequiredTypesAny()...),
		ExcludedTypes:    append([]types.Card(nil), selector.ExcludedTypes()...),
		Supertypes:       append([]types.Super(nil), selector.Supertypes()...),
		SubtypesAny:      append([]types.Sub(nil), selector.SubtypesAny()...),
		ColorsAny:        append([]color.Color(nil), selector.ColorsAny()...),
		ExcludedColors:   append([]color.Color(nil), selector.ExcludedColors()...),
		Colorless:        selector.Colorless,
		Multicolored:     selector.Multicolored,
		EnteredThisTurn:  selector.EnteredThisTurn,
		ExcludeSource:    selector.Another || selector.Other,
	}
	if excludedSupertypes := selector.ExcludedSupertypes(); len(excludedSupertypes) > 1 {
		return game.Selection{}, false
	} else if len(excludedSupertypes) == 1 {
		selection.ExcludedSupertype = excludedSupertypes[0]
	}
	for _, alternative := range selector.Alternatives {
		lowered, ok := massGroupSelection(alternative)
		if !ok {
			return game.Selection{}, false
		}
		selection.AnyOf = append(selection.AnyOf, lowered)
	}
	if len(selection.RequiredTypesAny) == 0 {
		if requiredType, ok := massGroupRequiredType(selector.Kind); ok {
			selection.RequiredTypes = []types.Card{requiredType}
		} else if selector.Kind == compiler.SelectorUnknown {
			// A bare subtype noun ("Destroy all Islands.") selects any permanent
			// carrying that subtype with no card-type restriction; the subtype
			// filter supplies the constraint. Without one, an unrecognized mass
			// noun has no representable predicate and fails closed.
			if len(selection.SubtypesAny) == 0 {
				return game.Selection{}, false
			}
		} else if selector.Kind != compiler.SelectorPermanent {
			return game.Selection{}, false
		}
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		selection.Controller = game.ControllerNotYou
	default:
		return game.Selection{}, false
	}
	switch {
	case selector.Attacking && selector.Blocking:
		selection.CombatState = game.CombatStateAttackingOrBlocking
	case selector.Attacking:
		selection.CombatState = game.CombatStateAttacking
	case selector.Blocking:
		selection.CombatState = game.CombatStateBlocking
	default:
	}
	switch {
	case selector.Tapped:
		selection.Tapped = game.TriTrue
	case selector.Untapped:
		selection.Tapped = game.TriFalse
	default:
	}
	if selector.MatchManaValue {
		if selector.ManaValueX {
			return game.Selection{}, false
		}
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower {
		selection.Power = opt.Val(selector.Power)
	}
	if selector.MatchToughness {
		selection.Toughness = opt.Val(selector.Toughness)
	}
	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if selector.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.ExcludedKeyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.ExcludedKeyword = keyword
	}
	if selector.MatchCounter {
		selection.MatchCounter = true
		selection.RequiredCounter = selector.RequiredCounter
	}
	if selector.MatchAnyCounter {
		selection.MatchAnyCounter = true
	}
	switch {
	case selector.SubtypeFromChosenTypeExcluded:
		selection.SubtypeChoice = game.SubtypeChoiceResolutionExcluded
	case selector.SubtypeFromChosenType:
		selection.SubtypeChoice = game.SubtypeChoiceResolution
	case selector.SubtypeFromEntryChoice:
		selection.SubtypeChoice = game.SubtypeChoiceSourceEntry
	default:
	}
	// "each artifact creature you control" names two card types the permanent
	// must carry at once, so its type set lowers to the conjunctive RequiredTypes
	// (all-of) filter rather than the default any-of RequiredTypesAny union.
	if selector.ConjunctiveTypes {
		selection.RequiredTypes = selection.RequiredTypesAny
		selection.RequiredTypesAny = nil
	}
	if len(selection.Validate()) != 0 {
		return game.Selection{}, false
	}
	return selection, true
}

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
	case effect.Amount.DynamicKind == compiler.DynamicAmountEventCardCount:
		dynamic, ok := lowerEventCardCountAmount(ctx, effect.Amount)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	case effect.Amount.DynamicKind == compiler.DynamicAmountTriggeringCounterCount:
		dynamic, ok := lowerEventCounterCountAmount(ctx, effect.Amount)
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
// kind drives the object reference kind: a permanent target yields a permanent
// reference, a spell on the stack yields a stack-object reference (so a
// counterspell's "its controller" resolves the countered spell's controller). It
// returns false (fail closed) for any other shape or antecedent kind. The
// embedded clause-local target index is rebased by the sequence machinery.
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
	switch ctx.content.Targets[0].Selector.Kind {
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment,
		compiler.SelectorLand, compiler.SelectorPermanent, compiler.SelectorPlaneswalker,
		compiler.SelectorBattle:
		return game.ObjectControllerReference(game.TargetPermanentReference(occ)), true
	case compiler.SelectorSpell:
		return game.ObjectControllerReference(game.TargetStackObjectReference(occ)), true
	default:
		return game.PlayerReference{}, false
	}
}
