package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// expressiveIterationDig returns the Dig shape that models Expressive Iteration:
// look at the top three cards, put one into your hand (the primary Take), put one
// on the bottom of your library, and exile one that may be played this turn. The
// slots route the looked-at cards into their printed-order destinations from a
// single look; no card name or Oracle text appears at runtime.
func expressiveIterationDig() game.Dig {
	return game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Slots: []game.DigSlot{
			{Count: game.Fixed(1), Destination: zone.Library, Bottom: true},
			{Count: game.Fixed(1), Destination: zone.Exile, Play: opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn})},
		},
	}
}

// playFromExileGrant returns the play-or-cast-from-exile rule effect recorded for
// cardID, or nil when none exists.
func playFromExileGrant(g *game.Game, cardID id.ID) *game.RuleEffect {
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.AffectedCardID != cardID || effect.CastFromZone != zone.Exile {
			continue
		}
		if effect.Kind == game.RuleEffectPlayFromZone || effect.Kind == game.RuleEffectCastFromZone {
			return effect
		}
	}
	return nil
}

// TestDigSlotsRouteExactLibraryByChoice verifies the full Expressive Iteration
// routing on a library with exactly the looked-at count: the digging player's
// choices send one card to hand, one to the bottom of the library, and one to
// exile with a play-this-turn permission, with no card duplicated or lost.
func TestDigSlotsRouteExactLibraryByChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Added bottom-to-top, so peekLibrary sees [c3, c2, c1].
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, expressiveIterationDig(), nil)
	log := TurnLog{}
	// Seen [c3, c2, c1]: take index 1 (c2) to hand; then from [c3, c1] bottom
	// index 0 (c3); then from [c1] exile index 0 (c1).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}, {0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) || player.Hand.Contains(c1) || player.Hand.Contains(c3) {
		t.Fatal("hand routing wrong: want only c2 in hand")
	}
	if !player.Library.Contains(c3) || player.Library.Contains(c1) || player.Library.Contains(c2) {
		t.Fatal("bottom routing wrong: want only c3 in library")
	}
	if !player.Exile.Contains(c1) || player.Exile.Contains(c2) || player.Exile.Contains(c3) {
		t.Fatal("exile routing wrong: want only c1 in exile")
	}
	if grant := playFromExileGrant(g, c1); grant == nil {
		t.Fatal("no play-from-exile permission recorded for the exiled card")
	} else if grant.Kind != game.RuleEffectPlayFromZone || grant.Duration != game.DurationThisTurn || grant.ExpiresFor != game.Player1 {
		t.Fatalf("play grant = %+v, want play-from-zone this turn expiring for Player1", *grant)
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c2 && event.FromZone == zone.Library && event.ToZone == zone.Hand
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c3 && event.FromZone == zone.Library && event.ToZone == zone.Library
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c1 && event.FromZone == zone.Library && event.ToZone == zone.Exile
	})
	// "Look at" is hidden: none of the looked-at cards is revealed.
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			t.Fatal("dig slots revealed a looked-at card, but the cards are only looked at")
		}
	}
	if len(log.Choices) != 3 {
		t.Fatalf("expected 3 dig choices (hand, bottom, exile), got %d", len(log.Choices))
	}
	for _, choice := range log.Choices {
		if choice.Request.Kind != game.ChoiceDig || choice.UsedFallback {
			t.Fatalf("choice = %+v, want non-fallback dig choice", choice)
		}
	}
}

// TestDigSlotsShortLibraryRoutesAsMuchAsPossible verifies that when the library
// holds fewer cards than the looked-at count, the ordered slots take as much as
// possible in printed order: with two cards the hand and library-bottom slots each
// take one and the exile slot takes none, so nothing is exiled and no play
// permission is granted.
func TestDigSlotsShortLibraryRoutesAsMuchAsPossible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, expressiveIterationDig(), nil)
	log := TurnLog{}
	// Seen [c2, c1]: take index 0 (c2) to hand; then bottom from [c1] index 0
	// (c1). The exile slot has no card left, so it requests no choice.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) {
		t.Fatal("short-library dig did not put a card into hand")
	}
	if !player.Library.Contains(c1) {
		t.Fatal("short-library dig did not bottom the second card")
	}
	if player.Exile.Contains(c1) || player.Exile.Contains(c2) {
		t.Fatal("short-library dig exiled a card even though no card remained for the exile slot")
	}
	if playFromExileGrant(g, c1) != nil || playFromExileGrant(g, c2) != nil {
		t.Fatal("short-library dig granted a play permission with nothing exiled")
	}
	if len(log.Choices) != 2 {
		t.Fatalf("expected 2 choices (hand, bottom) with an empty exile slot, got %d", len(log.Choices))
	}
}

