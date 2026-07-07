package rules

import (
	"testing"

	cardk "github.com/natefinch/council4/mtg/cards/k"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ksfArtifactDef builds a plain artifact permanent with the given name and mana
// value, the candidate pool King Solomon's Frogs distributes over ("permanent
// that player controls with mana value 3 or greater").
func ksfArtifactDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    []types.Card{types.Artifact},
	}}
}

// castKingSolomonSFrogs stages King Solomon's Frogs (from the registered card
// definition) in Player1's hand with enough Plains to pay {3}{W}, casts it, and
// resolves it onto the battlefield so it enters "because it was cast" (the enter
// event carries EnterWasCast). It returns the resolved permanent.
func castKingSolomonSFrogs(t *testing.T, g *game.Game, engine *Engine) *game.Permanent {
	t.Helper()
	frogsID := addCardToHand(g, game.Player1, cardk.KingSolomonSFrogs)
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(frogsID, nil, 0, nil)) {
		t.Fatal("applyAction(cast King Solomon's Frogs) = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	frogs := permanentByCardID(g, frogsID)
	if frogs == nil {
		t.Fatal("King Solomon's Frogs did not resolve onto the battlefield")
	}
	return frogs
}

// TestKingSolomonSFrogsExilesPerOpponentAndEachDraws drives the full engine path
// for the enters trigger when the artifact was cast: each of the three opponents
// controls one mana-value-3 permanent, and resolving the trigger exiles one per
// opponent (a genuine per-opponent distribution, not a single global target) and
// makes each affected opponent draw exactly one card. The controller neither
// loses a permanent nor draws.
func TestKingSolomonSFrogsExilesPerOpponentAndEachDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	opponents := []game.PlayerID{game.Player2, game.Player3, game.Player4}
	exiled := make(map[game.PlayerID]*game.Permanent, len(opponents))
	drawn := make(map[game.PlayerID]id.ID, len(opponents))
	for _, opp := range opponents {
		exiled[opp] = addCombatPermanent(g, opp, ksfArtifactDef("Big Relic", 3))
		drawn[opp] = addCardToLibrary(g, opp, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	}
	// The controller's own mana-value-3 permanent must never be a candidate.
	mine := addCombatPermanent(g, game.Player1, ksfArtifactDef("My Relic", 3))

	frogs := castKingSolomonSFrogs(t, g, engine)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}, {0}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("King Solomon's Frogs enters trigger was not put on the stack after being cast")
	}
	if top, ok := g.Stack.Peek(); !ok || top.SourceCardID != frogs.CardInstanceID {
		t.Fatalf("top of stack = %+v, want the frogs enters trigger", top)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	for _, opp := range opponents {
		relic := exiled[opp]
		if permanentByCardID(g, relic.CardInstanceID) != nil {
			t.Fatalf("opponent %v's permanent was not exiled", opp)
		}
		if !g.Players[opp].Exile.Contains(relic.CardInstanceID) {
			t.Fatalf("opponent %v's permanent did not reach its owner's exile zone", opp)
		}
		if !g.Players[opp].Hand.Contains(drawn[opp]) {
			t.Fatalf("opponent %v did not draw a card for its exiled permanent", opp)
		}
		if got := g.Players[opp].Hand.Size(); got != 1 {
			t.Fatalf("opponent %v hand size = %d, want exactly 1 (one draw)", opp, got)
		}
	}
	if permanentByCardID(g, mine.CardInstanceID) == nil {
		t.Fatal("the controller's own permanent was exiled, but only opponents' permanents are candidates")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("controller hand size = %d, want 0 (the controller does not draw)", got)
	}
}

// TestKingSolomonSFrogsIneligibleOpponentContributesNothing confirms an opponent
// who controls no eligible permanent (mana value below 3) neither loses a
// permanent nor draws, while the other opponents still each lose one and draw.
func TestKingSolomonSFrogsIneligibleOpponentContributesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	eligible := addCombatPermanent(g, game.Player2, ksfArtifactDef("Big Relic", 3))
	eligibleDraw := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	otherEligible := addCombatPermanent(g, game.Player4, ksfArtifactDef("Big Relic", 4))
	otherEligibleDraw := addCardToLibrary(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	// Player3 controls only a mana-value-1 permanent, so it is never a candidate.
	ineligible := addCombatPermanent(g, game.Player3, ksfArtifactDef("Small Relic", 1))
	ineligibleDraw := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})

	castKingSolomonSFrogs(t, g, engine)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}, {0}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("King Solomon's Frogs enters trigger was not put on the stack after being cast")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if permanentByCardID(g, eligible.CardInstanceID) != nil || !g.Players[game.Player2].Hand.Contains(eligibleDraw) {
		t.Fatal("Player2's eligible permanent was not exiled and drawn for")
	}
	if permanentByCardID(g, otherEligible.CardInstanceID) != nil || !g.Players[game.Player4].Hand.Contains(otherEligibleDraw) {
		t.Fatal("Player4's eligible permanent was not exiled and drawn for")
	}
	if permanentByCardID(g, ineligible.CardInstanceID) == nil {
		t.Fatal("Player3's mana-value-1 permanent was exiled, but it is not an eligible candidate")
	}
	if g.Players[game.Player3].Hand.Contains(ineligibleDraw) || g.Players[game.Player3].Hand.Size() != 0 {
		t.Fatal("Player3 drew a card despite contributing no exiled permanent")
	}
}

