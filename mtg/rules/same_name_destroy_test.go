package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSameNameGroupMatchesTargetAndSameNamedPermanents verifies that a
// SameNamePermanentGroup anchored on a chosen target enumerates the target
// together with every other battlefield permanent sharing its name, while
// sparing differently-named permanents ("Destroy target nonland permanent and
// all other permanents with the same name as that permanent" — Maelstrom Pulse,
// the Echoing cycle, Wake of Destruction).
func TestSameNameGroupMatchesTargetAndSameNamedPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Two copies of "Llanowar Elves" under different controllers plus a stray
	// "Forest" that shares no name and must be spared.
	anchor := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Llanowar Elves", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Llanowar Elves", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
	}})

	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(anchor.ObjectID)},
	}

	group := game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{})
	if got := countPermanentsMatchingGroup(g, obj, game.Player1, group); got != 2 {
		t.Fatalf("same-name group count = %d, want 2 (both Llanowar Elves, sparing the Forest)", got)
	}
}

// TestSameNameGroupTypeRestrictedAnchor verifies that the group's printed card
// type filter ("all other lands ...") still resolves the same-named members,
// since a shared name implies a shared card and therefore a shared type. The
// type filter on the group selection is fidelity only and must not drop the
// matching same-named permanents.
func TestSameNameGroupTypeRestrictedAnchor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cloudpost", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Locus},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Cloudpost", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Locus},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Glimmerpost", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Locus},
	}})

	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(anchor.ObjectID)},
	}

	group := game.SameNamePermanentGroup(
		game.TargetPermanentReference(0),
		game.Selection{RequiredTypes: []types.Card{types.Land}},
	)
	if got := countPermanentsMatchingGroup(g, obj, game.Player1, group); got != 2 {
		t.Fatalf("type-restricted same-name group count = %d, want 2 (both Cloudposts, sparing Glimmerpost)", got)
	}
}

// TestSameNameGroupFailsClosedWithoutAnchor verifies that a SameNamePermanentGroup
// whose anchor target was never resolved matches nothing rather than every
// permanent on the battlefield.
func TestSameNameGroupFailsClosedWithoutAnchor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Llanowar Elves", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})

	obj := &game.StackObject{Controller: game.Player1}
	group := game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{})
	if got := countPermanentsMatchingGroup(g, obj, game.Player1, group); got != 0 {
		t.Fatalf("same-name group count with no anchor = %d, want 0", got)
	}
}
