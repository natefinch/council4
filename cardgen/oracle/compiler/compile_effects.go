package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileTrigger(ability *parser.Ability, _ Context) CompiledTrigger {
	trigger := CompiledTrigger{
		Kind: TriggerUnknown,
	}
	if ability.Trigger == nil {
		return trigger
	}
	trigger.Span = ability.Trigger.Span
	trigger.Order = ability.Trigger.Order
	trigger.Text = ability.Trigger.Text
	trigger.Event = ability.Trigger.Event
	if frequency := ability.TriggerFrequency; frequency != nil {
		switch frequency.Kind {
		case parser.TriggerFrequencyOncePerTurn:
			trigger.MaxTriggersPerTurn = 1
			trigger.MaxTriggersPerTurnSpan = frequency.Span
		case parser.TriggerFrequencyTwicePerTurn:
			trigger.MaxTriggersPerTurn = 2
			trigger.MaxTriggersPerTurnSpan = frequency.Span
		default:
		}
	}
	if event := ability.Trigger.TriggerEvent; event != nil && event.FirstTimeEachTurn {
		trigger.MaxTriggersPerTurn = 1
		trigger.MaxTriggersPerTurnSpan = event.FirstTimeEachTurnSpan
	}
	switch ability.Trigger.Introduction.Kind {
	case parser.TriggerIntroductionWhen:
		trigger.Kind = TriggerWhen
	case parser.TriggerIntroductionWhenever:
		trigger.Kind = TriggerWhenever
	case parser.TriggerIntroductionAt:
		trigger.Kind = TriggerAt
	default:
	}
	conditions := compileConditions(ability.TriggerConditionSegments, ability.ConditionClauses, ability.EventHistoryConditions)
	for i := range conditions {
		if conditions[i].Intervening {
			condition := conditions[i]
			trigger.Condition = &condition
			break
		}
	}
	switch {
	case ability.Trigger.PhaseStep != nil:
		trigger.Pattern = compilePhaseStepTriggerPattern(
			ability.Trigger.PhaseStep,
			trigger.Kind,
			trigger.Condition,
		)
	case ability.Trigger.PlayerEvent != nil:
		trigger.Pattern = compilePlayerEventTriggerPattern(
			ability.Trigger.PlayerEvent,
			trigger.Kind,
			trigger.Condition,
		)
	case ability.Trigger.TriggerEvent != nil:
		trigger.Pattern = compileTriggerEventPattern(
			ability.Trigger.TriggerEvent,
			trigger.Kind,
			trigger.Condition,
		)
	default:
		trigger.Pattern = TriggerPattern{
			Span:                 ability.Trigger.EventSpan,
			Kind:                 trigger.Kind,
			InterveningCondition: trigger.Condition,
		}
	}
	return trigger
}

