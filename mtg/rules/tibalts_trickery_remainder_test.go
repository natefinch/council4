package rules

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// tibaltRecordingAgent answers the free-cast "may" prompt from mayAnswers and
// counts every ChoiceResolution request it receives, so a test can prove the
// random mill count and random-bottom order are never turned into a player
// choice that would reveal or let the player influence them.
type tibaltRecordingAgent struct {
	mayAnswers         []bool
	mayIndex           int
	resolutionRequests int
}

func (*tibaltRecordingAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *tibaltRecordingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
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
		a.resolutionRequests++
		return request.DefaultSelection
	default:
		return request.DefaultSelection
	}
}

// tibaltSeedsForMillCount returns n distinct PCG-seeded RNGs whose first IntN(3)
// draw yields want-1, so each fixes Tibalt's Trickery's mill count to want while
// producing an independent later random-bottom shuffle. It lets a test show the
// remainder order genuinely varies with the RNG.
func tibaltSeedsForMillCount(t *testing.T, want, n int) []*rand.Rand {
	t.Helper()
	out := make([]*rand.Rand, 0, n)
	for seed := uint64(1); seed < 1_000_000 && len(out) < n; seed++ {
		if rand.New(rand.NewPCG(seed, seed)).IntN(3) == want-1 {
			out = append(out, rand.New(rand.NewPCG(seed, seed)))
		}
	}
	if len(out) < n {
		t.Fatalf("found only %d seeds for mill count %d, want %d", len(out), want, n)
	}
	return out
}

// addTargetedSorceryToLibrary adds a Sorcery named name that must target a
// creature to the top of the player's library and returns its id. With a legal
// creature in play its free cast auto-targets that creature; with none it has no
// legal cast and stays exiled.
func addTargetedSorceryToLibrary(g *game.Game, playerID game.PlayerID, name string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability()),
		}},
		Owner: playerID,
	}
	g.Players[playerID].Library.Add(cardID)
	return cardID
}

// tibaltRemainderResult is the outcome of runTibaltRemainderScenario: the resolved
// game plus the three cards exiled this way and bottomed — a land, a same-named
// nonland, and the declined different-named found card.
type tibaltRemainderResult struct {
	game     *game.Game
	land     id.ID
	sameName id.ID
	found    id.ID
}

// runTibaltRemainderScenario resolves a declined-cast Tibalt's Trickery whose
// mill is one, leaving exactly a land, a same-named nonland, and the found
// different-named nonland exiled this way and bottomed. It returns the game plus
// those three remainder card ids for assertions on membership and order.
func runTibaltRemainderScenario(t *testing.T, rng *rand.Rand) tibaltRemainderResult {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rng)

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	return tibaltRemainderResult{game: g, land: land, sameName: sameName, found: found}
}

// TestTibaltsTrickeryRemainderExcludesMilledAndCast confirms the random-bottom
// remainder is exactly the cards exiled by the process — never the milled cards
// (which are in the graveyard) nor the cast card (which is on the stack).
func TestTibaltsTrickeryRemainderExcludesMilledAndCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	milled := addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if !p.Graveyard.Contains(milled) {
		t.Fatal("milled card should be in the graveyard, not the remainder")
	}
	if p.Library.Contains(milled) {
		t.Fatal("milled card must not be bottomed with the remainder")
	}
	if _, ok := tibaltCastSpell(g, found); !ok {
		t.Fatal("found card should have been cast to the stack")
	}
	if p.Library.Contains(found) {
		t.Fatal("cast found card must not be bottomed with the remainder")
	}
	bottom := libraryBottomIDs(g, game.Player2, 2)
	want := map[id.ID]bool{land: true, sameName: true}
	for _, cardID := range bottom {
		if !want[cardID] {
			t.Fatalf("unexpected remainder card %d in %v", cardID, bottom)
		}
	}
}

