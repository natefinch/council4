package parser

import "testing"

// zoneKindsEqual reports whether zones names exactly the given zone kinds in
// order, ignoring their source spans.
func zoneKindsEqual(zones []TriggerEventZone, want ...TriggerEventZoneKind) bool {
	if len(zones) != len(want) {
		return false
	}
	for i, zone := range zones {
		if zone.Kind != want[i] {
			return false
		}
	}
	return true
}

// TestMultiOriginZoneUnionFailsClosed verifies the multi-origin origin union
// parser fails closed: a mixed-owner union, a union naming an unrecognized
// origin, and a union that repeats a zone must not yield a FromZones zone-change
// clause, so no constraint is silently dropped.
func TestMultiOriginZoneUnionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever one or more cards are put into exile from your library and/or an opponent's graveyard, draw a card.",
		"Whenever one or more cards are put into exile from your library and/or the battlefield, draw a card.",
		"Whenever one or more cards are put into exile from your library and/or your library, draw a card.",
		"Whenever one or more cards are put into exile from your library and or your graveyard, draw a card.",
		"Whenever one or more cards are put into exile from your library and/or, draw a card.",
	} {
		document, diagnostics := Parse(source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		if len(document.Abilities) != 1 {
			t.Fatalf("%q abilities = %d", source, len(document.Abilities))
		}
		trigger := document.Abilities[0].Trigger
		if trigger != nil && trigger.TriggerEvent != nil && len(trigger.TriggerEvent.Zone.FromZones) > 0 {
			t.Fatalf("%q produced FromZones = %#v", source, trigger.TriggerEvent.Zone.FromZones)
		}
	}
}

// TestMultiOriginZoneUnionParsesSegments checks the segment splitter through the
// full parse, ensuring the "and/or" joiner (lexed "and", Slash, "or") separates
// exactly the two named origins.
func TestMultiOriginZoneUnionParsesSegments(t *testing.T) {
	t.Parallel()
	clause := parseTriggerEventFromSource(t,
		"Whenever one or more cards are put into exile from your library and/or your graveyard, draw a card.", "")
	if !zoneKindsEqual(clause.Zone.FromZones, TriggerEventZoneLibrary, TriggerEventZoneGraveyard) {
		t.Fatalf("FromZones = %#v", clause.Zone.FromZones)
	}
}