func runtimeCardTypeFromParser(cardType parser.CardType) (types.Card, bool) {
	switch cardType {
	case parser.CardTypeArtifact:
		return types.Artifact, true
	case parser.CardTypeBattle:
		return types.Battle, true
	case parser.CardTypeCreature:
		return types.Creature, true
	case parser.CardTypeEnchantment:
		return types.Enchantment, true
	case parser.CardTypeInstant:
		return types.Instant, true
	case parser.CardTypeLand:
		return types.Land, true
	case parser.CardTypePlaneswalker:
		return types.Planeswalker, true
	case parser.CardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}

func runtimeSupertypeFromParser(supertype parser.Supertype) (types.Super, bool) {
	switch supertype {
	case parser.SupertypeBasic:
		return types.Basic, true
	case parser.SupertypeLegendary:
		return types.Legendary, true
	case parser.SupertypeSnow:
		return types.Snow, true
	case parser.SupertypeWorld:
		return types.World, true
	default:
		return "", false
	}
}

func runtimeColorFromParser(colorValue parser.Color) (color.Color, bool) {
	switch colorValue {
	case parser.ColorWhite:
		return color.White, true
	case parser.ColorBlue:
		return color.Blue, true
	case parser.ColorBlack:
		return color.Black, true
	case parser.ColorRed:
		return color.Red, true
	case parser.ColorGreen:
		return color.Green, true
	default:
		return "", false
	}
}

// polymorphColors converts the parser's set polymorph colors to runtime colors,
// dropping any unrecognized color. The parser only yields colors recognized from
// atoms, so a well-formed EffectPolymorph never drops a color here.
func polymorphColors(colors []parser.Color) []color.Color {
	result := make([]color.Color, 0, len(colors))
	for _, parserColor := range colors {
		runtimeColor, ok := runtimeColorFromParser(parserColor)
		if !ok {
			continue
		}
		result = append(result, runtimeColor)
	}
	return result
}

// compilePreventDamageSourceColors maps the optional source color filter on a
// "next time a [color] source ... prevent that damage" clause, returning nil
// when the clause carries no color so unrelated effects keep an absent field.
func compilePreventDamageSourceColors(colors []parser.Color) []color.Color {
	if len(colors) == 0 {
		return nil
	}
	return compileParserColors(colors)
}

// compileParserColors converts a parser color list to runtime colors, dropping
// any unrecognized color. The parser only yields the five basic colors, so a
// well-formed effect never drops a color here.
func compileParserColors(colors []parser.Color) []color.Color {
	result := make([]color.Color, 0, len(colors))
	for _, parserColor := range colors {
		runtimeColor, ok := runtimeColorFromParser(parserColor)
		if !ok {
			continue
		}
		result = append(result, runtimeColor)
	}
	return result
}

// polymorphSupertypes converts the parser's added polymorph supertypes
// ("legendary") to runtime supertypes, dropping any unrecognized supertype. The
// parser only yields supertypes recognized from atoms, so a well-formed
// named-become polymorph never drops a supertype here.
func polymorphSupertypes(supertypes []parser.Supertype) []types.Super {
	result := make([]types.Super, 0, len(supertypes))
	for _, parserSupertype := range supertypes {
		runtimeSupertype, ok := compilerSupertype(parserSupertype)
		if !ok {
			continue
		}
		result = append(result, runtimeSupertype)
	}
	return result
}

// compileConditions builds the semantic conditions for an ability or mode from
// the parser's pre-segmented condition clauses. The parser owns introducer
// recognition, clause segmentation, and rendering; the compiler consumes each
// segment's typed kind, span, and rendered text mechanically and derives no
// meaning from Oracle wording.
func compileConditions(
	segments []parser.ConditionSegment,
	clauses []parser.ConditionClause,
	eventHistories []parser.EventHistoryCondition,
) []CompiledCondition {
	var conditions []CompiledCondition
	for _, segment := range segments {
		condition := CompiledCondition{
			Kind:                  compileConditionIntro(segment.Kind),
			Span:                  segment.Span,
			Text:                  segment.Text,
			Intervening:           segment.Intervening,
			Resolving:             segment.Resolving,
			ActivationKeywordSpan: segment.ActivationKeyword,
			NodeID:                segment.NodeID,
			ClauseIndex:           segment.ClauseIndex,
			EventHistoryIndex:     segment.EventHistoryIndex,
			Order:                 segment.Order,
		}
		recognizeCondition(&condition, clauses, eventHistories)
		conditions = append(conditions, condition)
	}
	return conditions
}

func compileConditionIntro(kind parser.ConditionIntroKind) ConditionKind {
	switch kind {
	case parser.ConditionIntroIf:
		return ConditionIf
	case parser.ConditionIntroUnless:
		return ConditionUnless
	case parser.ConditionIntroOnlyIf:
		return ConditionOnlyIf
	case parser.ConditionIntroAsLongAs:
		return ConditionAsLongAs
	default:
		return ConditionUnknown
	}
}

func compileEffects(sentences []parser.Sentence) []CompiledEffect {
	var effects []CompiledEffect
	for _, sentence := range sentences {
		if sentence.StaticRule != nil {
			if effect, ok := compileStaticRuleEffect(sentence); ok {
				effects = append(effects, effect)
			}
			continue
		}
		for syntaxIndex := range sentence.Effects {
			syntax := &sentence.Effects[syntaxIndex]
			effects = append(effects, CompiledEffect{
				Kind:                 compileEffectKind(syntax.Kind),
				Context:              syntax.Context,
				Connection:           syntax.Connection,
				ConnectionSpan:       syntax.ConnectionSpan,
				Span:                 syntax.Span,
				ClauseSpan:           syntax.ClauseSpan,
				Text:                 syntax.Text,
				VerbSpan:             syntax.VerbSpan,
				Player:               syntax.Player,
				CardSource:           syntax.CardSource,
				RequirePermanentCard: syntax.RequirePermanentCard,
				References:           compileTypedReferences(syntax.References),
				SubjectReferences:    compileTypedReferences(syntax.SubjectReferences),
				Targets:              compileTypedTargetList(syntax.Targets),
				SubjectTargets:       compileTypedTargetList(syntax.SubjectTargets),
				Duration:             compileEffectDuration(syntax.Duration),
				DelayedTiming:        compileDelayedTiming(syntax.DelayedTiming),
				Selector:             compileTypedSelection(syntax.Selection),
				DamageRecipient: CompiledDamageRecipient{
					GroupSelectors:  compileDamageRecipientSelectors(syntax.DamageRecipient.Groups),
					Reference:       syntax.DamageRecipient.Reference,
					EachSourceGroup: compileTypedSelection(syntax.DamageRecipient.EachSourceGroup),
					EachSourceRole:  syntax.DamageRecipient.EachSourceRole,
				},
				DamageRiders:                    syntax.DamageRiders,
				Amount:                          compileTypedAmount(syntax.Amount),
				PowerDelta:                      compileSignedAmount(syntax.PowerDelta),
				ToughnessDelta:                  compileSignedAmount(syntax.ToughnessDelta),
				TokenPower:                      syntax.TokenPower,
				TokenToughness:                  syntax.TokenToughness,
				TokenPTKnown:                    syntax.TokenPTKnown,
				TokenPTVariableX:                syntax.TokenPTVariableX,
				TokenPTDynamic:                  syntax.TokenPTDynamic,
				TokenKeywords:                   append([]parser.KeywordKind(nil), syntax.TokenKeywords...),
				TokenGrantedAbility:             syntax.TokenGrantedAbility,
				GainGrantedAbility:              syntax.GainGrantedAbility,
				EmblemAbilities:                 syntax.EmblemAbilities,
				DelayedTriggerAbility:           syntax.DelayedTriggerAbility,
				PayRepeatedlyAnimate:            syntax.PayRepeatedlyAnimate,
				AnimateSelf:                     syntax.AnimateSelf,
				AnimateTarget:                   syntax.AnimateTarget,
				DelayedTriggerOneShot:           syntax.DelayedTriggerOneShot,
				DelayedTriggerBindDamageSource:  syntax.DelayedTriggerBindDamageSource,
				TokenName:                       syntax.TokenName,
				TokenPredefinedName:             syntax.TokenPredefinedName,
				AmassSubtype:                    syntax.AmassSubtype,
				TokenCopyOfTarget:               syntax.TokenCopyOfTarget,
				TokenCopyOfReference:            syntax.TokenCopyOfReference,
				TokenCopyOfAttached:             syntax.TokenCopyOfAttached,
				TokenCopyOfTriggeringSet:        syntax.TokenCopyOfTriggeringSet,
				TokenCopyDropLegendary:          syntax.TokenCopyDropLegendary,
				TokenCopyEntersTapped:           syntax.TokenCopyEntersTapped,
				TokenCopyGrantKeywords:          append([]parser.KeywordKind(nil), syntax.TokenCopyGrantKeywords...),
				TokenCopyGrantRiderSpan:         syntax.TokenCopyGrantRiderSpan,
				TokenCopyOverride:               syntax.TokenCopyOverride,
				TokenCopyOverridePTKnown:        syntax.TokenCopyOverridePTKnown,
				TokenCopyOverridePower:          syntax.TokenCopyOverridePower,
				TokenCopyOverrideToughness:      syntax.TokenCopyOverrideToughness,
				TokenCopyOverrideColors:         compileParserColors(syntax.TokenCopyOverrideColors),
				TokenCopyOverrideSubtypes:       append([]types.Sub(nil), syntax.TokenCopyOverrideSubtypes...),
				TokenCopyOverrideTypes:          append([]types.Card(nil), syntax.TokenCopyOverrideTypes...),
				TokenCopyOverrideKeywords:       append([]parser.KeywordKind(nil), syntax.TokenCopyOverrideKeywords...),
				TokenCopyOverrideAdditiveTypes:  syntax.TokenCopyOverrideAdditiveTypes,
				TokenCopyOverrideAdditiveColors: syntax.TokenCopyOverrideAdditiveColors,
				TokenChoice:                     syntax.TokenChoice,
				AdditionalTokens:                compileEffects([]parser.Sentence{{Effects: syntax.AdditionalTokens}}),
				StaticSubject:                   compileStaticSubjectKind(syntax.StaticSubject.Kind),
				StaticSubjectSpan:               syntax.StaticSubject.Span,
				Details: compiledEffectDetails(
					staticSubjectType(syntax.StaticSubject.SubtypeText, syntax.StaticSubject.Subtype, syntax.StaticSubject.SubtypesAny, syntax.StaticSubject.SubtypeKnown, syntax.StaticSubject.ExcludedSubtype),
					staticSubjectColors(syntax.StaticSubject.Colors, syntax.StaticSubject.Colorless, syntax.StaticSubject.Multicolored, syntax.StaticSubject.ChosenColorFromEntry),
					staticSubjectKeyword(syntax.StaticSubject.Keyword, syntax.StaticSubject.ExcludedKeyword),
					staticSubjectCounter(syntax.StaticSubject.CounterRequired, syntax.StaticSubject.CounterKind, syntax.StaticSubject.CounterAny),
					syntax.Symbol,
				),
				CounterKind:                  syntax.CounterKind,
				CounterKindKnown:             syntax.CounterKnown,
				CounterKindChoices:           append([]counter.Kind(nil), syntax.CounterKindChoices...),
				CounterRecipientAttached:     syntax.CounterRecipientAttached,
				FightSubjectAttached:         syntax.FightSubjectAttached,
				CounterRecipientSingleChoice: syntax.CounterRecipientSingleChoice,
				RegenerateAttached:           syntax.RegenerateAttached,
				MoveCountersAll:              syntax.MoveCountersAll,
				RemoveCountersAll:            syntax.RemoveCountersAll,
				MoveCountersDistribute:       syntax.MoveCountersDistribute,
				MoveThoseCounters:            syntax.MoveThoseCounters,
				MoveCountersFromTarget:       syntax.MoveCountersFromTarget,
				MoveCountersAnyKind:          syntax.MoveCountersAnyKind,
				FromZone:                     syntax.FromZone,
				GraveyardZoneExile:           syntax.GraveyardZoneExile,
				ToZone:                       syntax.ToZone,
				Destination:                  syntax.Destination,
				EntersTapped:                 syntax.EntersTapped,
				EntersTappedSelf:             syntax.EntersTappedSelf,
				GroupEntryModification:       compileGroupEntryModification(syntax.GroupEntryModification),
				EntersColorChoice:            syntax.EntersColorChoice,
				EntersColorChoiceExclude:     syntax.EntersColorChoiceExclude,
				EntersTypeChoice:             syntax.EntersTypeChoice,
				EntersDevour:                 syntax.EntersDevour,
				EntersDevourMultiplier:       syntax.EntersDevourMultiplier,
				EntersDevourType:             syntax.EntersDevourType,
				EntersDevourSubtype:          syntax.EntersDevourSubtype,
				EntersTribute:                syntax.EntersTribute,
				EntersTributeCount:           syntax.EntersTributeCount,
				EntersAsCopy:                 syntax.EntersAsCopy,
				EntersAsCopyOptional:         syntax.EntersAsCopyOptional,
				EntersAsCopyNotLegendary:     syntax.EntersAsCopyNotLegendary,
				EntersAsCopyAddTypes:         slices.Clone(syntax.EntersAsCopyAddTypes),
				EntersAsCopyAddSubtypes:      slices.Clone(syntax.EntersAsCopyAddSubtypes),

				EntersAsCopyConditionalCounters: slices.Clone(syntax.EntersAsCopyConditionalCounters),
				EntersAsCopyUntilEndOfTurn:      syntax.EntersAsCopyUntilEndOfTurn,
				EntersAsCopyAddKeywords:         slices.Clone(syntax.EntersAsCopyAddKeywords),
				EntersAsCopyTapped:              syntax.EntersAsCopyTapped,

				BecomeCopyUntilEndOfTurn:     syntax.BecomeCopyUntilEndOfTurn,
				BecomeCopyRetainsThisAbility: syntax.BecomeCopyRetainsThisAbility,
				BecomeCopyAddKeywords:        slices.Clone(syntax.BecomeCopyAddKeywords),

				BecomeTypeAddTypes:       slices.Clone(syntax.BecomeTypeAddTypes),
				BecomeTypeAddColors:      polymorphColors(syntax.BecomeTypeAddColors),
				BecomeTypeUntilEndOfTurn: syntax.BecomeTypeUntilEndOfTurn,

				PolymorphColors:        polymorphColors(syntax.PolymorphColors),
				PolymorphColorless:     syntax.PolymorphColorless,
				PolymorphSubtypes:      slices.Clone(syntax.PolymorphSubtypes),
				PolymorphBasePower:     syntax.PolymorphBasePower,
				PolymorphBaseToughness: syntax.PolymorphBaseToughness,
				PolymorphName:          syntax.PolymorphName,
				PolymorphSupertypes:    polymorphSupertypes(syntax.PolymorphSupertypes),
				PolymorphPermanent:     syntax.PolymorphPermanent,

				SetBasePower:               syntax.SetBasePower,
				SetBaseToughness:           syntax.SetBaseToughness,
				SetBasePTVariableX:         syntax.SetBasePTVariableX,
				SetBasePTEveryCreatureType: syntax.SetBasePTEveryCreatureType,
				SetBasePTSource:            syntax.SetBasePTSource,
				SwitchPTSource:             syntax.SwitchPTSource,

				EntersWithCounters:        syntax.EntersWithCounters,
				UnderYourControl:          syntax.UnderYourControl,
				UnderOwnersControl:        syntax.UnderOwnersControl,
				TokenCopyOfForEach:        syntax.TokenCopyOfForEach,
				TokenCopyForEachGroup:     compileTokenCopyForEachGroup(syntax.TokenCopyForEachGroup),
				CastAsAdventure:           syntax.CastAsAdventure,
				CastWithoutPayingManaCost: syntax.CastWithoutPayingManaCost,
				PlayHideawayExiledCard:    syntax.PlayHideawayExiledCard,
				Negated:                   syntax.Negated,
				FallbackOnInability:       syntax.FallbackOnInability,
				Optional:                  syntax.Optional,
				Divided:                   syntax.Divided,
				OptionalSpan:              syntax.OptionalSpan,
				LifeObject:                syntax.LifeObject,
				Mana: CompiledEffectMana{
					Span:                     syntax.Mana.Span,
					Symbols:                  slices.Clone(syntax.Mana.Symbols),
					Colors:                   slices.Clone(syntax.Mana.Colors),
					ColorsKnown:              syntax.Mana.ColorsKnown,
					Choice:                   syntax.Mana.Choice,
					AnyColor:                 syntax.Mana.AnyColor,
					ChosenColor:              syntax.Mana.ChosenColor,
					ChosenColorFixed:         syntax.Mana.ChosenColorFixed,
					ChosenColorFixedKnown:    syntax.Mana.ChosenColorFixedKnown,
					ChosenColorDevotion:      syntax.Mana.ChosenColorDevotion,
					ChosenColorDynamic:       syntax.Mana.ChosenColorDynamic,
					CommanderIdentity:        syntax.Mana.CommanderIdentity,
					DynamicColorless:         syntax.Mana.DynamicColorless,
					LegacyBodyExact:          syntax.Mana.LegacyBodyExact,
					FilterPair:               syntax.Mana.FilterPair,
					FilterColors:             slices.Clone(syntax.Mana.FilterColors),
					LandsProduce:             syntax.Mana.LandsProduce,
					LandsProduceScope:        syntax.Mana.LandsProduceScope,
					LandsProduceAnyType:      syntax.Mana.LandsProduceAnyType,
					LinkedExileColors:        syntax.Mana.LinkedExileColors,
					ColorsAmongControlled:    syntax.Mana.ColorsAmongControlled,
					ColorsAmongSelector:      compileColorsAmongSelector(syntax.Mana.ColorsAmongSelection),
					EachColorAmongControlled: syntax.Mana.EachColorAmongControlled,
					AnyOneColorDynamic:       syntax.Mana.AnyOneColorDynamic,
					AnyColorCount:            syntax.Mana.AnyColorCount,
					Instead:                  syntax.Mana.Instead,
					TriggerLandProducedType:  syntax.Mana.TriggerLandProducedType,
				},
				Replacement:                            syntax.Replacement,
				Payment:                                compileEffectPayment(syntax.Payment),
				Exact:                                  syntax.Exact,
				KeywordGrantChoice:                     syntax.KeywordGrantChoice,
				RevealUntilThenPut:                     syntax.RevealUntilThenPut,
				PileSplitSequence:                      syntax.PileSplitSequence,
				PileSplitSeparatorOpponent:             syntax.PileSplitSeparatorOpponent,
				PileSplitChooserOpponent:               syntax.PileSplitChooserOpponent,
				PileSplitOtherZone:                     syntax.PileSplitOtherZone,
				PileSplitAmount:                        syntax.PileSplitAmount,
				PileSplitMiddleSpan:                    syntax.PileSplitMiddleSpan,
				SourceSpellCostReduction:               syntax.SourceSpellCostReduction,
				SourceSpellCostReductionAmount:         syntax.SourceSpellCostReductionAmount,
				SourceSpellCostReductionDynamic:        syntax.SourceSpellCostReductionDynamic,
				SourceSpellCostReductionConditional:    syntax.SourceSpellCostReductionConditional,
				RequiresOrderedLowering:                syntax.RequiresOrderedLowering,
				HasUnrecognizedSibling:                 syntax.HasUnrecognizedSibling,
				UnsupportedDetail:                      syntax.UnsupportedDetail,
				Order:                                  syntax.Order,
				VerbOrder:                              syntax.VerbOrder,
				PreventRegeneration:                    syntax.PreventRegeneration,
				RegenerationRiderSpan:                  syntax.RegenerationRiderSpan,
				CopyMayChooseNewTargets:                syntax.CopyMayChooseNewTargets,
				CopyChooseNewTargetsRiderSpan:          syntax.CopyChooseNewTargetsRiderSpan,
				Dig:                                    syntax.Dig,
				HandLibraryPut:                         syntax.HandLibraryPut,
				HandDiscard:                            syntax.HandDiscard,
				RevealChooseDiscard:                    syntax.RevealChooseDiscard,
				HandChoiceDiscard:                      syntax.HandChoiceDiscard,
				DiscardThenDraw:                        syntax.DiscardThenDraw,
				DiscardThenDrawMax:                     syntax.DiscardThenDrawMax,
				DiscardThenDrawOffset:                  syntax.DiscardThenDrawOffset,
				SearchSplit:                            syntax.SearchSplit,
				ManaSpendRider:                         compileManaSpendRider(syntax.ManaSpendRider),
				SearchSharedSubtype:                    syntax.SearchSharedSubtype,
				SearchDestination:                      syntax.SearchDestination,
				SearchControl:                          syntax.SearchControl,
				SearchSlots:                            syntax.SearchSlots,
				DiscardEntireHand:                      syntax.DiscardEntireHand,
				CounteredSpellExileReplacement:         syntax.CounteredSpellExileReplacement,
				ExileUntilSourceLeaves:                 syntax.ExileUntilSourceLeaves,
				ReturnExiledCard:                       syntax.ReturnExiledCard,
				ExileEntireHand:                        syntax.ExileEntireHand,
				ReturnExiledCardsToHand:                syntax.ReturnExiledCardsToHand,
				BottomLinkedExiledCards:                syntax.BottomLinkedExiledCards,
				ExileForEachPlayerUntilSourceLeaves:    syntax.ExileForEachPlayerUntilSourceLeaves,
				ReturnLinkedExiledToBattlefieldPartial: syntax.ReturnLinkedExiledToBattlefieldPartial,
				PutLinkedExiledRestOnLibraryBottom:     syntax.PutLinkedExiledRestOnLibraryBottom,
				DestroyForEachPlayer:                   syntax.DestroyForEachPlayer,
				CreateTokenForEachDestroyedThisWay:     syntax.CreateTokenForEachDestroyedThisWay,
				CreateTokenForEachExiledThisWay:        syntax.CreateTokenForEachExiledThisWay,
				CounterExiledCardManaValue:             syntax.CounterExiledCardManaValue,
				ReturnSourceAndExiledCardToHand:        syntax.ReturnSourceAndExiledCardToHand,
				CantCastSpellsAllPlayers:               syntax.CantCastSpellsAllPlayers,
				CantCastSpellsRequiredTypes:            compilerCardTypes(syntax.CantCastSpellsRequiredTypes),
				CantCastSpellsExcludedTypes:            compilerCardTypes(syntax.CantCastSpellsExcludedTypes),
				SpellCostModifierCaster:                syntax.SpellCostModifierCaster,
				SpellCostModifierAmount:                syntax.SpellCostModifierAmount,
				SpellCostModifierIncrease:              syntax.SpellCostModifierIncrease,
				SpellCostModifierRequiredTypes:         compilerCardTypes(syntax.SpellCostModifierRequiredTypes),
				SpellCostModifierExcludedTypes:         compilerCardTypes(syntax.SpellCostModifierExcludedTypes),
				AttackTaxGeneric:                       syntax.AttackTaxGeneric,
				PlayFromTopPayLife:                     syntax.PlayFromTopPayLife,
				PlayFromTopPayLifeRiderSpan:            syntax.PlayFromTopPayLifeRiderSpan,
				PreventDamageTo:                        syntax.PreventDamageTo,
				PreventDamageBy:                        syntax.PreventDamageBy,
				PreventDamageGlobal:                    syntax.PreventDamageGlobal,
				PreventDamageNextRecipient:             syntax.PreventDamageNextRecipient,
				PreventDamageThatAmount:                syntax.PreventDamageThatAmount,
				PreventDamageNextFromSource:            syntax.PreventDamageNextFromSource,
				PreventDamageSourceColors:              compilePreventDamageSourceColors(syntax.PreventDamageSourceColors),
				SpellsCantBeCounteredNextOnly:          syntax.SpellsCantBeCounteredNextOnly,
				DoublePower:                            syntax.DoublePower,
				DoubleToughness:                        syntax.DoubleToughness,
				DoubleSourceCounters:                   syntax.DoubleSourceCounters,
				DoubleSourceCounterKind:                syntax.DoubleSourceCounterKind,
				DoubleCountersTarget:                   syntax.DoubleCountersTarget,
				DoubleCountersAllKinds:                 syntax.DoubleCountersAllKinds,
				PunisherSacrifice:                      syntax.PunisherSacrifice,
				PunisherDiscard:                        syntax.PunisherDiscard,
				RepeatBody:                             compileEffects([]parser.Sentence{{Effects: syntax.RepeatBody}}),
				ReturnAsEnchantment:                    syntax.ReturnAsEnchantment,
				ReturnAsEnchantmentRiderSpan:           syntax.ReturnAsEnchantmentRiderSpan,
				AdditionalCombatPhase:                  syntax.AdditionalCombatPhase,
				AdditionalMainPhase:                    syntax.AdditionalMainPhase,
				DieSides:                               syntax.DieSides,
			})
		}
	}
	return effects
}

// appendDiceTableEffects compiles each die-roll outcome-table row's resolving
// sentences and appends them to effects, stamping every row effect with the
// row's inclusive result interval. Lowering groups the appended effects by
// interval and gates each on the rolled value. It returns effects unchanged
// when the ability carries no outcome table.
func appendDiceTableEffects(effects []CompiledEffect, table *parser.DiceTable) []CompiledEffect {
	if table == nil {
		return effects
	}
	for _, row := range table.Rows {
		rowEffects := compileEffects(row.Sentences)
		for i := range rowEffects {
			rowEffects[i].DiceRow = true
			rowEffects[i].DiceRowMin = row.Min
			rowEffects[i].DiceRowMax = row.Max
		}
		effects = append(effects, rowEffects...)
	}
	return effects
}

// appendCoinFlipEffects compiles each branch of a recognized "Flip a coin."
// outcome and appends the branch effects to effects, stamping every effect with
// the branch it belongs to. Lowering groups the appended effects by branch and
// gates each on the flip result. It returns effects unchanged when the ability
// carries no coin flip.
func appendCoinFlipEffects(effects []CompiledEffect, flip *parser.CoinFlip) []CompiledEffect {
	if flip == nil {
		return effects
	}
	effects = appendCoinFlipBranch(effects, flip.Win, CoinFlipBranchWin, flip.ConstructSpan)
	effects = appendCoinFlipBranch(effects, flip.Lose, CoinFlipBranchLose, flip.ConstructSpan)
	return effects
}

// appendCoinFlipBranch compiles one coin-flip branch's resolving sentences and
// appends them to effects, stamping each with branch and the construct span the
// parser recorded. The re-parsed branch effects carry consequence-clause spans
// that omit the flip line and may be out of source order, so the shared
// construct span keeps the backend's body-span machinery covering the whole
// construct, exactly as it treats a same-span effect group.
func appendCoinFlipBranch(
	effects []CompiledEffect,
	sentences []parser.Sentence,
	branch CoinFlipBranch,
	construct shared.Span,
) []CompiledEffect {
	branchEffects := compileEffects(sentences)
	for i := range branchEffects {
		branchEffects[i].CoinFlipBranch = branch
		branchEffects[i].Span = construct
		branchEffects[i].VerbSpan = construct
	}
	return append(effects, branchEffects...)
}

// appendVoteEffects compiles each arm of a recognized vote and appends the arm
// effects to effects, stamping every effect with the arm's gating data (the
// option index and tie-inclusiveness). Lowering groups the appended effects by
// arm and gates each on the vote tally. It returns effects unchanged when the
// ability carries no vote.
func appendVoteEffects(effects []CompiledEffect, vote *parser.VoteClause) []CompiledEffect {
	if vote == nil {
		return effects
	}
	for _, arm := range vote.Arms {
		armEffects := compileEffects(arm.Sentences)
		for i := range armEffects {
			armEffects[i].VoteArm = true
			armEffects[i].VoteArmOption = arm.Option
			armEffects[i].VoteArmTieInclusive = arm.TieInclusive
			armEffects[i].Span = vote.ConstructSpan
			armEffects[i].VerbSpan = vote.ConstructSpan
		}
		effects = append(effects, armEffects...)
	}
	return effects
}

func compileStaticRuleEffect(sentence parser.Sentence) (CompiledEffect, bool) {
	rule, _, ok := semanticStaticRuleForSyntax(*sentence.StaticRule)
	if !ok {
		return CompiledEffect{}, false
	}
	kind := effectKindForStaticRule(rule)
	if kind == EffectUnknown {
		return CompiledEffect{}, false
	}
	selector := CompiledSelector{}
	switch sentence.StaticRule.Subject.Kind {
	case parser.StaticRuleSubjectSourceCreature, parser.StaticRuleSubjectAttachedObject:
		selector.Kind = SelectorCreature
	case parser.StaticRuleSubjectSourcePermanent, parser.StaticRuleSubjectAttachedPermanent:
		selector.Kind = SelectorPermanent
	default:
	}
	return CompiledEffect{
		Kind:      kind,
		Span:      sentence.StaticRule.Span,
		Text:      sentence.Text,
		VerbSpan:  sentence.StaticRule.Operation.Span,
		Selector:  selector,
		Negated:   sentence.StaticRule.Constraint.Kind == parser.StaticRuleConstraintProhibition,
		Order:     sentence.StaticRule.Order,
		VerbOrder: sentence.StaticRule.Operation.Order,
	}, true
}

func effectKindForStaticRule(rule StaticRuleKind) EffectKind {
	switch rule {
	case StaticRuleCantBlock:
		return EffectCantBlock
	case StaticRuleCantBeBlocked:
		return EffectCantBeBlocked
	case StaticRuleCantAttack:
		return EffectCantAttack
	case StaticRuleMustAttack:
		return EffectMustAttack
	case StaticRuleMustBeBlocked:
		return EffectMustBeBlocked
	case StaticRuleMustBeBlockedByAllAble:
		return EffectMustBeBlockedByAllAble
	case StaticRuleAssignDamageAsUnblocked:
		return EffectAssignDamageAsUnblocked
	case StaticRuleCantBeCountered:
		return EffectCantBeCountered
	case StaticRuleCantBeBlockedByCreaturesWith:
		return EffectCantBeBlockedByCreaturesWith
	case StaticRuleCantBeBlockedExceptBy:
		return EffectCantBeBlockedExceptBy
	case StaticRuleAssignsCombatDamageByToughness:
		return EffectAssignsCombatDamageByToughness
	case StaticRuleCantBeBlockedByMoreThanOne:
		return EffectCantBeBlockedByMoreThanOne
	case StaticRuleCantAttackOrBlock:
		return EffectCantAttackOrBlock
	case StaticRuleCantAttackAlone:
		return EffectCantAttackAlone
	case StaticRuleCantBlockAlone:
		return EffectCantBlockAlone
	case StaticRuleCantAttackOrBlockAlone:
		return EffectCantAttackOrBlockAlone
	case StaticRuleCantBlockAndCantBeBlocked:
		return EffectCantBlockAndCantBeBlocked
	case StaticRuleDoesntUntap:
		return EffectDoesntUntap
	case StaticRuleCanBlockOnlyCreaturesWithFlying:
		return EffectCanBlockOnlyCreaturesWithFlying
	default:
		return EffectUnknown
	}
}

func staticRuleSentencesOnly(sentences []parser.Sentence) bool {
	if len(sentences) == 0 {
		return false
	}
	for _, sentence := range sentences {
		if sentence.StaticRule == nil {
			return false
		}
	}
	return true
}

// staticRuleSentencesHaveGuard reports whether any static-rule sentence carries
// a trailing guard clause (its GuardSpan), e.g. "... unless you control seven or
// more lands." Only guarded static-rule abilities compile their condition
// segments, preserving the prior behavior (no conditions) for unguarded rules
// whose sentences may carry condition-like timing phrases ("each combat").
func staticRuleSentencesHaveGuard(sentences []parser.Sentence) bool {
	for _, sentence := range sentences {
		if sentence.StaticRule != nil && sentence.StaticRule.Guarded {
			return true
		}
	}
	return false
}

func compileStaticRuleReferences(sentences []parser.Sentence) []CompiledReference {
	references := make([]CompiledReference, 0, len(sentences))
	for i, sentence := range sentences {
		references = append(references, CompiledReference{
			Kind:       ReferenceThisObject,
			Span:       sentence.StaticRule.Subject.Span,
			Binding:    ReferenceBindingSource,
			Occurrence: i,
			NodeID:     i,
			Order:      sentence.StaticRule.Subject.Order,
		})
	}
	return references
}

// compileGroupEntryModification mirrors the parser's typed static group
// entry-modification payload into its compiler form, cloning the tapped form's
// card-type restriction.
func compileGroupEntryModification(syntax parser.GroupEntryModificationSyntax) CompiledGroupEntryModification {
	return CompiledGroupEntryModification{
		Kind:            syntax.Kind,
		ControllerScope: syntax.ControllerScope,
		Types:           slices.Clone(syntax.Types),
	}
}
