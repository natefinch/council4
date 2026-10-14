package rules

import (
	"slices"
	"strings"
	"testing"

	cardregistry "github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

const abundanceID = "eccc9a54-2b56-4a02-9926-258d5b2e25fb"

type abundanceChoices struct {
	accepts     []bool
	kinds       []int
	reverse     bool
	replacement string
	requests    []game.ChoiceRequest
}

func (*abundanceChoices) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *abundanceChoices) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if obs.Player != request.Player {
		panic("replacement choice observation belongs to the wrong player")
	}
	switch request.Kind {
	case game.ChoiceMay:
		accept := true
		if len(a.accepts) > 0 {
			accept, a.accepts = a.accepts[0], a.accepts[1:]
		}
		if accept {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceResolution:
		if request.Prompt == "Choose land or nonland." {
			kind := 0
			if len(a.kinds) > 0 {
				kind, a.kinds = a.kinds[0], a.kinds[1:]
			}
			return []int{kind}
		}
	case game.ChoiceReplacement:
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.replacement) {
				return []int{option.Index}
			}
		}
	case game.ChoiceOrder:
		order := make([]int, len(request.Options))
		for index := range order {
			order[index] = index
			if a.reverse {
				order[index] = len(order) - 1 - index
			}
		}
		return order
	default:
		return slices.Clone(request.DefaultSelection)
	}
	return slices.Clone(request.DefaultSelection)
}

func abundanceGame(t *testing.T) (*game.Game, *Engine, *cardregistry.Registry) {
	t.Helper()
	registry := cardregistry.NewDefaultRegistry()
	g, _ := setupBestowMain(t)
	e := NewEngineWithPrograms(nil, registry.Programs())
	def, err := registry.LookupOracleID(abundanceID)
	if err != nil {
		t.Fatal(err)
	}
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	if !e.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("Abundance could not be cast through the actual action path")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, card) == nil || len(g.ReplacementEffects) != 1 {
		t.Fatal("Abundance did not enter and register its replacement")
	}
	return g, e, registry
}

func abundanceLibrary(g *game.Game, names ...string) []id.ID {
	result := make([]id.ID, len(names))
	for index := len(names) - 1; index >= 0; index-- {
		cardType := types.Instant
		if strings.HasPrefix(names[index], "Land") {
			cardType = types.Land
		}
		result[index] = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: names[index], Types: []types.Card{cardType},
		}})
	}
	return result
}

func abundanceEvents(g *game.Game, kind game.EventKind) []id.ID {
	var cards []id.ID
	for _, event := range g.Events {
		if event.Kind == kind {
			cards = append(cards, event.CardID)
		}
	}
	return cards
}

func TestAbundanceIndividualDrawsDifferentKindsDeclineAndBottomOrder(t *testing.T) {
	t.Parallel()
	g, e, _ := abundanceGame(t)
	cards := abundanceLibrary(g, "A", "B", "Land1", "Land2", "C", "Keep")
	a := &abundanceChoices{accepts: []bool{true, true, false}, kinds: []int{0, 1}, reverse: true}
	log := &TurnLog{}
	if !e.drawCards(g, game.Player1, 3, [game.NumPlayers]PlayerAgent{a}, log) {
		t.Fatal("third, declined replacement did not draw")
	}
	player := g.Players[game.Player1]
	if !slices.Equal(player.Hand.All(), []id.ID{cards[5], cards[4], cards[2]}) ||
		!slices.Equal(player.Library.All(), []id.ID{cards[1], cards[0], cards[3]}) {
		t.Fatalf("hand/library = %v / %v", player.Hand.All(), player.Library.All())
	}
	if !slices.Equal(abundanceEvents(g, game.EventCardRevealed), cards[:5]) ||
		!slices.Equal(abundanceEvents(g, game.EventCardDrawn), []id.ID{cards[5]}) ||
		g.FailedDraws[game.Player1] {
		t.Fatal("replacement reveal/draw facts differ from actual actions")
	}
	var prompts []string
	for _, request := range a.requests {
		prompts = append(prompts, request.Prompt)
		if request.Player != game.Player1 {
			t.Fatal("someone other than the affected player chose")
		}
		if request.Kind == game.ChoiceOrder {
			for _, option := range request.Options {
				if !option.Card.Exists || option.Card.Val.Name == "" {
					t.Fatal("bottom order omitted the revealed card information")
				}
			}
		}
	}
	if len(prompts) != 6 || prompts[1] != "Choose land or nonland." || prompts[4] != "Choose land or nonland." {
		t.Fatalf("choice timing = %v", prompts)
	}
}

