package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// displacedDinosaursDef returns an Enchantment carrying the reusable group ETB
// characteristic replacement modeled on Displaced Dinosaurs ("As a historic
// permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to
// its other types.").
func displacedDinosaursDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Displaced Dinosaurs",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersBecomesGroupReplacement(
				"As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types.",
				game.EntersBecomesGroupParams{
					Controller:    game.TriggerControllerYou,
					Historic:      true,
					AddTypes:      []types.Card{types.Creature},
					AddSubtypes:   []types.Sub{types.Dinosaur},
					BasePower:     opt.Val(7),
					BaseToughness: opt.Val(7),
				},
			),
		},
	}}
}

func assertBecameDinosaur(t *testing.T, g *game.Game, permanent *game.Permanent) {
	t.Helper()
	if !permanentHasType(g, permanent, types.Creature) {
		t.Error("permanent did not gain the Creature type")
	}
	if !permanentHasSubtype(g, permanent, types.Dinosaur) {
		t.Error("permanent did not gain the Dinosaur subtype")
	}
	if got := effectivePower(g, permanent); got != 7 {
		t.Errorf("effective power = %d, want 7", got)
	}
	if got, ok := effectiveToughness(g, permanent); !ok || got != 7 {
		t.Errorf("effective toughness = (%d, %t), want (7, true)", got, ok)
	}
}

func assertNotDinosaur(t *testing.T, g *game.Game, permanent *game.Permanent) {
	t.Helper()
	if permanentHasSubtype(g, permanent, types.Dinosaur) {
		t.Error("permanent gained the Dinosaur subtype, want it unaffected")
	}
	if got := effectivePower(g, permanent); got == 7 {
		if got, ok := effectiveToughness(g, permanent); ok && got == 7 {
			t.Error("permanent became 7/7, want it unaffected")
		}
	}
}

// TestDisplacedDinosaursMakesHistoricArtifactA7x7Dinosaur proves an artifact
// entering under a controller's Displaced Dinosaurs becomes a 7/7 Dinosaur
// creature while retaining its artifact type.
func TestDisplacedDinosaursMakesHistoricArtifactA7x7Dinosaur(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, displacedDinosaursDef())
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("registered replacement effects = %d, want 1", len(g.ReplacementEffects))
	}

	artifact := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mox Opal",
		Types: []types.Card{types.Artifact},
	}})

	assertBecameDinosaur(t, g, artifact)
	if !permanentHasType(g, artifact, types.Artifact) {
		t.Error("permanent lost its artifact type, want it retained in addition")
	}
}

// TestDisplacedDinosaursMakesLegendaryAndSagaDinosaurs proves a legendary
// permanent and a Saga enchantment both qualify as historic and become 7/7
// Dinosaurs.
func TestDisplacedDinosaursMakesLegendaryAndSagaDinosaurs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, displacedDinosaursDef())

	legendary := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Sol Ring, Legendary",
		Types:      []types.Card{types.Artifact},
		Supertypes: []types.Super{types.Legendary},
	}})
	assertBecameDinosaur(t, g, legendary)

	saga := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "History of Benalia",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Saga},
	}})
	assertBecameDinosaur(t, g, saga)
	if !permanentHasSubtype(g, saga, types.Saga) {
		t.Error("Saga lost its Saga subtype, want it retained in addition")
	}
}

// TestDisplacedDinosaursIgnoresNonHistoricAndOpponents proves a non-historic
// permanent (a plain land) and an opponent's historic permanent are both left
// unaffected by a "you control" scoped replacement.
func TestDisplacedDinosaursIgnoresNonHistoricAndOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, displacedDinosaursDef())

	land := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	assertNotDinosaur(t, g, land)
	if !permanentHasType(g, land, types.Land) {
		t.Error("land lost its land type")
	}

	opponentArtifact := addReplacementPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mox Opal",
		Types: []types.Card{types.Artifact},
	}})
	assertNotDinosaur(t, g, opponentArtifact)
}

// TestDisplacedDinosaursAffectsHistoricSourceItself proves that when the source
// carrying the replacement is itself historic, it becomes a 7/7 Dinosaur as it
// enters, because the replacement is registered before it applies to the entrant.
func TestDisplacedDinosaursAffectsHistoricSourceItself(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := displacedDinosaursDef()
	def.Name = "Displaced Dinosaurs, Legendary"
	def.Supertypes = []types.Super{types.Legendary}

	source := addReplacementPermanent(t, g, game.Player1, def)
	assertBecameDinosaur(t, g, source)
	if !permanentHasType(g, source, types.Enchantment) {
		t.Error("source lost its enchantment type, want it retained in addition")
	}
}
