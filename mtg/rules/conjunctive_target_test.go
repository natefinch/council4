package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestConjunctivePermanentTargetRequiresAllTypes covers the conjunctive
// "target artifact creature" filter that Modular's dies-trigger uses: a
// TargetPredicate.PermanentTypesAll requires a permanent to carry every listed
// type at once, so it matches an artifact creature but rejects a permanent that
// is only an artifact or only a creature.
func TestConjunctivePermanentTargetRequiresAllTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Arcbound Golem", Types: []types.Card{types.Artifact, types.Creature}},
	})
	plainCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature}},
	})
	plainArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Relic", Types: []types.Card{types.Artifact}},
	})

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Predicate:  game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact, types.Creature}, PermanentTypesConjunctive: true},
	}

	if !permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, artifactCreature.ObjectID) {
		t.Fatal("artifact creature should be a legal target for an artifact-creature filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, plainCreature.ObjectID) {
		t.Fatal("plain creature must not match a conjunctive artifact-creature filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, plainArtifact.ObjectID) {
		t.Fatal("plain artifact must not match a conjunctive artifact-creature filter")
	}
}