func TestAbundanceExhaustionEmptyAllMatchingAndDeclinedEmpty(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		library []string
		draws   int
		accept  bool
		hand    int
		reveal  int
		failed  bool
	}{
		{name: "no match", library: []string{"A", "B", "C"}, draws: 1, accept: true, reveal: 3},
		{name: "empty replacement", draws: 1, accept: true},
		{name: "all matching", library: []string{"Land1", "Land2"}, draws: 2, accept: true, hand: 2, reveal: 2},
		{name: "declined empty", draws: 1, failed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e, _ := abundanceGame(t)
			cards := abundanceLibrary(g, test.library...)
			a := &abundanceChoices{accepts: []bool{test.accept, test.accept}, reverse: true}
			drew := e.drawCards(g, game.Player1, test.draws, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
			if drew || g.Players[game.Player1].Hand.Size() != test.hand ||
				len(abundanceEvents(g, game.EventCardRevealed)) != test.reveal ||
				len(abundanceEvents(g, game.EventCardDrawn)) != 0 ||
				g.FailedDraws[game.Player1] != test.failed {
				t.Fatal("exhaustion/empty-library replacement was mistaken for a draw")
			}
			if test.name == "no match" {
				slices.Reverse(cards)
				if !slices.Equal(g.Players[game.Player1].Library.All(), cards) {
					t.Fatal("nonmatching cards were not all bottomed in chosen order")
				}
			}
			e.applyStateBasedActions(g)
			if g.Players[game.Player1].Eliminated != test.failed {
				t.Fatal("failed-draw loss did not distinguish decline from executed empty replacement")
			}
		})
	}
}

func TestAbundanceAndExistingDrawDoublerAndDrawTrigger(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		first     string
		accepts   []bool
		wantHand  int
		wantDraws int
	}{
		{name: "Abundance first consumes original", first: "choose land or nonland", wantHand: 1},
		{name: "doubler first replaces both descendants", first: "draw two cards", wantHand: 2},
		{name: "one descendant declined", first: "draw two cards", accepts: []bool{true, false}, wantHand: 2, wantDraws: 1},
		{name: "Abundance declined then doubler", first: "choose land or nonland", accepts: []bool{false, true, true}, wantHand: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e, _ := abundanceGame(t)
			abundanceLibrary(g, "Land1", "Land2", "Land3", "Land4")
			// Pinned Oracle shapes through existing shared declarations; these
			// fixtures do not admit either neighboring card to production.
			doubler := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Teferi's Ageless Insight",
				ReplacementAbilities: []game.ReplacementAbility{game.DrawCardMultiplierReplacement(
					"If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.", 2, true)},
			}})
			registerPermanentReplacementEffects(g, doubler)
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
			}, []game.Instruction{{Primitive: game.LoseLife{
				Amount: game.Fixed(1), PlayerGroup: game.OpponentsReference(),
			}}}, nil)
			a := &abundanceChoices{replacement: test.first, accepts: slices.Clone(test.accepts)}
			e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
			if g.Players[game.Player1].Hand.Size() != test.wantHand ||
				len(abundanceEvents(g, game.EventCardDrawn)) != test.wantDraws {
				t.Fatal("replacement ordering or ancestor lineage is incorrect")
			}
			e.putTriggeredAbilitiesOnStack(g)
			if g.Stack.Size() != test.wantDraws {
				t.Fatalf("Psychosis Crawler triggers = %d, want %d", g.Stack.Size(), test.wantDraws)
			}
			for g.Stack.Size() > 0 {
				e.resolveTopOfStack(g, &TurnLog{})
			}
			for player := game.Player2; player <= game.Player4; player++ {
				if g.Players[player].Life != 40-test.wantDraws {
					t.Fatal("actual draw trigger did not resolve exactly once per real draw")
				}
			}
			if len(g.ReplacementDecisions) != 1 || g.ReplacementDecisions[0].Player != game.Player1 {
				t.Fatal("CR 616 affected-player decision was not recorded")
			}
		})
	}
}

