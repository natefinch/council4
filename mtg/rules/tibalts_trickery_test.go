package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// tibaltMillCountKey is the shared choice key Tibalt's Trickery's random-number
// choose publishes and its mill consumes. It matches the lowerer's
// tibaltsTrickeryMillCountKey so these runtime tests exercise the exact key
// wiring the generated card relies on.
const tibaltMillCountKey = game.ChoiceKey("tibalts-trickery-mill-count")

const (
	tibaltExiledKey = game.LinkedKey("tibalts-trickery-exiled")
	tibaltFoundKey  = game.ResultKey("tibalts-trickery-found")
)

// tibaltTrickerySequence returns the six-instruction resolution sequence exactly
// as lowerTibaltsTrickerySequence emits it: counter the target spell, choose
// 1..3 at random, mill that many from the countered spell's controller, then run
// the different-name-nonland iterative library process, optionally cast its
// found card, and random-bottom the linked remainder.
func tibaltTrickerySequence() []game.Instruction {
	targetRef := game.TargetStackObjectReference(0)
	controller := game.ObjectControllerReference(targetRef)
	return []game.Instruction{
		{Primitive: game.CounterObject{Object: targetRef}},
		{Primitive: game.Choose{
			Choice: game.ResolutionChoice{
				Kind:      game.ResolutionChoiceNumber,
				MinNumber: 1,
				MaxNumber: 3,
				AtRandom:  true,
			},
			PublishChoice: tibaltMillCountKey,
		}},
		{Primitive: game.Mill{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountChosenNumber,
				ResultKey: game.ResultKey(tibaltMillCountKey),
			}),
			Player: controller,
		}},
		{Primitive: game.IterativeLibraryProcess{
			Player:            controller,
			Stop:              game.IterativeLibraryStopDifferentNameNonland,
			DifferentNameFrom: targetRef,
			PublishLinked:     tibaltExiledKey,
		}, PublishResult: tibaltFoundKey},
		{
			Primitive: game.CastForFree{
				Player: controller,
				Card: game.CardReference{
					Kind:   game.CardReferenceLinked,
					LinkID: string(tibaltExiledKey),
				},
				Zone: zone.Exile,
			},
			Optional:      true,
			OptionalActor: opt.Val(controller),
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       tibaltFoundKey,
				Succeeded: game.TriTrue,
			}),
		},
		{Primitive: game.PutLinkedExiledCardsInLibrary{
			LinkedKey:   tibaltExiledKey,
			Bottom:      true,
			RandomOrder: true,
		}},
	}
}

// tibaltSpellTargetSpec is the "target spell" selector Tibalt's Trickery carries,
// matching stackSpellTargetSpec's lowering so resolution-time target legality and
// the all-targets-illegal fizzle behave exactly as they do for the real card.
func tibaltSpellTargetSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
			Controller:       game.ControllerAny,
		},
		Constraint: "target spell",
	}
}

// pushTibaltVictim puts a nonland Sorcery spell named name on the stack under
// controllerID's control and returns its stack object and card instance id, so a
// test can target it with Tibalt's Trickery and later inspect where the countered
// card went.
func pushTibaltVictim(g *game.Game, controllerID game.PlayerID, name string) (*game.StackObject, id.ID) {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}},
		Owner: controllerID,
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controllerID,
	}
	g.Stack.Push(obj)
	return obj, cardID
}

// pushTibaltsTrickery puts Tibalt's Trickery on the stack under casterID's
// control targeting victimObjID and returns its source card id.
func pushTibaltsTrickery(g *game.Game, casterID game.PlayerID, victimObjID id.ID) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Tibalt's Trickery",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets:  []game.TargetSpec{tibaltSpellTargetSpec()},
				Sequence: tibaltTrickerySequence(),
			}.Ability()),
		}},
		Owner: casterID,
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		Controller:   casterID,
		Targets:      []game.Target{game.StackObjectTarget(victimObjID)},
		TargetCounts: []int{1},
	})
	return sourceID
}

