package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Eradicate (782ff740-ff43-4c07-af0b-433cb9770661): "Exile target nonblack
// creature. Search its controller's graveyard, hand, and library for all cards
// with the same name as that creature and exile them. Then that player
// shuffles." Rulings: use the name as it leaves the battlefield (2004-10-04); a
// face-down creature has no name, so nothing shares it (2012-05-01); copies
// must be found in publicly viewable zones, finding them in private zones is
// optional (2005-02-01); battlefield copies are untouched (2004-10-04).
// Test of Talents (d808262d-43ca-46f8-ab8b-b584773f9c58): "Counter target
// instant or sorcery spell. Search its controller's graveyard, hand, and
// library for any number of cards with the same name as that spell and exile
// them. That player shuffles, then draws a card for each card exiled from their
// hand this way." Rulings 2021-04-16: the countered card is in the graveyard to
// find; a countered modal double-faced back face matches nothing, because those
// cards have their front face's name. Surgical Extraction ruling 2011-06-01:
// "any number" may leave every card, the target included. Kassandra, Eagle
// Bearer (745b9a2c-bcd7-4db9-8cd0-d38be2ef6270) has no rulings; "a card named"
// must be found in the public graveyard if one is there (CR 701.23b).
// Opposition Agent ruling 2020-11-10: in a search of several zones including
// your library, its controller makes every decision and exiles every find.
// CR 2026-11-06 verified copy: 201.2a (same name), 401.2 and 402.3 (library and
// hand are hidden), 608.2h (last known information), 701.23a-b and e (searching
// hidden zones may fail to find; finds are not revealed), 708.2 (face-down has
// no name), 709.4a (a split card has both names), 712.6a (a double-faced card
// has only its front face's name outside the stack and battlefield). These
// test-local programs prove the shared operation; the cards stay unregistered.
var (
	eradicateTestReference      = game.ProgramReference{ID: "test/eradicate-same-name", Version: "1"}
	testOfTalentsTestReference  = game.ProgramReference{ID: "test/test-of-talents-same-name", Version: "1"}
	sameNameSearchTestPrograms  = []cardprogram.Program{eradicateTestProgram(), testOfTalentsTestProgram()}
	sameNameSearchFillerDefName = "Llanowar Elves"
)

func eradicateTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: eradicateTestReference, Resolve: func(ctx cardprogram.Context) error {
		exiled, err := ctx.Execute(game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile})
		if err != nil || len(exiled.Objects) != 1 {
			return err
		}
		creature := exiled.Objects[0]
		player := game.CapturedPlayerReference(creature.Controller)
		found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
			Player: player, Name: cardprogram.SameName{Snapshot: opt.Val(creature)}, Quantity: cardprogram.SameNameAll,
		})
		if err != nil || !found.Available {
			return err
		}
		if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Exile}); err != nil {
			return err
		}
		_, err = ctx.Execute(game.ShuffleLibrary{Player: player})
		return err
	}}
}

func testOfTalentsTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: testOfTalentsTestReference, Resolve: func(ctx cardprogram.Context) error {
		spell := game.TargetStackObjectReference(0)
		controller, available, err := ctx.View().Player(game.ObjectControllerReference(spell))
		if err != nil || !available {
			return err
		}
		if _, err := ctx.Execute(game.CounterObject{Object: spell}); err != nil {
			return err
		}
		player := game.CapturedPlayerReference(controller)
		found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
			Player: player, Name: cardprogram.SameName{Spell: opt.Val(spell)}, Quantity: cardprogram.SameNameAnyNumber,
		})
		if err != nil || !found.Available {
			return err
		}
		moved, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Exile})
		if err != nil {
			return err
		}
		fromHand := 0
		for _, card := range moved {
			if card.Zone == zone.Exile && slices.ContainsFunc(found.Cards, func(handle cardprogram.FoundCard) bool {
				return handle.Identity.CardID == card.Identity.CardID && handle.Zone == zone.Hand
			}) {
				fromHand++
			}
		}
		if _, err := ctx.Execute(game.ShuffleLibrary{Player: player}); err != nil {
			return err
		}
		if fromHand > 0 {
			_, err = ctx.Execute(game.Draw{Player: player, Amount: game.Fixed(fromHand)})
		}
		return err
	}}
}

func sameNameSearchEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(sameNameSearchTestPrograms)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func sameNameSpellDef(name string, reference game.ProgramReference, target game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{target}, Program: reference}.Ability()),
	}}
}

func eradicateTestDef() *game.CardDef {
	return sameNameSpellDef("Eradicate", eradicateTestReference, game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	})
}

func testOfTalentsTestDef() *game.CardDef {
	return sameNameSpellDef("Test of Talents", testOfTalentsTestReference, game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
	})
}

func sameNameInstant(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Instant}}}
}

// sameNameSearchBatch returns the simultaneous ID shared by every zone change of
// cards, failing unless each moved exactly once into exile in one batch.
func assertSameNameExileBatch(t *testing.T, g *game.Game, start int, cards []id.ID) {
	t.Helper()
	var batch id.ID
	moved := 0
	for _, event := range g.Events[start:] {
		if event.Kind != game.EventZoneChanged || event.ToZone != zone.Exile || !slices.Contains(cards, event.CardID) {
			continue
		}
		moved++
		if event.SimultaneousID == 0 || batch != 0 && event.SimultaneousID != batch {
			t.Fatalf("same-name finds were not exiled as one simultaneous batch: %+v", event)
		}
		batch = event.SimultaneousID
	}
	if moved != len(cards) {
		t.Fatalf("moved %d found cards, want %d", moved, len(cards))
	}
}

// sameNameOptionCards returns the cards a same-name search choice offered, in
// graveyard, hand, library zone order.
func sameNameOptionCards(t *testing.T, req game.ChoiceRequest) []id.ID {
	t.Helper()
	cards := make([]id.ID, len(req.Options))
	for index, option := range req.Options {
		if !option.Card.Exists {
			t.Fatalf("option %d does not show its card to the searcher", index)
		}
		cards[index] = option.Card.Val.CardID
	}
	return cards
}

// sameNameSelect selects the options offering cards.
func sameNameSelect(t *testing.T, req game.ChoiceRequest, cards ...id.ID) []int {
	t.Helper()
	var selected []int
	for index, card := range sameNameOptionCards(t, req) {
		if slices.Contains(cards, card) {
			selected = append(selected, index)
		}
	}
	return selected
}

func cardViewIDs(views []CardView) []id.ID {
	ids := make([]id.ID, len(views))
	for index, view := range views {
		ids[index] = view.CardInstanceID
	}
	return sortedIDs(ids)
}

// assertSearcherSeesHiddenZones checks, during a same-name search of player,
// that the deciding observation shows every card in player's hand and library,
// matching or not (CR 701.23a).
func assertSearcherSeesHiddenZones(t *testing.T, g *game.Game, obs PlayerObservation, player game.PlayerID) {
	t.Helper()
	if !slices.Equal(cardViewIDs(obs.HandOf(player)), sortedIDs(g.Players[player].Hand.All())) ||
		!slices.Equal(cardViewIDs(obs.LibraryLookable(player)), sortedIDs(g.Players[player].Library.All())) {
		t.Fatalf("the searcher did not see all of player %d's hand and library", player)
	}
}

func librarySearchedEvents(g *game.Game, start int) int {
	count := 0
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventLibrarySearched {
			count++
		}
	}
	return count
}

