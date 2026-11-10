package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned-corpus Oracle text for Balance (LEA
// ce648aa3-098b-4af0-a433-fd290bc85904, {1}{W} Sorcery, "Each player chooses a
// number of lands they control equal to the number of lands controlled by the
// player who controls the fewest, then sacrifices the rest. Players discard
// cards and sacrifice creatures the same way."), Balancing Act
// (7d6cb368-43f7-4d06-9cd5-492df5766567, {2}{W}{W} Sorcery, "Each player
// chooses a number of permanents they control equal to the number of permanents
// controlled by the player who controls the fewest, then sacrifices the rest.
// Each player discards cards the same way."), Restore Balance
// (0febc7fe-4172-4285-9804-a62ce4506c39, no mana cost, white Sorcery, "Suspend
// 6—{W} / Each player chooses a number of lands ... then sacrifices the rest.
// Players sacrifice creatures and discard cards the same way.") and Dryad Arbor
// (e3ddbebf-72cd-4d1b-ba0d-d94934654ab7, Land Creature — Forest Dryad).
// Scryfall /cards/<id>/rulings: Balance 2016-06-08 (APNAP land choices with
// knowledge of earlier choices, then a simultaneous sacrifice; cards in hand
// are not revealed until all have chosen, then discarded simultaneously; each
// type is counted during its own part, so creatures are counted after earlier
// steps and a land creature sacrificed as a land is gone; no targets);
// Restore Balance 2021-03-19 (lands, then creatures, then cards; triggers wait
// until it finishes resolving); Balancing Act has none. CR 101.4/101.4b (APNAP
// choices, hidden choices stay hidden until all are made, then simultaneous),
// 608.2c (instructions in order), 701.9a/701.9b (discard from hand, chosen by
// the discarding player), 701.21a (sacrifice) and 800.4a (a player who left
// the game is gone) apply. These fixtures prove the shared fewest-count amounts
// and KeepInHand discard; no card is admitted.

var (
	balanceLands     = game.Selection{RequiredTypes: []types.Card{types.Land}}
	balanceCreatures = game.Selection{RequiredTypes: []types.Card{types.Creature}}
)

func balanceKeep(sel *game.Selection) game.Instruction {
	keepSel := game.Selection{}
	if sel != nil {
		keepSel = *sel
	}
	return game.Instruction{Primitive: game.KeepPermanents{
		PlayerGroup:       game.AllPlayersReference(),
		AffectedSelection: keepSel,
		Keep:              game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountFewestPermanentsControlled, Selection: sel}),
	}}
}

func balanceDiscard() game.Instruction {
	return game.Instruction{Primitive: game.Discard{
		PlayerGroup: game.AllPlayersReference(),
		KeepInHand:  true,
		Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountFewestCardsInHand}),
	}}
}

func balanceFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Balance",
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			balanceKeep(&balanceLands),
			balanceDiscard(),
			balanceKeep(&balanceCreatures),
		}}.Ability()),
	}}
}

func balanceHandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Sorcery},
	}}
}

func addBalanceHand(g *game.Game, playerID game.PlayerID, count int) []id.ID {
	cards := make([]id.ID, count)
	for i := range cards {
		cards[i] = addCardToHand(g, playerID, balanceHandCard("Hand Card"))
	}
	return cards
}

// balanceAgent answers each choice request with the next scripted selection and
// records what it was shown.
type balanceAgent struct {
	simPassPolicy

	script       [][]int
	requests     []game.ChoiceRequest
	observations []PlayerObservation
	onChoose     func(PlayerObservation, game.ChoiceRequest)
}

func (a *balanceAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	a.observations = append(a.observations, obs)
	if a.onChoose != nil {
		a.onChoose(obs, request)
	}
	if len(a.script) == 0 {
		return request.DefaultSelection
	}
	next := a.script[0]
	a.script = a.script[1:]
	return next
}

func countControlled(g *game.Game, playerID game.PlayerID, sel game.Selection) int {
	return len(playerControlledSelectionCandidates(g, newReferenceResolver(g, &game.StackObject{Controller: playerID}), nil, playerID, sel))
}

