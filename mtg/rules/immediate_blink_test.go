package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// immediateBlinkInstructions builds the exile-then-return instruction pair that
// the cardgen immediate-blink lowerer emits for "Exile target creature you
// control, then return that card to the battlefield ...".
func immediateBlinkInstructions(put game.PutOnBattlefield) []game.Instruction {
	put.Source = game.LinkedBattlefieldSource("blink")
	return []game.Instruction{
		{Primitive: game.Exile{Object: game.TargetPermanentReference(0), ExileLinkedKey: "blink"}},
		{
			Primitive: put,
			CardCondition: opt.Val(game.CardCondition{
				Card:                 game.CardReference{Kind: game.CardReferenceLinked, LinkID: "blink"},
				RequirePermanentCard: true,
			}),
		},
	}
}

func resolveImmediateBlink(t *testing.T, put game.PutOnBattlefield) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Blinked Creature",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blink Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}
	instrs := immediateBlinkInstructions(put)
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("original creature object remained on battlefield after exile")
	}
	var returned *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == target.CardInstanceID {
			returned = permanent
		}
	}
	if returned == nil {
		t.Fatal("blinked creature card did not return to the battlefield")
	}
	if returned.ObjectID == target.ObjectID {
		t.Fatal("returned permanent reused the original object identity")
	}
	return g, returned
}

func TestImmediateBlinkReturnsCreatureUnderOwnersControl(t *testing.T) {
	_, returned := resolveImmediateBlink(t, game.PutOnBattlefield{})
	if returned.Controller != game.Player1 || returned.Owner != game.Player1 {
		t.Fatalf("returned permanent controller/owner = %v/%v, want owner Player1", returned.Controller, returned.Owner)
	}
	if returned.Tapped {
		t.Fatal("returned permanent entered tapped, want untapped")
	}
}

func TestImmediateBlinkReturnsCreatureTappedWithCounter(t *testing.T) {
	_, returned := resolveImmediateBlink(t, game.PutOnBattlefield{
		EntryTapped:   true,
		EntryCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
	})
	if !returned.Tapped {
		t.Fatal("returned permanent did not enter tapped")
	}
	if got := returned.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("returned permanent +1/+1 counters = %d, want 1", got)
	}
}

func TestImmediateBlinkReturnsCreatureUnderYourControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The exiled creature is owned by Player2; "under your control" must hand
	// control to the spell's controller (Player1) while preserving ownership.
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Stolen Creature",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blink Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}
	put := game.PutOnBattlefield{Recipient: opt.Val(game.ControllerReference())}
	instrs := immediateBlinkInstructions(put)
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
	var returned *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == target.CardInstanceID {
			returned = permanent
		}
	}
	if returned == nil {
		t.Fatal("blinked creature card did not return to the battlefield")
	}
	if returned.Controller != game.Player1 {
		t.Fatalf("returned permanent controller = %v, want Player1 (your control)", returned.Controller)
	}
	if returned.Owner != game.Player2 {
		t.Fatalf("returned permanent owner = %v, want Player2 (ownership preserved)", returned.Owner)
	}
}

// selfBlinkInstructions builds the exile-then-return instruction pair the
// cardgen immediate-blink lowerer emits for the self-blink "Exile this
// creature, then return it to the battlefield under its owner's control." The
// exile names the source permanent rather than a target.
func selfBlinkInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Exile{Object: game.SourcePermanentReference(), ExileLinkedKey: "blink"}},
		{
			Primitive: game.PutOnBattlefield{Source: game.LinkedBattlefieldSource("blink")},
			CardCondition: opt.Val(game.CardCondition{
				Card:                 game.CardReference{Kind: game.CardReferenceLinked, LinkID: "blink"},
				RequirePermanentCard: true,
			}),
		},
	}
}

func TestImmediateSelfBlinkReturnsSourceUnderOwnersControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Flickering Spirit",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   source.ObjectID,
		Controller: game.Player1,
	}
	log := TurnLog{}
	instrs := selfBlinkInstructions()
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}
	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source creature object remained on battlefield after self-exile")
	}
	var returned *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == source.CardInstanceID {
			returned = permanent
		}
	}
	if returned == nil {
		t.Fatal("self-blinked creature card did not return to the battlefield")
	}
	if returned.ObjectID == source.ObjectID {
		t.Fatal("returned permanent reused the original object identity")
	}
	if returned.Controller != game.Player1 || returned.Owner != game.Player1 {
		t.Fatalf("returned permanent controller/owner = %v/%v, want Player1", returned.Controller, returned.Owner)
	}
}
