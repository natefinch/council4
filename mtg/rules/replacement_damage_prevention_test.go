package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func damagePreventionCardDef(spec *game.DamagePreventionSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Sphere of Law",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionReplacement("damage prevention", spec),
		},
	}}
}

func addTypedSourceCard(g *game.Game, owner game.PlayerID, cardType types.Card) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Typed Source",
			Types: []types.Card{cardType},
		}},
		Owner: owner,
	}
	return cardID
}

func TestDamagePreventionCapsColoredSourceToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damagePreventionCardDef(&game.DamagePreventionSpec{
		Amount:       2,
		SourceColors: []color.Color{color.Red},
	}))
	redSource := addColoredSourceCard(g, game.Player2, color.Red)
	blueSource := addColoredSourceCard(g, game.Player2, color.Blue)

	if dealt := dealPlayerDamage(g, redSource, 0, game.Player2, game.Player1, 3, false); dealt != 1 {
		t.Fatalf("red-source damage to controller = %d, want 1", dealt)
	}
	if dealt := dealPlayerDamage(g, redSource, 0, game.Player2, game.Player1, 1, false); dealt != 0 {
		t.Fatalf("small red-source damage to controller = %d, want 0", dealt)
	}
	if dealt := dealPlayerDamage(g, blueSource, 0, game.Player2, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("blue-source damage = %d, want 3 (color mismatch)", dealt)
	}
}

func TestDamagePreventionOnlyProtectsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damagePreventionCardDef(&game.DamagePreventionSpec{
		Amount:       2,
		SourceColors: []color.Color{color.Red},
	}))
	redSource := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPlayerDamage(g, redSource, 0, game.Player2, game.Player2, 3, false); dealt != 3 {
		t.Fatalf("damage to non-controller player = %d, want 3", dealt)
	}
}

func TestDamagePreventionArtifactSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damagePreventionCardDef(&game.DamagePreventionSpec{
		Amount:      1,
		SourceTypes: []types.Card{types.Artifact},
	}))
	artifactSource := addTypedSourceCard(g, game.Player2, types.Artifact)
	instantSource := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPlayerDamage(g, artifactSource, 0, game.Player2, game.Player1, 3, false); dealt != 2 {
		t.Fatalf("artifact-source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, instantSource, 0, game.Player2, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("non-artifact-source damage = %d, want 3", dealt)
	}
}

func TestDamagePreventionOpponentSourceOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damagePreventionCardDef(&game.DamagePreventionSpec{
		Amount:                   1,
		SourceControllerOpponent: true,
	}))
	opponentSource := addColoredSourceCard(g, game.Player2, color.Red)
	ownSource := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPlayerDamage(g, opponentSource, 0, game.Player2, game.Player1, 3, false); dealt != 2 {
		t.Fatalf("opponent-source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, ownSource, 0, game.Player1, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("own-source damage = %d, want 3 (not opponent-controlled)", dealt)
	}
}

func TestDamagePreventionAnySourceToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damagePreventionCardDef(&game.DamagePreventionSpec{
		Amount: 1,
	}))
	anySource := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPlayerDamage(g, anySource, 0, game.Player2, game.Player1, 3, false); dealt != 2 {
		t.Fatalf("any-source damage to controller = %d, want 2", dealt)
	}
}
