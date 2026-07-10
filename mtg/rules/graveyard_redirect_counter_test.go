package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// graveyardRedirectCounterPermanent registers a Dauthi Voidwalker-style
// continuous redirect that exiles a card an opponent owns with a named counter
// on it ("If a card would be put into an opponent's graveyard from anywhere,
// instead exile it with a void counter on it.").
func graveyardRedirectCounterPermanent(g *game.Game, controller game.PlayerID, kind counter.Kind) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Voidwalker",
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectExileWithCounterReplacement(
					"redirect",
					game.TriggerControllerOpponent,
					game.TriggerControllerAny,
					false,
					kind,
				),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

// TestGraveyardRedirectCounterExilesOpponentCardWithCounter verifies that an
// opponent-owned card headed to its owner's graveyard from hand is exiled with
// the named counter placed on it, while the redirect's controller's own card is
// left untouched (opponent scope) and never receives a counter.
func TestGraveyardRedirectCounterExilesOpponentCardWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	oppCard := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Their Card"}})
	if !moveCardBetweenZones(g, game.Player2, oppCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(oppCard) {
		t.Fatal("redirect did not move the opponent's card away from the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(oppCard) {
		t.Fatal("redirect did not exile the opponent's card")
	}
	if !g.HasExileCounter(oppCard, counter.Void) {
		t.Fatal("redirect exiled the opponent's card without a void counter")
	}

	ownCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "My Card"}})
	if !moveCardBetweenZones(g, game.Player1, ownCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCard) {
		t.Fatal("opponent-scoped redirect wrongly exiled the controller's own card")
	}
	if g.HasExileCounter(ownCard, counter.Void) {
		t.Fatal("opponent-scoped redirect placed a counter on the controller's own card")
	}
}

// TestGraveyardRedirectCounterExilesDyingPermanentWithCounter verifies that the
// counter rider is placed when the redirected card leaves the battlefield (a
// dying creature an opponent controls), covering the permanent-death move path.
func TestGraveyardRedirectCounterExilesDyingPermanentWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	permanent := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Dies", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player2].Exile.Contains(permanent.CardInstanceID) {
		t.Fatal("redirect did not exile the dying permanent")
	}
	if !g.HasExileCounter(permanent.CardInstanceID, counter.Void) {
		t.Fatal("redirect exiled the dying permanent without a void counter")
	}
}

// TestGraveyardRedirectCounterExilesMilledOpponentCardWithCounter drives a real
// mill through millCards (an opponent milling while the redirect is in play, e.g.
// Dauthi Voidwalker vs. a mill spell). "From anywhere" covers library-to-
// graveyard mills, so the milled card must be exiled with a void counter instead
// of reaching the graveyard, and therefore be selectable by Dauthi's {T},
// Sacrifice ability.
func TestGraveyardRedirectCounterExilesMilledOpponentCardWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	milledCard := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Milled"}})
	if got := millCards(g, game.Player2, 1); len(got) != 0 {
		t.Fatalf("millCards() = %v, want no cards reaching the graveyard", got)
	}
	if g.Players[game.Player2].Graveyard.Contains(milledCard) {
		t.Fatal("redirect did not keep the milled card out of the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(milledCard) {
		t.Fatal("redirect did not exile the milled card")
	}
	if !g.HasExileCounter(milledCard, counter.Void) {
		t.Fatal("redirect exiled the milled card without a void counter")
	}

	// The milled card now bears a void counter, so Dauthi's activated ability
	// must offer it and grant its controller a free-play permission bound to it.
	log := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want exactly one resolution choice", log.Choices)
	}
	if got := len(log.Choices[0].Request.Options); got != 1 {
		t.Fatalf("choice options = %d, want 1 (the milled void-countered card)", got)
	}
	effect, ok := playFromZoneRuleEffect(g, milledCard)
	if !ok {
		t.Fatal("Dauthi's ability granted no play permission for the milled card")
	}
	if !effect.WithoutPayingManaCost {
		t.Fatal("granted permission is not flagged without paying its mana cost")
	}
	if !castFromZoneWithoutPayingManaCost(g, game.Player1, milledCard, zone.Exile, game.FaceFront) {
		t.Fatal("controller should be able to play the milled card without paying its mana cost")
	}
}

// TestGraveyardRedirectCounterLeavesControllerSelfMillUntouched confirms the
// opponent scope for mill: the redirect's own controller milling their own
// library is unaffected — the card reaches the controller's graveyard normally
// and never receives a counter.
func TestGraveyardRedirectCounterLeavesControllerSelfMillUntouched(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	ownCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "My Milled"}})
	if got := millCards(g, game.Player1, 1); len(got) != 1 || got[0] != ownCard {
		t.Fatalf("millCards() = %v, want [%d]", got, ownCard)
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCard) {
		t.Fatal("opponent-scoped redirect wrongly diverted the controller's own milled card")
	}
	if g.Players[game.Player1].Exile.Contains(ownCard) {
		t.Fatal("opponent-scoped redirect wrongly exiled the controller's own milled card")
	}
	if g.HasExileCounter(ownCard, counter.Void) {
		t.Fatal("opponent-scoped redirect placed a counter on the controller's own milled card")
	}
}

