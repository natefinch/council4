package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Castigate (28a972a6-cf67-4ee3-aa91-f7b2549f6c48): "Target opponent reveals
// their hand. You choose a nonland card from it and exile that card." No
// rulings. Soul Search (58f12f0d-e2c1-48f9-9e27-d1de8db743cb): "Target opponent
// reveals their hand. You choose a nonland card from it. Exile that card. If the
// card's mana value is 1 or less, create a 1/1 white and black Spirit creature
// token with flying." Rulings 2024-02-02: {X} is 0 and no mana cost is mana
// value 0. Appetite for Brains (4ffc16e0-1778-4aba-ade3-58146c800f5f): "Target
// opponent reveals their hand. You choose a card from it with mana value 4 or
// greater and exile that card." Ruling 2018-12-07: {X} in a hand card is 0.
// Severance Priest (3a38bf69-511f-4851-870f-8a65641d832e) ruling 2025-04-04:
// "If you chose not to exile a card, the effect of the last ability won't do
// anything." CR 2026-11-06 verified copy: 400.2 and 402.3 (a hand is hidden),
// 406.2-3 (exile is public), 608.2c-d (a mandatory choice takes a legal option
// when one exists and otherwise nothing), 701.20a-b (a revealed card is shown to
// all players for as long as the effect needs it and stays in its zone). These
// test-local programs prove the shared operation; the cards stay unregistered.
var (
	castigateTestReference       = game.ProgramReference{ID: "test/castigate-reveal-hand", Version: "1"}
	soulSearchTestReference      = game.ProgramReference{ID: "test/soul-search-reveal-hand", Version: "1"}
	appetiteTestReference        = game.ProgramReference{ID: "test/appetite-for-brains-reveal-hand", Version: "1"}
	severancePriestTestReference = game.ProgramReference{ID: "test/severance-priest-reveal-hand", Version: "1"}
)

func nonlandSelection() game.Selection {
	return game.Selection{ExcludedTypes: []types.Card{types.Land}}
}

// revealHandExileProgram reveals the target opponent's hand, then exiles the
// controller's chosen matching card, recording the operation's result.
func revealHandExileProgram(reference game.ProgramReference, filter game.Selection, optional bool, results *[]cardprogram.HandRevealResult) cardprogram.Program {
	return cardprogram.Program{Reference: reference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Filter: filter, Max: 1, Optional: optional,
		})
		if err != nil {
			return err
		}
		*results = append(*results, revealed)
		if len(revealed.Chosen) == 0 {
			return nil
		}
		_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile})
		return err
	}}
}

func soulSearchTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: soulSearchTestReference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Filter: nonlandSelection(), Max: 1,
		})
		if err != nil || len(revealed.Chosen) == 0 {
			return err
		}
		exiled, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile})
		if err != nil || len(exiled) != 1 || exiled[0].Zone != zone.Exile || exiled[0].Info.ManaValue > 1 {
			return err
		}
		spirit := &game.CardDef{CardFace: game.CardFace{
			Name: "Spirit", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Spirit},
			Colors: []color.Color{color.White, color.Black},
			Power:  opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Flying)}},
		}}
		_, err = ctx.CreateTokens(spirit, cardprogram.TokenRecipient{Player: ctx.View().Source().Controller, Count: 1})
		return err
	}}
}

func revealHandEngine(t *testing.T, programs ...cardprogram.Program) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func revealHandSpellDef(name string, reference game.ProgramReference) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Constraint: "target opponent", Allow: game.TargetAllowPlayer,
			Selection: opt.Val(game.Selection{Player: game.PlayerOpponent}),
		}}, Program: reference}.Ability()),
	}}
}

func revealHandLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}}
}

func handRevealEvents(g *game.Game, start int) []id.ID {
	var revealed []id.ID
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventCardRevealed && event.FromZone == zone.Hand {
			revealed = append(revealed, event.CardID)
		}
	}
	return revealed
}

// cardViewNames returns views' names in order and fails if any view carries a
// card instance ID, which would track a revealed card (CR 406.3).
func cardViewNames(t *testing.T, views []CardView) []string {
	t.Helper()
	names := make([]string, len(views))
	for index, view := range views {
		if view.CardInstanceID != 0 {
			t.Fatalf("revealed card view %q carries card instance ID %d", view.Name, view.CardInstanceID)
		}
		names[index] = view.Name
	}
	return names
}

