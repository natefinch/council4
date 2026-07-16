package rules

import (
	"testing"

	cardsh "github.com/natefinch/council4/mtg/cards/h"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addTypedArtifact places a noncreature artifact permanent with the given name
// and optional subtypes under controller. It exercises Hellkite Tyrant's mass
// artifact theft across distinct artifact identities (plain artifacts, Equipment,
// and so on).
func addTypedArtifact(g *game.Game, controller game.PlayerID, name string, subtypes ...types.Sub) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Artifact},
		Subtypes: subtypes,
	}})
}

// addArtifactToken places an artifact token permanent under controller. Token
// permanents read their characteristics from TokenDef rather than a card
// instance (CR 111), so both Token and TokenDef are set.
func addArtifactToken(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	permanent := addTypedArtifact(g, controller, name)
	permanent.Token = true
	permanent.TokenDef = &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
	}}
	return permanent
}

// resolveHellkiteCombatDamageTrigger puts a real Hellkite Tyrant on the
// battlefield under Player1, has it deal unblocked combat damage to victim, then
// puts the resulting "gain control of all artifacts that player controls" trigger
// on the stack and resolves it. It returns the Hellkite permanent so callers can
// assert control transfer and post-resolution source departure. The card comes
// straight from the compiled corpus, so this drives the real combat-damage
// trigger end to end rather than a hand-built stand-in.
func resolveHellkiteCombatDamageTrigger(t *testing.T, g *game.Game, victim game.PlayerID) *game.Permanent {
	t.Helper()
	engine := NewEngine(nil)
	hellkite := addCombatPermanent(g, game.Player1, cardsh.HellkiteTyrant())
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: hellkite.ObjectID,
			Target:   game.AttackTarget{Player: victim},
		}},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Hellkite Tyrant combat-damage trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	return hellkite
}

// TestHellkiteTyrantStealsEveryVictimArtifactIncludingTokens proves the real
// card's trigger gains control of every artifact the damaged player controls —
// plain artifacts, subtyped artifacts (Equipment), artifact creatures, and
// artifact tokens alike — while leaving the victim's nonartifact permanents and
// every other player's artifacts untouched.
func TestHellkiteTyrantStealsEveryVictimArtifactIncludingTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	plain := addTypedArtifact(g, game.Player2, "Sol Ring")
	equipment := addTypedArtifact(g, game.Player2, "Bonesplitter", types.Sub("Equipment"))
	artifactCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ornithopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	token := addArtifactToken(g, game.Player2, "Treasure")

	victimCreature := addCreaturePermanent(g, game.Player2)
	otherArtifact := addTypedArtifact(g, game.Player3, "Star Compass")

	resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	for _, stolen := range []struct {
		name      string
		permanent *game.Permanent
	}{
		{"plain artifact", plain},
		{"Equipment", equipment},
		{"artifact creature", artifactCreature},
		{"artifact token", token},
	} {
		if got := effectiveController(g, stolen.permanent); got != game.Player1 {
			t.Fatalf("%s effective controller = %v, want Player1 (stolen)", stolen.name, got)
		}
	}
	if got := effectiveController(g, victimCreature); got != game.Player2 {
		t.Fatalf("victim nonartifact controller = %v, want Player2 (not an artifact)", got)
	}
	if got := effectiveController(g, otherArtifact); got != game.Player3 {
		t.Fatalf("other player's artifact controller = %v, want Player3 (not the damaged player)", got)
	}
}

// TestHellkiteTyrantLocksArtifactSetAtResolution proves the CR 611.2c snapshot:
// the permanent-duration control effect fixes its affected set when it begins, so
// an artifact the victim gains control of after resolution is not swept in. This
// also models copies: a token copy of a stolen artifact made later is a distinct
// object outside the locked set and stays with its controller.
func TestHellkiteTyrantLocksArtifactSetAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	before := addTypedArtifact(g, game.Player2, "Mind Stone")

	resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	if got := effectiveController(g, before); got != game.Player1 {
		t.Fatalf("artifact present at resolution controller = %v, want Player1 (stolen)", got)
	}

	after := addTypedArtifact(g, game.Player2, "Fellwar Stone")
	if got := effectiveController(g, after); got != game.Player2 {
		t.Fatalf("artifact added after resolution controller = %v, want Player2 (outside the locked set)", got)
	}
}

