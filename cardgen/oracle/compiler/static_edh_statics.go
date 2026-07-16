package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeStaticOpeningHandPlayDeclaration maps the parser-owned pre-game
// permission "If this card is in your opening hand, you may begin the game with
// it on the battlefield." (the Leyline cycle) onto its typed semantic payload.
// The permission is a special action taken before the game begins (CR 103.6a);
// it lowers to a static ability marked BeginsGameOnBattlefield that the runtime
// honors during the pregame opening-hand action window. The residual body
// content is the "if" guard and the self/"it" references of the permission; any
// other shell fails closed.
func recognizeStaticOpeningHandPlayDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationOpeningHandPlay) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!openingHandPlayContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	return StaticDeclaration{
		Kind:            StaticDeclarationOpeningHandPlay,
		Span:            ability.Span,
		OperationSpan:   node.OperationSpan,
		OpeningHandPlay: &StaticOpeningHandPlayDeclaration{},
	}, true
}

// openingHandPlayContent reports whether the residual body of the opening-hand
// permission is only its "if" guard and the self/"it" references it names.
func openingHandPlayContent(content AbilityContent) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 2 {
		return false
	}
	condition := content.Conditions[0]
	if condition.Kind != ConditionIf ||
		condition.Predicate != ConditionPredicateUnsupported ||
		condition.Intervening ||
		condition.Resolving {
		return false
	}
	card := content.References[0]
	pronoun := content.References[1]
	return card.Kind == ReferenceThisObject &&
		card.Binding == ReferenceBindingSource &&
		pronoun.Kind == ReferencePronoun &&
		pronoun.Binding == ReferenceBindingSource
}

// recognizeStaticOpponentEnteringTriggerSuppressionDeclaration maps the
// parser-owned static "Permanents entering don't cause abilities of permanents
// your opponents control to trigger." (Elesh Norn, Mother of Machines) onto its
// fixed semantic payload. It suppresses the entering-caused triggered abilities
// of permanents the controller's opponents control. The whole sentence is
// consumed as the declaration, so the residual body must be empty.
func recognizeStaticOpponentEnteringTriggerSuppressionDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationOpponentEnteringTriggerSuppression) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	return StaticDeclaration{
		Kind:                        StaticDeclarationOpponentEnteringTriggerSuppression,
		Span:                        ability.Span,
		OperationSpan:               node.OperationSpan,
		OpponentEnteringSuppression: &StaticOpponentEnteringTriggerSuppressionDeclaration{},
	}, true
}

// recognizeStaticControlOpponentSearchesDeclaration maps the parser-owned static
// "You control your opponents while they're searching their libraries."
// (Opposition Agent) onto its fixed semantic payload. While an opponent of the
// controller is searching their library, the controller makes that search's
// choices. The parser matches the exact wording; the residual the generic scanner
// derives from it is one "as long as" search condition and the ambiguous "their"
// pronoun. Any other shell fails closed.
func recognizeStaticControlOpponentSearchesDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationControlOpponentSearches) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		!controlOpponentSearchesContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	return StaticDeclaration{
		Kind:                    StaticDeclarationControlOpponentSearches,
		Span:                    ability.Span,
		OperationSpan:           node.OperationSpan,
		ControlOpponentSearches: &StaticControlOpponentSearchesDeclaration{},
	}, true
}

// controlOpponentSearchesContent reports whether the leftover content of "You
// control your opponents while they're searching their libraries." is only the
// generic scanner's fixed residual: a single "as long as" search condition and
// the lone ambiguous "their" pronoun the sentence names. Any other content fails
// closed.
func controlOpponentSearchesContent(content AbilityContent) bool {
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Effects) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Conditions) != 1 ||
		len(content.References) != 1 {
		return false
	}
	return searchControlResidualCondition(content.Conditions[0]) &&
		content.References[0].Kind == ReferencePronoun &&
		content.References[0].Binding == ReferenceBindingAmbiguous
}

// recognizeStaticExileOpponentSearchFindsDeclaration maps the parser-owned static
// "While an opponent is searching their library, they exile each card they find.
// You may play those cards for as long as they remain exiled, and you may spend
// mana as though it were mana of any color to cast them." (Opposition Agent) onto
// its fixed semantic payload. Every card an opponent finds while searching their
// library is exiled, and the controller may afterward play those exiled cards
// spending mana of any color. The parser matches the exact wording; the residual
// the generic scanner derives from it is one "as long as" search condition, the
// exile-then-cast effect pair, and the sentence's ambiguous pronouns. Any other
// shell fails closed.
func recognizeStaticExileOpponentSearchFindsDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationExileOpponentSearchFinds) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		!exileOpponentSearchFindsContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	return StaticDeclaration{
		Kind:                     StaticDeclarationExileOpponentSearchFinds,
		Span:                     ability.Span,
		OperationSpan:            node.OperationSpan,
		ExileOpponentSearchFinds: &StaticExileOpponentSearchFindsDeclaration{},
	}, true
}

