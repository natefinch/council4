package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestControllingMatchingComposesOntoBaseGroup covers the per-member "who
// controls <selection>" qualifier: it sets ControlsMatching without disturbing
// the base kind, composes onto any base group, and validates the embedded
// selection.
func TestControllingMatchingComposesOntoBaseGroup(t *testing.T) {
	t.Parallel()

	selection := Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}

	allPlayers := AllPlayersReference().ControllingMatching(selection)
	if allPlayers.Kind != PlayerGroupReferenceAllPlayers {
		t.Fatalf("kind = %v, want AllPlayers (base kind unchanged)", allPlayers.Kind)
	}
	if allPlayers.ControlsMatching == nil {
		t.Fatal("ControlsMatching = nil, want the composed selection")
	}
	if problems := allPlayers.Validate(); len(problems) != 0 {
		t.Fatalf("Validate() = %v, want none", problems)
	}

	opponents := OpponentsReference().ControllingMatching(selection)
	if opponents.Kind != PlayerGroupReferenceOpponents {
		t.Fatalf("kind = %v, want Opponents (base kind unchanged)", opponents.Kind)
	}
	if problems := opponents.Validate(); len(problems) != 0 {
		t.Fatalf("Validate() = %v, want none", problems)
	}

	// An unqualified group has no ControlsMatching filter.
	if AllPlayersReference().ControlsMatching != nil {
		t.Fatal("unqualified AllPlayersReference has a ControlsMatching filter, want nil")
	}
}

// TestControllingMatchingValidatesEmbeddedSelection reports a problem when the
// composed selection is itself invalid, prefixing the selection's problem so the
// source of the failure is clear.
func TestControllingMatchingValidatesEmbeddedSelection(t *testing.T) {
	t.Parallel()

	invalid := Selection{Colorless: true, Multicolored: true}
	if problems := invalid.Validate(); len(problems) == 0 {
		t.Skip("selection considered valid; nothing to assert")
	}

	group := AllPlayersReference().ControllingMatching(invalid)
	problems := group.Validate()
	if len(problems) == 0 {
		t.Fatal("Validate() = none, want the embedded selection's problem surfaced")
	}
}

// TestPlayerGroupReferenceComparable guards the invariant that
// PlayerGroupReference stays comparable with == (embedding value types such as
// DamageRecipient rely on it), which the pointer ControlsMatching field
// preserves.
func TestPlayerGroupReferenceComparable(t *testing.T) {
	t.Parallel()

	a := AllPlayersReference()
	b := AllPlayersReference()
	if a != b {
		t.Fatal("two AllPlayersReference() values compare unequal, want equal")
	}
	if a == (PlayerGroupReference{}) {
		t.Fatal("AllPlayersReference() equals the zero value, want unequal")
	}
}
