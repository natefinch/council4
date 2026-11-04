package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3602. Official CR effective 2026-09-25 (verified local
// copy), read 109.4, 302.6, 303.4e, 611.2a-c, 613.1b, 613.7, 613.8a-c. 613.1b:
// control-changing effects apply in layer 2. 613.8a: an effect depends on
// another in its layer when applying the other would change what it applies to
// or what it does to those things; 613.8b: a dependent effect waits until just
// after the effects it depends on, and a dependency loop applies in timestamp
// order; 613.8c: the order is reevaluated after each effect applies. 302.6: a
// creature can't attack or use {T} abilities unless controlled continuously
// since its controller's most recent turn began. Cards (pinned corpus): Control
// Magic (oracle cd0d7141-46d2-4aa3-bc77-6b3b4513803e, {2}{U}{U} Aura) "Enchant
// creature / You control enchanted creature." (ruling 2016-06-08: gaining
// control of a creature doesn't gain control of attached Auras or Equipment).
// Confiscate (oracle d599a38c-7719-443e-bb59-55bd43a8fee6, {4}{U}{U} Aura)
// "Enchant permanent / You control enchanted permanent." (rulings 2022-12-08
// and 2024-11-08: gaining control of an Aura doesn't change what it's attached
// to). Word of Seizing (oracle b80d2841-03b0-42d3-ac9b-51e793b1dc24,
// {3}{R}{R} instant) "Split second ... Untap target permanent and gain control
// of it until end of turn. It gains haste until end of turn." (rulings
// 2021-03-19 cover split second only). Llanowar Elves (oracle
// 68954295-54e3-4303-a6bc-fc4547a4e3a3, {G} 1/1) "{T}: Add {G}." Disenchant
// (oracle a7e97fa9-4b72-4548-b854-5be5f18a6f1a, {1}{W} instant) "Destroy target
// artifact or enchantment." Context: Agent of Treachery (oracle
// 4cd82c7a-86df-4d29-b578-c009208c0d4b) is blocked on this gap. Control
// Magic's "you" is the Aura's current controller, so its effect depends on
// every layer-2 effect that changes the Aura's controller.

// stolenControlMagicOnElves leaves Player2's Control Magic on Player1's
// Llanowar Elves, with Player2 having controlled both since Player2's turn
// began, in Player1's precombat main phase.
func stolenControlMagicOnElves(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]
	beginSicknessTurn(g, engine, game.Player3)
	beginSicknessTurn(g, engine, game.Player4)
	beginSicknessTurn(g, engine, game.Player1)
	if effectiveController(g, elves) != game.Player2 || effectiveController(g, aura) != game.Player2 {
		t.Fatal("fixture: Player2 does not control Control Magic and the Elves")
	}
	return g, engine, elves, aura
}

func observedController(t *testing.T, g *game.Game, observer game.PlayerID, permanent *game.Permanent) game.PlayerID {
	t.Helper()
	view := NewObservation(g, observer).Battlefield()
	i := slices.IndexFunc(view, func(p PermanentView) bool { return p.ObjectID == permanent.ObjectID })
	if i < 0 {
		t.Fatalf("%s is not observed on the battlefield", permanentName(g, permanent))
	}
	return view[i].Controller
}

// TestWordOfSeizingOnControlMagicTakesEnchantedCreature is the #3602 scenario.
// During Player1's turn Player1 casts Word of Seizing on Player2's Control
// Magic. Control Magic's "you" is now Player1, so Player1 controls the
// enchanted Elves too. The Elves hasn't been under Player1's control since
// Player1's turn began and only the Aura gained haste, so it can't attack or
// tap for mana. When Word of Seizing's effect ends in cleanup, both return to
// Player2 before Player2's turn begins, so Player2 may use the Elves that turn.
func TestWordOfSeizingOnControlMagicTakesEnchantedCreature(t *testing.T) {
	g, engine, elves, aura := stolenControlMagicOnElves(t)

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, aura)
	if got := effectiveController(g, elves); got != game.Player1 {
		t.Fatalf("gaining control of Control Magic left the enchanted Elves under player %d", got)
	}
	for _, observer := range []game.PlayerID{game.Player1, game.Player2} {
		if got := observedController(t, g, observer, elves); got != game.Player1 {
			t.Fatalf("player %d observes the Elves under player %d", observer, got)
		}
	}
	if !permanentSummoningSick(g, elves) {
		t.Fatal("the Elves taken with its Aura this turn is not summoning sick")
	}
	if creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the Elves taken this turn attacked without haste")
	}
	if elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("the Elves taken this turn tapped for mana without haste")
	}

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if effectiveController(g, aura) != game.Player2 || effectiveController(g, elves) != game.Player2 {
		t.Fatal("Control Magic and the Elves did not return to Player2 when Word of Seizing ended")
	}
	beginSicknessTurn(g, engine, game.Player2)
	if !creatureCanAttack(t, g, engine, game.Player2, elves) {
		t.Fatal("the Elves returned before Player2's turn began cannot attack")
	}
	if !elvesPayForElves(t, g, engine, game.Player2) {
		t.Fatal("the Elves returned before Player2's turn began cannot tap for mana")
	}
}