// exileOpponentSearchFindsContent reports whether the leftover content of the
// exile-finds sentences is only the generic scanner's fixed residual: a single
// "as long as" search condition, the exile-then-cast effect pair the wording
// names ("they exile each card they find" and "to cast them"), and the sentences'
// pronouns, each bound only ambiguously or to the prior instruction's result
// ("those cards"). Any other content fails closed.
func exileOpponentSearchFindsContent(content AbilityContent) bool {
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Conditions) != 1 ||
		len(content.Effects) != 2 ||
		len(content.References) != 7 {
		return false
	}
	if !searchControlResidualCondition(content.Conditions[0]) ||
		content.Effects[0].Kind != EffectExile ||
		content.Effects[1].Kind != EffectCast {
		return false
	}
	for i := range content.References {
		reference := content.References[i]
		if reference.Kind != ReferencePronoun ||
			(reference.Binding != ReferenceBindingAmbiguous &&
				reference.Binding != ReferenceBindingPriorInstructionResult) {
			return false
		}
	}
	return true
}

// searchControlResidualCondition reports whether a leftover condition is the
// "while/for as long as ... searching" continuous scope the generic scanner emits
// for Opposition Agent's static abilities: an unsupported, non-intervening,
// non-resolving "as long as" condition.
func searchControlResidualCondition(condition CompiledCondition) bool {
	return condition.Kind == ConditionAsLongAs &&
		condition.Predicate == ConditionPredicateUnsupported &&
		!condition.Intervening &&
		!condition.Resolving &&
		!condition.Negated
}

// recognizeStaticCreatureAttackTaxDeclaration maps the parser-owned per-creature
// attack tax ("Creatures can't attack you[ or planeswalkers you control] unless
// their controller pays {COST} for each ...", Baird, Archon of Absolution,
// Sphere of Safety, Collective Restraint) onto its semantic payload. The
// protected defending player is the controller; the per-attacker amount is a
// fixed generic value, the controller's enchantment count, or domain. The
// residual body is the "unless" clause and the ambiguous pronouns the sentence
// names; any other shell fails closed.
func recognizeStaticCreatureAttackTaxDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCreatureAttackTax) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	amount, ok := creatureAttackTaxAmount(node.AttackTaxAmountKind)
	if !ok ||
		(amount == StaticCreatureAttackTaxFixed) != (node.AttackTaxGeneric > 0) ||
		!creatureAttackTaxContent(ability.Content, amount) {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCreatureAttackTax,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		CreatureAttackTax: &StaticCreatureAttackTaxDeclaration{
			Amount:               amount,
			FixedGeneric:         node.AttackTaxGeneric,
			IncludePlaneswalkers: node.AttackTaxIncludesPlaneswalkers,
		},
	}, true
}

// creatureAttackTaxAmount maps a parser per-creature attack-tax amount kind onto
// its compiler counterpart, failing closed on an unrecognized kind.
func creatureAttackTaxAmount(kind parser.StaticAttackTaxAmountKind) (StaticCreatureAttackTaxAmountKind, bool) {
	switch kind {
	case parser.StaticAttackTaxAmountFixed:
		return StaticCreatureAttackTaxFixed, true
	case parser.StaticAttackTaxAmountEnchantments:
		return StaticCreatureAttackTaxEnchantments, true
	case parser.StaticAttackTaxAmountDomain:
		return StaticCreatureAttackTaxDomain, true
	}
	return StaticCreatureAttackTaxFixed, false
}

// creatureAttackTaxContent reports whether the residual body of a per-creature
// attack tax is only its "unless" clause and the ambiguous pronouns the sentence
// names. The planeswalker-inclusive family (fixed and enchantment-scaled) names
// "their" and "those"; the player-only domain form names "their" and "they".
func creatureAttackTaxContent(content AbilityContent, amount StaticCreatureAttackTaxAmountKind) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 2 {
		return false
	}
	condition := content.Conditions[0]
	if condition.Kind != ConditionUnless ||
		condition.Predicate != ConditionPredicateUnsupported ||
		condition.Intervening ||
		condition.Resolving {
		return false
	}
	first := content.References[0]
	second := content.References[1]
	if first.Kind != ReferencePronoun ||
		first.Pronoun != ReferencePronounTheir ||
		second.Kind != ReferencePronoun {
		return false
	}
	if amount == StaticCreatureAttackTaxDomain {
		return second.Pronoun == ReferencePronounThey
	}
	return second.Pronoun == ReferencePronounThose
}

func recognizeStaticManaProductionMultiplierDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationManaProductionMultiplier) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.ManaMultiplier < 2 || !manaProductionMultiplierContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:                     StaticDeclarationManaProductionMultiplier,
		Span:                     ability.Span,
		OperationSpan:            node.OperationSpan,
		ManaProductionMultiplier: &StaticManaProductionMultiplierDeclaration{Factor: node.ManaMultiplier},
	}, true
}

// manaProductionMultiplierContent reports whether the residual body of the
// mana-production replacement is only its "if" guard, the tap clause it
// replaces, and the ambiguous "it" pronoun the sentence names ("If you tap a
// permanent for mana, it produces twice as much of that mana instead.").
func manaProductionMultiplierContent(content AbilityContent) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 1 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 1 {
		return false
	}
	condition := content.Conditions[0]
	if condition.Kind != ConditionIf ||
		condition.Predicate != ConditionPredicateUnsupported ||
		condition.Intervening ||
		condition.Resolving {
		return false
	}
	if content.Effects[0].Kind != EffectTap {
		return false
	}
	pronoun := content.References[0]
	return pronoun.Kind == ReferencePronoun &&
		pronoun.Binding == ReferenceBindingAmbiguous &&
		pronoun.Pronoun == ReferencePronounIt
}