func TestSearchSameNameEradicateFindsPublicCopiesAndOptionalHiddenCopies(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := sameNameSearchEngine(t)
	bears := vanillaCreature("Grizzly Bears", 2, 2)
	elves := vanillaCreature(sameNameSearchFillerDefName, 1, 1)
	target := addCombatPermanent(g, game.Player2, bears)
	graveyard := []id.ID{addCardToGraveyard(g, game.Player2, bears), addCardToGraveyard(g, game.Player2, bears)}
	graveElves := addCardToGraveyard(g, game.Player2, elves)
	hand := addCardToHand(g, game.Player2, bears)
	handElves := addCardToHand(g, game.Player2, elves)
	keptInLibrary := addCardToLibrary(g, game.Player2, bears)
	addCardToLibrary(g, game.Player2, elves)
	foundInLibrary := addCardToLibrary(g, game.Player2, bears)
	battlefieldCopy := addCombatPermanent(g, game.Player3, bears)
	addImplementationSpellToStack(g, game.Player1, eradicateTestDef(), []game.Target{game.PermanentTarget(target.ObjectID)})
	epoch := g.Players[game.Player2].Library.ReorderVersion()

	// The clone declines every optional find: only the graveyard copies, which
	// must be found, follow the creature into exile.
	clone := g.Clone()
	declined := &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int { return nil }}
	cloneLog := &TurnLog{}
	engine.resolveTopOfStackWithChoices(clone, [game.NumPlayers]PlayerAgent{declined}, cloneLog)
	if clone.Players[game.Player2].Exile.Size() != 3 || !clone.Players[game.Player2].Hand.Contains(hand) ||
		clone.Players[game.Player2].Library.Size() != 3 || clone.Players[game.Player2].Library.ReorderVersion() != epoch+1 {
		t.Fatalf("declining hidden finds did not still exile the graveyard copies: exile=%v", clone.Players[game.Player2].Exile.All())
	}
	if g.Stack.Size() != 1 || g.Players[game.Player2].Exile.Size() != 0 || g.Players[game.Player2].Library.ReorderVersion() != epoch {
		t.Fatal("resolving the clone changed the original game")
	}

	start := len(g.Events)
	searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind != game.ChoiceSearch || req.Player != game.Player1 || !req.Private ||
			req.MinChoices != 0 || req.MaxChoices != 3 || len(req.Options) != 3 {
			t.Fatalf("hidden same-name finds were not an optional private search choice: %+v", req)
		}
		offered := sameNameOptionCards(t, req)
		if offered[0] != hand || !slices.Equal(sortedIDs(offered[1:]), sortedIDs([]id.ID{keptInLibrary, foundInLibrary})) {
			t.Fatalf("offered %v, want the hand copy, then the library copies", offered)
		}
		assertSearcherSeesHiddenZones(t, g, obs, game.Player2)
		return sameNameSelect(t, req, hand, foundInLibrary)
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher}, log)
	exiled := append(slices.Clone(graveyard), hand, foundInLibrary)
	for _, card := range append(slices.Clone(exiled), target.CardInstanceID) {
		if !g.Players[game.Player2].Exile.Contains(card) {
			t.Fatalf("card %d was not exiled", card)
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(graveElves) || !g.Players[game.Player2].Hand.Contains(handElves) ||
		!g.Players[game.Player2].Library.Contains(keptInLibrary) || g.Players[game.Player2].Library.Size() != 2 ||
		permanentByCardID(g, battlefieldCopy.CardInstanceID) == nil {
		t.Fatal("same-name search exiled a differently named card, a declined copy, or a battlefield copy")
	}
	assertSameNameExileBatch(t, g, start, exiled)
	if g.Players[game.Player2].Library.ReorderVersion() != epoch+1 {
		t.Fatal("the searched player did not shuffle")
	}
	if librarySearchedEvents(g, start) != 0 {
		t.Fatal("searching another player's library counted as that player searching their own library")
	}
	for _, decision := range log.Choices {
		if decision.Request.Kind == game.ChoiceSearch {
			t.Fatal("the hidden-zone search choice was published in the log")
		}
	}
}

func TestSearchSameNameFaceDownTargetFindsNothingButShuffles(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := sameNameSearchEngine(t)
	bears := vanillaCreature("Grizzly Bears", 2, 2)
	target := addCombatPermanent(g, game.Player2, bears)
	target.FaceDown = true
	graveyard := addCardToGraveyard(g, game.Player2, bears)
	addCardToHand(g, game.Player2, bears)
	addCardToLibrary(g, game.Player2, bears)
	addImplementationSpellToStack(g, game.Player1, eradicateTestDef(), []game.Target{game.PermanentTarget(target.ObjectID)})
	epoch := g.Players[game.Player2].Library.ReorderVersion()
	asked := 0
	searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		asked++
		if req.Kind != game.ChoiceSearch || len(req.Options) != 0 || req.MaxChoices != 0 {
			t.Fatalf("a nameless creature offered same-name finds: %+v", req)
		}
		assertSearcherSeesHiddenZones(t, g, obs, game.Player2)
		return nil
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher}, &TurnLog{})
	if asked != 1 {
		t.Fatalf("the searcher looked through the zones %d times, want 1", asked)
	}
	if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) || g.Players[game.Player2].Exile.Size() != 1 ||
		!g.Players[game.Player2].Graveyard.Contains(graveyard) || g.Players[game.Player2].Library.ReorderVersion() != epoch+1 {
		t.Fatal("a face-down creature's search found a named card or skipped the shuffle")
	}
}