// choiceCardNames returns the cards' names in order and fails if any carries a
// CardID.
func choiceCardNames(t *testing.T, cards []game.ChoiceCardInfo) []string {
	t.Helper()
	names := make([]string, len(cards))
	for index, card := range cards {
		if card.CardID != 0 {
			t.Fatalf("revealed choice card %q carries card ID %d", card.Name, card.CardID)
		}
		names[index] = card.Name
	}
	return names
}

func optionCardNames(t *testing.T, options []game.ChoiceOption) []string {
	t.Helper()
	cards := make([]game.ChoiceCardInfo, len(options))
	for index, option := range options {
		cards[index] = option.Card.Val
	}
	return choiceCardNames(t, cards)
}

func cardNames(g *game.Game, cards []id.ID) []string {
	names := make([]string, len(cards))
	for index, cardID := range cards {
		card, _ := g.GetCardInstance(cardID)
		names[index] = card.Def.Name
	}
	return names
}

func handleCardIDs(handles []cardprogram.FoundCard) []id.ID {
	cards := make([]id.ID, len(handles))
	for index, handle := range handles {
		cards[index] = handle.Identity.CardID
	}
	return cards
}

func TestRevealHandCastigateShowsWholeHandAndExilesChosenNonland(t *testing.T) {
	t.Parallel()
	var results []cardprogram.HandRevealResult
	engine := revealHandEngine(t, revealHandExileProgram(castigateTestReference, nonlandSelection(), false, &results))
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCardToHand(g, game.Player2, revealHandLand())
	bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	hand := slices.Clone(g.Players[game.Player2].Hand.All())
	nonland := slices.DeleteFunc(slices.Clone(hand), func(card id.ID) bool { return card == land })
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Castigate", castigateTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})

	// A clone resolves independently with the default choice.
	clone := g.Clone()
	engine.resolveTopOfStackWithChoices(clone, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !clone.Players[game.Player2].Exile.Contains(nonland[0]) || g.Players[game.Player2].Exile.Size() != 0 ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Hand.All()), sortedIDs(hand)) {
		t.Fatal("resolving the clone did not exile its default choice independently of the original")
	}
	results = nil

	start := len(g.Events)
	asked := 0
	chooser := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		asked++
		if req.Player != game.Player1 || req.Kind != game.ChoiceZoneSelection || req.Private ||
			req.MinChoices != 1 || req.MaxChoices != 1 {
			t.Fatalf("the nonland choice was not one public mandatory choice by the caster: %+v", req)
		}
		offered := optionCardNames(t, req.Options)
		if !slices.Equal(offered, cardNames(g, nonland)) {
			t.Fatalf("offered %v, want only the nonland cards %v", offered, cardNames(g, nonland))
		}
		// The whole hand, land included, is shown to every player (CR 701.20a).
		for _, viewer := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
			if !slices.Equal(cardViewNames(t, NewObservation(g, viewer).HandOf(game.Player2)), cardNames(g, hand)) {
				t.Fatalf("player %d did not see the revealed hand", viewer)
			}
		}
		if !slices.Equal(cardViewNames(t, obs.HandOf(game.Player2)), cardNames(g, hand)) {
			t.Fatal("the chooser's decision observation did not show the revealed hand")
		}
		return []int{slices.Index(offered, "Grave Titan")}
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{chooser}, &TurnLog{})

	if asked != 1 {
		t.Fatalf("asked %d choices, want 1", asked)
	}
	if !slices.Equal(handRevealEvents(g, start), hand) {
		t.Fatalf("revealed %v, want every card in hand %v", handRevealEvents(g, start), hand)
	}
	if !g.Players[game.Player2].Exile.Contains(titan) ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Hand.All()), sortedIDs([]id.ID{land, bolt})) {
		t.Fatalf("exile=%v hand=%v, want only the chosen card exiled", g.Players[game.Player2].Exile.All(), g.Players[game.Player2].Hand.All())
	}
	if len(results) != 1 || !results[0].Available || results[0].Player != game.Player2 ||
		!slices.Equal(handleCardIDs(results[0].Revealed), hand) ||
		!slices.Equal(handleCardIDs(results[0].Matching), nonland) ||
		!slices.Equal(handleCardIDs(results[0].Chosen), []id.ID{titan}) || results[0].Chosen[0].Zone != zone.Hand {
		t.Fatalf("unexpected reveal result %+v", results)
	}
	// The hand is hidden again once the effect no longer needs it.
	if len(g.RevealedHands) != 0 || len(NewObservation(g, game.Player3).HandOf(game.Player2)) != 0 {
		t.Fatal("the hand stayed revealed after the effect")
	}
}