// TestDigSlotsSingleCardFillsFirstSlotOnly verifies that a one-card library fills
// only the first (hand) slot; the bottom and exile slots take nothing.
func TestDigSlotsSingleCardFillsFirstSlotOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only"}})
	addEffectSpellToStack(g, game.Player1, expressiveIterationDig(), nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c1) {
		t.Fatal("single-card dig did not put the only card into hand")
	}
	if player.Exile.Contains(c1) || player.Library.Contains(c1) {
		t.Fatal("single-card dig routed the card to more than one zone")
	}
	if len(log.Choices) != 1 {
		t.Fatalf("expected 1 choice for a single-card library, got %d", len(log.Choices))
	}
}

// TestDigSlotsEmptyLibraryDoesNothing verifies an empty library resolves with no
// routing, no choices, and no play permission rather than panicking.
func TestDigSlotsEmptyLibraryDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, expressiveIterationDig(), nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if player.Hand.Size() != 0 || player.Exile.Size() != 0 {
		t.Fatal("empty-library dig moved a card")
	}
	if len(g.RuleEffects) != 0 {
		t.Fatalf("empty-library dig recorded rule effects: %+v", g.RuleEffects)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("empty-library dig asked for choices: %+v", log.Choices)
	}
}

// TestDigSlotsFallbackRoutesTopCards verifies that, without a choosing agent, the
// slots route deterministically from the top: the top card to hand, the next to
// the bottom, and the last to exile, so the routing never stalls.
func TestDigSlotsFallbackRoutesTopCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, expressiveIterationDig(), nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	player := g.Players[game.Player1]
	// Seen [c3, c2, c1]: fallback takes c3 to hand, c2 to the bottom, c1 to exile.
	if !player.Hand.Contains(c3) {
		t.Fatal("fallback dig did not take the top card to hand")
	}
	if !player.Library.Contains(c2) {
		t.Fatal("fallback dig did not bottom the middle card")
	}
	if !player.Exile.Contains(c1) {
		t.Fatal("fallback dig did not exile the last card")
	}
}

// TestDigSlotExileGrantsLandPlayThatExpires verifies the exile slot's play
// permission lets the controller play an exiled land (not merely cast a spell)
// and expires at the turn's cleanup, matching "you may play the exiled card this
// turn."
func TestDigSlotExileGrantsLandPlayThatExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// c1 (deepest) is the land the fallback routing sends to the exile slot.
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}, expressiveIterationDig(), &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(c1) {
		t.Fatal("the exile slot did not exile the last looked-at card")
	}
	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, c1, zone.Exile) {
		t.Fatal("the exile slot did not grant land-play permission")
	}

	expireRuleEffects(g)
	if playFromExileGrant(g, c1) != nil {
		t.Fatal("the play-this-turn permission survived the turn's cleanup")
	}
}

// TestDigSlotExileCastGrantIsCastOnly verifies that an exile slot whose play
// grant sets Cast grants cast-only permission (an exiled spell may be cast but an
// exiled land may not be played), distinguishing it from the ordinary play grant.
func TestDigSlotExileCastGrantIsCastOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spell", Types: []types.Card{types.Sorcery}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	dig := expressiveIterationDig()
	dig.Slots[1].Play = opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn, Cast: true})
	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}, dig, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(c1) {
		t.Fatal("the exile slot did not exile the spell")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, c1, zone.Exile, game.FaceFront) {
		t.Fatal("the cast grant did not permit casting the exiled spell")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, c1, zone.Exile) {
		t.Fatal("the cast-only grant wrongly permitted playing a land")
	}
}

// TestDigSlotGraveyardRemovesFromLibrary proves that a graveyard slot moves the
// chosen card from the library into the graveyard without leaving a duplicate in
// the library: putLibraryCardIntoGraveyard places the card but does not remove
// it from the source, so the slot must remove it first.
func TestDigSlotGraveyardRemovesFromLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(2),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Slots:     []game.DigSlot{{Count: game.Fixed(1), Destination: zone.Graveyard}},
	}, nil)
	// Seen [c2, c1]: take index 0 (c2) to hand; then graveyard from [c1] index 0 (c1).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) {
		t.Fatal("graveyard-slot dig did not put the chosen card into hand")
	}
	if !player.Graveyard.Contains(c1) {
		t.Fatal("graveyard-slot dig did not put the card into the graveyard")
	}
	if player.Library.Contains(c1) {
		t.Fatal("graveyard-slot dig left a duplicate of the card in the library")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c1 && event.FromZone == zone.Library && event.ToZone == zone.Graveyard
	})
}

// TestDigWithoutSlotsUnchanged guards that an ordinary two-way dig (no slots)
// still takes the chosen card to hand and bottoms the rest, so existing dig
// callers are unaffected by the slot routing.
func TestDigWithoutSlotsUnchanged(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c3) {
		t.Fatal("plain dig did not take the chosen card to hand")
	}
	if !player.Library.Contains(c1) || !player.Library.Contains(c2) {
		t.Fatal("plain dig did not bottom the unchosen cards")
	}
	if player.Exile.Size() != 0 {
		t.Fatal("plain dig exiled a card despite having no slots")
	}
}
