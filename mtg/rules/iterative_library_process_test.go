package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// iterativeLibraryAgent drives the IterativeLibraryProcess runtime through the
// real choice pipeline. It answers each "you may put that card into your hand?"
// prompt from mayAnswers in order (defaulting to no once exhausted) and answers
// the "Choose a card name." naming prompt with the option whose label matches
// nameToChoose.
type iterativeLibraryAgent struct {
	nameToChoose string
	mayAnswers   []bool
	mayIndex     int

	// sawAbsentOption records whether any naming choice offered the
	// absent-name sentinel, so tests can assert it is (Demonic Consultation) or
	// is not (Tainted Pact) presented.
	sawAbsentOption bool
}

func (*iterativeLibraryAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *iterativeLibraryAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		answer := false
		if a.mayIndex < len(a.mayAnswers) {
			answer = a.mayAnswers[a.mayIndex]
		}
		a.mayIndex++
		if answer {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceResolution:
		selection := request.DefaultSelection
		// Select the last option whose label matches. The absent-name sentinel is
		// always appended last, so this reaches it even when a real card shares
		// its visible label, exercising the index-based (not label-based)
		// resolution in production.
		for _, option := range request.Options {
			if option.Label == absentLibraryNameLabel {
				a.sawAbsentOption = true
			}
			if option.Label == a.nameToChoose {
				selection = []int{option.Index}
			}
		}
		return selection
	default:
		return request.DefaultSelection
	}
}

// resolveIterativeLibrary resolves one IterativeLibraryProcess instruction for
// Player1 with the given agent and returns the game so callers can inspect the
// resulting zones and events.
func resolveIterativeLibrary(g *game.Game, prim game.IterativeLibraryProcess, agent *iterativeLibraryAgent) {
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	instr := &game.Instruction{Primitive: prim}
	engine.resolveInstructionWithChoices(g, obj, instr, agents, &TurnLog{})
}

func taintedPactPrimitive() game.IterativeLibraryProcess {
	return game.IterativeLibraryProcess{
		Player:       game.ControllerReference(),
		Stop:         game.IterativeLibraryStopDuplicateName,
		OptionalTake: true,
	}
}

func demonicConsultationPrimitive() game.IterativeLibraryProcess {
	return game.IterativeLibraryProcess{
		Player:          game.ControllerReference(),
		Stop:            game.IterativeLibraryStopChosenName,
		ChooseName:      true,
		Reveal:          true,
		PreExile:        game.Fixed(6),
		AllowAbsentName: true,
	}
}

// TestTaintedPactStopsOnDuplicateName exiles cards one at a time and stops the
// moment a name repeats, leaving the duplicate exiled and the rest of the
// library untouched.
func TestTaintedPactStopsOnDuplicateName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Library top -> bottom: Alpha, Beta, Alpha (duplicate), Gamma.
	gamma := addCardToLibraryNamed(g, game.Player1, "Gamma")
	dupAlpha := addCardToLibraryNamed(g, game.Player1, "Alpha")
	beta := addCardToLibraryNamed(g, game.Player1, "Beta")
	topAlpha := addCardToLibraryNamed(g, game.Player1, "Alpha")

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false, false}})

	p := g.Players[game.Player1]
	for _, cardID := range []id.ID{topAlpha, beta, dupAlpha} {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d not exiled", cardID)
		}
	}
	if !p.Library.Contains(gamma) {
		t.Fatal("Gamma past the duplicate was disturbed")
	}
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (all takes declined)", p.Hand.Size())
	}
}

// TestTaintedPactTakesCardIntoHand accepts the first optional take, moving the
// exiled card into hand and ending the process.
func TestTaintedPactTakesCardIntoHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Library top -> bottom: Alpha, Beta.
	beta := addCardToLibraryNamed(g, game.Player1, "Beta")
	alpha := addCardToLibraryNamed(g, game.Player1, "Alpha")

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{true}})

	p := g.Players[game.Player1]
	if !p.Hand.Contains(alpha) {
		t.Fatal("Alpha not put into hand")
	}
	if p.Exile.Contains(alpha) {
		t.Fatal("Alpha still exiled after being taken to hand")
	}
	if !p.Library.Contains(beta) {
		t.Fatal("Beta was disturbed after the take ended the process")
	}
}

