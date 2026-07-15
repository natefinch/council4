package rules

import (
	"math/rand/v2"
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// openingHandBattlefieldCard builds a minimal test card carrying only the
// pregame "begin the game on the battlefield" permission (CR 103.6a), distinct
// from any enters-the-battlefield replacement.
func openingHandBattlefieldCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{
			{BeginsGameOnBattlefield: true},
		},
	}}
}

// vanillaHandCard builds a plain creature with no pregame permission, used to
// prove ineligible cards are never offered or moved.
func vanillaHandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
	}}
}

// actionOnlyAgent is a PlayerAgent that does not implement ChoiceAgent, so the
// engine must fall back to the deterministic decline for opening-hand actions.
type actionOnlyAgent struct{}

func (actionOnlyAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

// orderRecordingAgent records, into a shared slice, the order in which players
// are consulted and the subject card each request carried, then answers with a
// scripted selection. Sharing one order slice across per-player agents lets a
// test assert the CR 103.6 processing order and that no request leaks another
// player's card.
type orderRecordingAgent struct {
	answer  []int
	order   *[]game.PlayerID
	seen    []id.ID
	players []game.PlayerID
}

func (*orderRecordingAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

func (a *orderRecordingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if a.order != nil {
		*a.order = append(*a.order, request.Player)
	}
	if request.Subject.Exists {
		a.seen = append(a.seen, request.Subject.Val.CardID)
	}
	a.players = append(a.players, request.Player)
	return a.answer
}

// cardRemovingAgent removes a card from a hand while answering, simulating a
// card that disappears between the engine offering it and resolving the move.
type cardRemovingAgent struct {
	g      *game.Game
	player game.PlayerID
	cardID id.ID
}

func (*cardRemovingAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

func (a *cardRemovingAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	a.g.Players[a.player].Hand.Remove(a.cardID)
	return []int{1}
}

// goldfishAcceptAgent plays the first legal action and accepts every
// opening-hand action, for end-to-end runs through RunGoldfish.
type goldfishAcceptAgent struct{}

func (goldfishAcceptAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

func (goldfishAcceptAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	return []int{1}
}

func agentsAll(agent PlayerAgent) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for i := range agents {
		agents[i] = agent
	}
	return agents
}

func TestOpeningHandBattlefieldAcceptPutsCardOnBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Test Leyline"))

	engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}))

	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("accepted card stayed in hand")
	}
	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatal("accepted card did not enter the battlefield")
	}
	if permanent.Owner != game.Player1 || permanent.Controller != game.Player1 {
		t.Fatalf("permanent owner/controller = %v/%v, want Player1/Player1", permanent.Owner, permanent.Controller)
	}
	var entered bool
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.CardID == cardID {
			entered = true
		}
	}
	if !entered {
		t.Fatal("no enters-the-battlefield event emitted for the accepted card")
	}
}

func TestOpeningHandBattlefieldDeclineKeepsCardInHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Test Leyline"))

	engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{0}}))

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("declined card left the hand")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield size = %d, want 0", len(g.Battlefield))
	}
}

func TestOpeningHandBattlefieldFallsBackToDeclineWithoutValidAnswer(t *testing.T) {
	for name, agent := range map[string]PlayerAgent{
		"no agent":         nil,
		"non-choice agent": actionOnlyAgent{},
		"invalid answer":   &recordingChoiceAgent{answer: []int{7}},
		"nil answer":       &recordingChoiceAgent{answer: nil},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Test Leyline"))

			engine.performOpeningHandBattlefieldActions(g, agentsAll(agent))

			if !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("card left the hand despite no valid acceptance")
			}
			if len(g.Battlefield) != 0 {
				t.Fatalf("battlefield size = %d, want 0", len(g.Battlefield))
			}
		})
	}
}

func TestOpeningHandBattlefieldMultipleEligibleCardsAndIneligibleUntouched(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Leyline A"))
	second := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Leyline B"))
	vanilla := addCardToHand(g, game.Player1, vanillaHandCard("Bear"))

	agent := &recordingChoiceAgent{answer: []int{1}}
	engine.performOpeningHandBattlefieldActions(g, agentsAll(agent))

	if permanentForCard(g, first) == nil {
		t.Fatal("first eligible card did not enter the battlefield")
	}
	if permanentForCard(g, second) == nil {
		t.Fatal("second eligible card did not enter the battlefield")
	}
	if !g.Players[game.Player1].Hand.Contains(vanilla) {
		t.Fatal("ineligible card left the hand")
	}
	if len(agent.requests) != 2 {
		t.Fatalf("agent consulted %d times, want 2 (only eligible cards)", len(agent.requests))
	}
	for _, request := range agent.requests {
		if request.Subject.Exists && request.Subject.Val.CardID == vanilla {
			t.Fatal("ineligible card was offered")
		}
	}
}

