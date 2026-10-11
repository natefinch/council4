package parser

func conditionSelectionHasObjectDomain(selection ConditionSelection) bool {
	return !conditionTypeSelectionEmpty(selection) ||
		len(selection.SubtypesAny) > 0 || len(selection.Supertypes) > 0 ||
		len(selection.ColorsAny) > 0 || selection.Colorless || selection.Multicolored
}