// TestTibaltsTrickeryRemainderRandomizedAndDeterministic proves the random-bottom
// order is produced by the engine RNG: the same seed yields the same order every
// time, distinct seeds yield differing orders, the bottomed set is always exactly
// the exiled cards, and no player is ever prompted to order them.
func TestTibaltsTrickeryRemainderRandomizedAndDeterministic(t *testing.T) {
	// Determinism: identical seeds produce identical library order.
	seeds := tibaltSeedsForMillCount(t, 1, 1)
	res1 := runTibaltRemainderScenario(t, seeds[0])
	res2 := runTibaltRemainderScenario(t, tibaltSeedsForMillCount(t, 1, 1)[0])
	if fmt.Sprint(res1.game.Players[game.Player2].Library.All()) !=
		fmt.Sprint(res2.game.Players[game.Player2].Library.All()) {
		t.Fatal("same seed produced different random-bottom orders")
	}

	// Randomization: across distinct seeds the bottomed order varies, and every
	// run bottoms exactly the land, same-name, and found cards.
	orders := map[string]bool{}
	for _, rng := range tibaltSeedsForMillCount(t, 1, 12) {
		res := runTibaltRemainderScenario(t, rng)
		bottom := libraryBottomIDs(res.game, game.Player2, 3)
		set := map[id.ID]bool{res.land: true, res.sameName: true, res.found: true}
		for _, cardID := range bottom {
			if !set[cardID] {
				t.Fatalf("unexpected remainder card %d in %v", cardID, bottom)
			}
			delete(set, cardID)
		}
		if len(set) != 0 {
			t.Fatalf("remainder missing cards: %v", set)
		}
		orders[fmt.Sprint(bottom)] = true
	}
	if len(orders) < 2 {
		t.Fatalf("random-bottom order never varied across seeds: %v", orders)
	}

	// No leakage: the whole resolution asks the player no ordering choice; the
	// only prompt is the free-cast may.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))
	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	addCardToLibraryNamed(g, game.Player2, "Different Name")
	addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")
	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agent := &tibaltRecordingAgent{mayAnswers: []bool{false}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
	if agent.resolutionRequests != 0 {
		t.Fatalf("player was prompted with %d resolution choices; the mill count and bottom order must not be chosen", agent.resolutionRequests)
	}
}

// TestTibaltsTrickeryLeavesUnrelatedExileUntouched confirms only the cards this
// process exiled are bottomed: a card already in the controller's exile before
// resolution stays exiled.
func TestTibaltsTrickeryLeavesUnrelatedExileUntouched(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	unrelated := g.IDGen.Next()
	g.CardInstances[unrelated] = &game.CardInstance{
		ID:    unrelated,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Unrelated", Types: []types.Card{types.Sorcery}}},
		Owner: game.Player2,
	}
	g.Players[game.Player2].Exile.Add(unrelated)

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	addCardToLibraryNamed(g, game.Player2, "Different Name")
	addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(unrelated) {
		t.Fatal("unrelated exiled card was wrongly bottomed with the remainder")
	}
	if g.Players[game.Player2].Library.Contains(unrelated) {
		t.Fatal("unrelated exiled card must not enter the library")
	}
}

// TestTibaltsTrickeryCommanderDivertedFromRemainder confirms the commander-zone
// replacement applies: when the different-named nonland that stops the process is
// a commander, its owner may put it in the command zone, so it is never cast and
// never part of the random-bottom remainder.
func TestTibaltsTrickeryCommanderDivertedFromRemainder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	commander := addCardToLibraryNamed(g, game.Player2, "Different Name")
	sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")
	g.CommanderIDs = map[id.ID]bool{commander: true}

	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if !p.CommandZone.Contains(commander) {
		t.Fatal("commander that stopped the process was not diverted to the command zone")
	}
	if p.Exile.Contains(commander) || p.Library.Contains(commander) {
		t.Fatal("diverted commander must not stay exiled or be bottomed")
	}
	if _, ok := tibaltCastSpell(g, commander); ok {
		t.Fatal("a commander diverted to the command zone cannot be cast from exile")
	}
	// The land and same-named nonland exiled before the commander are still the
	// whole remainder.
	bottom := libraryBottomIDs(g, game.Player2, 2)
	want := map[id.ID]bool{land: true, sameName: true}
	for _, cardID := range bottom {
		if !want[cardID] {
			t.Fatalf("unexpected remainder card %d in %v", cardID, bottom)
		}
	}
}

// TestTibaltsTrickeryUncastableFoundCardIsBottomed confirms a found card the
// controller cannot legally cast (its target is unavailable) is left in exile by
// the failed cast and so becomes part of the random-bottom remainder.
func TestTibaltsTrickeryUncastableFoundCardIsBottomed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	// A sorcery that must target a creature, but no creature exists, so it cannot
	// be cast for free.
	found := addTargetedSorceryToLibrary(g, game.Player2, "Needs A Target")
	land := addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p := g.Players[game.Player2]
	if _, ok := tibaltCastSpell(g, found); ok {
		t.Fatal("uncastable found card must not reach the stack")
	}
	if p.Exile.Contains(found) {
		t.Fatal("uncastable found card should have been bottomed, not left exiled")
	}
	bottom := libraryBottomIDs(g, game.Player2, 2)
	want := map[id.ID]bool{land: true, found: true}
	for _, cardID := range bottom {
		if !want[cardID] {
			t.Fatalf("unexpected remainder card %d in %v", cardID, bottom)
		}
		delete(want, cardID)
	}
	if len(want) != 0 {
		t.Fatalf("remainder missing cards: %v", want)
	}
}

