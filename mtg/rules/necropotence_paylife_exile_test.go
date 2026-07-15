package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// necropotencePayLifeAbility returns the registered card's activated ability and
// asserts it is the "Pay 1 life:" ability whose body exiles the top card of the
// controller's library face down under a link key and then schedules the delayed
// return keyed to that same object.
func necropotencePayLifeAbility(t *testing.T) game.ActivatedAbility {
	t.Helper()
	ability := necropotenceCardDef(t).ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
		ability.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("activated ability cost = %+v, want a single Pay 1 life", ability.AdditionalCosts)
	}
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("activated ability body has %d instructions, want 2 (exile, delayed return)", len(seq))
	}
	exile, ok := seq[0].Primitive.(game.ExileTopOfLibrary)
	if !ok {
		t.Fatalf("first instruction is %T, want game.ExileTopOfLibrary", seq[0].Primitive)
	}
	if !exile.FaceDown {
		t.Fatal("top-of-library exile must be face down")
	}
	if exile.Amount != game.Fixed(1) {
		t.Fatalf("exile amount = %v, want exactly the single top card", exile.Amount)
	}
	if exile.PublishLinked == "" {
		t.Fatal("exile must publish a link key so the delayed return can capture the exiled card")
	}
	if _, ok := seq[1].Primitive.(game.CreateDelayedTrigger); !ok {
		t.Fatalf("second instruction is %T, want game.CreateDelayedTrigger", seq[1].Primitive)
	}
	return ability
}

// setUpNecropotenceActivation puts Necropotence onto the battlefield for Player1
// and configures a precombat main phase in which Player1 holds priority with the
// given life total, the timing in which its "Pay 1 life:" ability is activatable.
func setUpNecropotenceActivation(g *game.Game, life int) *game.Permanent {
	source := addNecropotence(g, game.Player1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].Life = life
	return source
}

func TestNecropotencePayLifeExilesTopCardFaceDownAtResolution(t *testing.T) {
	necropotencePayLifeAbility(t) // structural assertions on the real definition
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Card", Types: []types.Card{types.Creature}}})

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Pay 1 life ability is not a legal action with life to spare")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the Pay 1 life ability failed")
	}
	if got := g.Players[game.Player1].Life; got != 19 {
		t.Fatalf("life = %d, want 19 (paid exactly 1 life on activation)", got)
	}
	// The exile is a resolution effect, not an activation cost: nothing leaves the
	// library until the ability resolves.
	if g.Players[game.Player1].Exile.Contains(topID) {
		t.Fatal("top card was exiled on activation; it must be exiled at resolution")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(topID) {
		t.Fatal("resolution did not exile the top card")
	}
	if !g.Players[game.Player1].Exile.IsFaceDown(topID) {
		t.Fatal("exiled top card is not face down")
	}
	if got := g.Players[game.Player1].Exile.Size(); got != 1 {
		t.Fatalf("exile size = %d, want exactly the single top card", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1 (only the top card left)", got)
	}
}

// TestNecropotencePayLifeExilesTopCardChosenAtResolution proves the exiled card
// is the top card at the moment of resolution, not activation: a card placed on
// top after the ability is on the stack is the one exiled.
func TestNecropotencePayLifeExilesTopCardChosenAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	originalTop := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Original Top"}})

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the Pay 1 life ability failed")
	}
	// A new card arrives on top of the library while the ability is on the stack.
	newTop := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "New Top"}})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(newTop) {
		t.Fatal("resolution exiled a card other than the top card at resolution time")
	}
	if g.Players[game.Player1].Exile.Contains(originalTop) {
		t.Fatal("resolution exiled the top card as of activation, not as of resolution")
	}
}

// TestNecropotencePayLifeUnpayableWithoutLife proves the Pay 1 life cost is a real
// payment gate: a player at 0 life cannot activate the ability (CR 119.4 forbids
// paying life a player does not have).
func TestNecropotencePayLifeUnpayableWithoutLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 0)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Pay 1 life ability was legal at 0 life; the life payment must be unpayable")
	}
}

// TestNecropotencePayLifeEmptyLibraryNoOps proves the ability is still activatable
// with an empty library (the life is paid) but resolves to nothing: an empty
// library exiles no card and schedules no returnable object.
func TestNecropotencePayLifeEmptyLibraryNoOps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the Pay 1 life ability with an empty library failed")
	}
	if got := g.Players[game.Player1].Life; got != 19 {
		t.Fatalf("life = %d, want 19 (life paid even with an empty library)", got)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Exile.Size(); got != 0 {
		t.Fatalf("exile size = %d, want 0 (empty library exiles nothing)", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want 0 (nothing to return later)", got)
	}
}

// TestNecropotencePayLifeExiledCardIdentityHidden proves the face-down exile keeps
// hidden information hidden: neither the owner nor the opponent sees the exiled
// card's identity in their public observation of the exile zone.
func TestNecropotencePayLifeExiledCardIdentityHidden(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Secret Card", Types: []types.Card{types.Creature}}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("activating the Pay 1 life ability failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, observer := range []game.PlayerID{game.Player1, game.Player2} {
		views := NewObservation(g, observer).Exile(game.Player1)
		var found bool
		for _, view := range views {
			if view.CardInstanceID != topID {
				continue
			}
			found = true
			if view.Name != "" {
				t.Fatalf("observer %v sees exiled card name %q; a face-down card's identity must be hidden", observer, view.Name)
			}
		}
		if !found {
			t.Fatalf("observer %v does not see the face-down card object at all in Player1's exile", observer)
		}
	}
}

// TestNecropotencePayLifeIsPerControllerLibrary proves two Necropotences on
// different controllers act independently: Player1's activation exiles the top
// card of Player1's own library, costs only Player1 life, and never reaches
// Player2's library or exile even though Player2 also controls a Necropotence.
func TestNecropotencePayLifeIsPerControllerLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source1 := setUpNecropotenceActivation(g, 20)
	addNecropotence(g, game.Player2)
	g.Players[game.Player2].Life = 20
	p1Top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "P1 Top"}})
	p2Top := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "P2 Top"}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source1.ObjectID, 0, nil, 0)) {
		t.Fatal("Player1 activation failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(p1Top) {
		t.Fatal("Player1's activation did not exile Player1's top card")
	}
	if g.Players[game.Player2].Exile.Contains(p1Top) || g.Players[game.Player1].Exile.Contains(p2Top) {
		t.Fatal("an activation reached the wrong player's library or exile")
	}
	if got := g.Players[game.Player2].Exile.Size(); got != 0 {
		t.Fatalf("Player2 exile size = %d, want 0 (untouched by Player1's activation)", got)
	}
	if got := g.Players[game.Player2].Life; got != 20 {
		t.Fatalf("Player2 life = %d, want 20 (unaffected by Player1's activation)", got)
	}
}
