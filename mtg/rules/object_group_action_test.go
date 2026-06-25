package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// creatureBattlefieldGroup is the "all creatures" group used by the mass
// object/group action parity tests.
func creatureBattlefieldGroup() game.GroupReference {
	return game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})
}

// simultaneousIDsOfKind returns the SimultaneousID of every event of kind that
// satisfies matches, in event order.
func simultaneousIDsOfKind(events []game.Event, kind game.EventKind, matches func(game.Event) bool) []id.ID {
	var ids []id.ID
	for _, event := range events {
		if event.Kind == kind && matches(event) {
			ids = append(ids, event.SimultaneousID)
		}
	}
	return ids
}

// TestGroupDestroyDeathsShareOneSimultaneousID proves that routing the group
// destroy resolution through the shared object/group executor preserves the
// simultaneous batching CR 603.3b requires: every creature a mass destroy puts
// into the graveyard dies under a single shared, non-zero SimultaneousID, so
// "whenever one or more creatures die" triggers coalesce.
func TestGroupDestroyDeathsShareOneSimultaneousID(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{Group: creatureBattlefieldGroup()}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	ids := simultaneousIDsOfKind(g.Events, game.EventPermanentDied, func(game.Event) bool { return true })
	if len(ids) != 2 {
		t.Fatalf("death events = %d, want 2 for mass destroy of two creatures", len(ids))
	}
	var zeroID id.ID
	if ids[0] == zeroID {
		t.Fatalf("mass destroy death SimultaneousID = %v, want a non-zero batch id", ids[0])
	}
	if ids[0] != ids[1] {
		t.Fatalf("mass destroy deaths have SimultaneousIDs %v and %v, want one shared batch id", ids[0], ids[1])
	}
}

// TestGroupExileMovesAreSequentialNotBatched proves the shared executor
// preserves the exile group form's pre-existing per-permanent (non-simultaneous)
// emission: each exiled permanent leaves the battlefield under its own zero
// SimultaneousID rather than a shared batch id. Exile is not a death and the
// group form historically did not coalesce, so the refactor must not introduce
// a shared batch id.
func TestGroupExileMovesAreSequentialNotBatched(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Exile{Group: creatureBattlefieldGroup()}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	ids := simultaneousIDsOfKind(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.FromZone == zone.Battlefield && event.ToZone == zone.Exile
	})
	if len(ids) != 2 {
		t.Fatalf("exile zone-change events = %d, want 2 for mass exile of two creatures", len(ids))
	}
	var zeroID id.ID
	for i, simID := range ids {
		if simID != zeroID {
			t.Fatalf("exile move %d SimultaneousID = %v, want zero (sequential, not batched)", i, simID)
		}
	}
}

// TestGroupBounceMovesShareOneSimultaneousID proves the shared executor
// preserves the bounce group form's simultaneous batching: every bounced
// permanent returns to hand under one shared, non-zero SimultaneousID.
func TestGroupBounceMovesShareOneSimultaneousID(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Bounce{Group: creatureBattlefieldGroup()}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	ids := simultaneousIDsOfKind(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.FromZone == zone.Battlefield && event.ToZone == zone.Hand
	})
	if len(ids) != 2 {
		t.Fatalf("bounce zone-change events = %d, want 2 for mass bounce of two creatures", len(ids))
	}
	var zeroID id.ID
	if ids[0] == zeroID {
		t.Fatalf("mass bounce SimultaneousID = %v, want a non-zero batch id", ids[0])
	}
	if ids[0] != ids[1] {
		t.Fatalf("mass bounce moves have SimultaneousIDs %v and %v, want one shared batch id", ids[0], ids[1])
	}
}

// TestSingleVsGroupTapBatchingPreserved proves the shared executor preserves the
// observable difference between the single and group tap forms: a single tap
// emits under a zero SimultaneousID, while a group tap batches every tap under
// one shared, non-zero id.
func TestSingleVsGroupTapBatchingPreserved(t *testing.T) {
	var zeroID id.ID

	single := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	perm := addCreaturePermanent(single, game.Player1)
	addEffectSpellToStack(single, game.Player1, game.Tap{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(perm.ObjectID)})
	engine.resolveTopOfStack(single, &TurnLog{})
	singleIDs := simultaneousIDsOfKind(single.Events, game.EventPermanentTapped, func(game.Event) bool { return true })
	if len(singleIDs) != 1 {
		t.Fatalf("single tap events = %d, want 1", len(singleIDs))
	}
	if singleIDs[0] != zeroID {
		t.Fatalf("single tap SimultaneousID = %v, want zero (unbatched)", singleIDs[0])
	}

	groupGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(groupGame, game.Player1)
	addCreaturePermanent(groupGame, game.Player2)
	addEffectSpellToStack(groupGame, game.Player1, game.Tap{Group: creatureBattlefieldGroup()}, nil)
	engine.resolveTopOfStack(groupGame, &TurnLog{})
	groupIDs := simultaneousIDsOfKind(groupGame.Events, game.EventPermanentTapped, func(game.Event) bool { return true })
	if len(groupIDs) != 2 {
		t.Fatalf("group tap events = %d, want 2", len(groupIDs))
	}
	if groupIDs[0] == zeroID {
		t.Fatalf("group tap SimultaneousID = %v, want a non-zero batch id", groupIDs[0])
	}
	if groupIDs[0] != groupIDs[1] {
		t.Fatalf("group taps have SimultaneousIDs %v and %v, want one shared batch id", groupIDs[0], groupIDs[1])
	}
}

// TestSingleObjectDestroyExileBounceResolveThroughSharedExecutor proves the
// single-object forms still resolve their target and apply the correct terminal
// after migrating resolution onto the shared object/group executor.
func TestSingleObjectDestroyExileBounceResolveThroughSharedExecutor(t *testing.T) {
	cases := []struct {
		name      string
		primitive func(target id.ID) game.Primitive
		toZone    zone.Type
	}{
		{"destroy", func(id.ID) game.Primitive { return game.Destroy{Object: game.TargetPermanentReference(0)} }, zone.Graveyard},
		{"exile", func(id.ID) game.Primitive { return game.Exile{Object: game.TargetPermanentReference(0)} }, zone.Exile},
		{"bounce", func(id.ID) game.Primitive { return game.Bounce{Object: game.TargetPermanentReference(0)} }, zone.Hand},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreaturePermanent(g, game.Player2)
			addEffectSpellToStack(g, game.Player1, tc.primitive(target.ObjectID), []game.Target{game.PermanentTarget(target.ObjectID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
				return event.PermanentID == target.ObjectID &&
					event.FromZone == zone.Battlefield &&
					event.ToZone == tc.toZone
			})
		})
	}
}