func TestSearchSameNameTestOfTalentsFindsCounteredCardAndDrawsForHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := sameNameSearchEngine(t)
	bolt := sameNameInstant("Lightning Bolt")
	for range 3 {
		addCardToLibrary(g, game.Player2, vanillaCreature(sameNameSearchFillerDefName, 1, 1))
	}
	libraryBolt := addCardToLibrary(g, game.Player2, bolt)
	handBolts := []id.ID{addCardToHand(g, game.Player2, bolt), addCardToHand(g, game.Player2, bolt)}
	countered := addImplementationSpellToStack(g, game.Player2, bolt, nil)
	spell, _ := g.Stack.Peek()
	addImplementationSpellToStack(g, game.Player1, testOfTalentsTestDef(), []game.Target{game.StackObjectTarget(spell.ID)})
	epoch := g.Players[game.Player2].Library.ReorderVersion()
	start := len(g.Events)
	searcher := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind != game.ChoiceSearch || req.MinChoices != 0 || req.MaxChoices != 4 {
			t.Fatalf("any-number search required finds: %+v", req)
		}
		offered := sameNameOptionCards(t, req)
		if offered[0] != countered || !slices.Equal(sortedIDs(offered[1:3]), sortedIDs(handBolts)) || offered[3] != libraryBolt {
			t.Fatalf("offered %v, want the countered card, then hand and library copies", offered)
		}
		return sameNameSelect(t, req, countered, handBolts[0], handBolts[1])
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher}, &TurnLog{})
	exiled := []id.ID{countered, handBolts[0], handBolts[1]}
	for _, card := range exiled {
		if !g.Players[game.Player2].Exile.Contains(card) {
			t.Fatalf("chosen card %d was not exiled", card)
		}
	}
	assertSameNameExileBatch(t, g, start, exiled)
	if g.Players[game.Player2].Exile.Contains(libraryBolt) || g.Players[game.Player2].Hand.Size() != 2 ||
		g.Players[game.Player2].Library.Size() != 2 || g.Players[game.Player2].Library.ReorderVersion() != epoch+1 {
		t.Fatalf("expected a shuffle and two draws for two hand exiles: hand=%d library=%d", g.Players[game.Player2].Hand.Size(), g.Players[game.Player2].Library.Size())
	}
}

func TestSearchSameNameCounteredModalBackFaceMatchesNothing(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := sameNameSearchEngine(t)
	mdfc := sameNameInstant("Front Spell")
	mdfc.Layout, mdfc.Back = game.LayoutModalDFC, opt.Val(sameNameInstant("Back Spell").CardFace)
	handCopy := addCardToHand(g, game.Player2, mdfc)
	countered := addImplementationSpellToStack(g, game.Player2, mdfc, nil)
	spell, _ := g.Stack.Peek()
	spell.Face = game.FaceBack
	addImplementationSpellToStack(g, game.Player1, testOfTalentsTestDef(), []game.Target{game.StackObjectTarget(spell.ID)})
	searcher := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if len(req.Options) != 0 {
			t.Fatalf("cards with the front face's name matched a countered back face: %+v", req)
		}
		return nil
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher}, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(countered) || !g.Players[game.Player2].Hand.Contains(handCopy) ||
		g.Players[game.Player2].Exile.Size() != 0 {
		t.Fatal("a countered modal double-faced back face exiled a card")
	}
}