func TestRevealHandAppetiteForBrainsForcedAndEmptyChoices(t *testing.T) {
	t.Parallel()
	var results []cardprogram.HandRevealResult
	filter := game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}
	engine := revealHandEngine(t, revealHandExileProgram(appetiteTestReference, filter, false, &results))
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	small := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	big := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	small3 := addCardToHand(g, game.Player3, evidenceCard("Divination", 3))
	def := revealHandSpellDef("Appetite for Brains", appetiteTestReference)
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})

	refuse := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		t.Fatalf("a forced or empty choice was asked: %+v", req)
		return nil
	}}
	// One matching card must be exiled; nothing is asked.
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse}, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(big) || !g.Players[game.Player2].Hand.Contains(small) {
		t.Fatal("the only matching card was not exiled")
	}

	// No matching card: the hand is still revealed, and nothing is exiled.
	start := len(g.Events)
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player3)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse}, &TurnLog{})
	if !slices.Equal(handRevealEvents(g, start), []id.ID{small3}) || g.Players[game.Player3].Exile.Size() != 0 ||
		!g.Players[game.Player3].Hand.Contains(small3) {
		t.Fatal("a hand with no matching card was not just revealed")
	}
	if len(results) != 2 || len(results[1].Matching) != 0 || len(results[1].Chosen) != 0 {
		t.Fatalf("unexpected reveal results %+v", results)
	}
}

func TestRevealHandSeverancePriestMayChooseNothing(t *testing.T) {
	t.Parallel()
	var results []cardprogram.HandRevealResult
	engine := revealHandEngine(t, revealHandExileProgram(severancePriestTestReference, nonlandSelection(), true, &results))
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	only := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Severance Priest", severancePriestTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})

	decline := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.MinChoices != 0 || req.MaxChoices != 1 || len(req.Options) != 1 {
			t.Fatalf("the optional choice was not up to one card: %+v", req)
		}
		return nil
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{decline}, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(only) || g.Players[game.Player2].Exile.Size() != 0 ||
		len(results) != 1 || len(results[0].Chosen) != 0 || len(results[0].Matching) != 1 {
		t.Fatalf("declining the optional choice still exiled a card: %+v", results)
	}
}

func TestRevealHandSoulSearchReadsExiledCardManaValue(t *testing.T) {
	t.Parallel()
	engine := revealHandEngine(t, soulSearchTestProgram())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cheap := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	pricey := addCardToHand(g, game.Player2, evidenceCard("Divination", 3))
	def := revealHandSpellDef("Soul Search", soulSearchTestReference)
	pick := func(name string) *controlScript {
		return &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
			if index := slices.Index(optionCardNames(t, req.Options), name); index >= 0 {
				return []int{index}
			}
			t.Fatalf("%s was not offered", name)
			return nil
		}}
	}
	spirits := func() int {
		count := 0
		for _, permanent := range g.Battlefield {
			if permanent.Token && permanent.Controller == game.Player1 {
				count++
			}
		}
		return count
	}

	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{pick("Divination")}, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(pricey) || spirits() != 0 {
		t.Fatal("exiling a mana value 3 card made a Spirit")
	}
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{pick("Opt")}, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(cheap) || spirits() != 1 {
		t.Fatal("exiling a mana value 1 card did not make a Spirit")
	}
}

