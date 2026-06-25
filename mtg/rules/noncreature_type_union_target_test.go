package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestNoncreatureTypeUnionTargetExcludesCreatures covers the "noncreature
// artifact or noncreature enchantment" filter that Haywire Mite's exile ability
// uses: a TargetPredicate combining a card-type union (PermanentTypes, matched
// disjunctively) with an ExcludedTypes creature qualifier. It targets a plain
// artifact or a plain enchantment but rejects an artifact creature, an
// enchantment creature, and an unrelated permanent.
func TestNoncreatureTypeUnionTargetExcludesCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	plainArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Relic", Types: []types.Card{types.Artifact}},
	})
	plainEnchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Aura", Types: []types.Card{types.Enchantment}},
	})
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Ornithopter", Types: []types.Card{types.Artifact, types.Creature}},
	})
	enchantmentCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Gods Willing", Types: []types.Card{types.Enchantment, types.Creature}},
	})
	plainLand := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}},
	})

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
			ExcludedTypes:    []types.Card{types.Creature},
		}),
	}

	if !permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, plainArtifact.ObjectID) {
		t.Fatal("plain artifact should be a legal target for a noncreature artifact-or-enchantment filter")
	}
	if !permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, plainEnchantment.ObjectID) {
		t.Fatal("plain enchantment should be a legal target for a noncreature artifact-or-enchantment filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, artifactCreature.ObjectID) {
		t.Fatal("artifact creature must not match a noncreature artifact-or-enchantment filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, enchantmentCreature.ObjectID) {
		t.Fatal("enchantment creature must not match a noncreature artifact-or-enchantment filter")
	}
	if permanentTargetMatchesSpec(g, game.Player1, plainArtifact.ObjectID, &spec, plainLand.ObjectID) {
		t.Fatal("land must not match a noncreature artifact-or-enchantment filter")
	}
}