func sameNameTestContext(t *testing.T, g *game.Game, agents [game.NumPlayers]PlayerAgent) *programContext {
	t.Helper()
	return &programContext{
		engine: sameNameSearchEngine(t), game: g, agents: agents,
		obj: &game.StackObject{Controller: game.Player1}, active: true, log: &TurnLog{},
	}
}

func TestSearchSameNameMatchesEveryNameAZoneCardHas(t *testing.T) {
	t.Parallel()
	split := &game.CardDef{
		CardFace: sameNameInstant("Fire").CardFace, Layout: game.LayoutSplit,
		Alternate: opt.Val(sameNameInstant("Ice").CardFace),
	}
	dfc := vanillaCreature("Delver of Secrets", 1, 1)
	dfc.Layout, dfc.Back = game.LayoutTransform, opt.Val(vanillaCreature("Insectile Aberration", 3, 2).CardFace)
	for _, test := range []struct {
		name string
		want func(splitCard, fire, ice, delver id.ID) []id.ID
	}{
		{"Ice", func(splitCard, _, ice, _ id.ID) []id.ID { return []id.ID{ice, splitCard} }},
		{"Fire // Ice", func(splitCard, fire, ice, _ id.ID) []id.ID { return []id.ID{fire, ice, splitCard} }},
		{"Insectile Aberration", func(id.ID, id.ID, id.ID, id.ID) []id.ID { return nil }},
		{"Delver of Secrets", func(_, _, _, delver id.ID) []id.ID { return []id.ID{delver} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			fire := addCardToGraveyard(g, game.Player1, sameNameInstant("Fire"))
			ice := addCardToGraveyard(g, game.Player1, sameNameInstant("Ice"))
			splitCard := addCardToHand(g, game.Player1, split)
			delver := addCardToLibrary(g, game.Player1, dfc)
			addCardToLibrary(g, game.Player1, sameNameInstant("Fire and Ice"))
			ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{&controlScript{}})
			start := len(g.Events)
			found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
				Player: game.ControllerReference(), Name: cardprogram.SameName{Name: test.name}, Quantity: cardprogram.SameNameAll,
			})
			if err != nil || !found.Available || found.Player != game.Player1 {
				t.Fatalf("search failed: %+v, %v", found, err)
			}
			var got []id.ID
			for _, card := range found.Cards {
				got = append(got, card.Identity.CardID)
				wantZone := zone.Graveyard
				if card.Identity.CardID == splitCard {
					wantZone = zone.Hand
				} else if card.Identity.CardID == delver {
					wantZone = zone.Library
				}
				if card.Zone != wantZone {
					t.Fatalf("found card %d reported origin %v, want %v", card.Identity.CardID, card.Zone, wantZone)
				}
			}
			if !slices.Equal(sortedIDs(got), sortedIDs(test.want(splitCard, fire, ice, delver))) {
				t.Fatalf("found %v", got)
			}
			if librarySearchedEvents(g, start) != 1 {
				t.Fatal("searching your own library did not fire one library-searched event")
			}
		})
	}
}