// TestTibaltsTrickeryCastsFoundCardWithTarget confirms additional targeting still
// applies to the free cast: a found card that must target a creature is cast with
// the only legal creature as its target.
func TestTibaltsTrickeryCastsFoundCardWithTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	creature := addCreaturePermanent(g, game.Player2)

	victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
	addCardToLibraryNamed(g, game.Player2, "Rest")
	found := addTargetedSorceryToLibrary(g, game.Player2, "Needs A Target")
	addLandToLibraryNamed(g, game.Player2, "Forest")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{true}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	castObj, ok := tibaltCastSpell(g, found)
	if !ok {
		t.Fatal("found card that needs a target was not cast")
	}
	if len(castObj.Targets) != 1 || castObj.Targets[0].StackObjectID != 0 {
		t.Fatalf("cast spell targets = %+v, want the single creature permanent", castObj.Targets)
	}
	if castObj.Targets[0].PermanentID != creature.ObjectID {
		t.Fatalf("cast spell target permanent = %d, want %d", castObj.Targets[0].PermanentID, creature.ObjectID)
	}
}

// TestTibaltsTrickeryResolutionsAreIsolated confirms two separate resolutions
// keep independent processed-card history: each mills and processes only its own
// target's controller, so one resolution's exiles never affect the other's stop.
func TestTibaltsTrickeryResolutionsAreIsolated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	// First resolution against Player2.
	victim2, _ := pushTibaltVictim(g, game.Player2, "Name Two")
	addCardToLibraryNamed(g, game.Player2, "Two Rest")
	found2 := addCardToLibraryNamed(g, game.Player2, "Two Different")
	addLandToLibraryNamed(g, game.Player2, "Two Land")
	addCardToLibraryNamed(g, game.Player2, "Two Milled")
	pushTibaltsTrickery(g, game.Player1, victim2.ID)
	NewEngine(tibaltSeedForMillCount(t, 1)).resolveTopOfStackWithChoices(
		g, [game.NumPlayers]PlayerAgent{game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}}}, &TurnLog{})

	// Second resolution against Player3, unrelated names.
	victim3, _ := pushTibaltVictim(g, game.Player3, "Name Three")
	addCardToLibraryNamed(g, game.Player3, "Three Rest")
	found3 := addCardToLibraryNamed(g, game.Player3, "Three Different")
	addLandToLibraryNamed(g, game.Player3, "Three Land")
	addCardToLibraryNamed(g, game.Player3, "Three Milled")
	pushTibaltsTrickery(g, game.Player1, victim3.ID)
	NewEngine(tibaltSeedForMillCount(t, 1)).resolveTopOfStackWithChoices(
		g, [game.NumPlayers]PlayerAgent{game.Player3: &iterativeLibraryAgent{mayAnswers: []bool{false}}}, &TurnLog{})

	// Each library was processed only for its own controller: the found card and
	// its exiled predecessors ended on that player's library bottom.
	if !g.Players[game.Player2].Library.Contains(found2) {
		t.Fatal("first resolution did not bottom its own found card")
	}
	if !g.Players[game.Player3].Library.Contains(found3) {
		t.Fatal("second resolution did not bottom its own found card")
	}
	// The second resolution never touched Player2, and vice versa.
	if g.Players[game.Player3].Library.Contains(found2) || g.Players[game.Player2].Library.Contains(found3) {
		t.Fatal("resolutions leaked cards across players")
	}
}

// TestTibaltsTrickeryComparesByCastFaceName confirms the different-name predicate
// uses the countered spell's cast-face name: a modal double-faced spell cast as
// its back face is compared by the back name, so a library card sharing that back
// name is skipped and one with the front name stops the process.
func TestTibaltsTrickeryComparesByCastFaceName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(tibaltSeedForMillCount(t, 1))

	victimCardID := g.IDGen.Next()
	g.CardInstances[victimCardID] = &game.CardInstance{
		ID: victimCardID,
		Def: &game.CardDef{
			CardFace: game.CardFace{Name: "Front Name", Types: []types.Card{types.Sorcery}},
			Layout:   game.LayoutModalDFC,
			Back:     optCardFace("Back Name"),
		},
		Owner: game.Player2,
	}
	victim := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   victimCardID,
		Controller: game.Player2,
		Face:       game.FaceBack,
	}
	g.Stack.Push(victim)

	addCardToLibraryNamed(g, game.Player2, "Rest")
	front := addCardToLibraryNamed(g, game.Player2, "Front Name")
	backSame := addCardToLibraryNamed(g, game.Player2, "Back Name")
	addCardToLibraryNamed(g, game.Player2, "Milled Top")

	pushTibaltsTrickery(g, game.Player1, victim.ID)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{false}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// "Back Name" matches the cast (back) face and is skipped; "Front Name"
	// differs and stops the process. Both are exiled this way and bottomed.
	bottom := libraryBottomIDs(g, game.Player2, 2)
	want := map[id.ID]bool{backSame: true, front: true}
	for _, cardID := range bottom {
		if !want[cardID] {
			t.Fatalf("unexpected remainder card %d in %v", cardID, bottom)
		}
		delete(want, cardID)
	}
	if len(want) != 0 {
		t.Fatalf("remainder missing cards: %v", want)
	}
}