func TestOpeningHandBattlefieldDuplicatesEachConsidered(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := openingHandBattlefieldCard("Duplicated Leyline")
	first := addCardToHand(g, game.Player1, def)
	second := addCardToHand(g, game.Player1, def)

	engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}))

	if len(g.Battlefield) != 2 {
		t.Fatalf("battlefield size = %d, want 2", len(g.Battlefield))
	}
	firstPerm := permanentForCard(g, first)
	if firstPerm == nil {
		t.Fatal("first duplicate did not enter")
	}
	secondPerm := permanentForCard(g, second)
	if secondPerm == nil {
		t.Fatal("second duplicate did not enter")
	}
	if firstPerm.ObjectID == secondPerm.ObjectID {
		t.Fatal("duplicates share an object id")
	}
}

func TestOpeningHandBattlefieldMultiplePlayersStartingPlayerOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	var order []game.PlayerID
	var agents [game.NumPlayers]PlayerAgent
	cardByPlayer := map[game.PlayerID]id.ID{}
	for pid := range game.NumPlayers {
		player := game.PlayerID(pid)
		cardByPlayer[player] = addCardToHand(g, player, openingHandBattlefieldCard("Leyline"))
		agents[player] = &orderRecordingAgent{answer: []int{1}, order: &order}
	}

	engine.performOpeningHandBattlefieldActions(g, agents)

	want := []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4}
	if len(order) != len(want) {
		t.Fatalf("consulted %d players, want %d", len(order), len(want))
	}
	for i, player := range want {
		if order[i] != player {
			t.Fatalf("order[%d] = %v, want %v (starting player then turn order)", i, order[i], player)
		}
	}
	for player, card := range cardByPlayer {
		if permanentForCard(g, card) == nil {
			t.Fatalf("player %v card did not enter the battlefield", player)
		}
	}
}

func TestOpeningHandBattlefieldDoesNotLeakOpponentCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	var agents [game.NumPlayers]PlayerAgent
	cardByPlayer := map[game.PlayerID]id.ID{}
	byPlayer := map[game.PlayerID]*orderRecordingAgent{}
	for pid := range game.NumPlayers {
		player := game.PlayerID(pid)
		cardByPlayer[player] = addCardToHand(g, player, openingHandBattlefieldCard("Leyline"))
		agent := &orderRecordingAgent{answer: []int{0}}
		agents[player] = agent
		byPlayer[player] = agent
	}

	engine.performOpeningHandBattlefieldActions(g, agents)

	for player, agent := range byPlayer {
		if len(agent.players) != 1 || agent.players[0] != player {
			t.Fatalf("player %v agent consulted for %v, want only itself", player, agent.players)
		}
		if len(agent.seen) != 1 || agent.seen[0] != cardByPlayer[player] {
			t.Fatalf("player %v agent saw subjects %v, want only its own card", player, agent.seen)
		}
	}
	// Declined cards are resolved with a nil log and emit no events, so nothing
	// about a hidden hand card is recorded in the public event stream.
	if len(g.Events) != 0 {
		t.Fatalf("declined opening-hand actions emitted %d events, want 0", len(g.Events))
	}
}

func TestConsiderOpeningHandCardNotInHandIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Leyline"))
	g.Players[game.Player1].Hand.Remove(cardID)

	agent := &recordingChoiceAgent{answer: []int{1}}
	engine.considerOpeningHandBattlefieldCard(g, agentsAll(agent), game.Player1, cardID)

	if len(agent.requests) != 0 {
		t.Fatalf("agent consulted %d times for a card not in hand, want 0", len(agent.requests))
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield size = %d, want 0", len(g.Battlefield))
	}
}

func TestConsiderOpeningHandCardDisappearingDuringChoiceNoPartialMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Leyline"))

	agent := &cardRemovingAgent{g: g, player: game.Player1, cardID: cardID}
	engine.considerOpeningHandBattlefieldCard(g, agentsAll(agent), game.Player1, cardID)

	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield size = %d, want 0 (card left hand mid-decision)", len(g.Battlefield))
	}
	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("engine re-added a card it never removed")
	}
}