// TestRevealHandRecordsWholeHandForEveryPlayer proves the whole revealed hand,
// nonmatching cards included, reaches every player and the public log whether
// the controller's choice is asked, forced or has no legal option (CR 701.20a).
func TestRevealHandRecordsWholeHandForEveryPlayer(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		hand    []*game.CardDef
		asked   bool
		matches int
	}{
		{name: "no matching card", hand: []*game.CardDef{revealHandLand()}},
		{name: "forced choice", hand: []*game.CardDef{revealHandLand(), evidenceCard("Opt", 1)}, matches: 1},
		{name: "asked choice", hand: []*game.CardDef{revealHandLand(), evidenceCard("Opt", 1), evidenceCard("Grave Titan", 6)}, asked: true, matches: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var results []cardprogram.HandRevealResult
			engine := revealHandEngine(t, revealHandExileProgram(castigateTestReference, nonlandSelection(), false, &results))
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToHand(g, game.Player3, evidenceCard("Divination", 3))
			for _, def := range test.hand {
				addCardToHand(g, game.Player2, def)
			}
			hand := slices.Clone(g.Players[game.Player2].Hand.All())
			addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Castigate", castigateTestReference),
				[]game.Target{game.PlayerTarget(game.Player2)})
			var requests []game.ChoiceRequest
			chooser := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				requests = append(requests, req)
				return []int{0}
			}}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{chooser}, &log)

			if len(results) != 1 || len(results[0].Matching) != test.matches || len(results[0].Chosen) != min(test.matches, 1) {
				t.Fatalf("unexpected reveal result %+v", results)
			}
			// The revealing player knows their own cards' identities.
			if announced := cardViewIDs(NewObservation(g, game.Player2).AnnouncedCards()); !slices.Equal(announced, sortedIDs(hand)) {
				t.Fatalf("the revealing player's observation announced %v, want %v", announced, sortedIDs(hand))
			}
			for _, viewer := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
				announced := cardViewNames(t, NewObservation(g, viewer).AnnouncedCards())
				if !slices.Equal(announced, cardNames(g, hand)) || slices.Contains(announced, "Divination") {
					t.Fatalf("player %d's observation announced %v, want the whole revealed hand %v", viewer, announced, cardNames(g, hand))
				}
			}
			if !test.asked {
				if len(requests) != 0 || len(log.Choices) != 0 {
					t.Fatalf("a forced or empty choice was asked: %+v", requests)
				}
				return
			}
			if len(requests) != 1 || len(log.Choices) != 1 {
				t.Fatalf("asked %d and logged %d choices, want 1", len(requests), len(log.Choices))
			}
			for _, request := range []game.ChoiceRequest{requests[0], log.Choices[0].Request} {
				shown := choiceCardNames(t, request.RevealedCards)
				optionCardNames(t, request.Options)
				if !slices.Equal(shown, cardNames(g, hand)) || len(request.Options) != test.matches {
					t.Fatalf("the choice listed revealed %v with %d options, want the whole hand %v with %d options",
						shown, len(request.Options), hand, test.matches)
				}
			}
		})
	}
}