func eventBatch(g *game.Game, kind game.EventKind) (ids []id.ID, first, last int) {
	first = -1
	for i, event := range g.Events {
		if event.Kind != kind {
			continue
		}
		ids = append(ids, event.SimultaneousID)
		if first < 0 {
			first = i
		}
		last = i
	}
	return ids, first, last
}

func singleBatch(ids []id.ID) bool {
	return len(ids) > 0 && ids[0] != 0 && !slices.ContainsFunc(ids, func(batch id.ID) bool { return batch != ids[0] })
}

func TestBalanceEqualizesLandsThenHandsThenCreaturesInFourPlayerGame(t *testing.T) {
	dryadArbor := &game.CardDef{CardFace: game.CardFace{
		Name:      "Dryad Arbor",
		Types:     []types.Card{types.Land, types.Creature},
		Subtypes:  []types.Sub{types.Forest, types.Dryad},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Lands: P1 4, P2 3 (including Dryad Arbor), P3 2, P4 5 -> keep 2.
	for range 4 {
		addLandPermanent(g, game.Player1, "P1 Land")
	}
	addLandPermanent(g, game.Player2, "P2 Land")
	addLandPermanent(g, game.Player2, "P2 Land")
	arbor := addCombatPermanent(g, game.Player2, dryadArbor)
	p3Lands := []*game.Permanent{addLandPermanent(g, game.Player3, "P3 Land"), addLandPermanent(g, game.Player3, "P3 Land")}
	for range 5 {
		addLandPermanent(g, game.Player4, "P4 Land")
	}
	// Creatures: P1 3, P2 1 plus Dryad Arbor, P3 4, P4 2. Counted after the
	// land step, P2 has only one creature, so every player keeps one.
	var p1Creatures []*game.Permanent
	for range 3 {
		p1Creatures = append(p1Creatures, addCombatCreaturePermanentWithPower(g, game.Player1, 2))
	}
	p2Creature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	for range 4 {
		addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	}
	addCombatCreaturePermanentWithPower(g, game.Player4, 2)
	addCombatCreaturePermanentWithPower(g, game.Player4, 2)
	// Hands after casting: P1 3, P2 5, P3 1, P4 2 -> keep 1.
	p1Hand := addBalanceHand(g, game.Player1, 3)
	addBalanceHand(g, game.Player2, 5)
	p3Hand := addBalanceHand(g, game.Player3, 1)
	addBalanceHand(g, game.Player4, 2)

	castKeepFixture(t, g, engine, balanceFixture(), 0)
	p1 := &balanceAgent{script: [][]int{{0, 1}, {0, 1}, {0}}}
	// Player2 keeps the two plain lands, so Dryad Arbor is sacrificed as a land.
	p2 := &balanceAgent{script: [][]int{{0, 1}, {0, 1, 2, 3}}}
	// Player3 holds the fewest lands and cards but must still keep one creature.
	p3 := &balanceAgent{script: [][]int{{0}}}
	p4 := &balanceAgent{script: [][]int{{0, 1}, {0}, {0}}}
	p4.onChoose = func(_ PlayerObservation, request game.ChoiceRequest) {
		if request.Prompt == "Choose cards to discard" && !g.Players[game.Player1].Hand.Contains(p1Hand[1]) {
			t.Fatal("Player1's discard happened before every player chose")
		}
	}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player3], agents[game.Player4] = p1, p2, p3, p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentStillPresent(g, arbor.ObjectID) {
		t.Fatal("Dryad Arbor survived the land step")
	}
	for _, permanent := range append(p3Lands, p2Creature) {
		if !permanentStillPresent(g, permanent.ObjectID) {
			t.Fatalf("permanent %v of the player with the fewest was sacrificed", permanent.ObjectID)
		}
	}
	if !g.Players[game.Player3].Hand.Contains(p3Hand[0]) {
		t.Fatal("the player with the fewest cards discarded")
	}
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := countControlled(g, playerID, balanceLands); got != 2 {
			t.Fatalf("player %v lands = %d, want 2", playerID, got)
		}
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %v hand = %d, want 1", playerID, got)
		}
		if got := countControlled(g, playerID, balanceCreatures); got != 1 {
			t.Fatalf("player %v creatures = %d, want 1 (Dryad Arbor not counted)", playerID, got)
		}
	}

	if len(p1.requests) != 3 || len(p2.requests) != 2 || len(p3.requests) != 1 || len(p4.requests) != 3 {
		t.Fatalf("request counts = %d/%d/%d/%d, want 3/2/1/3", len(p1.requests), len(p2.requests), len(p3.requests), len(p4.requests))
	}
	if req := p4.requests[0]; req.MinChoices != 2 || req.MaxChoices != 2 || len(req.KeepDecisions) != 3 {
		t.Fatalf("Player4 land request = %#v, want keep 2 after three earlier public decisions", req)
	}
	if got := p2.requests[1]; !got.Private || got.MinChoices != 4 || len(got.Options) != 5 || got.KeepDecisions != nil {
		t.Fatalf("Player2 discard request = %#v, want private discard of 4 from 5", got)
	}
	if got := p1.requests[2]; got.MinChoices != 1 || got.MaxChoices != 1 || len(got.Options) != 3 {
		t.Fatalf("Player1 creature request = %#v, want keep 1 of 3", got)
	}
	if got := p3.requests[0]; got.MaxChoices != 1 || len(got.Options) != 4 {
		t.Fatalf("Player3 creature request = %#v, want keep 1 of 4", got)
	}

	landBatch := sacrificeBatchIDs(g, arbor)
	discards, firstDiscard, lastDiscard := eventBatch(g, game.EventCardDiscarded)
	if len(discards) != 2+4+1 || !singleBatch(discards) {
		t.Fatalf("discard batches = %v, want 7 discards in one simultaneous batch", discards)
	}
	if len(landBatch) != 1 || landBatch[0] == discards[0] {
		t.Fatalf("land sacrifice batch %v must be separate from the discard batch", landBatch)
	}
	if arborIndex := keepEventIndex(g, game.EventPermanentSacrificed, arbor); arborIndex > firstDiscard {
		t.Fatal("lands were not sacrificed before the discard")
	}
	if creatureIndex := keepEventIndex(g, game.EventPermanentSacrificed, p1Creatures[2]); creatureIndex < lastDiscard {
		t.Fatal("a creature was sacrificed before the discard finished")
	}
}

