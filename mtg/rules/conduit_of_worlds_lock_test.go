package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestConduitInsufficientManaCastsNothingNoLock proves that accepting the
// optional cast without enough mana to pay the card's cost casts nothing, leaves
// the card in the graveyard, and applies no restriction: the lock is gated on an
// actual cast, which never happened.
func TestConduitInsufficientManaCastsNothingNoLock(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	// No mana is added to the pool, so the {1} cost cannot be paid.

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted creature left the graveyard even though its cost was unpayable")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied without a successful cast")
	}
}

// TestConduitTargetLeavesGraveyardFizzles proves the ability rechecks its single
// target at resolution: if the targeted card leaves the graveyard first, the
// ability is countered by the rules, nothing is cast, and no restriction applies.
func TestConduitTargetLeavesGraveyardFizzles(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)

	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	// The targeted card leaves the graveyard before the ability resolves.
	g.Players[game.Player1].Graveyard.Remove(bearID)
	g.Players[game.Player1].Exile.Add(bearID)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Exile.Contains(bearID) {
		t.Fatal("targeted creature was cast from the graveyard despite having moved to exile")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied even though the ability fizzled")
	}
}

// resolveConduitCastAndLock runs the successful accept-and-pay path and asserts
// the lock landed, returning the source permanent and cast-creature id so lock
// lifecycle tests can start from a known-good post-cast state.
func resolveConduitCastAndLock(t *testing.T, g *game.Game, engine *Engine) *game.Permanent {
	t.Helper()
	source := addConduit(g, game.Player1)
	bearID := addCardToGraveyard(g, game.Player1, graveyardBear())
	setUpConduitMainPhase(g)
	if !activateConduitTargeting(t, g, engine, source, bearID) {
		t.Fatal("activating Conduit failed")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("precondition failed: the can't-cast lock was not applied by the successful cast")
	}
	return source
}

// TestConduitLockBlocksEveryLaterSpellButSparesOthers proves the applied
// restriction is unqualified: it blocks any later spell the controller would
// cast — of any card type and, because the prohibition predicate is zone-blind,
// from any zone — while leaving the opponent free to cast. A leaked spell-type
// filter would let one of these through.
func TestConduitLockBlocksEveryLaterSpellButSparesOthers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	resolveConduitCastAndLock(t, g, engine)

	creature := vanillaCreatureDef()
	instant := &game.CardDef{CardFace: game.CardFace{
		Name: "Later Bolt", Types: []types.Card{types.Instant}}}
	for _, spell := range []*game.CardDef{creature, instant} {
		if !spellCastProhibited(g, game.Player1, spell) {
			t.Fatalf("controller may still cast %q despite the lock", spell.Name)
		}
		if spellCastProhibited(g, game.Player2, spell) {
			t.Fatalf("the opponent is wrongly restricted from casting %q by the controller's self lock", spell.Name)
		}
	}
}

// TestConduitLockExpiresAtCleanup proves the restriction is only for the turn:
// the turn's cleanup removes it and the controller may cast again.
func TestConduitLockExpiresAtCleanup(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	resolveConduitCastAndLock(t, g, engine)

	expireRuleEffects(g)

	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock survived the turn's cleanup")
	}
}

// TestConduitLockPersistsWhenCastCountered proves "if you do" is linked to the
// cast having happened, not to the cast surviving: countering the cast afterward
// does not lift the restriction.
func TestConduitLockPersistsWhenCastCountered(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	resolveConduitCastAndLock(t, g, engine)

	// Countering the cast creature removes it from the stack; the lock, gated on
	// the cast having occurred, must remain in force.
	if top, ok := g.Stack.Peek(); ok {
		g.Stack.RemoveByID(top.ID)
	}
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("countering the cast lifted the can't-cast lock; it must persist because a spell was cast")
	}
}

// TestConduitLockSurvivesSourceLeaving proves the one-shot restriction is not a
// continuous effect tied to the source: it persists this turn even after Conduit
// leaves the battlefield.
func TestConduitLockSurvivesSourceLeaving(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := resolveConduitCastAndLock(t, g, engine)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("failed to move Conduit off the battlefield")
	}
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was lifted when its source left the battlefield")
	}
}

// TestConduitLockSurvivesControlChange proves the restriction stays with the
// player who cast the spell even if control of Conduit changes afterward, and
// still does not leak to the new controller.
func TestConduitLockSurvivesControlChange(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := resolveConduitCastAndLock(t, g, engine)

	source.Controller = game.Player2

	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the lock left the original caster when control of Conduit changed")
	}
	if spellCastProhibited(g, game.Player2, vanillaCreatureDef()) {
		t.Fatal("the lock followed control of Conduit to the new controller")
	}
}