// TestRevealHandDoesNotTrackRevealedCardsIntoHiddenObjects proves a revealed
// hand's public record (announcements, the decision's views and options, and
// the public choice log) shows the cards without identities that would let
// another player correlate a revealed card with a later hidden object: here a
// card the revealing player then foretells face down (CR 406.3, 701.20a).
func TestRevealHandDoesNotTrackRevealedCardsIntoHiddenObjects(t *testing.T) {
	t.Parallel()
	f := newCascadeTriggerFixture(t)
	g := f.game
	doomskar := addCardToHand(g, game.Player2, f.def(t, "Doomskar"))
	behold := addCardToHand(g, game.Player2, f.def(t, "Behold the Multiverse"))
	opt := addCardToHand(g, game.Player2, f.def(t, "Opt"))
	tracked := []id.ID{doomskar, behold, opt}
	want := cardNames(g, g.Players[game.Player2].Hand.All())
	var results []cardprogram.HandRevealResult
	engine := revealHandEngine(t, revealHandExileProgram(castigateTestReference, nonlandSelection(), false, &results))
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Castigate", castigateTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})

	var requests []game.ChoiceRequest
	chooser := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		requests = append(requests, req)
		for _, viewer := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
			if shown := cardViewNames(t, NewObservation(g, viewer).HandOf(game.Player2)); len(shown) != 3 {
				t.Fatalf("player %d saw revealed hand %v", viewer, shown)
			}
		}
		cardViewNames(t, obs.HandOf(game.Player2))
		return []int{slices.Index(optionCardNames(t, req.Options), "Opt")}
	}}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{chooser}, &log)
	if len(requests) != 1 || !g.Players[game.Player2].Exile.Contains(opt) {
		t.Fatal("Castigate did not exile the chosen Opt")
	}

	addBasicLandPermanent(g, game.Player2, types.Island)
	addBasicLandPermanent(g, game.Player2, types.Island)
	setSorceryTiming(g, game.Player2)
	if !f.engine.applyAction(g, game.Player2, action.ForetellCard(doomskar)) || !g.Players[game.Player2].Exile.IsFaceDown(doomskar) {
		t.Fatal("Doomskar was not foretold face down")
	}

	for _, viewer := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		obs := NewObservation(g, viewer)
		if announced := cardViewNames(t, obs.AnnouncedCards()); !slices.Equal(announced, want) {
			t.Fatalf("player %d's public record announced %v, want %v", viewer, announced, want)
		}
		for _, card := range obs.Exile(game.Player2) {
			if card.CardInstanceID == doomskar && card.Name != "" {
				t.Fatalf("player %d sees the foretold card's identity", viewer)
			}
		}
	}
	for _, request := range []game.ChoiceRequest{requests[0], log.Choices[0].Request} {
		if !slices.Equal(choiceCardNames(t, request.RevealedCards), want) {
			t.Fatalf("the choice listed %v, want %v", choiceCardNames(t, request.RevealedCards), want)
		}
		optionCardNames(t, request.Options)
	}
	// The revealing player still knows which of their own cards were revealed.
	if announced := cardViewIDs(NewObservation(g, game.Player2).AnnouncedCards()); !slices.Equal(announced, sortedIDs(tracked)) {
		t.Fatalf("the revealing player's record announced %v, want %v", announced, sortedIDs(tracked))
	}
}

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Agonizing Remorse (1267dfda-eb1a-4963-9fe3-fa619d924d7a): "Target opponent
// reveals their hand. You choose a nonland card from it or a card from their
// graveyard. Exile that card. You lose 1 life." Rulings 2020-01-24: a
// graveyard card may be a land card; you must choose a card if able, from the
// graveyard when the hand has no nonland card and vice versa; no player may act
// between the choice and the exile; with neither, you just lose 1 life. Thought
// Distortion (5f089ac6-9e92-4ec2-bf46-a0b08d1e2979): "This spell can't be
// countered. Target opponent reveals their hand. Exile all noncreature, nonland
// cards from that player's hand and graveyard." Rulings 2019-07-12 cover only
// its uncounterability and an illegal target. CR 2026-11-06 verified copy:
// 400.2 (graveyard public, hand hidden even when revealed), 608.2c-d, 701.20a-b.
// These test-local programs prove the shared operation; the cards stay
// unregistered.
var (
	agonizingRemorseTestReference  = game.ProgramReference{ID: "test/agonizing-remorse-reveal-hand", Version: "1"}
	thoughtDistortionTestReference = game.ProgramReference{ID: "test/thought-distortion-reveal-hand", Version: "1"}
)

func agonizingRemorseTestProgram(results *[]cardprogram.HandRevealResult) cardprogram.Program {
	return cardprogram.Program{Reference: agonizingRemorseTestReference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Filter: nonlandSelection(), Graveyard: opt.Val(game.Selection{}), Max: 1,
		})
		if err != nil {
			return err
		}
		*results = append(*results, revealed)
		if len(revealed.Chosen) != 0 {
			if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile}); err != nil {
				return err
			}
		}
		_, err = ctx.Execute(game.LoseLife{Amount: game.Fixed(1), Player: game.ControllerReference()})
		return err
	}}
}

func thoughtDistortionTestProgram(results *[]cardprogram.HandRevealResult) cardprogram.Program {
	return cardprogram.Program{Reference: thoughtDistortionTestReference, Resolve: func(ctx cardprogram.Context) error {
		filter := game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}}
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Filter: filter, Graveyard: opt.Val(filter), All: true,
		})
		if err != nil {
			return err
		}
		*results = append(*results, revealed)
		_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile})
		return err
	}}
}