func TestBalancingActEqualizesPermanentsThenPrivateHands(t *testing.T) {
	balancingAct := &game.CardDef{CardFace: game.CardFace{
		Name:     "Balancing Act",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W, cost.W}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			balanceKeep(nil),
			balanceDiscard(),
		}}.Ability()),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	p1Kept := []*game.Permanent{
		addLandPermanent(g, game.Player1, "P1 Land"),
		addNamedTypedPermanent(g, game.Player1, "P1 Artifact", types.Artifact),
	}
	p1Sacrificed := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player1, 2),
		addNamedTypedPermanent(g, game.Player1, "P1 Enchantment", types.Enchantment),
	}
	addLandPermanent(g, game.Player2, "P2 Land")
	addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	for range 3 {
		addNamedTypedPermanent(g, game.Player3, "P3 Artifact", types.Artifact)
	}
	addLandPermanent(g, game.Player4, "P4 Land")
	addLandPermanent(g, game.Player4, "P4 Land")
	hands := map[game.PlayerID][]id.ID{
		game.Player1: addBalanceHand(g, game.Player1, 3),
		game.Player2: addBalanceHand(g, game.Player2, 2),
		game.Player3: addBalanceHand(g, game.Player3, 4),
		game.Player4: addBalanceHand(g, game.Player4, 2),
	}

	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	castKeepFixture(t, g, engine, balancingAct, 0)
	agents := simPassPolicies()
	recorders := map[game.PlayerID]*balanceAgent{}
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		recorder := &balanceAgent{script: [][]int{{0, 1}, {0, 1}}}
		recorder.onChoose = func(obs PlayerObservation, request game.ChoiceRequest) {
			if request.Prompt != "Choose cards to discard" {
				return
			}
			for _, other := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
				if other != playerID && obs.HandOf(other) != nil {
					t.Fatalf("player %v saw player %v's hand while choosing discards", playerID, other)
				}
			}
			for _, option := range request.Options {
				if !slices.Contains(hands[playerID], option.Card.Val.CardID) {
					t.Fatalf("player %v discard option %v is not from their own hand", playerID, option.Card.Val.CardID)
				}
			}
		}
		recorders[playerID] = recorder
		agents[playerID] = recorder
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, permanent := range p1Kept {
		if !permanentStillPresent(g, permanent.ObjectID) {
			t.Fatalf("kept permanent %v was sacrificed", permanent.ObjectID)
		}
	}
	for _, permanent := range p1Sacrificed {
		if permanentStillPresent(g, permanent.ObjectID) {
			t.Fatalf("unkept permanent %v survived", permanent.ObjectID)
		}
	}
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := countControlled(g, playerID, game.Selection{}); got != 2 {
			t.Fatalf("player %v permanents = %d, want 2", playerID, got)
		}
		if got := g.Players[playerID].Hand.Size(); got != 2 {
			t.Fatalf("player %v hand = %d, want 2", playerID, got)
		}
	}
	// Player2 and Player4 already hold the fewest and are never asked.
	if got := len(recorders[game.Player1].requests); got != 2 {
		t.Fatalf("Player1 requests = %d, want permanents and discard", got)
	}
	if got := len(recorders[game.Player3].requests); got != 2 {
		t.Fatalf("Player3 requests = %d, want permanents and discard", got)
	}
	if len(recorders[game.Player2].requests) != 0 || len(recorders[game.Player4].requests) != 0 {
		t.Fatal("a player holding the fewest was asked to choose")
	}
	if got := recorders[game.Player3].requests[1]; !got.Private || got.MinChoices != 2 || len(got.Options) != 4 {
		t.Fatalf("Player3 discard request = %#v, want private discard of 2 from 4", got)
	}
	if discards, _, _ := eventBatch(g, game.EventCardDiscarded); len(discards) != 3 || !singleBatch(discards) {
		t.Fatalf("discard batches = %v, want 3 discards in one simultaneous batch", discards)
	}
}

