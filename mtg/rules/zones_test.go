package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRemovePermanentFromBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)

	removed := removePermanentFromBattlefield(g, first.ObjectID)

	if removed != first {
		t.Fatalf("removed permanent = %+v, want %+v", removed, first)
	}
	if len(g.Battlefield) != 1 || g.Battlefield[0] != second {
		t.Fatalf("battlefield = %+v, want only second permanent", g.Battlefield)
	}
}

func TestMovePermanentToZoneMovesCardBackedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1)

	if !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if !g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("card-backed permanent did not move to owner's graveyard")
	}
}

func TestMovePermanentToZoneRemovesTokenWithoutAddingCardToZone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player1,
		Controller: game.Player1,
		Token:      true,
		TokenDef: &game.CardDef{
			Name:  "Token",
			Types: []game.CardType{game.TypeCreature},
		},
	}
	g.Battlefield = append(g.Battlefield, token)

	if !movePermanentToZone(g, token, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("graveyard size = %d, want 0 for token", g.Players[game.Player1].Graveyard.Size())
	}
}

func TestDestroyPermanentMovesToOwnersGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1)

	removed, ok := destroyPermanent(g, permanent.ObjectID)

	if !ok {
		t.Fatal("destroyPermanent() ok = false, want true")
	}
	if removed != permanent {
		t.Fatalf("destroyed permanent = %+v, want %+v", removed, permanent)
	}
	if !g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("destroyed permanent did not move to graveyard")
	}
}

func TestDestroyPermanentDoesNotMoveIndestructiblePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)

	removed, ok := destroyPermanent(g, permanent.ObjectID)

	if ok {
		t.Fatal("destroyPermanent() ok = true, want false")
	}
	if removed != nil {
		t.Fatalf("destroyed permanent = %+v, want nil", removed)
	}
	if permanentByObjectID(g, permanent.ObjectID) == nil {
		t.Fatal("indestructible permanent left the battlefield")
	}
	if g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("indestructible permanent moved to graveyard")
	}
}

func TestRemovePermanentFromBattlefieldMissingPermanentReturnsNil(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	if removed := removePermanentFromBattlefield(g, 999); removed != nil {
		t.Fatalf("removed permanent = %+v, want nil", removed)
	}
}