func TestRevealHandAgonizingRemorseChoosesFromHandOrGraveyard(t *testing.T) {
	t.Parallel()
	var results []cardprogram.HandRevealResult
	engine := revealHandEngine(t, agonizingRemorseTestProgram(&results))
	def := revealHandSpellDef("Agonizing Remorse", agonizingRemorseTestReference)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCardToHand(g, game.Player2, revealHandLand())
	opt1 := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	hand := slices.Clone(g.Players[game.Player2].Hand.All())
	graveLand := addCardToGraveyard(g, game.Player2, revealHandLand())
	titan := addCardToGraveyard(g, game.Player2, evidenceCard("Grave Titan", 6))
	graveyard := slices.Clone(g.Players[game.Player2].Graveyard.All())
	life := g.Players[game.Player1].Life
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})

	asked := 0
	chooser := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		asked++
		if req.Player != game.Player1 || req.Kind != game.ChoiceZoneSelection || req.Private ||
			req.MinChoices != 1 || req.MaxChoices != 1 || len(req.Options) != 3 {
			t.Fatalf("the choice was not one public mandatory choice among three cards: %+v", req)
		}
		// The hand's nonland card is offered without its identity; graveyard
		// cards, lands included, are public objects (CR 400.2).
		if card := req.Options[0].Card.Val; card.Name != "Opt" || card.CardID != 0 || req.Options[0].Label != "Opt (Hand)" {
			t.Fatalf("the hand option was %+v", req.Options[0])
		}
		for index, want := range graveyard {
			option := req.Options[index+1]
			if option.Card.Val.CardID != want || option.Label != cardNames(g, []id.ID{want})[0]+" (Graveyard)" {
				t.Fatalf("graveyard option %d was %+v, want card %d", index, option, want)
			}
		}
		if !slices.Equal(choiceCardNames(t, req.RevealedCards), cardNames(g, hand)) {
			t.Fatalf("the choice listed revealed %v, want the whole hand", choiceCardNames(t, req.RevealedCards))
		}
		return []int{1 + slices.Index(graveyard, graveLand)}
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{chooser}, &TurnLog{})
	if asked != 1 || !g.Players[game.Player2].Exile.Contains(graveLand) || g.Players[game.Player2].Exile.Size() != 1 ||
		!g.Players[game.Player2].Hand.Contains(land) || !g.Players[game.Player2].Hand.Contains(opt1) ||
		!g.Players[game.Player2].Graveyard.Contains(titan) || g.Players[game.Player1].Life != life-1 {
		t.Fatal("Agonizing Remorse did not exile only the chosen graveyard land and lose 1 life")
	}
	if len(results) != 1 || !slices.Equal(handleCardIDs(results[0].Matching), []id.ID{opt1}) ||
		!slices.Equal(handleCardIDs(results[0].GraveyardMatching), graveyard) ||
		len(results[0].Chosen) != 1 || results[0].Chosen[0].Zone != zone.Graveyard {
		t.Fatalf("unexpected reveal result %+v", results)
	}

	// With no nonland card in hand, the only graveyard card must be exiled
	// unasked; with neither, the caster just loses 1 life.
	refuse := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		t.Fatalf("a forced or empty choice was asked: %+v", req)
		return nil
	}}
	onlyLand := addCardToHand(g, game.Player3, revealHandLand())
	onlyGrave := addCardToGraveyard(g, game.Player3, evidenceCard("Divination", 3))
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player3)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse}, &TurnLog{})
	if !g.Players[game.Player3].Exile.Contains(onlyGrave) || !g.Players[game.Player3].Hand.Contains(onlyLand) ||
		g.Players[game.Player1].Life != life-2 {
		t.Fatal("the only graveyard card was not exiled")
	}
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player4)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse}, &TurnLog{})
	if g.Players[game.Player4].Exile.Size() != 0 || g.Players[game.Player1].Life != life-3 || len(results) != 3 ||
		len(results[2].Chosen) != 0 {
		t.Fatal("with no card to choose, Agonizing Remorse did more than lose 1 life")
	}
}