func TestRestoreBalanceFromSuspendEqualizesLandsCreaturesThenCards(t *testing.T) {
	restoreBalance := &game.CardDef{CardFace: game.CardFace{
		Name:   "Restore Balance",
		Types:  []types.Card{types.Sorcery},
		Colors: []color.Color{color.White},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.SuspendKeyword{Cost: cost.Mana{cost.W}, TimeCounters: 6}},
		}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			balanceKeep(&balanceLands),
			balanceKeep(&balanceCreatures),
			balanceDiscard(),
		}}.Ability()),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landCounts := [game.NumPlayers]int{3, 2, 4, 2}
	creatureCounts := [game.NumPlayers]int{3, 1, 2, 2}
	handCounts := [game.NumPlayers]int{3, 2, 1, 3}
	for playerID := range game.PlayerID(game.NumPlayers) {
		for range landCounts[playerID] {
			addLandPermanent(g, playerID, "Land")
		}
		for range creatureCounts[playerID] {
			addCombatCreaturePermanentWithPower(g, playerID, 2)
		}
		addBalanceHand(g, playerID, handCounts[playerID])
	}
	cardID := addSuspendedCard(g, game.Player1, restoreBalance, 1)

	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || !obj.Suspend {
		t.Fatalf("stack top = %+v, want Restore Balance cast from suspend", obj)
	}
	agents := simPassPolicies()
	recorders := [game.NumPlayers]*balanceAgent{}
	for playerID := range game.PlayerID(game.NumPlayers) {
		recorders[playerID] = &balanceAgent{}
		agents[playerID] = recorders[playerID]
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for playerID := range game.PlayerID(game.NumPlayers) {
		if got := countControlled(g, playerID, balanceLands); got != 2 {
			t.Fatalf("player %v lands = %d, want 2", playerID, got)
		}
		if got := countControlled(g, playerID, balanceCreatures); got != 1 {
			t.Fatalf("player %v creatures = %d, want 1", playerID, got)
		}
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %v hand = %d, want 1", playerID, got)
		}
	}
	prompts := func(playerID game.PlayerID) []string {
		var got []string
		for _, request := range recorders[playerID].requests {
			got = append(got, request.Prompt)
		}
		return got
	}
	keepPrompt := "Choose permanents to keep; the rest are sacrificed"
	discardPrompt := "Choose cards to discard"
	if got, want := prompts(game.Player1), []string{keepPrompt, keepPrompt, discardPrompt}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Player1 prompts = %v, want lands, creatures, then cards", got)
	}
	if got := prompts(game.Player2); !reflect.DeepEqual(got, []string{discardPrompt}) {
		t.Fatalf("Player2 prompts = %v, want only the discard", got)
	}
	_, firstDiscard, _ := eventBatch(g, game.EventCardDiscarded)
	_, _, lastSacrifice := eventBatch(g, game.EventPermanentSacrificed)
	if firstDiscard < 0 || lastSacrifice > firstDiscard {
		t.Fatal("cards were discarded before creatures were sacrificed")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Restore Balance did not finish resolving into its owner's graveyard")
	}
}

