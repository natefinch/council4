package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3491. Official CR effective 2026-09-25 (verified local
// copy), read 506.4, 509.1g-h, 510.1c-d, 510.4 and 702.26b. 506.4: a
// permanent is removed from combat if it leaves the battlefield, its
// controller changes, it phases out, an effect specifically removes it from
// combat, or it's an attacking or blocking creature that regenerates; a
// creature removed from combat stops being an attacking, blocking, blocked
// and/or unblocked creature. 509.1g: a blocking creature remains a blocking
// creature until it's removed from combat or the combat phase ends. 510.1c-d:
// a blocked or blocking creature assigns combat damage only to the creatures
// currently blocking it or that it's currently blocking, and none if there are
// none. 510.4: if any attacking or blocking creature has first strike, the step
// gets first-strike damage. 702.26b: a permanent that phases out is removed
// from combat. Cards (pinned corpus): Gustcloak Runner (oracle
// 0f485ea1-af00-4f6c-8e70-103ba172aa18, {W} 1/1 Human Soldier) "Whenever this
// creature becomes blocked, you may untap it and remove it from combat." (no
// rulings). Righteousness (oracle 3f6b2f76-0364-415e-a6a2-e9e5bf31b745, {W}
// instant) "Target blocking creature gets +7/+7 until end of turn." Ruling
// 2009-10-01: a blocking creature "continues to be a blocking creature through
// the end of combat step, even if the creature or creatures it was blocking are
// no longer on the battlefield or have otherwise left combat." Benalish Hero
// (oracle 4c81cfb7-8765-4e28-ae33-4287fa9a86cc, {W} 1/1 Banding) and Grizzly
// Bears (oracle 14c8f55d-d177-4c25-a931-ebeb9e6062a0, {1}{G} 2/2).

// blockedAttacker declares attacker attacking Player2 and blocker blocking it
// through the engine, leaving the game in the declare blockers step.
func blockedAttacker(t *testing.T, g *game.Game, engine *Engine, attacker, blocker *game.Permanent) {
	t.Helper()
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(attacker, game.AttackTarget{Player: game.Player2}),
	}}) {
		t.Fatal("attack declaration was rejected")
	}
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}}) {
		t.Fatal("block declaration was rejected")
	}
}

// assertRemovedAttackerLeftBlockerBlocking checks the combat state after an
// attacker that stays on the battlefield was removed from combat: its blocker
// is still a blocking creature (CR 509.1g), but the attacker is neither
// attacking, blocking nor blocked, and no block names it (CR 506.4).
func assertRemovedAttackerLeftBlockerBlocking(t *testing.T, g *game.Game, attacker, blocker *game.Permanent) {
	t.Helper()
	if !combatStateMatches(g, blocker, game.CombatStateBlocking) {
		t.Fatal("the blocker stopped blocking when the attacker it blocked was removed from combat")
	}
	if combatStateMatches(g, attacker, game.CombatStateAttackingOrBlocking) {
		t.Fatal("the removed attacker is still attacking or blocking")
	}
	if attackerWasBlocked(g, attacker.ObjectID) || blockingCreaturesOf(g, attacker.ObjectID) != 0 {
		t.Fatal("the removed attacker is still blocked")
	}
	if slices.ContainsFunc(g.Combat.Blockers, func(block game.BlockDeclaration) bool { return block.Blocking == attacker.ObjectID }) {
		t.Fatalf("blocks %+v still name the removed attacker", g.Combat.Blockers)
	}
	if snapshot := snapshotPermanent(g, blocker, zone.Battlefield); !snapshot.Blocking {
		t.Fatal("the blocker's snapshot no longer records it as blocking")
	}
}

