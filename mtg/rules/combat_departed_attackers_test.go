package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3485. A permanent that leaves the battlefield is
// removed from combat (CR 506.4) and an attacking creature stays attacking only
// until it is removed from combat (CR 508.1k). A blocking creature remains
// blocking until it is itself removed from combat (CR 509.1g), and a blocked
// creature remains blocked even if all its blockers are removed (CR 509.1h).
// A moved object is a new object (CR 400.7); look-back abilities use the
// departing object's last known information (CR 603.10a).
// Pitfall Trap (oracle 26e75f0b-3a02-4989-8429-b6276e98ac33): "If exactly one
// creature is attacking, you may pay {W} rather than pay this spell's mana
// cost. Destroy target attacking creature without flying." Lethargy Trap
// (oracle 7356319d-1485-426b-80a7-9700d6684ac9): "If three or more creatures are
// attacking, you may pay {U} rather than pay this spell's mana cost. Attacking
// creatures get -3/-0 until end of turn." Righteousness: "Target blocking
// creature gets +7/+7 until end of turn." No Scryfall rulings in the pinned
// corpus for the two Traps.

// declaredAttackers puts count registered Grizzly Bears onto Player1's
// battlefield and declares them all attacking Player2 through the engine, then
// gives Player2 priority in the declare attackers step.
func declaredAttackers(t *testing.T, count int) (*game.Game, *Engine, []*game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackers := make([]*game.Permanent, count)
	declarations := make([]game.AttackDeclaration, count)
	for i := range attackers {
		attackers[i] = registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
		declarations[i] = attackAt(attackers[i], game.AttackTarget{Player: game.Player2})
	}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: declarations}) {
		t.Fatal("attack declaration was rejected")
	}
	g.Turn.PriorityPlayer = game.Player2
	return g, engine, attackers
}

func currentAttackerIDs(g *game.Game) []game.ObjectID {
	var ids []game.ObjectID
	for _, declaration := range g.Combat.Attackers {
		ids = append(ids, declaration.Attacker)
	}
	return ids
}

// TestPitfallTrapAfterOtherAttackerDies is the #3485 scenario: two creatures
// attack, one is destroyed, and exactly one creature is now attacking, so the
// registered Pitfall Trap is cast for its {W} alternative cost and destroys the
// remaining attacker. The combat's history still records that creatures
// attacked.
func TestPitfallTrapAfterOtherAttackerDies(t *testing.T) {
	g, engine, attackers := declaredAttackers(t, 2)
	spell := addCardToHand(g, game.Player2, registryCardDef(t, "Pitfall Trap"))
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	cast := action.CastSpell(spell, []game.Target{game.PermanentTarget(attackers[1].ObjectID)}, 0, nil)
	if containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Pitfall Trap was castable for {W} with two creatures attacking")
	}

	if _, ok := destroyPermanent(g, attackers[0].ObjectID); !ok {
		t.Fatal("the first attacker was not destroyed")
	}
	if got := currentAttackerIDs(g); !slices.Equal(got, []game.ObjectID{attackers[1].ObjectID}) {
		t.Fatalf("attackers after the death = %v, want only %d", got, attackers[1].ObjectID)
	}
	if !g.Combat.AttackersDeclared || !g.Combat.PlayersAttacked[game.Player2] {
		t.Fatal("removing the departed attacker erased the combat's attack history")
	}
	declaredEvents := 0
	for _, event := range g.Events {
		if event.Kind == game.EventAttackerDeclared {
			declaredEvents++
		}
	}
	if declaredEvents != 2 {
		t.Fatalf("attacker-declared events = %d, want 2", declaredEvents)
	}
	if snapshot, ok := lastKnownObject(g, attackers[0].ObjectID); !ok || !snapshot.Attacking {
		t.Fatalf("departed attacker's last known information = %+v, %v; want attacking", snapshot, ok)
	}

	if !containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Pitfall Trap was not castable for {W} with exactly one creature attacking")
	}
	if !engine.applyAction(g, game.Player2, cast) {
		t.Fatal("casting Pitfall Trap for {W} failed")
	}
	if got := g.Players[game.Player2].ManaPool.Total(); got != 0 {
		t.Fatalf("mana remaining = %d, want 0 after paying {W}", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, attackers[1].ObjectID); ok {
		t.Fatal("Pitfall Trap did not destroy the remaining attacker")
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers = %+v, want none", g.Combat.Attackers)
	}
}

// TestDepartedAttackerLeavesCombatByEveryRoute moves one of two attackers off
// the battlefield by each shared departure route. Every route removes it from
// combat, so exactly one creature is attacking and Pitfall Trap's {W}
// alternative becomes available; the public combat view agrees.
func TestDepartedAttackerLeavesCombatByEveryRoute(t *testing.T) {
	routes := []struct {
		name   string
		depart func(*game.Game, *game.Permanent) bool
	}{
		{"destroy", func(g *game.Game, p *game.Permanent) bool { _, ok := destroyPermanent(g, p.ObjectID); return ok }},
		{"sacrifice", sacrificePermanent},
		{"exile", func(g *game.Game, p *game.Permanent) bool { return movePermanentToZone(g, p, zone.Exile) }},
		{"return to hand", func(g *game.Game, p *game.Permanent) bool { return movePermanentToZone(g, p, zone.Hand) }},
		{"put into library", func(g *game.Game, p *game.Permanent) bool { return movePermanentToZone(g, p, zone.Library) }},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			g, engine, attackers := declaredAttackers(t, 2)
			spell := addCardToHand(g, game.Player2, registryCardDef(t, "Pitfall Trap"))
			g.Players[game.Player2].ManaPool.Add(mana.W, 1)
			if !route.depart(g, attackers[0]) {
				t.Fatalf("%s failed", route.name)
			}
			if got := currentAttackerIDs(g); !slices.Equal(got, []game.ObjectID{attackers[1].ObjectID}) {
				t.Fatalf("attackers = %v, want only %d", got, attackers[1].ObjectID)
			}
			view := NewObservation(g, game.Player2).Combat()
			if len(view.Attackers) != 1 || view.Attackers[0].Attacker.ObjectID != attackers[1].ObjectID {
				t.Fatalf("observed attackers = %+v", view.Attackers)
			}
			cast := action.CastSpell(spell, []game.Target{game.PermanentTarget(attackers[1].ObjectID)}, 0, nil)
			if !containsAction(engine.legalActions(g, game.Player2), cast) {
				t.Fatal("Pitfall Trap was not castable for {W} after the other attacker left")
			}
		})
	}
}