// TestKingSolomonSFrogsCastGateSkipsWhenNotCast confirms the "if you cast it"
// intervening gate: when the artifact enters without being cast (put onto the
// battlefield), the enters trigger is never put on the stack, so no opponent
// loses a permanent or draws.
func TestKingSolomonSFrogsCastGateSkipsWhenNotCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	relic := addCombatPermanent(g, game.Player2, ksfArtifactDef("Big Relic", 3))
	reward := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	frogs := addCombatPermanent(g, game.Player1, cardk.KingSolomonSFrogs)
	// Put onto the battlefield: the enter event carries no cast information.
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		Player:      game.Player1,
		PermanentID: frogs.ObjectID,
		CardID:      frogs.CardInstanceID,
		ToZone:      zone.Battlefield,
	})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("enters trigger fired even though the artifact was not cast")
	}
	if permanentByCardID(g, relic.CardInstanceID) == nil {
		t.Fatal("an opponent's permanent was exiled without the cast gate being satisfied")
	}
	if g.Players[game.Player2].Hand.Contains(reward) {
		t.Fatal("an opponent drew a card without the cast gate being satisfied")
	}
}

// TestKingSolomonSFrogsBecomeMonarchAbility drives the {3}, {T}, Exile King
// Solomon's Frogs activated ability through the real activation path: paying the
// mana, tap, and exile-self additional cost, then resolving makes the controller
// the monarch and the artifact is exiled by the cost.
func TestKingSolomonSFrogsBecomeMonarchAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	frogs := addCombatPermanent(g, game.Player1, cardk.KingSolomonSFrogs)
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(frogs.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("the {3}, {T}, Exile-self become-monarch ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(become monarch ability) = false, want true")
	}
	if permanentByCardID(g, frogs.CardInstanceID) != nil {
		t.Fatal("King Solomon's Frogs was not exiled by its own activation cost")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("controller became the monarch before the ability resolved")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("controller did not become the monarch after the ability resolved")
	}
	if !g.Players[game.Player1].Exile.Contains(frogs.CardInstanceID) {
		t.Fatal("King Solomon's Frogs did not reach its owner's exile zone")
	}
}

// addKsfTokenPermanent puts a token artifact permanent with the given mana value
// onto the battlefield under owner's control. A token copy of a mana-value-3+
// permanent is a legal candidate for the distributive exile, and "For each
// permanent exiled this way, its controller draws a card" must still draw for it
// even though a token has CardInstanceID == 0.
func addKsfTokenPermanent(g *game.Game, owner game.PlayerID, name string, manaValue int) *game.Permanent {
	permanent := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      owner,
		Controller: owner,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:     name,
			ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
			Types:    []types.Card{types.Artifact},
		}},
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestKingSolomonSFrogsExiledTokenControllerDraws proves the draw payoff is
// text-faithful for tokens: exiling an opponent's mana-value-3 token permanent
// still makes that opponent draw a card. The link must preserve the token's
// ObjectID (permanentObjectBindingRef) rather than dropping it for having no card
// instance id, or the opponent would be denied their guaranteed draw.
func TestKingSolomonSFrogsExiledTokenControllerDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	token := addKsfTokenPermanent(g, game.Player2, "Relic Token", 3)
	reward := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})

	castKingSolomonSFrogs(t, g, engine)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("King Solomon's Frogs enters trigger was not put on the stack after being cast")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("the opponent's token permanent was not exiled")
	}
	if !g.Players[game.Player2].Hand.Contains(reward) {
		t.Fatal("the opponent did not draw a card for its exiled token (the draw payoff must be text-faithful for tokens)")
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("opponent hand size = %d, want exactly 1 (one draw for the exiled token)", got)
	}
}