func TestSearchSameNameCountedQuantities(t *testing.T) {
	t.Parallel()
	spear := &game.CardDef{CardFace: game.CardFace{
		Name: "The Spear of Leonidas", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Equipment},
	}}
	for _, test := range []struct {
		name             string
		quantity         cardprogram.SameNameQuantity
		graveyard        bool
		minimum, maximum int
	}{
		{"named card in graveyard must be found", cardprogram.SameNameCards, true, 1, 1},
		{"named card only in hidden zones may be missed", cardprogram.SameNameCards, false, 0, 1},
		{"up to", cardprogram.SameNameUpTo, true, 0, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var graveyard id.ID
			if test.graveyard {
				graveyard = addCardToGraveyard(g, game.Player1, spear)
			}
			hand := addCardToHand(g, game.Player1, spear)
			addCardToLibrary(g, game.Player1, spear)
			var asked game.ChoiceRequest
			searcher := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				asked = req
				return nil
			}}
			ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{searcher})
			found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
				Player: game.ControllerReference(), Name: cardprogram.SameName{Name: "The Spear of Leonidas"},
				Quantity: test.quantity, Max: test.maximum,
			})
			if err != nil || asked.MinChoices != test.minimum || asked.MaxChoices != test.maximum {
				t.Fatalf("asked min=%d max=%d, %v", asked.MinChoices, asked.MaxChoices, err)
			}
			if test.minimum == 0 {
				if len(found.Cards) != 0 {
					t.Fatal("declining an optional counted search still found cards")
				}
				return
			}
			// Declining a required graveyard find falls back to the default,
			// which takes the graveyard copy first.
			if len(found.Cards) != 1 || found.Cards[0].Identity.CardID != graveyard {
				t.Fatalf("required find took %+v", found.Cards)
			}
			moved, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Battlefield})
			if err != nil || len(moved) != 1 || moved[0].Zone != zone.Battlefield ||
				permanentByCardID(g, graveyard) == nil || !g.Players[game.Player1].Hand.Contains(hand) {
				t.Fatalf("found graveyard card did not enter the battlefield: %+v, %v", moved, err)
			}
		})
	}
}

func TestSearchSameNameOppositionAgentControlsAndExilesEveryFind(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player4, oppositionAgentDef())
	bears := vanillaCreature("Grizzly Bears", 2, 2)
	finds := []id.ID{
		addCardToGraveyard(g, game.Player1, bears), addCardToHand(g, game.Player1, bears), addCardToLibrary(g, game.Player1, bears),
	}
	addCardToLibrary(g, game.Player1, vanillaCreature(sameNameSearchFillerDefName, 1, 1))
	ordinary := &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int {
		t.Fatal("the searching player chose finds under Opposition Agent")
		return nil
	}}
	controller := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind != game.ChoiceSearch || obs.DecisionMaker != game.Player4 || len(req.Options) != 2 {
			t.Fatalf("Opposition Agent's controller did not choose the hidden finds: %+v", req)
		}
		assertSearcherSeesHiddenZones(t, g, obs, game.Player1)
		return []int{0, 1}
	}}
	ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{ordinary, ordinary, ordinary, controller})
	found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
		Player: game.ControllerReference(), Name: cardprogram.SameName{Name: "Grizzly Bears"}, Quantity: cardprogram.SameNameAll,
	})
	if err != nil || !found.Available || len(found.Cards) != 0 {
		t.Fatalf("controlled finds were returned for placement: %+v, %v", found, err)
	}
	for _, card := range finds {
		permission, exists := playFromExilePermissionFor(g, card)
		if !g.Players[game.Player1].Exile.Contains(card) || !exists || permission.Controller != game.Player4 {
			t.Fatalf("find %d was not exiled for Opposition Agent's controller to play", card)
		}
	}
	if len(g.SearchControlScopes) != 0 || len(g.HiddenZoneSearchScopes) != 0 || g.Players[game.Player1].Library.Size() != 1 {
		t.Fatal("the controlled search left a scope open or took the differently named card")
	}
}

