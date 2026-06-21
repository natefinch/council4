package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestMassDestroyExcludesResolutionChosenType verifies that a battlefield group
// whose Selection carries SubtypeChoiceResolutionExcluded matches every creature
// that does NOT share the creature subtype published under SpellChosenTypeChoiceKey
// earlier in the resolution, while sparing creatures of the chosen type ("Choose a
// creature type. Destroy all creatures that aren't of the chosen type." — Kindred
// Dominance). It exercises the same group-membership path the Destroy primitive
// uses to enumerate the permanents it removes.
func TestMassDestroyExcludesResolutionChosenType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Two Elves and a Goblin under the controller, plus an opposing Elf, an
	// opposing Goblin, and a noncreature artifact that the creature-restricted
	// group must never match.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Your Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Your Elf 2", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Your Goblin", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Their Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Their Goblin", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Their Artifact", Types: []types.Card{types.Artifact},
	}})

	obj := &game.StackObject{
		Controller: game.Player1,
		ResolutionChoices: map[string]game.ResolutionChoiceResult{
			string(game.SpellChosenTypeChoiceKey): {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
		},
	}

	// "Destroy all creatures that aren't of the chosen type." lowers to a
	// battlefield-wide creature group excluding the chosen subtype. Both Goblins
	// match; both Elves are spared and the artifact is never a creature.
	excluded := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		SubtypeChoice: game.SubtypeChoiceResolutionExcluded,
	})
	if got := countPermanentsMatchingGroup(g, obj, game.Player1, excluded); got != 2 {
		t.Fatalf("excluded chosen-type creature count = %d, want 2 (both Goblins, sparing both Elves)", got)
	}

	// The positive sibling ("of the chosen type") matches exactly the spared Elves.
	included := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		SubtypeChoice: game.SubtypeChoiceResolution,
	})
	if got := countPermanentsMatchingGroup(g, obj, game.Player1, included); got != 3 {
		t.Fatalf("chosen-type creature count = %d, want 3 (all three Elves)", got)
	}

	// With no published choice the excluded predicate fails closed and matches
	// nothing, so a stray destroy never fires without a chosen type.
	empty := &game.StackObject{Controller: game.Player1}
	if got := countPermanentsMatchingGroup(g, empty, game.Player1, excluded); got != 0 {
		t.Fatalf("excluded count with no chosen type = %d, want 0", got)
	}
}
