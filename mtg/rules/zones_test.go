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

func TestMovePermanentToZoneMovesTokenObjectIDToDestination(t *testing.T) {
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
	if !g.Players[game.Player1].Graveyard.Contains(token.ObjectID) {
		t.Fatal("token object ID did not move to graveyard")
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

func TestAttachPermanentAttachesAuraToLegalCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	aura := addAuraPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player2)

	if !attachPermanent(g, aura, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}
	if aura.AttachedTo == nil || *aura.AttachedTo != creature.ObjectID {
		t.Fatalf("aura attached to = %v, want %v", aura.AttachedTo, creature.ObjectID)
	}
	if len(creature.Attachments) != 1 || creature.Attachments[0] != aura.ObjectID {
		t.Fatalf("creature attachments = %+v, want aura %v", creature.Attachments, aura.ObjectID)
	}
}

func TestIllegalAuraStateBasedActionMovesAuraToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	aura := addAuraPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player2)
	if !attachPermanent(g, aura, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}

	movePermanentToZone(g, creature, game.ZoneGraveyard)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if permanentByObjectID(g, aura.ObjectID) != nil {
		t.Fatal("unattached aura remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("unattached aura did not move to graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != aura.ObjectID || deaths[0].Reason != PermanentDeathReasonIllegalAura {
		t.Fatalf("death logs = %+v, want illegal aura death", deaths)
	}
}

func TestEquipmentRemainsWhenEquippedCreatureLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addEquipmentPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}

	movePermanentToZone(g, creature, game.ZoneGraveyard)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("death logs = %+v, want no equipment death", deaths)
	}
	if permanentByObjectID(g, equipment.ObjectID) == nil {
		t.Fatal("equipment left battlefield when equipped creature left")
	}
	if equipment.AttachedTo != nil {
		t.Fatalf("equipment attached to = %v, want nil", *equipment.AttachedTo)
	}
	if len(creature.Attachments) != 0 {
		t.Fatalf("removed creature attachments = %+v, want none", creature.Attachments)
	}
}

func TestRemovePermanentFromBattlefieldMissingPermanentReturnsNil(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	if removed := removePermanentFromBattlefield(g, 999); removed != nil {
		t.Fatalf("removed permanent = %+v, want nil", removed)
	}
}

func addAuraPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:     "Test Aura",
		Types:    []game.CardType{game.TypeEnchantment},
		Subtypes: []string{"Aura"},
	})
}

func addEquipmentPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:     "Test Equipment",
		Types:    []game.CardType{game.TypeArtifact},
		Subtypes: []string{"Equipment"},
	})
}
