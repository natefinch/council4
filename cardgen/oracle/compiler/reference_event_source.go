package compiler

func exactLibraryGraveyardSourceEvent(pattern TriggerPattern) bool {
	return pattern.Event == TriggerEventZoneChanged &&
		pattern.UnionEvent == TriggerEventUnknown &&
		pattern.Source == TriggerSourceSelf && !pattern.ExcludeSelf &&
		!pattern.SubjectSelectionOrSelf && !pattern.OneOrMore &&
		pattern.Player == TriggerPlayerYou &&
		pattern.MatchFromZone && pattern.FromZone == TriggerZoneLibrary &&
		!pattern.ExcludeFromZone && len(pattern.FromZones) == 0 &&
		pattern.MatchToZone && pattern.ToZone == TriggerZoneGraveyard && !pattern.ExcludeToZone
}