// TestSimultaneousAttackerDeparturesCountForLethargyTrap kills two of four
// attackers simultaneously. Both leave combat, so only two creatures are
// attacking and the registered Lethargy Trap's {U} alternative ("three or more
// creatures are attacking") is no longer available, while each departed
// attacker's last known information still says it was attacking.
func TestSimultaneousAttackerDeparturesCountForLethargyTrap(t *testing.T) {
	g, engine, attackers := declaredAttackers(t, 4)
	spell := addCardToHand(g, game.Player2, registryCardDef(t, "Lethargy Trap"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	cast := action.CastSpell(spell, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Lethargy Trap was not castable for {U} with four creatures attacking")
	}

	if !movePermanentsToZoneSimultaneously(g, attackers[:2], zone.Graveyard) {
		t.Fatal("the simultaneous deaths failed")
	}
	if got := currentAttackerIDs(g); !slices.Equal(got, []game.ObjectID{attackers[2].ObjectID, attackers[3].ObjectID}) {
		t.Fatalf("attackers = %v, want the two survivors", got)
	}
	for _, departed := range attackers[:2] {
		if snapshot, ok := lastKnownObject(g, departed.ObjectID); !ok || !snapshot.Attacking {
			t.Fatalf("last known information for %d = %+v, %v; want attacking", departed.ObjectID, snapshot, ok)
		}
	}
	if containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Lethargy Trap was castable for {U} with only two creatures attacking")
	}
}

// TestDepartedCombatantsKeepBlockStatus covers the blocks around a departed
// combatant. The creature blocking a destroyed attacker remains a blocking
// creature (CR 509.1g), so Righteousness can still target it. A destroyed
// blocker leaves combat, but the attacker it blocked remains blocked (CR
// 509.1h) and so deals no combat damage to the defending player.
func TestDepartedCombatantsKeepBlockStatus(t *testing.T) {
	g, engine, attackers := declaredAttackers(t, 2)
	blockerOfDead := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	deadBlocker := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{
		{Blocker: blockerOfDead.ObjectID, Blocking: attackers[0].ObjectID},
		{Blocker: deadBlocker.ObjectID, Blocking: attackers[1].ObjectID},
	}}) {
		t.Fatal("block declaration was rejected")
	}

	if _, ok := destroyPermanent(g, attackers[0].ObjectID); !ok {
		t.Fatal("the first attacker was not destroyed")
	}
	if !combatStateMatches(g, blockerOfDead, game.CombatStateBlocking) {
		t.Fatal("the blocker stopped blocking when the attacker it blocked died")
	}
	righteousness := addCardToHand(g, game.Player2, registryCardDef(t, "Righteousness"))
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player2
	cast := action.CastSpell(righteousness, []game.Target{game.PermanentTarget(blockerOfDead.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Righteousness could not target the creature still blocking")
	}

	if _, ok := destroyPermanent(g, deadBlocker.ObjectID); !ok {
		t.Fatal("the second blocker was not destroyed")
	}
	if combatStateMatches(g, &game.Permanent{ObjectID: deadBlocker.ObjectID}, game.CombatStateBlocking) {
		t.Fatal("the destroyed blocker is still blocking")
	}
	if order := g.Combat.BlockerOrder[attackers[1].ObjectID]; slices.Contains(order, deadBlocker.ObjectID) {
		t.Fatalf("blocker order %v still lists the destroyed blocker", order)
	}
	view := NewObservation(g, game.Player1).Combat()
	if len(view.Attackers) != 1 || !view.Attackers[0].Blocked || len(view.Attackers[0].Blockers) != 0 {
		t.Fatalf("observed attackers = %+v, want one blocked attacker without blockers", view.Attackers)
	}
	life := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != life {
		t.Fatalf("defending player life = %d, want %d: a blocked attacker whose blocker died dealt damage", got, life)
	}
}

// TestDepartedPlaneswalkerLeavesAttackersAttacking destroys a planeswalker
// being attacked. It stops being attacked (CR 506.4), but its attacker remains
// an attacking creature that attacks nothing, still in the direction of the
// defending player, so exactly one creature is still attacking.
func TestDepartedPlaneswalkerLeavesAttackersAttacking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	target := game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(attacker, target)}}) {
		t.Fatal("attack on the planeswalker was rejected")
	}
	if _, ok := destroyPermanent(g, planeswalker.ObjectID); !ok {
		t.Fatal("the planeswalker was not destroyed")
	}
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player2}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
	spell := addCardToHand(g, game.Player2, registryCardDef(t, "Pitfall Trap"))
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player2
	cast := action.CastSpell(spell, []game.Target{game.PermanentTarget(attacker.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Pitfall Trap was not castable for {W} against the creature still attacking")
	}
}