// TestMillWithoutRedirectReachesGraveyard verifies mill's default behavior is
// unchanged when no graveyard-redirect replacement is active: the milled card
// reaches its owner's graveyard and is reported as milled.
func TestMillWithoutRedirectReachesGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	milledCard := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Milled"}})
	if got := millCards(g, game.Player2, 1); len(got) != 1 || got[0] != milledCard {
		t.Fatalf("millCards() = %v, want [%d]", got, milledCard)
	}
	if !g.Players[game.Player2].Graveyard.Contains(milledCard) {
		t.Fatal("mill without a redirect did not put the card into the graveyard")
	}
	if g.Players[game.Player2].Exile.Contains(milledCard) {
		t.Fatal("mill without a redirect wrongly exiled the card")
	}
}

// TestGraveyardRedirectCounterExilesDiscardedOpponentCardWithCounter drives a
// real discard through discardCardFromHandInBatch (an opponent discarding while
// the redirect is in play, e.g. Dauthi Voidwalker vs. a Mind Rot). "From
// anywhere" covers hand-to-graveyard discards, so the card must be exiled with a
// void counter and therefore be selectable by Dauthi's {T}, Sacrifice ability.
func TestGraveyardRedirectCounterExilesDiscardedOpponentCardWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	discarded := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	if !discardCardFromHand(g, game.Player2, discarded) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(discarded) {
		t.Fatal("redirect did not keep the discarded card out of the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(discarded) {
		t.Fatal("redirect did not exile the discarded card")
	}
	if !g.HasExileCounter(discarded, counter.Void) {
		t.Fatal("redirect exiled the discarded card without a void counter")
	}

	// The discarded card now bears a void counter, so Dauthi's activated ability
	// must offer it and grant its controller a free-play permission bound to it.
	log := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want exactly one resolution choice", log.Choices)
	}
	if got := len(log.Choices[0].Request.Options); got != 1 {
		t.Fatalf("choice options = %d, want 1 (the discarded void-countered card)", got)
	}
	effect, ok := playFromZoneRuleEffect(g, discarded)
	if !ok {
		t.Fatal("Dauthi's ability granted no play permission for the discarded card")
	}
	if !effect.WithoutPayingManaCost {
		t.Fatal("granted permission is not flagged without paying its mana cost")
	}
	if !castFromZoneWithoutPayingManaCost(g, game.Player1, discarded, zone.Exile, game.FaceFront) {
		t.Fatal("controller should be able to play the discarded card without paying its mana cost")
	}
}

// TestGraveyardRedirectCounterExilesPileSplitLosingPileWithCounter drives a real
// Fact-or-Fiction pile split through the engine while an opponent controls the
// redirect (e.g. Dauthi Voidwalker vs. a Fact or Fiction cast). "From anywhere"
// covers the losing pile's library-to-graveyard move, so those cards must be
// exiled with a void counter instead of reaching the graveyard, and therefore be
// selectable by Dauthi's {T}, Sacrifice ability.
func TestGraveyardRedirectCounterExilesPileSplitLosingPileWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	// Player2 casts Fact or Fiction; its losing pile heads to Player2's graveyard,
	// which Player1's Dauthi (an opponent of Player2) redirects to exile.
	// Add bottom-to-top: c1 deepest, c3 top. peekLibrary is top-first: [c3,c2,c1].
	c1 := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	prim := game.PileSplit{
		Player:            game.ControllerReference(),
		Amount:            game.Fixed(3),
		SeparatorOpponent: true,
		ChooserOpponent:   false,
		Kept:              zone.Hand,
		Other:             zone.Graveyard,
	}
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		// The separating opponent (Player3, next after Player2) puts c3 into the
		// first pile; the controller (Player2) keeps the second pile {c2,c1}, so
		// the first pile {c3} is the losing pile bound for the graveyard.
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	if !engine.pileSplitCards(g, agents, &log, game.Player2, 3, prim) {
		t.Fatal("pileSplitCards() = false, want true")
	}

	player := g.Players[game.Player2]
	if !player.Hand.Contains(c2) || !player.Hand.Contains(c1) {
		t.Fatal("pile split did not put the kept pile into hand")
	}
	if player.Graveyard.Contains(c3) {
		t.Fatal("redirect did not keep the losing pile out of the graveyard")
	}
	if !player.Exile.Contains(c3) {
		t.Fatal("redirect did not exile the losing pile")
	}
	if !g.HasExileCounter(c3, counter.Void) {
		t.Fatal("redirect exiled the losing pile without a void counter")
	}

	// The exiled losing-pile card bears a void counter, so Dauthi's controller
	// (Player1) must be offered it and granted a free-play permission bound to it.
	playLog := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	if len(playLog.Choices) != 1 {
		t.Fatalf("choices = %+v, want exactly one resolution choice", playLog.Choices)
	}
	if got := len(playLog.Choices[0].Request.Options); got != 1 {
		t.Fatalf("choice options = %d, want 1 (the exiled void-countered pile card)", got)
	}
	effect, ok := playFromZoneRuleEffect(g, c3)
	if !ok {
		t.Fatal("Dauthi's ability granted no play permission for the exiled pile card")
	}
	if !effect.WithoutPayingManaCost {
		t.Fatal("granted permission is not flagged without paying its mana cost")
	}
	if !castFromZoneWithoutPayingManaCost(g, game.Player1, c3, zone.Exile, game.FaceFront) {
		t.Fatal("controller should be able to play the exiled pile card without paying its mana cost")
	}
}
