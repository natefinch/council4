package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: empower Jace is defined by CR 701.71a in the verified official
// CR effective 2026-09-25, read with tokens (111.1-111.4), entering loyalty
// (306.5b), loyalty abilities once per turn (306.5d, 606.3, 606.4, 606.6),
// putting counters (122.6), surveil looking privately at library cards
// (701.25a) and state-based actions only when a player would
// receive priority (704.3, 704.5i). Rulings for the Reality Fracture printings
// that empower Jace, all dated 2026-08-21: Theorist's Proxy
// 710302ca-c4be-4069-8ce1-f531414c74e9, Protege's Awakening
// a08c7ec2-4c6a-4db2-85a7-41afe8731523, Avatar of Burgeoning Echoes
// 5905995b-7a20-4602-a7cc-90aa5089a082 and Way of the Warlord
// 6d86e410-20c4-4248-96bf-5780ece6274a
// (https://api.scryfall.com/cards/<id>/rulings). They say an existing Jace
// planeswalker token must receive the counters instead of a new one being
// created, several tokens give the controller a choice, nontoken Jace
// planeswalkers don't count but a Jace planeswalker token obtained another way
// does, and only one loyalty ability per planeswalker may be activated each
// turn. These scenarios use synthetic empower spells; no card is admitted.

type empowerJacePolicy struct {
	simPassPolicy

	choices  [][]int
	requests []game.ChoiceRequest
}

func (p *empowerJacePolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.requests = append(p.requests, request)
	if len(p.choices) > 0 {
		choice := p.choices[0]
		p.choices = p.choices[1:]
		return choice
	}
	return request.DefaultSelection
}

func empowerJaceAgents(policy *empowerJacePolicy) [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	agents[game.Player1] = policy
	return agents
}

func resolveEmpowerJaceSpell(e *Engine, g *game.Game, amount int, agents [game.NumPlayers]PlayerAgent) {
	addEffectSpellToStack(g, game.Player1, game.EmpowerJace{Amount: game.Fixed(amount)}, nil)
	e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

func loyaltyAbilityIndex(t *testing.T, g *game.Game, permanent *game.Permanent, loyaltyCost int) int {
	t.Helper()
	for idx, ability := range permanentEffectiveAbilitiesView(g, permanent) {
		if body, ok := ability.(*game.LoyaltyAbility); ok && body.LoyaltyCost == loyaltyCost {
			return idx
		}
	}
	t.Fatalf("no loyalty ability costing %d on %s", loyaltyCost, permanentEffectiveName(g, permanent))
	return -1
}

func TestEmpowerJaceCreatesLoyalJaceTokenWithUsableAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom Card"}})
	topCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})
	policy := &empowerJacePolicy{}

	resolveEmpowerJaceSpell(e, g, 4, empowerJaceAgents(policy))
	e.applyStateBasedActions(g)

	tokens := controlledJacePlaneswalkerTokens(g, game.Player1)
	if len(tokens) != 1 {
		t.Fatalf("Jace planeswalker tokens = %d, want 1 created", len(tokens))
	}
	jace := tokens[0]
	if permanentEffectiveName(g, jace) != "Jace" || !slices.Equal(permanentEffectiveColors(g, jace), []color.Color{color.Blue}) {
		t.Fatalf("created token = %q %v, want blue Jace", permanentEffectiveName(g, jace), permanentEffectiveColors(g, jace))
	}
	if got := jace.Counters.Get(counter.Loyalty); got != 4 {
		t.Fatalf("loyalty = %d, want 4 put on the 0-loyalty token before state-based actions", got)
	}
	for _, request := range policy.requests {
		if request.Kind == game.ChoiceResolution {
			t.Fatalf("empower asked %q with a single Jace token", request.Prompt)
		}
	}

	setMainPhasePriority(g, game.Player1)
	surveil := loyaltyAbilityIndex(t, g, jace, -1)
	draw := loyaltyAbilityIndex(t, g, jace, -3)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(jace.ObjectID, surveil, nil, 0)) {
		t.Fatal("token's −1 ability could not be activated")
	}
	if got := jace.Counters.Get(counter.Loyalty); got != 3 {
		t.Fatalf("loyalty after −1 = %d, want 3", got)
	}
	if e.applyAction(g, game.Player1, action.ActivateAbility(jace.ObjectID, draw, nil, 0)) {
		t.Fatal("a second loyalty ability was activated in the same turn")
	}
	surveilLog := &TurnLog{}
	e.resolveTopOfStackWithChoices(g, empowerJaceAgents(policy), surveilLog)
	if !slices.ContainsFunc(policy.requests, func(request game.ChoiceRequest) bool {
		return request.Kind == game.ChoiceSurveil && request.Subject.Exists && request.Subject.Val.CardID == topCard
	}) {
		t.Fatal("−1 resolved without showing its controller the top card to surveil")
	}
	if !g.Players[game.Player1].Library.Contains(topCard) {
		t.Fatal("default surveil choice did not keep the top card")
	}
	for _, decision := range surveilLog.Choices {
		if decision.Request.Kind == game.ChoiceSurveil && (!decision.Request.Private || decision.Request.Subject.Exists) {
			t.Fatalf("public log reveals the surveilled library card: %+v", decision.Request)
		}
	}

	g.ActivatedAbilitiesThisTurn = nil
	handBefore := g.Players[game.Player1].Hand.Size()
	if !e.applyAction(g, game.Player1, action.ActivateAbility(jace.ObjectID, draw, nil, 0)) {
		t.Fatal("token's −3 ability could not be activated on a later turn")
	}
	e.resolveTopOfStackWithChoices(g, empowerJaceAgents(policy), &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand = %d after −3, want %d", got, handBefore+1)
	}
	e.applyStateBasedActions(g)
	if len(controlledJacePlaneswalkerTokens(g, game.Player1)) != 0 {
		t.Fatal("Jace token with 0 loyalty survived state-based actions")
	}
}

