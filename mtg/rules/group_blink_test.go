package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// resolveGroupBlink exiles every permanent chosen for target spec 0 under one
// linked key and returns the whole group with a single put, mirroring the
// instructions the cardgen group-blink lowerer emits for "Exile any number of
// target creatures you control. Return those cards to the battlefield ...".
func resolveGroupBlink(t *testing.T, put game.PutOnBattlefield, count int) (g *game.Game, originals []*game.Permanent, returned []*game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	originals = make([]*game.Permanent, 0, count)
	targets := make([]game.Target, 0, count)
	for range count {
		permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Blinked Creature",
			Types: []types.Card{types.Creature},
		}})
		originals = append(originals, permanent)
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	put.Source = game.LinkedBattlefieldSource("group-blink")
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Group Blink Spell"}}),
		Controller:   game.Player1,
		Targets:      targets,
		TargetCounts: []int{count},
	}
	log := TurnLog{}
	instrs := []game.Instruction{
		{Primitive: game.Exile{Object: game.AllTargetPermanentsReference(0), ExileLinkedKey: "group-blink"}},
		{Primitive: put},
	}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
	for _, original := range originals {
		if _, ok := permanentByObjectID(g, original.ObjectID); ok {
			t.Fatal("original creature object remained on battlefield after exile")
		}
	}
	returned = make([]*game.Permanent, 0, count)
	for _, original := range originals {
		for _, permanent := range g.Battlefield {
			if permanent.CardInstanceID == original.CardInstanceID {
				returned = append(returned, permanent)
			}
		}
	}
	return g, originals, returned
}

func TestGroupBlinkReturnsEveryTargetedPermanent(t *testing.T) {
	_, originals, returned := resolveGroupBlink(t, game.PutOnBattlefield{}, 3)
	if len(returned) != len(originals) {
		t.Fatalf("returned %d permanents, want %d", len(returned), len(originals))
	}
	for _, permanent := range returned {
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 {
			t.Fatalf("returned permanent controller/owner = %v/%v, want Player1", permanent.Controller, permanent.Owner)
		}
		if permanent.Tapped {
			t.Fatal("returned permanent entered tapped, want untapped")
		}
	}
	for _, original := range originals {
		for _, permanent := range returned {
			if permanent.ObjectID == original.ObjectID {
				t.Fatal("returned permanent reused the original object identity")
			}
		}
	}
}

func TestGroupBlinkReturnsTappedWithCounter(t *testing.T) {
	_, _, returned := resolveGroupBlink(t, game.PutOnBattlefield{
		EntryTapped:   true,
		EntryCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
	}, 2)
	if len(returned) != 2 {
		t.Fatalf("returned %d permanents, want 2", len(returned))
	}
	for _, permanent := range returned {
		if !permanent.Tapped {
			t.Fatal("returned permanent did not enter tapped")
		}
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
			t.Fatalf("returned permanent +1/+1 counters = %d, want 1", got)
		}
	}
}

// resolveMassGroupBlink exiles every creature its controller controls under one
// linked key and returns them, mirroring the instructions the cardgen
// mass-group-blink lowerer emits for "Exile each creature you control. Return
// those cards to the battlefield ...".
func resolveMassGroupBlink(t *testing.T, count int) []*game.Permanent {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	originals := make([]*game.Permanent, 0, count)
	for range count {
		permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Blinked Creature",
			Types: []types.Card{types.Creature},
		}})
		originals = append(originals, permanent)
	}
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypesAny: []types.Card{types.Creature},
		Controller:       game.ControllerYou,
	})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mass Blink Spell"}}),
		Controller: game.Player1,
	}
	log := TurnLog{}
	put := game.PutOnBattlefield{Source: game.LinkedBattlefieldSource("group-blink")}
	put.Recipient = opt.Val(game.ControllerReference())
	instrs := []game.Instruction{
		{Primitive: game.Exile{Group: group, ExileLinkedKey: "group-blink"}},
		{Primitive: put},
	}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
	for _, original := range originals {
		if _, ok := permanentByObjectID(g, original.ObjectID); ok {
			t.Fatal("original creature object remained on battlefield after mass exile")
		}
	}
	returned := make([]*game.Permanent, 0, count)
	for _, original := range originals {
		for _, permanent := range g.Battlefield {
			if permanent.CardInstanceID == original.CardInstanceID {
				returned = append(returned, permanent)
			}
		}
	}
	return returned
}

func TestMassGroupBlinkReturnsEveryControlledPermanent(t *testing.T) {
	returned := resolveMassGroupBlink(t, 3)
	if len(returned) != 3 {
		t.Fatalf("returned %d permanents, want 3", len(returned))
	}
	for _, permanent := range returned {
		if permanent.Controller != game.Player1 {
			t.Fatalf("returned permanent controller = %v, want Player1", permanent.Controller)
		}
	}
}