func TestAbundanceCloneSimulatorAndConcurrentGames(t *testing.T) {
	t.Parallel()
	g, e, registry := abundanceGame(t)
	cards := abundanceLibrary(g, "Land1", "A", "Land2", "B")
	addEffectSpellToStack(g, game.Player1, game.Draw{Amount: game.Fixed(2), Player: game.ControllerReference()}, nil)
	land := e.Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{&abundanceChoices{}})
	nonland := e.Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{&abundanceChoices{kinds: []int{1, 1}}})
	if !slices.Equal(land.Players[game.Player1].Hand.All(), []id.ID{cards[2], cards[0]}) ||
		!slices.Equal(nonland.Players[game.Player1].Hand.All(), []id.ID{cards[3], cards[1]}) ||
		g.Players[game.Player1].Hand.Size() != 0 || !slices.Equal(g.Players[game.Player1].Library.All(), cards) ||
		g.Stack.Size() != 1 || land.Stack.Size() != 0 || nonland.Stack.Size() != 0 {
		t.Fatal("Simulator replacement branches leaked state or failed to execute")
	}

	for index := range 8 {
		t.Run(string(rune('A'+index)), func(t *testing.T) {
			t.Parallel()
			clone := g.Clone()
			engine := NewEngineWithPrograms(nil, registry.Programs())
			result := engine.Simulator().ResolvePriority(clone, [game.NumPlayers]PlayerAgent{&abundanceChoices{}})
			if !slices.Equal(result.Players[game.Player1].Hand.All(), land.Players[game.Player1].Hand.All()) ||
				clone.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player1].Hand.Size() != 0 {
				t.Fatal("concurrent engines sharing immutable code leaked choices or event state")
			}
		})
	}
}

func TestAbundanceMultiPlayerDrawsUseEngineAPNAPOrder(t *testing.T) {
	t.Parallel()
	g, e, registry := abundanceGame(t)
	g.Turn.ActivePlayer = game.Player3
	def, err := registry.LookupOracleID(abundanceID)
	if err != nil {
		t.Fatal(err)
	}
	for player := game.Player1; player <= game.Player4; player++ {
		if player != game.Player1 {
			p := addCombatPermanent(g, player, def)
			registerPermanentReplacementEffects(g, p)
		}
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
			Name: "Land", Types: []types.Card{types.Land},
		}})
	}
	a := &abundanceChoices{}
	addEffectSpellToStack(g, game.Player1, game.Draw{Amount: game.Fixed(1), PlayerGroup: game.AllPlayersReference()}, nil)
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{a, a, a, a}, &TurnLog{})
	var order []game.PlayerID
	for _, request := range a.requests {
		if request.Kind == game.ChoiceMay {
			order = append(order, request.Player)
		}
	}
	if !slices.Equal(order, []game.PlayerID{game.Player3, game.Player4, game.Player1, game.Player2}) ||
		len(abundanceEvents(g, game.EventCardDrawn)) != 0 {
		t.Fatalf("replacement APNAP order = %v", order)
	}
	for player := game.Player1; player <= game.Player4; player++ {
		if g.Players[player].Hand.Size() != 1 {
			t.Fatal("one player's replacement consumed another player's pending draw")
		}
	}
}

type abundanceInformationChoices struct {
	abundanceChoices

	t *testing.T
}

func (a *abundanceInformationChoices) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if _, exposed := obs.LibraryTopRevealed(request.Player); exposed {
		a.t.Fatal("replacement choice exposed an otherwise hidden library")
	}
	if _, exposed := obs.LibraryTopLookable(request.Player); exposed {
		a.t.Fatal("replacement choice granted unrelated library inspection")
	}
	for _, option := range request.Options {
		if request.Kind != game.ChoiceOrder && option.Card.Exists {
			a.t.Fatal("optional/kind choice leaked library cards before reveal")
		}
		if request.Kind == game.ChoiceOrder && (!option.Card.Exists ||
			option.Card.Val.Name != "A" && option.Card.Val.Name != "B") {
			a.t.Fatal("bottom choice contained an unrevealed or matching card")
		}
	}
	return a.abundanceChoices.ChooseChoice(obs, request)
}

func TestAbundanceChoiceInformationRevealsOnlyTheExaminedPrefix(t *testing.T) {
	t.Parallel()
	g, e, _ := abundanceGame(t)
	cards := abundanceLibrary(g, "A", "B", "Land1", "Unseen")
	a := &abundanceInformationChoices{t: t}
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
	if !slices.Equal(abundanceEvents(g, game.EventCardRevealed), cards[:3]) || len(a.requests) != 3 {
		t.Fatal("Abundance revealed beyond the first match or omitted its choices")
	}
}