// addLandToLibraryNamed adds a Land named name to the top of the player's library
// and returns its instance id, the land counterpart of addCardToLibraryNamed.
func addLandToLibraryNamed(g *game.Game, playerID game.PlayerID, name string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}},
		Owner: playerID,
	}
	g.Players[playerID].Library.Add(cardID)
	return cardID
}

// tibaltSeedForMillCount returns a PCG-seeded RNG whose first IntN(3) draw yields
// want-1, so Tibalt's Trickery's "Choose 1, 2, or 3 at random." resolves to
// exactly want. The random-number choose is the first RNG use in the sequence, so
// this fixes the mill count while leaving the later random-bottom shuffle
// deterministic for the returned stream. Tests double-check the resulting count
// through the milled cards themselves.
func tibaltSeedForMillCount(t *testing.T, want int) *rand.Rand {
	t.Helper()
	for seed := uint64(1); seed < 1_000_000; seed++ {
		if rand.New(rand.NewPCG(seed, seed)).IntN(3) == want-1 {
			return rand.New(rand.NewPCG(seed, seed))
		}
	}
	t.Fatalf("no PCG seed produced mill count %d", want)
	return nil
}

// libraryBottomIDs returns the bottom n card ids of the player's library in
// top-to-bottom order, the cards Tibalt's Trickery's random-bottom remainder
// appends.
func libraryBottomIDs(g *game.Game, playerID game.PlayerID, n int) []id.ID {
	all := g.Players[playerID].Library.All()
	if n <= 0 || n > len(all) {
		return all
	}
	return all[len(all)-n:]
}

// tibaltCastSpell returns the stack object whose source is cardID, the spell
// Tibalt's Trickery cast from exile for free, and whether it is on the stack.
func tibaltCastSpell(g *game.Game, cardID id.ID) (*game.StackObject, bool) {
	for _, obj := range g.Stack.Objects() {
		if obj.SourceID == cardID {
			return obj, true
		}
	}
	return nil, false
}

// TestTibaltsTrickeryProcessesUntilDifferentNameNonland runs the whole real
// sequence with the free cast declined: the target is countered, its controller
// mills the random number of cards, then exiles a land and a same-named nonland
// before stopping on the first different-named nonland, and every card exiled
// this way goes to the bottom of the library.
func TestTibaltsTrickeryProcessesUntilDifferentNameNonland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, victimCardID := pushTibaltVictim(g, game.Player2, "Tibalt Victim")

	// Library top -> bottom after the single mill: milled, land, same-name
	// nonland, different-name nonland (stop), and an untouched rest.
	rest := addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	milled := addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if !p.Graveyard.Contains(victimCardID) {
		t.Fatal("countered victim spell did not move to its controller's graveyard")
	}
	if !p.Graveyard.Contains(milled) {
		t.Fatal("milled top card did not reach the graveyard")
	}
	if _, ok := stackObjectByID(g, victim.ID); ok {
		t.Fatal("countered victim remained on the stack")
	}
	// The land, same-name nonland, and different-name nonland were all exiled
	// this way, then bottomed. Nothing else moved.
	remainder := map[id.ID]bool{land: true, sameName: true, found: true}
	bottom := libraryBottomIDs(g, game.Player2, 3)
	for _, cardID := range bottom {
		if !remainder[cardID] {
			t.Fatalf("unexpected card %d among the bottomed remainder %v", cardID, bottom)
		}
		delete(remainder, cardID)
	}
	if len(remainder) != 0 {
		t.Fatalf("cards missing from the bottomed remainder: %v", remainder)
	}
	if top, _ := p.Library.Top(); top != rest {
		t.Fatalf("library top = %d, want the untouched Rest card %d", top, rest)
	}
	for _, cardID := range []id.ID{land, sameName, found} {
		if p.Exile.Contains(cardID) {
			t.Fatalf("card %d stayed exiled instead of being bottomed", cardID)
		}
	}
}

