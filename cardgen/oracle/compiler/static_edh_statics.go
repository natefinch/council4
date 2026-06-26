package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeStaticOpeningHandPlayDeclaration maps the parser-owned pre-game
// permission "If this card is in your opening hand, you may begin the game with
// it on the battlefield." (the Leyline cycle) onto its inert semantic payload.
// The permission is a special action taken before the game begins; this engine
// starts every game from a fixed setup and never models opening hands, so the
// declaration carries no runtime effect. The residual body content is the "if"
// guard and the self/"it" references of the permission; any other shell fails
// closed.
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