func TestOpeningHandBattlefieldAppliesEnterTappedReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := openingHandBattlefieldCard("Tapped Leyline")
	def.ReplacementAbilities = []game.ReplacementAbility{game.EntersTappedReplacement("This permanent enters tapped.")}
	cardID := addCardToHand(g, game.Player1, def)

	engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}))

	permanent := permanentForCard(g, cardID)
	if permanent == nil {
		t.Fatal("card did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("enters-tapped replacement did not apply to a pregame battlefield entry")
	}
}

func TestOpeningHandBattlefieldEnterDoesNotTriggerBeforeFirstTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := openingHandBattlefieldCard("Triggering Leyline")
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhen,
			Pattern: game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}}
	agents := agentsAll(&recordingChoiceAgent{answer: []int{1}})

	// Mirror the engine order: opening-hand actions run before the trigger
	// cursor advances at the start of the first turn.
	cardID := addCardToHand(g, game.Player1, def)
	engine.performOpeningHandBattlefieldActions(g, agents)
	if permanentForCard(g, cardID) == nil {
		t.Fatal("card did not enter the battlefield")
	}
	markCurrentTurnEventStart(g)

	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("a pregame battlefield entry put a triggered ability on the stack")
	}
	if !g.Stack.IsEmpty() {
		t.Fatalf("stack size = %d after pregame entry, want 0", g.Stack.Size())
	}

	// Contrast: the same card's enter trigger does fire for an in-game entry
	// after the cursor, proving the pregame suppression is due to event ordering
	// rather than an unmatched trigger.
	inGame := addCardToHand(g, game.Player1, def)
	card, _ := g.GetCardInstance(inGame)
	if _, ok := createCardPermanentFaceWithChoices(engine, g, card, game.Player1, zone.Hand, game.FaceFront, agents, nil); !ok {
		t.Fatal("in-game entry failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("in-game entry did not trigger the enter ability")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d after in-game entry, want 1", g.Stack.Size())
	}
}

func TestPerformPlayerOpeningHandBattlefieldEliminatedPlayerNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, openingHandBattlefieldCard("Leyline"))
	g.Players[game.Player1].Eliminated = true

	engine.performPlayerOpeningHandBattlefield(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}), game.Player1)

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("eliminated player's card was moved")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield size = %d, want 0", len(g.Battlefield))
	}
}

func TestOpeningHandActionOrderSkipsEliminatedSeats(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.TurnOrder.Eliminate(game.Player2)

	order := openingHandActionOrder(g)

	want := []game.PlayerID{game.Player1, game.Player3, game.Player4}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i, player := range want {
		if order[i] != player {
			t.Fatalf("order[%d] = %v, want %v", i, order[i], player)
		}
	}
}

func TestRunGoldfishOpeningHandBattlefieldAcceptsRealLeyline(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: repeatedCard(cards.LeylineOfSanctity(), 99)}

	engine := NewEngine(rand.New(rand.NewPCG(7, 11)))
	g := engine.NewGoldfishGame(config)
	result := engine.RunGoldfish(g, goldfishAcceptAgent{}, 1)

	if len(result.OpeningHand) != openingHandSize {
		t.Fatalf("opening hand size = %d, want %d", len(result.OpeningHand), openingHandSize)
	}
	leylines := 0
	for _, permanent := range g.Battlefield {
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if !ok || card.Def.Name != "Leyline of Sanctity" {
			continue
		}
		leylines++
		if permanent.Controller != game.Player1 {
			t.Fatalf("permanent controller = %v, want Player1", permanent.Controller)
		}
	}
	if leylines != openingHandSize {
		t.Fatalf("Leyline of Sanctity permanents = %d, want %d accepted from opening hand", leylines, openingHandSize)
	}
}

func TestRunGoldfishOpeningHandBattlefieldDeclinesWithoutChoiceAgent(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: repeatedCard(cards.LeylineOfSanctity(), 99)}

	engine := NewEngine(rand.New(rand.NewPCG(7, 11)))
	g := engine.NewGoldfishGame(config)
	engine.RunGoldfish(g, goldfishTestAgent{}, 1)

	for _, permanent := range g.Battlefield {
		card, ok := g.GetCardInstance(permanent.CardInstanceID)
		if ok && card.Def.Name == "Leyline of Sanctity" {
			t.Fatal("a Leyline entered the battlefield without a ChoiceAgent to accept it")
		}
	}
}