// TestGustcloakRunnerBlockerRemainsBlockingForRighteousness is the #3491
// scenario. Registered Gustcloak Runner attacks and is blocked by Grizzly
// Bears; its trigger untaps it and removes it from combat. The Bears remain a
// blocking creature, so the registered Righteousness can target it and make it
// 9/9, but neither creature deals or is dealt combat damage from the block, and
// the Runner, no longer attacking, deals none to the defending player.
func TestGustcloakRunnerBlockerRemainsBlockingForRighteousness(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	runner := registryCombatPermanent(t, g, game.Player1, "Gustcloak Runner")
	bears := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	blockedAttacker(t, g, engine, runner, bears)
	if !runner.Tapped {
		t.Fatal("the attacking Runner was not tapped")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("the Runner's becomes-blocked trigger was not put on the stack (stack size %d)", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if runner.Tapped {
		t.Fatal("the Runner's trigger did not untap it")
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers = %+v, want the Runner removed from combat", g.Combat.Attackers)
	}
	assertRemovedAttackerLeftBlockerBlocking(t, g, runner, bears)

	righteousness := addCardToHand(g, game.Player2, registryCardDef(t, "Righteousness"))
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	g.Turn.PriorityPlayer = game.Player2
	cast := action.CastSpell(righteousness, []game.Target{game.PermanentTarget(bears.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player2), cast) {
		t.Fatal("Righteousness could not target the creature still blocking")
	}
	if containsAction(engine.legalActions(g, game.Player2), action.CastSpell(righteousness, []game.Target{game.PermanentTarget(runner.ObjectID)}, 0, nil)) {
		t.Fatal("Righteousness could target the removed attacker")
	}
	if !engine.applyAction(g, game.Player2, cast) {
		t.Fatal("casting Righteousness failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if power := effectivePower(g, bears); power != 9 {
		t.Fatalf("blocker power = %d, want 9 after Righteousness", power)
	}

	life := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if markedDamage(t, g, runner) != 0 || markedDamage(t, g, bears) != 0 {
		t.Fatalf("marked damage runner=%d bears=%d, want none from a block whose attacker left combat", runner.MarkedDamage, bears.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != life {
		t.Fatalf("defending player life = %d, want %d", got, life)
	}
}

// TestAttackerRemovedWithoutLeavingKeepsBlockerBlocking removes a blocked
// attacker from combat by each route that keeps it on the battlefield: an
// effect that removes it from combat, regeneration and phasing out. Each time
// its blocker stays blocking and neither creature is dealt combat damage. When
// the blocker is then removed from combat itself, it stops blocking.
func TestAttackerRemovedWithoutLeavingKeepsBlockerBlocking(t *testing.T) {
	routes := []struct {
		name   string
		remove func(*testing.T, *game.Game, *Engine, *game.Permanent)
	}{
		{"remove from combat effect", func(t *testing.T, g *game.Game, engine *Engine, attacker *game.Permanent) {
			addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
				{Primitive: game.RemoveFromCombat{Object: game.TargetPermanentReference(0)}},
			}, []game.Target{game.PermanentTarget(attacker.ObjectID)})
			engine.resolveTopOfStack(g, &TurnLog{})
		}},
		{"regeneration", func(t *testing.T, g *game.Game, _ *Engine, attacker *game.Permanent) {
			attacker.RegenerationShields = 1
			destroyPermanent(g, attacker.ObjectID)
			if _, ok := permanentByObjectID(g, attacker.ObjectID); !ok || attacker.RegenerationShields != 0 {
				t.Fatal("the attacker was not regenerated")
			}
		}},
		{"phasing out", func(t *testing.T, g *game.Game, _ *Engine, attacker *game.Permanent) {
			if !phaseOutPermanentTree(g, attacker, game.Player1, make(map[game.ObjectID]bool)) || !attacker.PhasedOut {
				t.Fatal("the attacker did not phase out")
			}
		}},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			g := bandCombatGame()
			engine := NewEngine(nil)
			attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
			blocker := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
			blockedAttacker(t, g, engine, attacker, blocker)

			route.remove(t, g, engine, attacker)
			assertRemovedAttackerLeftBlockerBlocking(t, g, attacker, blocker)
			clone := g.Clone()
			if !combatStateMatches(clone, blocker, game.CombatStateBlocking) {
				t.Fatal("a cloned game lost the blocker's blocking status")
			}

			life := g.Players[game.Player2].Life
			engine.resolveCombatDamage(clone, &TurnLog{})
			if attacker.MarkedDamage != 0 || markedDamage(t, clone, blocker) != 0 {
				t.Fatal("a creature was dealt combat damage from a block whose attacker left combat")
			}
			if after, ok := permanentByObjectID(clone, attacker.ObjectID); ok && after.MarkedDamage != 0 {
				t.Fatalf("removed attacker was dealt %d combat damage", after.MarkedDamage)
			}
			if got := clone.Players[game.Player2].Life; got != life {
				t.Fatalf("defending player life = %d, want %d", got, life)
			}

			removePermanentFromCombat(g, blocker.ObjectID)
			if combatStateMatches(g, blocker, game.CombatStateBlocking) || len(g.Combat.BlockersOfRemovedAttackers) != 0 {
				t.Fatal("the blocker kept blocking after it was itself removed from combat")
			}
		})
	}
}

// TestBandMemberRemovedFromCombatLeavesSharedBlockerOnOtherMember: Benalish
// Hero and Grizzly Bears attack as a band and a 2/4 blocks the Hero, so it
// blocks both (CR 702.22h). The Hero is then removed from combat. The blocker
// still blocks the Bears and assigns all its 2 damage to them (CR 510.1d) and
// takes the Bears' 2 damage; the Hero neither deals nor is dealt any.
func TestBandMemberRemovedFromCombatLeavesSharedBlockerOnOtherMember(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := bandingTestCreature(g, game.Player2, 2, 4)
	if !engine.applyDeclareAttackers(g, game.Player1, bandAttack(hero, bears)) {
		t.Fatal("band declaration rejected")
	}
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: hero.ObjectID},
	}}) {
		t.Fatal("block declaration was rejected")
	}

	removePermanentFromCombat(g, hero.ObjectID)
	assertRemovedAttackerLeftBlockerBlocking(t, g, hero, blocker)
	if want := []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: bears.ObjectID}}; !slices.Equal(g.Combat.Blockers, want) {
		t.Fatalf("blocks = %+v, want %+v", g.Combat.Blockers, want)
	}
	if !attackerWasBlocked(g, bears.ObjectID) || blockingCreaturesOf(g, bears.ObjectID) != 1 {
		t.Fatal("the remaining band member is no longer blocked by the shared blocker")
	}

	engine.resolveCombatDamage(g, &TurnLog{})
	if markedDamage(t, g, bears) != 2 || markedDamage(t, g, hero) != 0 || markedDamage(t, g, blocker) != 2 {
		t.Fatalf("marked damage bears=%d hero=%d blocker=%d, want 2, 0 and 2", bears.MarkedDamage, hero.MarkedDamage, blocker.MarkedDamage)
	}
}

// TestFirstStrikeBlockerOfRemovedAttackerStillCountsForFirstStrikeStep: a
// first-strike creature remains a blocking creature after the attacker it
// blocked is removed from combat, so the combat damage phase still has a
// first-strike damage step (CR 510.4). Once it is itself removed from combat it
// no longer counts.
func TestFirstStrikeBlockerOfRemovedAttackerStillCountsForFirstStrikeStep(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := bandingTestCreature(g, game.Player2, 1, 1, game.FirstStrike)
	blockedAttacker(t, g, engine, attacker, blocker)

	removePermanentFromCombat(g, attacker.ObjectID)
	if !combatHasFirstStrikeDamage(g) {
		t.Fatal("the first-strike blocking creature no longer counts for a first-strike damage step")
	}
	removePermanentFromCombat(g, blocker.ObjectID)
	if combatHasFirstStrikeDamage(g) {
		t.Fatal("a creature removed from combat still counts for a first-strike damage step")
	}
}
