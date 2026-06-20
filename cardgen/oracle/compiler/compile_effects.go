package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
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
				Kind:                                 compileEffectKind(syntax.Kind),
				Context:                              syntax.Context,
				Connection:                           syntax.Connection,
				ConnectionSpan:                       syntax.ConnectionSpan,
				Span:                                 syntax.Span,
				ClauseSpan:                           syntax.ClauseSpan,
				Text:                                 syntax.Text,
				VerbSpan:                             syntax.VerbSpan,
				Player:                               syntax.Player,
				CardSource:                           syntax.CardSource,
				RequirePermanentCard:                 syntax.RequirePermanentCard,
				References:                           compileTypedReferences(syntax.References),
				SubjectReferences:                    compileTypedReferences(syntax.SubjectReferences),
				Targets:                              compileTypedTargetList(syntax.Targets),
				SubjectTargets:                       compileTypedTargetList(syntax.SubjectTargets),
				Duration:                             compileEffectDuration(syntax.Duration),
				DelayedTiming:                        compileDelayedTiming(syntax.DelayedTiming),
				Selector:                             compileTypedSelection(syntax.Selection),
				DamageRecipientSelectors:             compileDamageRecipientSelectors(syntax.DamageRecipientPair),
				DamageRecipientReference:             syntax.DamageRecipientReference,
				HasSelfDamageRider:                   syntax.HasSelfDamageRider,
				SelfDamageRiderValue:                 syntax.SelfDamageRiderValue,
				TargetControllerDamageRiderRecipient: syntax.TargetControllerDamageRiderRecipient,
				TargetControllerDamageRiderValue:     syntax.TargetControllerDamageRiderValue,
				HasSecondTargetDamageRider:           syntax.HasSecondTargetDamageRider,
				SecondTargetDamageRiderValue:         syntax.SecondTargetDamageRiderValue,
				Amount:                               compileTypedAmount(syntax.Amount),
				PowerDelta:                           compileSignedAmount(syntax.PowerDelta),
				ToughnessDelta:                       compileSignedAmount(syntax.ToughnessDelta),
				TokenPower:                           syntax.TokenPower,
				TokenToughness:                       syntax.TokenToughness,
				TokenPTKnown:                         syntax.TokenPTKnown,
				TokenName:                            syntax.TokenName,
				TokenCopyOfTarget:                    syntax.TokenCopyOfTarget,
				TokenChoice:                          syntax.TokenChoice,
				StaticSubject:                        compileStaticSubjectKind(syntax.StaticSubject.Kind),
				StaticSubjectSpan:                    syntax.StaticSubject.Span,
				Details: compiledEffectDetails(
					staticSubjectType(syntax.StaticSubject.SubtypeText, syntax.StaticSubject.Subtype, syntax.StaticSubject.SubtypeKnown),
					staticSubjectColors(syntax.StaticSubject.Colors, syntax.StaticSubject.Colorless, syntax.StaticSubject.Multicolored),
					staticSubjectKeyword(syntax.StaticSubject.Keyword, syntax.StaticSubject.ExcludedKeyword),
					syntax.Symbol,
				),
				CounterKind:              syntax.CounterKind,
				CounterKindKnown:         syntax.CounterKnown,
				CounterRecipientAttached: syntax.CounterRecipientAttached,
				FromZone:                 syntax.FromZone,
				GraveyardZoneExile:       syntax.GraveyardZoneExile,
				ToZone:                   syntax.ToZone,
				Destination:              syntax.Destination,
				EntersTapped:             syntax.EntersTapped,
				EntersTappedSelf:         syntax.EntersTappedSelf,
				EntersColorChoice:        syntax.EntersColorChoice,
				EntersColorChoiceExclude: syntax.EntersColorChoiceExclude,
				EntersTypeChoice:         syntax.EntersTypeChoice,
				EntersWithCounters:       syntax.EntersWithCounters,
				UnderYourControl:         syntax.UnderYourControl,
				CastAsAdventure:          syntax.CastAsAdventure,
				Negated:                  syntax.Negated,
				Optional:                 syntax.Optional,
				Divided:                  syntax.Divided,
				OptionalSpan:             syntax.OptionalSpan,
				LifeObject:               syntax.LifeObject,
				Mana: CompiledEffectMana{
					Span:                  syntax.Mana.Span,
					Symbols:               slices.Clone(syntax.Mana.Symbols),
					Colors:                slices.Clone(syntax.Mana.Colors),
					ColorsKnown:           syntax.Mana.ColorsKnown,
					Choice:                syntax.Mana.Choice,
					AnyColor:              syntax.Mana.AnyColor,
					ChosenColor:           syntax.Mana.ChosenColor,
					ChosenColorFixed:      syntax.Mana.ChosenColorFixed,
					ChosenColorFixedKnown: syntax.Mana.ChosenColorFixedKnown,
					CommanderIdentity:     syntax.Mana.CommanderIdentity,
					DynamicColorless:      syntax.Mana.DynamicColorless,
					LegacyBodyExact:       syntax.Mana.LegacyBodyExact,
					FilterPair:            syntax.Mana.FilterPair,
					FilterColors:          slices.Clone(syntax.Mana.FilterColors),
					LandsProduce:          syntax.Mana.LandsProduce,
					LandsProduceScope:     syntax.Mana.LandsProduceScope,
					LandsProduceAnyType:   syntax.Mana.LandsProduceAnyType,
					LinkedExileColors:     syntax.Mana.LinkedExileColors,
					ColorsAmongControlled: syntax.Mana.ColorsAmongControlled,
					ColorsAmongSelector:   compileColorsAmongSelector(syntax.Mana.ColorsAmongSelection),
				},
				Replacement:                    syntax.Replacement,
				Payment:                        compileEffectPayment(syntax.Payment),
				Exact:                          syntax.Exact,
				SourceSpellCostReduction:       syntax.SourceSpellCostReduction,
				SourceSpellCostReductionAmount: syntax.SourceSpellCostReductionAmount,
				RequiresOrderedLowering:        syntax.RequiresOrderedLowering,
				HasUnrecognizedSibling:         syntax.HasUnrecognizedSibling,
				UnsupportedDetail:              syntax.UnsupportedDetail,
				Order:                          syntax.Order,
				VerbOrder:                      syntax.VerbOrder,
				PreventRegeneration:            syntax.PreventRegeneration,
				RegenerationRiderSpan:          syntax.RegenerationRiderSpan,
				Dig:                            syntax.Dig,
				HandLibraryPut:                 syntax.HandLibraryPut,
				HandDiscard:                    syntax.HandDiscard,
				SearchSplit:                    syntax.SearchSplit,
				ManaSpendRider:                 compileManaSpendRider(syntax.ManaSpendRider),
				SearchSharedSubtype:            syntax.SearchSharedSubtype,
				SearchDestination:              syntax.SearchDestination,
				DiscardEntireHand:              syntax.DiscardEntireHand,
			})
		}
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
	case parser.StaticRuleSubjectSourcePermanent:
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
	case StaticRuleCantBeCountered:
		return EffectCantBeCountered
	case StaticRuleCantBeBlockedByCreaturesWith:
		return EffectCantBeBlockedByCreaturesWith
	case StaticRuleCantBeBlockedByMoreThanOne:
		return EffectCantBeBlockedByMoreThanOne
	case StaticRuleCantAttackOrBlock:
		return EffectCantAttackOrBlock
	case StaticRuleDoesntUntap:
		return EffectDoesntUntap
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
