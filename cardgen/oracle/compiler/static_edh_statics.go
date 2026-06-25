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
