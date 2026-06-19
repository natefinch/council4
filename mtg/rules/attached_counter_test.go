package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestAddCounterOnSourceAttachedPermanentTargetsEnchantedCreature proves the
// runtime resolution of AddCounter{Object: SourceAttachedPermanentReference()}
// places the counters on the creature the Aura source is attached to and on no
// other permanent. This is the recipient the "put a +1/+1 counter on enchanted
// creature" Aura lowering relies on.
func TestAddCounterOnSourceAttachedPermanentTargetsEnchantedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchanted := addCombatCreaturePermanent(g, game.Player1)
	bystander := addCombatCreaturePermanent(g, game.Player1)
	aura := addAuraPermanent(g, game.Player1)
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent() = false, want true")
	}

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   aura.ObjectID,
	}

	resolveInstruction(engine, g, obj, game.AddCounter{
		Amount:      game.Fixed(2),
		Object:      game.SourceAttachedPermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
	}, &TurnLog{})

	if got := enchanted.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("enchanted creature +1/+1 counters = %d, want 2", got)
	}
	if got := bystander.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("bystander creature +1/+1 counters = %d, want 0", got)
	}
}

// TestAddCounterOnSourceAttachedPermanentNoOpWhenUnattached proves the runtime
// fails closed: an unattached source places no counter anywhere, matching the
// reference resolver returning false when AttachedTo is empty.
func TestAddCounterOnSourceAttachedPermanentNoOpWhenUnattached(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bystander := addCombatCreaturePermanent(g, game.Player1)
	aura := addAuraPermanent(g, game.Player1)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   aura.ObjectID,
	}

	resolveInstruction(engine, g, obj, game.AddCounter{
		Amount:      game.Fixed(1),
		Object:      game.SourceAttachedPermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
	}, &TurnLog{})

	if got := bystander.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("bystander creature +1/+1 counters = %d, want 0", got)
	}
}