// TestDestroyingSeizedControlMagicReturnsCreatureToOwner ends both control
// effects: after Player1 seizes Player2's Control Magic, Player2 destroys the
// Aura with Disenchant. The Elves returns to its owner, Player1, but Player1's
// control has not been continuous since Player1's turn began (Player2 had it),
// so it can't attack or tap until Player1's next turn.
func TestDestroyingSeizedControlMagicReturnsCreatureToOwner(t *testing.T) {
	g, engine, elves, aura := stolenControlMagicOnElves(t)
	castAndResolveWordOfSeizing(t, g, engine, game.Player1, aura)

	castAndResolveRegistered(t, g, engine, game.Player2, "Disenchant", mana.W, 2, aura)
	if slices.Contains(g.Battlefield, aura) {
		t.Fatal("fixture: Disenchant did not destroy Control Magic")
	}
	if got := effectiveController(g, elves); got != game.Player1 {
		t.Fatalf("the Elves is under player %d after both control effects ended", got)
	}
	if creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the Elves controlled only since this turn attacked")
	}
	if elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("the Elves controlled only since this turn tapped for mana")
	}

	beginSicknessTurn(g, engine, game.Player2)
	if effectiveController(g, elves) != game.Player1 {
		t.Fatal("the Elves left its owner's control")
	}
	beginSicknessTurn(g, engine, game.Player3)
	beginSicknessTurn(g, engine, game.Player4)
	beginSicknessTurn(g, engine, game.Player1)
	if !creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the Elves controlled since Player1's turn began cannot attack")
	}
	if !elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("the Elves controlled since Player1's turn began cannot tap for mana")
	}
}

// controlMagicDependencyGame has Player1 enchant Player3's Elves with Control
// Magic, Player3 take the Elves back with Word of Seizing, and then Player2
// take Control Magic with Word of Seizing. Player3's later-timestamp steal
// would win in timestamp order, but Control Magic's effect depends on Player2's
// Word of Seizing (it changes what Control Magic does), so it applies after
// both and Player2 controls the Elves (CR 613.8b).
func controlMagicDependencyGame(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player3, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player1, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]
	castAndResolveWordOfSeizing(t, g, engine, game.Player3, elves)
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, aura)
	return g, engine, elves, aura
}

func TestControlMagicAppliesAfterControlEffectsOnItself(t *testing.T) {
	g, engine, elves, aura := controlMagicDependencyGame(t)
	if got := effectiveController(g, elves); got != game.Player2 {
		t.Fatalf("Control Magic's dependent effect left the Elves under player %d", got)
	}
	if got := observedController(t, g, game.Player3, elves); got != game.Player2 {
		t.Fatalf("Player3 observes the Elves under player %d", got)
	}

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if effectiveController(g, aura) != game.Player1 || effectiveController(g, elves) != game.Player1 {
		t.Fatal("Control Magic and the Elves did not return to Player1 when both Words of Seizing ended")
	}
}

// confiscateLoop attaches Player1's Confiscate and Player2's Confiscate to
// each other. Each effect's new controller is the other's enchanted Aura's
// controller, so they form a dependency loop and apply in timestamp order
// (CR 613.8b): the earlier one gives the later Aura to its controller, and the
// later one then gives the earlier Aura to that same player.
func confiscateLoop(t *testing.T, firstController game.PlayerID) (*game.Game, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := sicknessMainPhaseGame(game.Player1)
	def := registryCardDef(t, "Confiscate")
	first := addCombatPermanent(g, firstController, def)
	second := addCombatPermanent(g, game.Player1+game.Player2-firstController, def)
	first.AttachedTo, first.Attachments = opt.Val(second.ObjectID), []id.ID{second.ObjectID}
	second.AttachedTo, second.Attachments = opt.Val(first.ObjectID), []id.ID{first.ObjectID}
	return g, first, second
}