// A search of another player's hand and library shows the searcher, and only
// the searcher, every card in both zones, matching or not, for the whole search
// and no longer (CR 401.2, 402.3, 701.23a).
func TestSearchSameNameSearcherPrivatelySeesSearchedHandAndLibrary(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToGraveyard(g, game.Player2, vanillaCreature("Grizzly Bears", 2, 2))
	for range 2 {
		addCardToHand(g, game.Player2, vanillaCreature(sameNameSearchFillerDefName, 1, 1))
		addCardToLibrary(g, game.Player2, vanillaCreature("Hill Giant", 3, 3))
	}
	addCardToHand(g, game.Player1, vanillaCreature("Hill Giant", 3, 3))
	hand, library := g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Library.All()
	asked := 0
	searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		asked++
		if req.Kind != game.ChoiceSearch || len(req.Options) != 0 || !req.Private {
			t.Fatalf("the search offered hidden finds with no matches there: %+v", req)
		}
		assertSearcherSeesHiddenZones(t, g, obs, game.Player2)
		for _, other := range []game.PlayerID{game.Player2, game.Player3} {
			view := NewObservation(g, other)
			if len(view.HandOf(game.Player1)) != 0 || other != game.Player2 && (len(view.HandOf(game.Player2)) != 0 || len(view.LibraryLookable(game.Player2)) != 0) {
				t.Fatalf("player %d saw a hidden zone during another player's search", other)
			}
		}
		if len(NewObservation(g, game.Player2).LibraryLookable(game.Player2)) != 0 {
			t.Fatal("the searched player gained a look at their own library order")
		}
		world := g.Clone()
		determinize(world, game.Player1, rand.New(rand.NewPCG(7, 7)))
		if !slices.Equal(world.Players[game.Player2].Hand.All(), hand) || !slices.Equal(world.Players[game.Player2].Library.All(), library) {
			t.Fatal("a determinized search world resampled the hand and library the searcher was looking at")
		}
		return nil
	}}
	ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{searcher})
	found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
		Player: game.CapturedPlayerReference(game.Player2), Name: cardprogram.SameName{Name: "Grizzly Bears"}, Quantity: cardprogram.SameNameAll,
	})
	if err != nil || len(found.Cards) != 1 || asked != 1 {
		t.Fatalf("search: %+v, asked %d, %v", found, asked, err)
	}
	after := NewObservation(g, game.Player1)
	if len(g.HiddenZoneSearchScopes) != 0 || len(after.HandOf(game.Player2)) != 0 || len(after.LibraryLookable(game.Player2)) != 0 {
		t.Fatal("the searcher kept looking at the hand and library after the search")
	}
}

func TestSearchSameNameLibraryHandlesExpireOnShuffle(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bears := vanillaCreature("Grizzly Bears", 2, 2)
	graveyard := addCardToGraveyard(g, game.Player2, bears)
	library := addCardToLibrary(g, game.Player2, bears)
	ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{&controlScript{}})
	player := game.CapturedPlayerReference(game.Player2)
	found, err := ctx.SearchSameName(cardprogram.SameNameSearch{
		Player: player, Name: cardprogram.SameName{Name: "Grizzly Bears"}, Quantity: cardprogram.SameNameAll,
	})
	if err != nil || len(found.Cards) != 2 {
		t.Fatalf("search: %+v, %v", found, err)
	}
	if _, err := ctx.Execute(game.ShuffleLibrary{Player: player}); err != nil {
		t.Fatal(err)
	}
	moved, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Exile})
	if err != nil || len(moved) != 1 || moved[0].Identity.CardID != graveyard || moved[0].Zone != zone.Exile ||
		!g.Players[game.Player2].Library.Contains(library) {
		t.Fatalf("a library find survived its library's shuffle: %+v, %v", moved, err)
	}
}

func TestSearchSameNameRejectsMalformedRequests(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := sameNameTestContext(t, g, [game.NumPlayers]PlayerAgent{})
	name := cardprogram.SameName{Name: "Grizzly Bears"}
	for _, search := range []cardprogram.SameNameSearch{
		{Player: game.ControllerReference(), Quantity: cardprogram.SameNameAll},
		{Player: game.ControllerReference(), Name: cardprogram.SameName{Name: "A", Snapshot: opt.Val(game.ObjectSnapshot{})}},
		{Player: game.ControllerReference(), Name: name, Quantity: cardprogram.SameNameAll, Max: 1},
		{Player: game.ControllerReference(), Name: name, Quantity: cardprogram.SameNameUpTo},
		{Player: game.ControllerReference(), Name: name, Quantity: cardprogram.SameNameCards, Max: -1},
		{Player: game.ControllerReference(), Name: cardprogram.SameName{Spell: opt.Val(game.TargetPermanentReference(0))}},
	} {
		if _, err := ctx.SearchSameName(search); err == nil {
			t.Fatalf("malformed same-name search was accepted: %+v", search)
		}
	}
	ctx.active = false
	if _, err := ctx.SearchSameName(cardprogram.SameNameSearch{Player: game.ControllerReference(), Name: name}); err == nil {
		t.Fatal("an expired invocation searched")
	}
}