func TestRevealHandThoughtDistortionExilesAllMatchingFromHandAndGraveyard(t *testing.T) {
	t.Parallel()
	var results []cardprogram.HandRevealResult
	engine := revealHandEngine(t, thoughtDistortionTestProgram(&results))
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	handLand := addCardToHand(g, game.Player2, revealHandLand())
	handBear := addCardToHand(g, game.Player2, greenCreature())
	handOpt := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	handDivination := addCardToHand(g, game.Player2, evidenceCard("Divination", 3))
	hand := slices.Clone(g.Players[game.Player2].Hand.All())
	graveBear := addCardToGraveyard(g, game.Player2, greenCreature())
	graveLand := addCardToGraveyard(g, game.Player2, revealHandLand())
	graveTitan := addCardToGraveyard(g, game.Player2, evidenceCard("Grave Titan", 6))
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Thought Distortion", thoughtDistortionTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})

	start := len(g.Events)
	refuse := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		t.Fatalf("exiling all matching cards asked a choice: %+v", req)
		return nil
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse}, &TurnLog{})

	if !slices.Equal(handRevealEvents(g, start), hand) {
		t.Fatalf("revealed %v, want the whole hand %v", handRevealEvents(g, start), hand)
	}
	// Chosen lists the hand's matches in hand order, then the graveyard's.
	exiled := slices.DeleteFunc(slices.Clone(hand), func(card id.ID) bool { return card == handLand || card == handBear })
	exiled = append(exiled, graveTitan)
	if !slices.Equal(sortedIDs(g.Players[game.Player2].Exile.All()), sortedIDs([]id.ID{handOpt, handDivination, graveTitan})) ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Hand.All()), sortedIDs([]id.ID{handLand, handBear})) ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Graveyard.All()), sortedIDs([]id.ID{graveBear, graveLand})) {
		t.Fatalf("exile=%v hand=%v graveyard=%v, want every noncreature, nonland card from both exiled",
			g.Players[game.Player2].Exile.All(), g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Graveyard.All())
	}
	// Both zones' cards are exiled in one simultaneous event.
	batches := make(map[game.ObjectID]bool)
	from := make(map[zone.Type]int)
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventZoneChanged && event.ToZone == zone.Exile {
			batches[event.SimultaneousID] = true
			from[event.FromZone]++
		}
	}
	if len(batches) != 1 || batches[0] || from[zone.Hand] != 2 || from[zone.Graveyard] != 1 {
		t.Fatalf("exile batches %v from %v, want one simultaneous exile of two hand cards and one graveyard card", batches, from)
	}
	if len(results) != 1 || !slices.Equal(handleCardIDs(results[0].Chosen), exiled) {
		t.Fatalf("unexpected reveal result %+v", results)
	}
}

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Struggle for Sanity (5c0f63ad-4bb1-479e-9760-5421265da40b), {2}{B}{B}
// Sorcery: "Target opponent reveals their hand. That player exiles a card from
// it, then you exile a card from it. Repeat this process until all cards in
// that hand have been exiled. That player returns the cards they exiled this
// way to their hand and puts the rest into their graveyard." No rulings. CR
// 2026-11-06 verified copy: 701.20a (a revealed card stays revealed as long as
// necessary to complete the parts of the effect it's relevant to), 402.3 and
// 406.3 (the hand is hidden again afterwards; returned cards are new objects,
// CR 400.7). This test-local program proves the shared operation; the card
// stays unregistered.
var struggleForSanityTestReference = game.ProgramReference{ID: "test/struggle-for-sanity-reveal-hand", Version: "1"}

func struggleForSanityTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: struggleForSanityTestReference, Resolve: func(ctx cardprogram.Context) error {
		target := game.TargetPlayerReference(0)
		var theirs, mine []cardprogram.FoundCard
		for opponentChooses := true; ; opponentChooses = !opponentChooses {
			reveal := cardprogram.HandReveal{Player: target, Max: 1, KeepRevealed: true}
			if opponentChooses {
				reveal.Chooser = opt.Val(target)
			}
			revealed, err := ctx.RevealHand(reveal)
			if err != nil {
				return err
			}
			if len(revealed.Chosen) == 0 {
				break
			}
			exiled, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile})
			if err != nil {
				return err
			}
			if opponentChooses {
				theirs = append(theirs, exiled...)
			} else {
				mine = append(mine, exiled...)
			}
		}
		var placements []cardprogram.CardPlacement
		if len(theirs) > 0 {
			placements = append(placements, cardprogram.CardPlacement{Cards: theirs, Destination: zone.Hand})
		}
		if len(mine) > 0 {
			placements = append(placements, cardprogram.CardPlacement{Cards: mine, Destination: zone.Graveyard})
		}
		if len(placements) == 0 {
			return nil
		}
		_, err := ctx.MoveCards(placements...)
		return err
	}}
}