// TestTaintedPactDeclinesUntilLibraryEmpty exiles every card of a unique
// singleton deck when each take is declined, terminating on the empty library.
func TestTaintedPactDeclinesUntilLibraryEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	gamma := addCardToLibraryNamed(g, game.Player1, "Gamma")
	beta := addCardToLibraryNamed(g, game.Player1, "Beta")
	alpha := addCardToLibraryNamed(g, game.Player1, "Alpha")

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false, false, false}})

	p := g.Players[game.Player1]
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0", p.Library.Size())
	}
	for _, cardID := range []id.ID{alpha, beta, gamma} {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d not exiled", cardID)
		}
	}
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0", p.Hand.Size())
	}
}

// TestTaintedPactEmptyLibrary resolves against an empty library without panic
// and produces no zone changes.
func TestTaintedPactEmptyLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{})
	p := g.Players[game.Player1]
	if p.Exile.Size() != 0 || p.Hand.Size() != 0 {
		t.Fatalf("exile=%d hand=%d, want both 0", p.Exile.Size(), p.Hand.Size())
	}
}

// TestTaintedPactMatchesFrontFaceName treats two double-faced cards that share a
// front-face name as duplicates even though their back faces differ, proving the
// name predicate reads the front face (split/DFC names).
func TestTaintedPactMatchesFrontFaceName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	twinB := addDoubleFacedCardToLibrary(g, game.Player1, "Twin", "Back Two")
	twinA := addDoubleFacedCardToLibrary(g, game.Player1, "Twin", "Back One")

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false}})

	p := g.Players[game.Player1]
	if !p.Exile.Contains(twinA) || !p.Exile.Contains(twinB) {
		t.Fatal("both same-front-name DFCs should be exiled by the duplicate stop")
	}
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0", p.Hand.Size())
	}
}

// TestTaintedPactEmitsNoRevealEvents proves the duplicate-name process exiles
// without revealing: none of its cards are shown to opponents (hidden info).
func TestTaintedPactEmitsNoRevealEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibraryNamed(g, game.Player1, "Beta")
	addCardToLibraryNamed(g, game.Player1, "Alpha")

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false, false}})

	assertNoEvent(t, g.Events, game.EventCardRevealed, func(game.Event) bool { return true })
}

// TestTaintedPactHistoryIsPerResolution proves each resolution starts with an
// empty processed-name history: a name exiled in the first resolution does not
// pre-seed a duplicate stop in a second, independent resolution (source/copy
// independence).
func TestTaintedPactHistoryIsPerResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// First resolution exiles a single "Echo".
	firstEcho := addCardToLibraryNamed(g, game.Player1, "Echo")
	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false}})
	if !g.Players[game.Player1].Exile.Contains(firstEcho) {
		t.Fatal("first Echo not exiled")
	}

	// Second resolution over a fresh library with the same name must not stop
	// immediately as a duplicate; it declines, exiles, and empties the library.
	secondEcho := addCardToLibraryNamed(g, game.Player1, "Echo")
	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false}})
	if !g.Players[game.Player1].Exile.Contains(secondEcho) {
		t.Fatal("second Echo not exiled — history leaked across resolutions")
	}
	if g.Players[game.Player1].Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0", g.Players[game.Player1].Library.Size())
	}
}

