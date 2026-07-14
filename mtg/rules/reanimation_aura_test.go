package rules

import (
	"testing"

	cardsa "github.com/natefinch/council4/mtg/cards/a"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// reanimationVanillaCreature builds a plain vanilla creature card used as a
// reanimation target sitting in a graveyard.
func reanimationVanillaCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: toughness}),
		},
	}
}

// reanimationArtifact builds a non-creature card used to prove Animate Dead
// cannot target a card that is not a creature card.
func reanimationArtifact(name string) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Artifact},
		},
	}
}

// castAndResolveAnimateDead casts the real, compiler-generated Animate Dead card
// at a graveyard creature card and resolves it, returning the resulting Aura
// permanent. The caster is Player1.
func castAndResolveAnimateDead(t *testing.T, g *game.Game, engine *Engine, auraID, targetCardID id.ID) *game.Permanent {
	t.Helper()
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	targets := []game.Target{game.CardTarget(targetCardID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, targets, 0, nil)) {
		t.Fatal("cast of Animate Dead failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatal("Animate Dead did not enter the battlefield")
	}
	return aura
}

// TestAnimateDeadReturnsAttachesAndModifies proves the core behavior of the
// real generated card: casting Animate Dead on a creature card in a graveyard
// returns that creature to the battlefield under the caster's control, attaches
// the Aura to it, applies -1/-0, and leaves the Aura a non-creature Aura.
func TestAnimateDeadReturnsAttachesAndModifies(t *testing.T) {
	g, engine := setupBestowMain(t)
	graveCardID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Grizzly Bears", 2, 2))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	aura := castAndResolveAnimateDead(t, g, engine, auraID, graveCardID)

	creature, ok := findPermanentByCardID(g, graveCardID)
	if !ok {
		t.Fatal("enchanted creature was not returned to the battlefield")
	}
	if effectiveController(g, creature) != game.Player1 {
		t.Fatalf("returned creature controller = %v, want Player1", effectiveController(g, creature))
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("Aura attached to = %v, want creature %v", aura.AttachedTo, creature.ObjectID)
	}
	if aura.ReanimationLinkedObject != creature.ObjectID {
		t.Fatalf("Aura ReanimationLinkedObject = %v, want %v", aura.ReanimationLinkedObject, creature.ObjectID)
	}
	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("returned creature power = %d, want 1 (2 -1/-0)", got)
	}
	if got, _ := effectiveToughness(g, creature); got != 2 {
		t.Fatalf("returned creature toughness = %d, want 2", got)
	}
	if !permanentIsEnchanted(g, creature) {
		t.Fatal("returned creature is not enchanted by the Aura")
	}
	if permanentHasType(g, aura, types.Creature) {
		t.Fatal("Animate Dead is a creature; want a non-creature Aura")
	}
	if !permanentHasSubtype(g, aura, types.Aura) {
		t.Fatal("Animate Dead is not an Aura")
	}
}

// TestAnimateDeadAuraLeavesSacrificesCreature proves that when the Aura leaves
// the battlefield while the creature remains, the creature's controller
// sacrifices it and it goes to its owner's graveyard (CR: "When this Aura leaves
// the battlefield, that creature's controller sacrifices it").
func TestAnimateDeadAuraLeavesSacrificesCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	graveCardID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Grizzly Bears", 2, 2))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	aura := castAndResolveAnimateDead(t, g, engine, auraID, graveCardID)
	creature, ok := findPermanentByCardID(g, graveCardID)
	if !ok {
		t.Fatal("enchanted creature was not returned")
	}

	movePermanentToZone(g, aura, zone.Graveyard)
	settleReanimation(engine, g)

	if _, ok := findPermanentByCardID(g, graveCardID); ok {
		t.Fatalf("creature %v still on battlefield; expected sacrifice after Aura left", creature.ObjectID)
	}
	if !g.Players[game.Player2].Graveyard.Contains(graveCardID) {
		t.Fatal("sacrificed creature did not go to its owner's graveyard")
	}
}