func TestEmpowerJaceGrowsExistingTokenAndIgnoresNontokenJace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	nontoken := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Jace, Printed Walker",
		Types:    []types.Card{types.Planeswalker},
		Subtypes: []types.Sub{types.Jace},
		Loyalty:  opt.Val(4),
	}})
	nontoken.Counters.Add(counter.Loyalty, 4)
	agents := empowerJaceAgents(&empowerJacePolicy{})

	resolveEmpowerJaceSpell(e, g, 2, agents)
	tokens := controlledJacePlaneswalkerTokens(g, game.Player1)
	if len(tokens) != 1 || tokens[0] == nontoken {
		t.Fatalf("a nontoken Jace planeswalker prevented token creation: %d tokens", len(tokens))
	}
	resolveEmpowerJaceSpell(e, g, 3, agents)

	if got := len(controlledJacePlaneswalkerTokens(g, game.Player1)); got != 1 {
		t.Fatalf("Jace planeswalker tokens = %d, want the existing token reused", got)
	}
	if got := tokens[0].Counters.Get(counter.Loyalty); got != 5 {
		t.Fatalf("token loyalty = %d, want 2+3", got)
	}
	if got := nontoken.Counters.Get(counter.Loyalty); got != 4 {
		t.Fatalf("nontoken Jace loyalty = %d, want untouched 4", got)
	}
}

func TestEmpowerJaceControllerChoosesAmongJaceTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	agents := empowerJaceAgents(&empowerJacePolicy{})
	resolveEmpowerJaceSpell(e, g, 1, agents)
	empowered := g.Battlefield[len(g.Battlefield)-1]
	resolveInstruction(e, g, &game.StackObject{Controller: game.Player2}, game.EmpowerJace{Amount: game.Fixed(1)}, &TurnLog{})
	if len(controlledJacePlaneswalkerTokens(g, game.Player2)) != 1 {
		t.Fatal("another player's Jace token stopped the opponent's empower from creating one")
	}
	resolveInstruction(e, g, &game.StackObject{Controller: game.Player1}, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenDef(&game.CardDef{CardFace: game.CardFace{
			Name:     "Jace, Copied Walker",
			Types:    []types.Card{types.Planeswalker},
			Subtypes: []types.Sub{types.Jace},
			Loyalty:  opt.Val(2),
		}}),
	}, &TurnLog{})
	tokens := controlledJacePlaneswalkerTokens(g, game.Player1)
	if len(tokens) != 2 || tokens[0] != empowered {
		t.Fatalf("Player 1 Jace planeswalker tokens = %d, want the empower token and the copied token", len(tokens))
	}
	copied := tokens[1]

	policy := &empowerJacePolicy{choices: [][]int{{1}}}
	resolveEmpowerJaceSpell(e, g, 3, empowerJaceAgents(policy))

	var choice *game.ChoiceRequest
	for i := range policy.requests {
		if policy.requests[i].Kind == game.ChoiceResolution {
			choice = &policy.requests[i]
		}
	}
	if choice == nil || choice.Player != game.Player1 || len(choice.Options) != 2 {
		t.Fatalf("empower choice = %+v, want Player 1 choosing between two tokens", choice)
	}
	if got := copied.Counters.Get(counter.Loyalty); got != 5 {
		t.Fatalf("chosen copied token loyalty = %d, want 2+3", got)
	}
	if got := empowered.Counters.Get(counter.Loyalty); got != 1 {
		t.Fatalf("unchosen token loyalty = %d, want 1", got)
	}
	if got := len(controlledJacePlaneswalkerTokens(g, game.Player1)); got != 2 {
		t.Fatalf("Jace planeswalker tokens = %d, want no new token", got)
	}
}

func TestEmpowerJaceChoosesAmongReplacementCreatedTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, cards.NewDefaultRegistry().Lookup("Doubling Season"))
	policy := &empowerJacePolicy{choices: [][]int{{1}}}

	resolveEmpowerJaceSpell(e, g, 3, empowerJaceAgents(policy))

	tokens := controlledJacePlaneswalkerTokens(g, game.Player1)
	if len(tokens) != 2 {
		t.Fatalf("Jace planeswalker tokens = %d, want 2 from Doubling Season", len(tokens))
	}
	if tokens[0].Counters.Get(counter.Loyalty) != 0 || tokens[1].Counters.Get(counter.Loyalty) != 6 {
		t.Fatalf("loyalty = %d/%d, want the chosen token doubled to 6",
			tokens[0].Counters.Get(counter.Loyalty), tokens[1].Counters.Get(counter.Loyalty))
	}
	e.applyStateBasedActions(g)
	if remaining := controlledJacePlaneswalkerTokens(g, game.Player1); len(remaining) != 1 || remaining[0] != tokens[1] {
		t.Fatal("state-based actions did not leave only the empowered Jace token")
	}
}