// TestTaintedPactCommanderExileRedirects proves the per-iteration exile honors
// the commander-zone replacement (CR 903.9): a commander exiled this way goes to
// the command zone instead of exile (replacement effects).
func TestTaintedPactCommanderExileRedirects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCardToLibraryNamed(g, game.Player1, "General")
	if g.CommanderIDs == nil {
		g.CommanderIDs = map[id.ID]bool{}
	}
	g.CommanderIDs[commander] = true

	resolveIterativeLibrary(g, taintedPactPrimitive(), &iterativeLibraryAgent{mayAnswers: []bool{false}})

	p := g.Players[game.Player1]
	if !p.CommandZone.Contains(commander) {
		t.Fatal("commander not redirected to the command zone")
	}
	if p.Exile.Contains(commander) {
		t.Fatal("commander wrongly landed in exile")
	}
}

// TestDemonicConsultationNamedOnTop names a card, exiles the top six, then
// reveals the named card immediately and puts it into hand.
func TestDemonicConsultationNamedOnTop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Library top -> bottom: C1..C6, Named, Rest.
	rest := addCardToLibraryNamed(g, game.Player1, "Rest")
	named := addCardToLibraryNamed(g, game.Player1, "Named")
	filler := addSixFillerCardsOnTop(g, game.Player1)

	resolveIterativeLibrary(g, demonicConsultationPrimitive(), &iterativeLibraryAgent{nameToChoose: "Named"})

	p := g.Players[game.Player1]
	if !p.Hand.Contains(named) {
		t.Fatal("Named card not put into hand")
	}
	for _, cardID := range filler {
		if !p.Exile.Contains(cardID) {
			t.Fatal("a pre-exiled top-six card is not in exile")
		}
	}
	if !p.Library.Contains(rest) {
		t.Fatal("card past the named card was disturbed")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == named
	})
}

// TestDemonicConsultationNamedAfterSix reveals several non-matching cards before
// the named card, exiling each until the named card is found and taken to hand.
func TestDemonicConsultationNamedAfterSix(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Library top -> bottom: C1..C6, Filler1, Filler2, Named, Rest.
	rest := addCardToLibraryNamed(g, game.Player1, "Rest")
	named := addCardToLibraryNamed(g, game.Player1, "Named")
	filler2 := addCardToLibraryNamed(g, game.Player1, "Filler2")
	filler1 := addCardToLibraryNamed(g, game.Player1, "Filler1")
	topSix := addSixFillerCardsOnTop(g, game.Player1)

	resolveIterativeLibrary(g, demonicConsultationPrimitive(), &iterativeLibraryAgent{nameToChoose: "Named"})

	p := g.Players[game.Player1]
	if !p.Hand.Contains(named) {
		t.Fatal("Named card not put into hand")
	}
	for _, cardID := range append(topSix, filler1, filler2) {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d should be exiled", cardID)
		}
	}
	if !p.Library.Contains(rest) {
		t.Fatal("card past the named card was disturbed")
	}
	for _, cardID := range []id.ID{filler1, filler2, named} {
		assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
			return event.CardID == cardID
		})
	}
	// The pre-exiled six are not revealed — only the reveal-until phase is.
	for _, cardID := range topSix {
		assertNoEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
			return event.CardID == cardID
		})
	}
}

// TestDemonicConsultationNamedCardExiledNotFound proves that when the named card
// sits among the six pre-exiled cards, the reveal-until phase never finds it and
// the remaining library is exiled entirely (the classic self-mill risk).
func TestDemonicConsultationNamedCardExiledNotFound(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Library top -> bottom: Named, C2..C6, R1, R2.
	r2 := addCardToLibraryNamed(g, game.Player1, "R2")
	r1 := addCardToLibraryNamed(g, game.Player1, "R1")
	c := []id.ID{}
	for i := 6; i >= 2; i-- {
		c = append(c, addCardToLibraryNamed(g, game.Player1, "Filler"))
	}
	named := addCardToLibraryNamed(g, game.Player1, "Named")

	resolveIterativeLibrary(g, demonicConsultationPrimitive(), &iterativeLibraryAgent{nameToChoose: "Named"})

	p := g.Players[game.Player1]
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (named card was pre-exiled)", p.Hand.Size())
	}
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0 (whole library exiled)", p.Library.Size())
	}
	for _, cardID := range append(append(c, r1, r2), named) {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d should be exiled", cardID)
		}
	}
}