// TestAnimateDeadAuraLeavesSacrificesCreatureAfterControlChange proves the
// "that creature's controller sacrifices it" semantics when control of the
// reanimated creature has changed away from the Aura's controller. Animate Dead
// is controlled by Player1, but a control-change effect hands the reanimated
// creature to Player3 before the Aura leaves. The leaves trigger must still
// sacrifice the creature — performed by its current controller (Player3), not
// gated on the ability's controller — and the card goes to its owner's
// (Player2's) graveyard. Without ByItsController the creature would incorrectly
// survive because its controller no longer matched the trigger's controller.
func TestAnimateDeadAuraLeavesSacrificesCreatureAfterControlChange(t *testing.T) {
	g, engine := setupBestowMain(t)
	graveCardID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Grizzly Bears", 2, 2))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	aura := castAndResolveAnimateDead(t, g, engine, auraID, graveCardID)
	creature, ok := findPermanentByCardID(g, graveCardID)
	if !ok {
		t.Fatal("enchanted creature was not returned")
	}
	if effectiveController(g, creature) != game.Player1 {
		t.Fatalf("reanimated creature controller = %v, want Player1", effectiveController(g, creature))
	}

	// Control of the reanimated creature changes to a third player who is
	// neither the Aura's controller nor the creature's owner.
	creature.Controller = game.Player3
	if effectiveController(g, creature) != game.Player3 {
		t.Fatalf("after control change, controller = %v, want Player3", effectiveController(g, creature))
	}

	movePermanentToZone(g, aura, zone.Graveyard)
	settleReanimation(engine, g)

	if _, ok := findPermanentByCardID(g, graveCardID); ok {
		t.Fatalf("creature %v still on battlefield; expected its current controller to sacrifice it after Aura left", creature.ObjectID)
	}
	if !g.Players[game.Player2].Graveyard.Contains(graveCardID) {
		t.Fatal("sacrificed creature did not go to its owner's graveyard")
	}
}

// TestAnimateDeadCreatureLeavesSendsAuraToGraveyard proves that when the
// enchanted creature leaves, the Aura becomes unattached and is put into its
// owner's graveyard by the state-based action (CR 704.5m).
func TestAnimateDeadCreatureLeavesSendsAuraToGraveyard(t *testing.T) {
	g, engine := setupBestowMain(t)
	graveCardID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Grizzly Bears", 2, 2))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	castAndResolveAnimateDead(t, g, engine, auraID, graveCardID)
	creature, ok := findPermanentByCardID(g, graveCardID)
	if !ok {
		t.Fatal("enchanted creature was not returned")
	}

	movePermanentToZone(g, creature, zone.Exile)
	settleReanimation(engine, g)

	if _, ok := findPermanentByCardID(g, auraID); ok {
		t.Fatal("Aura still on battlefield; expected it to be put into the graveyard as an illegal Aura")
	}
	if !g.Players[game.Player1].Graveyard.Contains(auraID) {
		t.Fatal("unattached Aura did not go to its owner's graveyard")
	}
}

// TestAnimateDeadCannotTargetNonCreatureCard proves the graveyard-card enchant
// restriction: Animate Dead can only target a creature card in a graveyard, so
// casting it at a non-creature (artifact) card in a graveyard is rejected.
func TestAnimateDeadCannotTargetNonCreatureCard(t *testing.T) {
	g, engine := setupBestowMain(t)
	artifactID := addCardToGraveyard(g, game.Player2, reanimationArtifact("Bottle Gnomes Relic"))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.CardTarget(artifactID)}
	if engine.applyAction(g, game.Player1, action.CastSpell(auraID, targets, 0, nil)) {
		t.Fatal("cast of Animate Dead at a non-creature card was accepted; want rejected")
	}
}

// TestAnimateDeadFizzlesWhenTargetLeavesGraveyard proves target legality is
// re-checked on resolution: if the targeted creature card leaves the graveyard
// before Animate Dead resolves, the spell is countered on resolution and the
// Aura never enters the battlefield.
func TestAnimateDeadFizzlesWhenTargetLeavesGraveyard(t *testing.T) {
	g, engine := setupBestowMain(t)
	graveCardID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Grizzly Bears", 2, 2))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	targets := []game.Target{game.CardTarget(graveCardID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, targets, 0, nil)) {
		t.Fatal("cast of Animate Dead failed")
	}
	// The targeted card leaves the graveyard (exiled) before resolution.
	g.Players[game.Player2].Graveyard.Remove(graveCardID)
	g.Players[game.Player2].Exile.Add(graveCardID)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := findPermanentByCardID(g, auraID); ok {
		t.Fatal("Animate Dead entered the battlefield despite an illegal target; want countered")
	}
	if _, ok := findPermanentByCardID(g, graveCardID); ok {
		t.Fatal("target creature entered the battlefield; want no reanimation")
	}
}

