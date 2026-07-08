package rules

import (
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// locthwainUpkeepSequence extracts the two-instruction upkeep sequence from the
// registered Court of Locthwain card definition. Sourcing the instructions from
// the real card proves the generated ability — not a hand-written stand-in —
// drives the runtime behavior, and that the link key the exile publishes matches
// the one the monarch free cast reads.
func locthwainUpkeepSequence(t *testing.T) []game.Instruction {
	t.Helper()
	card := cardc.CourtOfLocthwain
	if len(card.TriggeredAbilities) != 2 {
		t.Fatalf("Court of Locthwain has %d triggered abilities, want 2", len(card.TriggeredAbilities))
	}
	upkeep := card.TriggeredAbilities[1]
	if upkeep.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("triggered ability 1 is not the upkeep trigger: %+v", upkeep.Trigger)
	}
	if len(upkeep.Content.Modes) != 1 {
		t.Fatalf("upkeep content has %d modes, want 1", len(upkeep.Content.Modes))
	}
	seq := upkeep.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("upkeep sequence has %d instructions, want 2", len(seq))
	}
	return seq
}

// TestCourtOfLocthwainCardDefUpkeepEndToEnd resolves the real Court of Locthwain
// upkeep ability from the registered card definition against a live game: it
// exiles the top card of the target opponent's library (leaving it in that
// opponent's exile, since the opponent owns it), grants the controller an
// any-mana play-from-exile permission recorded under the enchantment's linked
// pool, and — while the controller is the monarch — installs the one-shot free
// cast that lets the controller cast that opponent-owned exiled card without
// paying its cost.
func TestCourtOfLocthwainCardDefUpkeepEndToEnd(t *testing.T) {
	seq := locthwainUpkeepSequence(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Opponent Spell",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}})

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}

	// Clauses 1 and 2: exile the target opponent's top card and grant the
	// controller an any-mana play-from-exile permission over it.
	engine.resolveInstructionWithChoices(g, obj, &seq[0], [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("upkeep did not exile the top card of the target opponent's library")
	}
	if !castFromZoneAllowsAnyMana(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("controller did not receive an any-mana play-from-exile permission for the exiled card")
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, courtOfLocthwainLink))
	if len(refs) != 1 || refs[0].CardID != topID {
		t.Fatalf("linked pool = %v, want one ref to the exiled card %v", refs, topID)
	}

	// Clause 3: while the controller is the monarch, the free-cast permission is
	// installed over the enchantment's linked pool and lets them cast the exiled
	// card — which lives in the opponent's exile — without paying its cost.
	g.Players[game.Player1].IsMonarch = true
	engine.resolveInstructionWithChoices(g, obj, &seq[1], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !castLinkedExileForFree(g, game.Player1, topID) {
		t.Fatal("monarch controller was not granted the free cast over the linked pool")
	}
	setSorcerySpeedTurn(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("monarch could not cast the opponent-owned exiled card for free")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d, want 0 (the free cast paid no mana)", got)
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != topID {
		t.Fatalf("stack top = %#v, want the exiled card %v cast for free", top, topID)
	}
}

// TestCourtOfLocthwainCardDefFreeCastGatedOffWhenNotMonarch proves the monarch
// gate on the real card definition: when the controller is not the monarch, the
// upkeep's conditional instruction installs no free-cast permission over the
// enchantment's linked pool.
func TestCourtOfLocthwainCardDefFreeCastGatedOffWhenNotMonarch(t *testing.T) {
	seq := locthwainUpkeepSequence(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Opponent Spell",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}})

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}

	// The controller is not the monarch, so only the exile/permission clause runs.
	engine.resolveInstructionWithChoices(g, obj, &seq[0], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, obj, &seq[1], [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if castLinkedExileForFree(g, game.Player1, topID) {
		t.Fatal("a non-monarch controller must not receive the free cast over the linked pool")
	}
}
