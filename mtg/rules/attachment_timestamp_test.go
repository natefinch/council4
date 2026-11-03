package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Felidar Umbra PCA 6 printing
// a4986689-def4-4c8f-836b-4f89a25790ae (oracle
// 278574f6-e083-461b-8bac-52ed51f0cf63); its rulings endpoint returned twelve
// rulings (2012-06-01, 2024-06-07), all about umbra armor or moving lifelink
// mid-combat. Frogify 2XM 52 printing d89be5f5-75d4-42a0-81e9-5da5bf868e73
// (oracle 32249228-e300-4865-b977-5e1f285d02f2); its 2020-08-07 ruling says an
// ability the creature gains after Frogify becomes attached is kept. CR
// 2026-09-25: an Aura, Equipment, or Fortification receives a new timestamp
// each time it becomes attached (613.7e, 701.3c); static-ability effects use
// their object's timestamp (613.7a); attaching to the current object does
// nothing (701.3b); only the most recent same-controller Role stays (704.5z).
func attachmentTimestampCardDef(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup(name)
	if def == nil || def.Name != name {
		t.Fatalf("production %s is missing", name)
	}
	return def
}

func attachmentTimestampBear(g *game.Game, name string) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// TestMovedFelidarUmbraLifelinkOutlastsOlderFrogify activates the registered
// Felidar Umbra's attach ability onto a creature already enchanted by a newer
// Frogify. The move gives Umbra a new timestamp (CR 613.7e), so its lifelink
// grant applies after Frogify's ability loss in layer 6 and the Frog keeps it,
// matching Frogify's ruling about abilities gained after it became attached.
func TestMovedFelidarUmbraLifelinkOutlastsOlderFrogify(t *testing.T) {
	umbraDef := attachmentTimestampCardDef(t, "Felidar Umbra")
	frogifyDef := attachmentTimestampCardDef(t, "Frogify")

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := attachmentTimestampBear(g, "First Bear")
	frog := attachmentTimestampBear(g, "Frogified Bear")
	umbra := addCombatPermanent(g, game.Player1, umbraDef)
	frogify := addCombatPermanent(g, game.Player2, frogifyDef)
	if !attachPermanent(g, umbra, first) || !attachPermanent(g, frogify, frog) {
		t.Fatal("initial Aura attachment failed")
	}
	if !hasKeyword(g, first, game.Lifelink) || hasKeyword(g, frog, game.Lifelink) {
		t.Fatal("initial lifelink state is wrong")
	}
	if umbra.Timestamp() >= frogify.Timestamp() {
		t.Fatalf("Umbra timestamp %d is not older than Frogify timestamp %d", umbra.Timestamp(), frogify.Timestamp())
	}

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	act := action.ActivateAbility(umbra.ObjectID, 0, []game.Target{game.PermanentTarget(frog.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Felidar Umbra's attach ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating Felidar Umbra's attach ability failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !umbra.AttachedTo.Exists || umbra.AttachedTo.Val != frog.ObjectID {
		t.Fatalf("Felidar Umbra attached to %v, want %v", umbra.AttachedTo, frog.ObjectID)
	}
	if umbra.Timestamp() <= frogify.Timestamp() {
		t.Fatalf("moved Umbra timestamp %d did not become newer than Frogify %d", umbra.Timestamp(), frogify.Timestamp())
	}
	if !hasKeyword(g, frog, game.Lifelink) {
		t.Fatal("Frogified creature lost lifelink granted by the later-attached Umbra")
	}
	if !permanentHasSubtype(g, frog, types.Frog) || effectivePower(g, frog) != 1 {
		t.Fatal("Frogify stopped making the creature a 1/1 Frog")
	}
	if hasKeyword(g, first, game.Lifelink) {
		t.Fatal("former Umbra host kept lifelink")
	}
}

// TestMovedFrogifyRemovesLifelinkFromOlderUmbra moves registered Frogify onto
// a creature with a newer Felidar Umbra. Frogify's new timestamp (CR 613.7e)
// makes its ability loss apply after Umbra's lifelink grant.
func TestMovedFrogifyRemovesLifelinkFromOlderUmbra(t *testing.T) {
	umbraDef := attachmentTimestampCardDef(t, "Felidar Umbra")
	frogifyDef := attachmentTimestampCardDef(t, "Frogify")

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := attachmentTimestampBear(g, "First Bear")
	second := attachmentTimestampBear(g, "Second Bear")
	frogify := addCombatPermanent(g, game.Player2, frogifyDef)
	umbra := addCombatPermanent(g, game.Player1, umbraDef)
	if !attachPermanent(g, frogify, first) || !attachPermanent(g, umbra, second) {
		t.Fatal("initial Aura attachment failed")
	}
	if !hasKeyword(g, second, game.Lifelink) {
		t.Fatal("Umbra host lacks lifelink before Frogify moves")
	}

	if !attachPermanent(g, frogify, second) {
		t.Fatal("moving Frogify failed")
	}
	if hasKeyword(g, second, game.Lifelink) {
		t.Fatal("later-attached Frogify did not remove the older Umbra's lifelink")
	}
	if permanentHasSubtype(g, first, types.Frog) || !permanentHasSubtype(g, second, types.Frog) {
		t.Fatal("Frogify's type change did not follow the move")
	}
}

// TestReattachingToSameObjectOrPlayerKeepsTimestamp proves that attaching an
// Aura to what it already enchants does nothing (CR 701.3b), while attaching
// it to a different player gives it a new timestamp (CR 613.7e).
func TestReattachingToSameObjectOrPlayerKeepsTimestamp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addRoleTestCreature(g, game.Player1, "Target")
	role, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
	attachPermanent(g, role, target)
	attached := role.Timestamp()
	if !attachPermanent(g, role, target) || role.Timestamp() != attached {
		t.Fatalf("reattaching to the same creature changed timestamp %d to %d", attached, role.Timestamp())
	}

	curse := addEnchantPlayerAura(g, game.Player1, game.PlayerAny)
	if !attachAuraToPlayer(g, curse, game.Player2) {
		t.Fatal("attaching curse to Player2 failed")
	}
	first := curse.Timestamp()
	if first <= game.Timestamp(curse.ObjectID) {
		t.Fatalf("player attachment timestamp %d is not newer than entry %d", first, curse.ObjectID)
	}
	if !attachAuraToPlayer(g, curse, game.Player2) || curse.Timestamp() != first {
		t.Fatalf("reattaching to the same player changed timestamp %d to %d", first, curse.Timestamp())
	}
	if !attachAuraToPlayer(g, curse, game.Player3) || curse.Timestamp() <= first {
		t.Fatalf("attaching to a different player kept timestamp %d", curse.Timestamp())
	}
	clone := g.Clone()
	cloned, ok := permanentByObjectID(clone, curse.ObjectID)
	if !ok || cloned.Timestamp() != curse.Timestamp() {
		t.Fatal("game clone did not preserve the attachment timestamp")
	}
}
