package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestNextCastEntersWithCounterReplacementBindsToCardInstance verifies the
// runtime half of Summon: Fenrir chapter II. The chapter's resolved ability
// registers an enters-the-battlefield replacement bound to a specific card
// instance (the future-cast creature spell) by AffectedCardID. The binding must
// be by card instance, not object ID, because a permanent spell gains a fresh
// object ID as it resolves onto the battlefield. The bound creature enters with
// the extra +1/+1 counter; an unrelated creature sharing the same card
// definition does not.
func TestNextCastEntersWithCounterReplacementBindsToCardInstance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Howled Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	boundID := addCardInstance(g, game.Player1, creatureDef)
	otherID := addCardInstance(g, game.Player1, creatureDef)

	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID:                 g.IDGen.Next(),
		Controller:         game.Player1,
		Duration:           game.DurationUntilEndOfTurn,
		CreatedTurn:        g.Turn.TurnNumber,
		MatchEvent:         game.EventPermanentEnteredBattlefield,
		AffectedCardID:     boundID,
		EntersWithCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
	})

	otherPerm, ok := createCardPermanentFace(g, g.CardInstances[otherID], game.Player1, zone.Stack, game.FaceFront)
	if !ok {
		t.Fatal("create unrelated creature permanent failed")
	}
	if got := otherPerm.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("unrelated creature +1/+1 counters = %d, want 0", got)
	}

	boundPerm, ok := createCardPermanentFace(g, g.CardInstances[boundID], game.Player1, zone.Stack, game.FaceFront)
	if !ok {
		t.Fatal("create bound creature permanent failed")
	}
	if got := boundPerm.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("bound creature +1/+1 counters = %d, want 1", got)
	}
}

// TestCreateReplacementBindsEventStackObjectToCardInstance proves the resolution
// half of Summon: Fenrir chapter II: resolving CreateReplacement with an
// event-stack-object reference ("that creature") reads the triggering spell's
// card instance ID off the stack and stores it on the replacement's
// AffectedCardID, so the floating replacement later matches that specific
// creature as it enters.
func TestCreateReplacementBindsEventStackObjectToCardInstance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Howled Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	castCardID := addCardInstance(g, game.Player1, creatureDef)
	castSpell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   castCardID,
		Controller: game.Player1,
	}
	g.Stack.Push(castSpell)

	trigger := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:          game.EventSpellCast,
			Controller:    game.Player1,
			CardID:        castCardID,
			StackObjectID: castSpell.ID,
		},
	}

	resolveInstruction(engine, g, trigger, game.CreateReplacement{
		Object:   game.EventStackObjectReference(),
		Duration: game.DurationUntilEndOfTurn,
		Replacement: &game.ReplacementEffect{
			MatchEvent:         game.EventPermanentEnteredBattlefield,
			EntersWithCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
		},
	}, &TurnLog{})

	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("registered replacement effects = %d, want 1", len(g.ReplacementEffects))
	}
	if got := g.ReplacementEffects[0].AffectedCardID; got != castCardID {
		t.Fatalf("replacement AffectedCardID = %v, want triggering spell card %v", got, castCardID)
	}
	if got := g.ReplacementEffects[0].AffectedObjectID; got != 0 {
		t.Fatalf("replacement AffectedObjectID = %v, want 0 (bound by card instance, not object)", got)
	}
}