func TestConfiscateLoopAppliesInTimestampOrder(t *testing.T) {
	for _, firstController := range []game.PlayerID{game.Player1, game.Player2} {
		for _, queryFirstFirst := range []bool{true, false} {
			g, first, second := confiscateLoop(t, firstController)
			g.BeginStaticSourceFrame()
			var got [2]game.PlayerID
			if queryFirstFirst {
				got[0], got[1] = effectiveController(g, first), effectiveController(g, second)
			} else {
				got[1], got[0] = effectiveController(g, second), effectiveController(g, first)
			}
			g.EndStaticSourceFrame()
			if got != [2]game.PlayerID{firstController, firstController} {
				t.Fatalf("earlier Confiscate's controller %d, query first first %v: controllers %v", firstController, queryFirstFirst, got)
			}
			values := effectivePermanentValues(g, second)
			if values.controller != firstController {
				t.Fatalf("full layer values put the later Confiscate under player %d", values.controller)
			}
		}
	}
}

// TestControlMagicDependencyIsCloneAndReplayDeterministic replays the
// dependency scenario and clones the result: every copy computes the same
// controller for every permanent.
func TestControlMagicDependencyIsCloneAndReplayDeterministic(t *testing.T) {
	controllers := func(g *game.Game) []game.PlayerID {
		var result []game.PlayerID
		for _, permanent := range g.Battlefield {
			result = append(result, effectiveController(g, permanent))
		}
		return result
	}
	g, _, elves, _ := controlMagicDependencyGame(t)
	want := controllers(g)
	if want[slices.Index(g.Battlefield, elves)] != game.Player2 {
		t.Fatal("fixture: the Elves is not under Player2")
	}
	replay, _, _, _ := controlMagicDependencyGame(t)
	if got := controllers(replay); !slices.Equal(got, want) {
		t.Fatalf("replayed controllers %v, want %v", got, want)
	}
	clone := g.Clone()
	if got := controllers(clone); !slices.Equal(got, want) {
		t.Fatalf("cloned controllers %v, want %v", got, want)
	}
	clone.ContinuousEffects = nil
	if got := controllers(g); !slices.Equal(got, want) {
		t.Fatalf("changing the clone changed the original's controllers to %v", got)
	}
}

// TestControlEffectNotChangingControlMagicIsNotADependency has Player1 enchant
// their Elves with Control Magic and Player2 seize the Elves. Player1 then
// casts Word of Seizing on their own Control Magic. That effect gives the Aura
// to the player already controlling it, so it doesn't change what Control
// Magic does and isn't a dependency (CR 613.8a): Control Magic still applies
// in timestamp order, before Player2's later steal, and Player2 keeps the
// Elves.
func TestControlEffectNotChangingControlMagicIsNotADependency(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player1, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, elves)
	if got := effectiveController(g, elves); got != game.Player2 {
		t.Fatalf("fixture: the seized Elves is under player %d", got)
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, aura)
	if got := effectiveController(g, aura); got != game.Player1 {
		t.Fatalf("Control Magic is under player %d", got)
	}
	if got := effectiveController(g, elves); got != game.Player2 {
		t.Fatalf("seizing one's own Control Magic took the Elves back to player %d", got)
	}
	if got := observedController(t, g, game.Player1, elves); got != game.Player2 {
		t.Fatalf("Player1 observes the Elves under player %d", got)
	}
}

// TestControlDependencyIsReevaluatedAfterEachEffect has Player2 seize
// Player1's Control Magic and then Player1 seize it back. Before Player2's
// effect applies, Player1's would give the Aura to its current controller and
// isn't a dependency; once Player2's effect applies, Player1's would change
// the Aura's controller, so Control Magic now depends on it too (CR 613.8c).
// Control Magic applies after both, and Player1 controls the Elves.
func TestControlDependencyIsReevaluatedAfterEachEffect(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player3, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player1, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, aura)
	if effectiveController(g, aura) != game.Player2 || effectiveController(g, elves) != game.Player2 {
		t.Fatal("fixture: Player2 does not control Control Magic and the Elves")
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, aura)
	if got := effectiveController(g, aura); got != game.Player1 {
		t.Fatalf("Control Magic is under player %d", got)
	}
	if got := effectiveController(g, elves); got != game.Player1 {
		t.Fatalf("the Elves is under player %d after Player1 seized Control Magic back", got)
	}
	values := effectivePermanentValues(g, elves)
	if values.controller != game.Player1 {
		t.Fatalf("full layer values put the Elves under player %d", values.controller)
	}
}