// TestHellkiteTyrantControlPersistsAfterSourceLeaves proves the resolved
// permanent-duration control effect is a free one-shot effect (CR 611.2b): it is
// not tied to Hellkite's continued presence, so moving Hellkite off the
// battlefield and running state-based actions does not revert control of the
// already-stolen artifacts.
func TestHellkiteTyrantControlPersistsAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	artifact := addTypedArtifact(g, game.Player2, "Coldsteel Heart")
	hellkite := resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	if got := effectiveController(g, artifact); got != game.Player1 {
		t.Fatalf("controller after theft = %v, want Player1", got)
	}

	removePermanentFromBattlefield(g, hellkite.ObjectID)
	engine.applyStateBasedActions(g)
	expireConditionalControlDurations(g)

	if got := effectiveController(g, artifact); got != game.Player1 {
		t.Fatalf("controller after Hellkite left = %v, want Player1 (permanent one-shot effect does not revert)", got)
	}
}

// TestHellkiteTyrantEmptyArtifactSetIsNoOp proves the trigger resolves cleanly
// when the damaged player controls no artifacts: no control effect is created and
// the victim's other permanents are unaffected.
func TestHellkiteTyrantEmptyArtifactSetIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	creature := addCreaturePermanent(g, game.Player2)
	before := len(g.ContinuousEffects)

	resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	if got := len(g.ContinuousEffects); got != before {
		t.Fatalf("continuous effects = %d, want %d (empty group is a no-op)", got, before)
	}
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("victim creature controller = %v, want Player2 (no artifacts to steal)", got)
	}
}

// TestHellkiteTyrantExcludesPhasedOutArtifact proves a phased-out artifact is
// treated as nonexistent (CR 702.26e) and is left out of the snapshot, while the
// victim's active artifacts are still stolen.
func TestHellkiteTyrantExcludesPhasedOutArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	active := addTypedArtifact(g, game.Player2, "Everflowing Chalice")
	phased := addTypedArtifact(g, game.Player2, "Worn Powerstone")
	phased.PhasedOut = true

	resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	if got := effectiveController(g, active); got != game.Player1 {
		t.Fatalf("active artifact controller = %v, want Player1 (stolen)", got)
	}
	if got := effectiveController(g, phased); got != game.Player2 {
		t.Fatalf("phased-out artifact controller = %v, want Player2 (nonexistent at snapshot)", got)
	}
}

// TestHellkiteTyrantStealingTwentyArtifactsFeedsUpkeepWin proves the theft and
// the separate "if you control twenty or more artifacts, you win the game" upkeep
// trigger compose: after stealing twenty of the victim's artifacts, Player1's
// beginning-of-upkeep trigger fires its intervening-if and wins the game, marking
// every opponent to lose. This confirms the mass gain-control change leaves the
// win ability intact.
func TestHellkiteTyrantStealingTwentyArtifactsFeedsUpkeepWin(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	for range 20 {
		addTypedArtifact(g, game.Player2, "Mox")
	}

	resolveHellkiteCombatDamageTrigger(t, g, game.Player2)

	controlled := 0
	for _, permanent := range g.Battlefield {
		if permanentHasType(g, permanent, types.Artifact) && effectiveController(g, permanent) == game.Player1 {
			controlled++
		}
	}
	if controlled != 20 {
		t.Fatalf("artifacts controlled by Player1 = %d, want 20", controlled)
	}

	emitEvent(g, game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: game.Player1,
		Player:     game.Player1,
		Step:       game.StepUpkeep,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("upkeep win trigger did not fire with twenty controlled artifacts")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if !g.MarkedToLoseGame[opponent] {
			t.Fatalf("opponent %v not marked to lose; upkeep win did not resolve", opponent)
		}
	}
}