// TestDemonicConsultationAbsentNameExilesWholeLibrary names a card that is not
// in the library (the absent-name sentinel). With more than six cards present,
// the top six are pre-exiled and, since the sentinel never matches, every
// remaining card is revealed and exiled, leaving an empty library and empty
// hand — the defining "exile your whole library" Consultation line.
func TestDemonicConsultationAbsentNameExilesWholeLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Nine distinctly named cards: six pre-exiled, three revealed-and-exiled.
	var all []id.ID
	for i := 9; i >= 1; i-- {
		all = append(all, addCardToLibraryNamed(g, game.Player1, "Card"+string(rune('0'+i))))
	}

	agent := &iterativeLibraryAgent{nameToChoose: absentLibraryNameLabel}
	resolveIterativeLibrary(g, demonicConsultationPrimitive(), agent)

	if !agent.sawAbsentOption {
		t.Fatal("naming choice did not offer the absent-name sentinel")
	}
	p := g.Players[game.Player1]
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (absent name never matches)", p.Hand.Size())
	}
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0 (whole library exiled)", p.Library.Size())
	}
	for _, cardID := range all {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d should be exiled", cardID)
		}
	}
}

// TestDemonicConsultationAbsentNameFewerThanSix proves the absent-name line is
// safe when the library holds fewer than six cards: the pre-exile empties the
// library, the reveal loop finds nothing, and the process ends with everything
// exiled and no card in hand.
func TestDemonicConsultationAbsentNameFewerThanSix(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	a := addCardToLibraryNamed(g, game.Player1, "Aardvark")
	b := addCardToLibraryNamed(g, game.Player1, "Badger")
	c := addCardToLibraryNamed(g, game.Player1, "Cheetah")

	agent := &iterativeLibraryAgent{nameToChoose: absentLibraryNameLabel}
	resolveIterativeLibrary(g, demonicConsultationPrimitive(), agent)

	if !agent.sawAbsentOption {
		t.Fatal("naming choice did not offer the absent-name sentinel")
	}
	p := g.Players[game.Player1]
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0", p.Hand.Size())
	}
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0", p.Library.Size())
	}
	for _, cardID := range []id.ID{a, b, c} {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d should be exiled", cardID)
		}
	}
}

// TestDemonicConsultationAbsentNameEmptyLibrary proves the naming choice — and
// its absent-name sentinel — is still offered against an empty library, and the
// process resolves without panic and touches nothing.
func TestDemonicConsultationAbsentNameEmptyLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	agent := &iterativeLibraryAgent{nameToChoose: absentLibraryNameLabel}
	resolveIterativeLibrary(g, demonicConsultationPrimitive(), agent)

	if !agent.sawAbsentOption {
		t.Fatal("empty library did not offer the absent-name sentinel")
	}
	p := g.Players[game.Player1]
	if p.Exile.Size() != 0 || p.Hand.Size() != 0 {
		t.Fatalf("exile=%d hand=%d, want both 0", p.Exile.Size(), p.Hand.Size())
	}
}

