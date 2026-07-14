package game

import "testing"

// TestOpponentsDealtCombatDamageThisGameByNamedReferenceValidate covers the
// name-parameterized player-group recipient: it validates only with a non-empty
// name, and every other kind fails closed if a name is set.
func TestOpponentsDealtCombatDamageThisGameByNamedReferenceValidate(t *testing.T) {
	t.Parallel()

	valid := OpponentsDealtCombatDamageThisGameByNamedReference("Gollum, Obsessed Stalker")
	if valid.Kind != PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed {
		t.Fatalf("kind = %v, want OpponentsDealtCombatDamageThisGameByNamed", valid.Kind)
	}
	if valid.Name != "Gollum, Obsessed Stalker" {
		t.Fatalf("name = %q, want %q", valid.Name, "Gollum, Obsessed Stalker")
	}
	if problems := valid.Validate(); len(problems) != 0 {
		t.Fatalf("Validate() = %v, want none", problems)
	}

	missingName := PlayerGroupReference{Kind: PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed}
	if problems := missingName.Validate(); len(problems) == 0 {
		t.Fatal("Validate() with empty name = none, want a problem")
	}

	namedOpponents := PlayerGroupReference{Kind: PlayerGroupReferenceOpponents, Name: "Gollum, Obsessed Stalker"}
	if problems := namedOpponents.Validate(); len(problems) == 0 {
		t.Fatal("Validate() for opponents with a stray name = none, want a problem")
	}
}