// TestTibaltsTrickeryCastsFoundCardWithoutPaying accepts the free cast: the
// different-named nonland is cast from exile onto the stack and so is excluded
// from the random-bottom remainder, which keeps only the land and same-named
// nonland exiled this way.
func TestTibaltsTrickeryCastsFoundCardWithoutPaying(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")

	addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	castObj, ok := tibaltCastSpell(g, found)
	if !ok {
		t.Fatal("found card was not cast onto the stack")
	}
	if castObj.Controller != game.Player2 {
		t.Fatalf("cast spell controller = %d, want the countered spell's controller", castObj.Controller)
	}
	if p.Library.Contains(found) || p.Exile.Contains(found) {
		t.Fatal("cast found card must not be in the library or exile")
	}
	// Only the land and same-name nonland were bottomed; the cast card was not.
	remainder := map[id.ID]bool{land: true, sameName: true}
	bottom := libraryBottomIDs(g, game.Player2, 2)
	for _, cardID := range bottom {
		if !remainder[cardID] {
			t.Fatalf("unexpected card %d among the bottomed remainder %v", cardID, bottom)
		}
		delete(remainder, cardID)
	}
	if len(remainder) != 0 {
		t.Fatalf("cards missing from the bottomed remainder: %v", remainder)
	}
}

// TestTibaltsTrickeryFaceDownTargetHasNoName verifies that a face-down targeted
// spell (Morph/Disguise/Cloak) is treated as nameless per CR 708.2: its concealed
// real name is never used as the reference name, so every named nonland differs
// from it and the process stops at the very first nonland. Here the hidden card and
// the first library nonland share the name "Hidden Ogre"; a nameless reference
// still makes that first nonland a different-named stop rather than a match, so it
// becomes the found card instead of being skipped.
func TestTibaltsTrickeryFaceDownTargetHasNoName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Hidden Ogre")
	victim.FaceDown = true

	untouched := addCardToLibraryNamed(g, game.Player2, "Untouched Sorcery")
	hiddenName := addCardToLibraryNamed(g, game.Player2, "Hidden Ogre")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// A nameless target makes the same-spelled "Hidden Ogre" nonland the first
	// different-named stop, so it is the found card and is cast from exile.
	if _, ok := tibaltCastSpell(g, hiddenName); !ok {
		t.Fatal("face-down target was not treated as nameless: the first nonland was skipped as a same-name match")
	}
	// Stopping at the first nonland leaves the card beneath it untouched.
	if !g.Players[game.Player2].Library.Contains(untouched) {
		t.Fatal("process exiled past the first nonland; the card below it must stay in the library")
	}
}

// TestTibaltsTrickeryUncounterableTargetStillProcesses confirms the mill, exile,
// free cast, and remainder all happen even when the targeted spell cannot be
// countered: the counter simply fails, the spell stays on the stack, and its
// controller is still milled and has their library processed by name.
func TestTibaltsTrickeryUncounterableTargetStillProcesses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, victimCardID := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	victim.RuleEffects = []game.RuleEffect{{Kind: game.RuleEffectCantBeCountered}}

	addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	milled := addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if _, ok := stackObjectByID(g, victim.ID); !ok {
		t.Fatal("uncounterable target should remain on the stack")
	}
	if p.Graveyard.Contains(victimCardID) {
		t.Fatal("uncounterable target must not be put into the graveyard by the failed counter")
	}
	if !p.Graveyard.Contains(milled) {
		t.Fatal("uncounterable target's controller was not milled")
	}
	// The name captured from the still-live spell still drives the different-name
	// stop: the same-named nonland is skipped and the different-named one stops it.
	remainder := map[id.ID]bool{land: true, sameName: true, found: true}
	bottom := libraryBottomIDs(g, game.Player2, 3)
	for _, cardID := range bottom {
		if !remainder[cardID] {
			t.Fatalf("unexpected card %d among the bottomed remainder %v", cardID, bottom)
		}
		delete(remainder, cardID)
	}
	if len(remainder) != 0 {
		t.Fatalf("cards missing from the bottomed remainder: %v", remainder)
	}
}