// TestDemonicConsultationSentinelNeverMatchesRealCardName proves the sentinel is
// resolved structurally by option index, not by its label: a real card literally
// named the sentinel label is not treated as the chosen card, so the whole
// library is still exiled when the sentinel is chosen.
func TestDemonicConsultationSentinelNeverMatchesRealCardName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A real card whose name collides with the sentinel's visible label, plus
	// enough cards to exceed the pre-exile so the reveal loop runs.
	var all []id.ID
	all = append(all, addCardToLibraryNamed(g, game.Player1, "Tail"))
	trap := addCardToLibraryNamed(g, game.Player1, absentLibraryNameLabel)
	all = append(all, trap)
	filler := addSixFillerCardsOnTop(g, game.Player1)
	all = append(all, filler...)

	agent := &iterativeLibraryAgent{nameToChoose: absentLibraryNameLabel}
	resolveIterativeLibrary(g, demonicConsultationPrimitive(), agent)

	p := g.Players[game.Player1]
	if p.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (sentinel must not match the real card)", p.Hand.Size())
	}
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0", p.Library.Size())
	}
	for _, cardID := range all {
		if !p.Exile.Contains(cardID) {
			t.Fatalf("card %d should be exiled (whole library exiled by the sentinel)", cardID)
		}
	}
	if !p.Exile.Contains(trap) {
		t.Fatal("the real card sharing the sentinel label should be exiled, not taken to hand")
	}
}

// TestDemonicConsultationPresentNameStillTakesToHand proves offering the absent
// sentinel does not break naming a card that is present: choosing a real library
// name still reveals and takes that card to hand.
func TestDemonicConsultationPresentNameStillTakesToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rest := addCardToLibraryNamed(g, game.Player1, "Rest")
	named := addCardToLibraryNamed(g, game.Player1, "Named")
	filler := addSixFillerCardsOnTop(g, game.Player1)

	agent := &iterativeLibraryAgent{nameToChoose: "Named"}
	resolveIterativeLibrary(g, demonicConsultationPrimitive(), agent)

	if !agent.sawAbsentOption {
		t.Fatal("naming choice did not offer the absent-name sentinel alongside real names")
	}
	p := g.Players[game.Player1]
	if !p.Hand.Contains(named) {
		t.Fatal("named present card not put into hand")
	}
	for _, cardID := range filler {
		if !p.Exile.Contains(cardID) {
			t.Fatal("a pre-exiled top-six card is not in exile")
		}
	}
	if !p.Library.Contains(rest) {
		t.Fatal("card past the named card was disturbed")
	}
}

// TestTaintedPactOffersNoAbsentName proves the duplicate-name process (Tainted
// Pact) never names a card and therefore never offers the absent-name sentinel:
// AllowAbsentName is scoped to the chosen-name shape only.
func TestTaintedPactOffersNoAbsentName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibraryNamed(g, game.Player1, "Beta")
	addCardToLibraryNamed(g, game.Player1, "Alpha")

	agent := &iterativeLibraryAgent{mayAnswers: []bool{false, false}}
	resolveIterativeLibrary(g, taintedPactPrimitive(), agent)

	if agent.sawAbsentOption {
		t.Fatal("Tainted Pact must not offer the absent-name sentinel")
	}
}

// addDoubleFacedCardToLibrary adds a DFC with the given front and back names to
// the top of the player's library and returns its instance id.
func addDoubleFacedCardToLibrary(g *game.Game, playerID game.PlayerID, front, back string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{
			CardFace: game.CardFace{Name: front, Types: []types.Card{types.Creature}},
			Layout:   game.LayoutModalDFC,
			Back:     optCardFace(back),
		},
		Owner: playerID,
	}
	g.Players[playerID].Library.Add(cardID)
	return cardID
}

// addSixFillerCardsOnTop adds six distinctly named cards to the top of the
// player's library and returns them in top -> bottom order.
func addSixFillerCardsOnTop(g *game.Game, playerID game.PlayerID) []id.ID {
	names := []string{"C6", "C5", "C4", "C3", "C2", "C1"}
	added := make([]id.ID, 0, len(names))
	for _, name := range names {
		added = append(added, addCardToLibraryNamed(g, playerID, name))
	}
	// added is bottom -> top; reverse to top -> bottom.
	for i, j := 0, len(added)-1; i < j; i, j = i+1, j-1 {
		added[i], added[j] = added[j], added[i]
	}
	return added
}

func optCardFace(name string) opt.V[game.CardFace] {
	return opt.Val(game.CardFace{Name: name, Types: []types.Card{types.Creature}})
}