// TestAnimateDeadTargetsSpecificGraveyardAmongMany proves the Aura returns
// exactly the targeted creature card even when multiple creature cards sit in
// different players' graveyards.
func TestAnimateDeadTargetsSpecificGraveyardAmongMany(t *testing.T) {
	g, engine := setupBestowMain(t)
	decoyID := addCardToGraveyard(g, game.Player1, reanimationVanillaCreature("Runeclaw Bear", 2, 2))
	targetID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Hill Giant", 3, 3))
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	aura := castAndResolveAnimateDead(t, g, engine, auraID, targetID)

	creature, ok := findPermanentByCardID(g, targetID)
	if !ok {
		t.Fatal("targeted creature was not returned")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("Aura attached to = %v, want targeted creature %v", aura.AttachedTo, creature.ObjectID)
	}
	if _, ok := findPermanentByCardID(g, decoyID); ok {
		t.Fatal("non-targeted graveyard creature was returned")
	}
	if !g.Players[game.Player1].Graveyard.Contains(decoyID) {
		t.Fatal("non-targeted creature card left its graveyard")
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("returned creature power = %d, want 2 (3 -1/-0)", got)
	}
}

// TestAnimateDeadReanimatesCommanderThenReturnsToCommandZone proves commander
// interactions: Animate Dead reanimates an opponent's commander from their
// graveyard under the caster's control, and when the Aura later leaves, the
// forced sacrifice diverts the commander to its owner's command zone via the
// commander-zone replacement (CR 903.9a) rather than the graveyard.
func TestAnimateDeadReanimatesCommanderThenReturnsToCommandZone(t *testing.T) {
	g, engine := setupBestowMain(t)
	commanderID := addCardToGraveyard(g, game.Player2, reanimationVanillaCreature("Opposing General", 4, 4))
	g.Players[game.Player2].CommanderInstanceID = commanderID
	auraID := addCardToHand(g, game.Player1, cardsa.AnimateDead())

	aura := castAndResolveAnimateDead(t, g, engine, auraID, commanderID)

	creature, ok := findPermanentByCardID(g, commanderID)
	if !ok {
		t.Fatal("commander was not reanimated onto the battlefield")
	}
	if effectiveController(g, creature) != game.Player1 {
		t.Fatalf("reanimated commander controller = %v, want Player1", effectiveController(g, creature))
	}

	movePermanentToZone(g, aura, zone.Graveyard)
	settleReanimation(engine, g)

	if _, ok := findPermanentByCardID(g, commanderID); ok {
		t.Fatal("commander still on battlefield after Aura left; expected sacrifice")
	}
	if g.Players[game.Player2].Graveyard.Contains(commanderID) {
		t.Fatal("commander went to graveyard; expected command-zone replacement")
	}
	if !g.Players[game.Player2].CommandZone.Contains(commanderID) {
		t.Fatal("sacrificed commander did not return to its owner's command zone")
	}
}

// settleReanimation runs the engine to a stable state: repeatedly apply
// state-based actions (including the unattached-Aura and leaves-the-battlefield
// paths), put triggered abilities on the stack, and resolve the stack until
// nothing changes. It lets a test observe the downstream consequences of a
// forced zone move without driving a full priority loop.
func settleReanimation(engine *Engine, g *game.Game) {
	log := &TurnLog{}
	for range 32 {
		engine.applyStateBasedActionsWithDeaths(g)
		placed := engine.putTriggeredAbilitiesOnStack(g)
		resolved := false
		if !g.Stack.IsEmpty() {
			engine.resolveTopOfStack(g, log)
			resolved = true
		}
		if !placed && !resolved {
			return
		}
	}
}