// TestTibaltsTrickeryIllegalTargetFizzles confirms the ordinary all-targets
// fizzle: when the targeted spell has left the stack before resolution, Tibalt's
// Trickery does nothing at all — no mill, no exile, library untouched.
func TestTibaltsTrickeryIllegalTargetFizzles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	pushTibaltsTrickery(g, game.Player1, victim.ID)

	// The target leaves the stack (resolved or countered) before Tibalt resolves.
	g.Stack.RemoveByID(victim.ID)

	// Populate the library so a stray mill or exile would be observable.
	keep := addCardToLibraryNamed(g, game.Player2, "Untouched")

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if p.Graveyard.Size() != 0 {
		t.Fatal("fizzled Tibalt's Trickery still milled cards")
	}
	if p.Exile.Size() != 0 {
		t.Fatal("fizzled Tibalt's Trickery still exiled cards")
	}
	if got := p.Library.All(); len(got) != 1 || got[0] != keep {
		t.Fatalf("fizzled Tibalt's Trickery disturbed the library: %v", got)
	}
}

// TestTibaltsTrickeryMillsEntireShortLibrary confirms that when the random mill
// count exceeds the library size, the whole library is milled and the following
// iterative process simply finds an empty library and does nothing.
func TestTibaltsTrickeryMillsEntireShortLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 3))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	a := addCardToLibraryNamed(g, game.Player2, "A")
	b := addCardToLibraryNamed(g, game.Player2, "B")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if p.Library.Size() != 0 {
		t.Fatalf("library size = %d, want fully milled", p.Library.Size())
	}
	for _, cardID := range []id.ID{a, b} {
		if !p.Graveyard.Contains(cardID) {
			t.Fatalf("card %d was not milled from the short library", cardID)
		}
	}
	if p.Exile.Size() != 0 {
		t.Fatal("nothing should have been exiled from an empty post-mill library")
	}
}

// TestTibaltsTrickeryEmptyLibraryDoesNothingAfterCounter confirms an empty
// library is handled cleanly: the target is still countered but there is nothing
// to mill, exile, cast, or bottom.
func TestTibaltsTrickeryEmptyLibraryDoesNothingAfterCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 2))

	victim, victimCardID := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if !p.Graveyard.Contains(victimCardID) {
		t.Fatal("target was not countered when the library was empty")
	}
	if p.Exile.Size() != 0 || p.Library.Size() != 0 {
		t.Fatal("empty-library resolution should leave exile and library empty")
	}
}

// TestTibaltsTrickeryExilesWholeLibraryWhenNoDifferentName confirms that when no
// nonland has a different name than the countered spell, the predicate never
// fires, so the entire remaining library is exiled and then bottomed with no card
// found or cast.
func TestTibaltsTrickeryExilesWholeLibraryWhenNoDifferentName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")

	// After one mill, the rest are lands and same-name nonlands only, so nothing
	// stops the process.
	sameA := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	sameB := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	milled := addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if !p.Graveyard.Contains(milled) {
		t.Fatal("top card was not milled")
	}
	if p.Exile.Size() != 0 {
		t.Fatal("all exiled cards should have been bottomed, leaving exile empty")
	}
	remainder := map[id.ID]bool{sameA: true, land: true, sameB: true}
	bottom := libraryBottomIDs(g, game.Player2, 3)
	for _, cardID := range bottom {
		if !remainder[cardID] {
			t.Fatalf("unexpected card %d among the bottomed remainder %v", cardID, bottom)
		}
		delete(remainder, cardID)
	}
	if len(remainder) != 0 {
		t.Fatalf("cards missing from the bottomed remainder: %v", remainder)
	}
}