func TestBalanceResolvesIdenticallyInClonedGame(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landCounts := [game.NumPlayers]int{4, 2, 3, 5}
	creatureCounts := [game.NumPlayers]int{2, 3, 1, 4}
	handCounts := [game.NumPlayers]int{4, 2, 3, 5}
	for playerID := range game.PlayerID(game.NumPlayers) {
		for range landCounts[playerID] {
			addLandPermanent(g, playerID, "Land")
		}
		for range creatureCounts[playerID] {
			addCombatCreaturePermanentWithPower(g, playerID, 2)
		}
		addBalanceHand(g, playerID, handCounts[playerID])
	}
	castKeepFixture(t, g, engine, balanceFixture(), 0)
	clone := g.Clone()

	resolve := func(state *game.Game) {
		agents := simPassPolicies()
		for playerID := range game.PlayerID(game.NumPlayers) {
			// Keep the last options so the outcome differs from the defaults.
			agents[playerID] = &balanceAgent{onChoose: nil, script: [][]int{{1, 2}, {1, 2}, {1}}}
		}
		NewEngine(nil).resolveTopOfStackWithChoices(state, agents, &TurnLog{})
	}
	resolve(g)
	resolve(clone)

	summary := func(state *game.Game) [][]id.ID {
		var out [][]id.ID
		for playerID := range game.PlayerID(game.NumPlayers) {
			var battlefield []id.ID
			for _, permanent := range state.Battlefield {
				if permanent.Controller == playerID {
					battlefield = append(battlefield, permanent.ObjectID)
				}
			}
			player := state.Players[playerID]
			out = append(out, battlefield, player.Hand.All(), player.Graveyard.All())
		}
		return out
	}
	if got, want := summary(clone), summary(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("clone resolved to %v, original to %v", got, want)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2 (fewest was 2)", got)
	}
	if len(clone.Events) != len(g.Events) {
		t.Fatalf("clone events = %d, original = %d", len(clone.Events), len(g.Events))
	}
}

func TestFewestCountsExcludeEliminatedPlayersButCountEmptyLivePlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for playerID := range game.PlayerID(game.NumPlayers - 1) {
		for range int(playerID) + 2 {
			addLandPermanent(g, playerID, "Land")
		}
		addBalanceHand(g, playerID, int(playerID)+1)
	}
	// Player4 controls nothing and holds no cards.
	lands := game.DynamicAmount{Kind: game.DynamicAmountFewestPermanentsControlled, Selection: &balanceLands}
	hand := game.DynamicAmount{Kind: game.DynamicAmountFewestCardsInHand}
	obj := &game.StackObject{Controller: game.Player1}
	if got := dynamicAmountValue(g, obj, game.Player1, lands); got != 0 {
		t.Fatalf("fewest lands with an empty live player = %d, want 0", got)
	}
	if got := dynamicAmountValue(g, obj, game.Player1, hand); got != 0 {
		t.Fatalf("fewest hand with an empty live player = %d, want 0", got)
	}

	g.Players[game.Player4].Eliminated = true
	if got := dynamicAmountValue(g, obj, game.Player1, lands); got != 2 {
		t.Fatalf("fewest lands after elimination = %d, want 2", got)
	}
	if got := dynamicAmountValue(g, obj, game.Player1, hand); got != 1 {
		t.Fatalf("fewest hand after elimination = %d, want 1", got)
	}
}