func TestRevealHandStruggleForSanityAlternatesChoicesFromOneRevealedHand(t *testing.T) {
	t.Parallel()
	engine := revealHandEngine(t, struggleForSanityTestProgram())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addCardToHand(g, game.Player2, revealHandLand())
	bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	opt1 := addCardToHand(g, game.Player2, evidenceCard("Opt", 1))
	divination := addCardToHand(g, game.Player2, evidenceCard("Divination", 3))
	hand := slices.Clone(g.Players[game.Player2].Hand.All())
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Struggle for Sanity", struggleForSanityTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})

	start := len(g.Events)
	var choosers []game.PlayerID
	ordered := false
	picks := []string{"Grave Titan", "Lightning Bolt", "Divination", "Opt"}
	choose := func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Player == game.Player2 && len(choosers) == len(picks) {
			// Ordering the caster's cards into the graveyard happens as the
			// opponent's cards return to the emptied hand, which is no longer
			// revealed (CR 402.3).
			if g.HandRevealedToAll(game.Player2) || len(NewObservation(g, game.Player3).HandOf(game.Player2)) != 0 {
				t.Fatal("cards returned to the emptied hand were revealed")
			}
			ordered = true
			return req.DefaultSelection
		}
		if req.Kind != game.ChoiceZoneSelection || req.Private || req.MinChoices != 1 || req.MaxChoices != 1 {
			t.Fatalf("the exile choice was not one public mandatory choice: %+v", req)
		}
		current := cardNames(g, g.Players[game.Player2].Hand.All())
		// The hand stays revealed to every player through every choice (CR 701.20a).
		if !g.HandRevealedToAll(game.Player2) ||
			!slices.Equal(cardViewNames(t, NewObservation(g, game.Player3).HandOf(game.Player2)), current) ||
			req.Player == game.Player1 && !slices.Equal(cardViewNames(t, obs.HandOf(game.Player2)), current) ||
			!slices.Equal(choiceCardNames(t, req.RevealedCards), current) ||
			!slices.Equal(optionCardNames(t, req.Options), current) {
			t.Fatalf("choice %d did not show the remaining revealed hand %v", len(choosers), current)
		}
		pick := picks[len(choosers)]
		choosers = append(choosers, req.Player)
		return []int{slices.Index(current, pick)}
	}
	agents := [game.NumPlayers]PlayerAgent{&controlScript{choice: choose}, &controlScript{choice: choose}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// The opponent chooses first, then the caster, alternating; the last card is
	// the opponent's forced choice.
	if !ordered || !slices.Equal(choosers, []game.PlayerID{game.Player2, game.Player1, game.Player2, game.Player1}) {
		t.Fatalf("choosers %v, want the opponent and caster alternating", choosers)
	}
	// One reveal: each card is announced once, not again on every repetition.
	if !slices.Equal(handRevealEvents(g, start), hand) {
		t.Fatalf("revealed %v, want each card of the hand %v once", handRevealEvents(g, start), hand)
	}
	if g.Players[game.Player2].Exile.Size() != 0 ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Hand.All()), sortedIDs([]id.ID{titan, divination, forest})) ||
		!slices.Equal(sortedIDs(g.Players[game.Player2].Graveyard.All()), sortedIDs([]id.ID{bolt, opt1})) {
		t.Fatalf("hand=%v graveyard=%v exile=%v, want the opponent's exiled cards returned and the caster's in the graveyard",
			g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Graveyard.All(), g.Players[game.Player2].Exile.All())
	}
	// The returns to hand and graveyard are one simultaneous event.
	batches := make(map[game.ObjectID]int)
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile {
			batches[event.SimultaneousID]++
		}
	}
	if len(batches) != 1 {
		t.Fatalf("exile departures in batches %v, want one simultaneous placement", batches)
	}
	for batch, count := range batches {
		if batch == 0 || count != len(hand) {
			t.Fatalf("exile departures in batches %v, want all %d cards together", batches, len(hand))
		}
	}
	// The hand is hidden again once the effect finishes (CR 402.3).
	if len(g.RevealedHands) != 0 || len(NewObservation(g, game.Player3).HandOf(game.Player2)) != 0 {
		t.Fatal("the hand stayed revealed after the effect")
	}
}

func TestRevealHandStruggleForSanityWithEmptyHandDoesNothing(t *testing.T) {
	t.Parallel()
	engine := revealHandEngine(t, struggleForSanityTestProgram())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Struggle for Sanity", struggleForSanityTestReference),
		[]game.Target{game.PlayerTarget(game.Player2)})
	refuse := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		t.Fatalf("an empty hand asked a choice: %+v", req)
		return nil
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{refuse, refuse}, &TurnLog{})
	if len(g.RevealedHands) != 0 || g.Players[game.Player2].Graveyard.Size() != 0 {
		t.Fatal("an empty revealed hand changed the game")
	}
}
