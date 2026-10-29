package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Rebound casts follow the alternative-cost rules (CR 702.88b): "without paying
// its mana cost" replaces only the mana cost, while additional costs, cost
// increases and reductions still apply (CR 118.9d, 601.2f-h). These scenarios
// cover Engine gap #3458 on the rebound route.

// castIntoReboundExile casts cardID from Player1's hand, paying from the pool
// already supplied, and resolves it into rebound exile.
func castIntoReboundExile(t *testing.T, e *Engine, g *game.Game, cardID id.ID) {
	t.Helper()
	if !e.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting the rebound spell from hand failed")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rebound spell was not exiled on resolution")
	}
	if _, ok := g.ReboundCards[cardID]; !ok {
		t.Fatal("rebound spell was not tracked for the next upkeep")
	}
}

func reboundCastEvents(g *game.Game, cardID id.ID) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID && event.FromZone == zone.Exile {
			events = append(events, event)
		}
	}
	return events
}

// requireReboundNotCast asserts the upkeep consumed the rebound permission
// without casting the card, leaving it in exile.
func requireReboundNotCast(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if got := reboundCastEvents(g, cardID); len(got) != 0 {
		t.Fatalf("rebound cast events = %d, want none", len(got))
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("uncast rebound card left exile")
	}
	if _, ok := g.ReboundCards[cardID]; ok {
		t.Fatal("rebound permission survived the upkeep")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want empty after an unpaid rebound cast", g.Stack.Size())
	}
}

// requireReboundCast asserts the card was cast from exile, returning its stack
// object and the cast event.
func requireReboundCast(t *testing.T, g *game.Game, cardID id.ID) (*game.StackObject, game.Event) {
	t.Helper()
	events := reboundCastEvents(g, cardID)
	if len(events) != 1 {
		t.Fatalf("rebound cast events = %d, want 1", len(events))
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rebound card is still in exile after being cast")
	}
	for _, obj := range g.Stack.Objects() {
		if !obj.Copy && obj.SourceID == cardID {
			if obj.SourceZone != zone.Exile {
				t.Fatalf("rebound spell source zone = %v, want exile", obj.SourceZone)
			}
			return obj, events[0]
		}
	}
	t.Fatal("rebound spell is not on the stack")
	return nil, game.Event{}
}

func reboundTestSpell(name string, extra ...game.StaticAbility) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Instant},
		ManaCost:        opt.Val(cost.Mana{cost.G}),
		StaticAbilities: append([]game.StaticAbility{game.ReboundStaticBody}, extra...),
		SpellAbility:    opt.Val(game.AbilityContent{}),
	}}
}

func reboundTestArtifact(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Rebound Fodder",
		Types: []types.Card{types.Artifact},
	}})
}

// TestReboundCastPaysThaliaTax uses the default-registry Faithless Salvaging
// and Thalia, Guardian of Thraben: the rebound cast must pay Thalia's {1} tax,
// so it cannot be cast without that mana, and spends it when cast.
func TestReboundCastPaysThaliaTax(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name      string
		thalia    bool
		upkeepC   int
		wantCast  bool
		wantSpent int
	}{
		{name: "no tax control", upkeepC: 0, wantCast: true, wantSpent: 0},
		{name: "unpayable tax", thalia: true, upkeepC: 0, wantCast: false},
		{name: "payable tax", thalia: true, upkeepC: 1, wantCast: true, wantSpent: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Faithless Salvaging"))
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			castIntoReboundExile(t, e, g, cardID)
			if tc.thalia {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)

			e.processReboundUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if !tc.wantCast {
				requireReboundNotCast(t, g, cardID)
				if got := g.Players[game.Player1].ManaPool.Total(); got != tc.upkeepC {
					t.Fatalf("mana pool = %d, want %d untouched", got, tc.upkeepC)
				}
				return
			}
			obj, event := requireReboundCast(t, g, cardID)
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana pool = %d after rebound cast, want Thalia tax spent", got)
			}
			if obj.ManaSpentToCast != tc.wantSpent || !event.ManaSpentToCast.Exists || event.ManaSpentToCast.Val != tc.wantSpent {
				t.Fatalf("mana spent obj=%d event=%+v, want %d", obj.ManaSpentToCast, event.ManaSpentToCast, tc.wantSpent)
			}
		})
	}
}

// TestReboundCastPaysAdditionalCosts proves a mandatory additional cost
// (sacrifice an artifact) is still paid on the free rebound cast, alone and
// combined with Thalia's tax, and that an unpaid total rolls back without
// sacrificing anything.
func TestReboundCastPaysAdditionalCosts(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name     string
		artifact bool
		thalia   bool
		upkeepC  int
		wantCast bool
	}{
		{name: "artifact sacrificed", artifact: true, wantCast: true},
		{name: "no artifact", wantCast: false},
		{name: "artifact and tax paid", artifact: true, thalia: true, upkeepC: 1, wantCast: true},
		{name: "artifact but tax unpayable", artifact: true, thalia: true, wantCast: false},
		{name: "tax but no artifact", thalia: true, upkeepC: 1, wantCast: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := reboundTestSpell("Rebound Sacrifice Spell")
			def.AdditionalCosts = []cost.Additional{{
				Kind:               cost.AdditionalSacrifice,
				Text:               "Sacrifice an artifact",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
			}}
			cardID := addCardToHand(g, game.Player1, def)
			reboundTestArtifact(g)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			castIntoReboundExile(t, e, g, cardID)
			var fodder *game.Permanent
			if tc.artifact {
				fodder = reboundTestArtifact(g)
			}
			if tc.thalia {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)

			e.processReboundUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if !tc.wantCast {
				requireReboundNotCast(t, g, cardID)
				if fodder != nil {
					if _, onBattlefield := permanentByObjectID(g, fodder.ObjectID); !onBattlefield {
						t.Fatal("rolled-back rebound cast sacrificed the artifact")
					}
				}
				return
			}
			obj, _ := requireReboundCast(t, g, cardID)
			if len(obj.SacrificedAsCostIDs) != 1 || !g.Players[game.Player1].Graveyard.Contains(fodder.CardInstanceID) {
				t.Fatalf("sacrificed as cost = %v, want the artifact sacrificed", obj.SacrificedAsCostIDs)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana pool = %d, want Thalia tax spent", got)
			}
		})
	}
}

// reboundKickerAgent answers the rebound free-cast choice by preferring a kicked
// option when one is offered, recording each request.
type reboundKickerAgent struct {
	requests []game.ChoiceRequest
}

func (*reboundKickerAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *reboundKickerAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceResolution || !strings.Contains(request.Prompt, "without paying its mana cost") {
		return request.DefaultSelection
	}
	a.requests = append(a.requests, request)
	for _, option := range request.Options {
		if strings.Contains(option.Label, "with kicker") {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

func reboundKickerAgents(agent PlayerAgent) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	return agents
}

// TestReboundCastOptionalCostsAndStormCopies combines Thalia's tax with an
// optional kicker cost on the rebound cast: the kicked option is offered only
// when the combined mana is affordable, and storm copies the rebound cast with
// its paid kicker (CR 707.10).
func TestReboundCastOptionalCostsAndStormCopies(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name       string
		upkeepC    int
		wantKicked bool
	}{
		{name: "kicker and tax affordable", upkeepC: 2, wantKicked: true},
		{name: "only tax affordable", upkeepC: 1, wantKicked: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := reboundTestSpell("Rebound Storm Kicker Spell", game.StaticAbility{KeywordAbilities: []game.KeywordAbility{
				game.KickerKeyword{Cost: cost.Mana{cost.O(1)}},
				game.SimpleKeyword{Kind: game.Storm},
			}})
			cardID := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			castIntoReboundExile(t, e, g, cardID)
			addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)
			agent := &reboundKickerAgent{}

			e.processReboundUpkeep(g, game.Player1, reboundKickerAgents(agent), &TurnLog{})

			if len(agent.requests) != 1 {
				t.Fatalf("free-cast requests = %d, want 1", len(agent.requests))
			}
			kickedOffered := false
			for _, option := range agent.requests[0].Options {
				kickedOffered = kickedOffered || strings.Contains(option.Label, "with kicker")
			}
			if kickedOffered != tc.wantKicked {
				t.Fatalf("kicked option offered = %t, want %t", kickedOffered, tc.wantKicked)
			}
			obj, _ := requireReboundCast(t, g, cardID)
			if obj.KickerPaid != tc.wantKicked || obj.ManaSpentToCast != tc.upkeepC {
				t.Fatalf("kicker paid=%t mana spent=%d, want %t/%d", obj.KickerPaid, obj.ManaSpentToCast, tc.wantKicked, tc.upkeepC)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana pool = %d, want all spent", got)
			}
			resolveStormTrigger(t, e, g, nil)
			copies := 0
			for _, stackObj := range g.Stack.Objects() {
				if stackObj.Copy && stackObj.SourceID == cardID {
					copies++
					if stackObj.KickerPaid != tc.wantKicked {
						t.Fatalf("storm copy kicker paid = %t, want %t", stackObj.KickerPaid, tc.wantKicked)
					}
				}
			}
			if copies != 1 {
				t.Fatalf("storm copies = %d, want 1 for the earlier hand cast", copies)
			}
		})
	}
}

// TestReboundCastMayBeDeclined keeps rebound optional: declining casts nothing,
// spends nothing and leaves the card in exile.
func TestReboundCastMayBeDeclined(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, reboundTestSpell("Rebound Decline Spell"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	castIntoReboundExile(t, e, g, cardID)
	agent := &cascadeCastChoiceAgent{decline: true}

	e.processReboundUpkeep(g, game.Player1, reboundKickerAgents(agent), &TurnLog{})

	if len(agent.requests) != 1 || declineOptionCount(agent.requests[0]) != 1 {
		t.Fatalf("free-cast requests = %+v, want one request offering a decline", agent.requests)
	}
	requireReboundNotCast(t, g, cardID)
}

// TestReboundCastCostsSurviveClone proves the rebound permission and its cost
// planning are clone-safe: a cloned game pays Thalia's tax independently while
// the original keeps its pending rebound and mana.
func TestReboundCastCostsSurviveClone(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, reg.Lookup("Faithless Salvaging"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castIntoReboundExile(t, e, g, cardID)
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	clone := g.Clone()
	e.processReboundUpkeep(clone, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	obj, _ := requireReboundCast(t, clone, cardID)
	if obj.ManaSpentToCast != 1 || clone.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("clone rebound mana spent = %d pool = %d, want Thalia tax paid", obj.ManaSpentToCast, clone.Players[game.Player1].ManaPool.Total())
	}
	if _, ok := g.ReboundCards[cardID]; !ok || !g.Players[game.Player1].Exile.Contains(cardID) ||
		g.Players[game.Player1].ManaPool.Total() != 1 || g.Stack.Size() != 0 {
		t.Fatal("casting the rebound in the clone changed the original game")
	}
}

// paymentManaAgent activates, once, the payment-time mana ability whose label
// contains ability, and otherwise takes each default choice.
type paymentManaAgent struct {
	ability   string
	windows   int
	activated int
}

func (*paymentManaAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *paymentManaAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt != "Activate a mana ability while paying?" {
		return request.DefaultSelection
	}
	a.windows++
	if a.activated == 0 {
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.ability) {
				a.activated++
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func setReboundUpkeep(g *game.Game) {
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseBeginning, game.StepUpkeep
}

func reboundPermanentNamed(t *testing.T, g *game.Game, name string) *game.Permanent {
	t.Helper()
	for _, permanent := range g.Battlefield {
		if card, ok := g.GetCardInstance(permanent.CardInstanceID); ok && card.Def.Name == name {
			return permanent
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return nil
}

// TestReboundCastPaysWithPaymentTimeManaAbility uses Waterlogged Grove, whose
// only mana ability costs life and is never activated automatically, as the
// sole source for Thalia's tax: the free rebound proposal is offered and the
// caster activates the ability while paying (CR 601.2g-h). The cast also runs
// in a clone without changing the original game.
func TestReboundCastPaysWithPaymentTimeManaAbility(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, reg.Lookup("Faithless Salvaging"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castIntoReboundExile(t, e, g, cardID)
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	setReboundUpkeep(g)

	clone := g.Clone()
	agent := &paymentManaAgent{ability: "Pay 1 life"}
	e.processReboundUpkeep(clone, game.Player1, reboundKickerAgents(agent), &TurnLog{})

	obj, event := requireReboundCast(t, clone, cardID)
	grove := reboundPermanentNamed(t, clone, "Waterlogged Grove")
	if agent.activated != 1 || !grove.Tapped || clone.Players[game.Player1].Life != g.Players[game.Player1].Life-1 {
		t.Fatalf("activated=%d grove tapped=%t life=%d, want the Grove paid for Thalia's tax", agent.activated, grove.Tapped, clone.Players[game.Player1].Life)
	}
	if obj.ManaSpentToCast != 1 || event.ManaSpentToCast.Val != 1 || clone.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana spent obj=%d event=%+v pool=%d, want the Grove's mana spent on the tax", obj.ManaSpentToCast, event.ManaSpentToCast, clone.Players[game.Player1].ManaPool.Total())
	}
	if _, ok := g.ReboundCards[cardID]; !ok || !g.Players[game.Player1].Exile.Contains(cardID) ||
		reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || g.Stack.Size() != 0 {
		t.Fatal("casting the rebound in the clone changed the original game")
	}
}

// TestReboundCastLocksCostBeforeManaAbilities proves the total cost is locked
// before payment-time mana abilities (CR 601.2f-g): sacrificing the caster's
// own Thalia to Ashnod's Altar while paying does not remove the {1} tax already
// included in the total.
func TestReboundCastLocksCostBeforeManaAbilities(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addCardToHand(g, game.Player1, reg.Lookup("Faithless Salvaging"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castIntoReboundExile(t, e, g, cardID)
	thalia := addCombatPermanent(g, game.Player1, reg.Lookup("Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, reg.Lookup("Ashnod's Altar"))
	setReboundUpkeep(g)
	agent := &paymentManaAgent{ability: "Sacrifice a creature"}

	e.processReboundUpkeep(g, game.Player1, reboundKickerAgents(agent), &TurnLog{})

	obj, _ := requireReboundCast(t, g, cardID)
	if _, onBattlefield := permanentByObjectID(g, thalia.ObjectID); agent.activated != 1 || onBattlefield {
		t.Fatalf("activated=%d thalia on battlefield=%t, want Thalia sacrificed for mana", agent.activated, onBattlefield)
	}
	if obj.ManaSpentToCast != 1 || g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatalf("mana spent=%d pool=%d, want the locked {1} tax paid and {C} left", obj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total())
	}
}

// TestReboundCastRollsBackPaymentTimeManaAbilities combines a payment-time
// mana ability with Thalia's tax and a mandatory sacrifice cost. Mana
// abilities only add mana, so without an artifact the proposal is not offered
// and no payment window opens. With a second Thalia the {2} tax exceeds the
// Grove's mana: the proposal is offered, but the unpaid cast is reversed along
// with the reversible mana activation (CR 733), leaving the card in exile.
func TestReboundCastRollsBackPaymentTimeManaAbilities(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name          string
		artifact      bool
		thalias       int
		wantWindows   int
		wantActivated int
		wantCast      bool
	}{
		{name: "all costs paid", artifact: true, thalias: 1, wantWindows: 1, wantActivated: 1, wantCast: true},
		{name: "sacrifice impossible", artifact: false, thalias: 1},
		{name: "mana short", artifact: true, thalias: 2, wantWindows: 1, wantActivated: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := reboundTestSpell("Rebound Grove Sacrifice Spell")
			def.AdditionalCosts = []cost.Additional{{
				Kind:               cost.AdditionalSacrifice,
				Text:               "Sacrifice an artifact",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
			}}
			cardID := addCardToHand(g, game.Player1, def)
			reboundTestArtifact(g)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			castIntoReboundExile(t, e, g, cardID)
			var fodder *game.Permanent
			if tc.artifact {
				fodder = reboundTestArtifact(g)
			}
			for range tc.thalias {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			setReboundUpkeep(g)
			life := g.Players[game.Player1].Life
			agent := &paymentManaAgent{ability: "Pay 1 life"}

			e.processReboundUpkeep(g, game.Player1, reboundKickerAgents(agent), &TurnLog{})

			if agent.windows != tc.wantWindows || agent.activated != tc.wantActivated {
				t.Fatalf("payment windows=%d activations=%d, want %d/%d", agent.windows, agent.activated, tc.wantWindows, tc.wantActivated)
			}
			if !tc.wantCast {
				requireReboundNotCast(t, g, cardID)
				if grove.Tapped || g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatalf("grove tapped=%t life=%d pool=%d, want no mana activation left behind", grove.Tapped, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
				}
				if fodder != nil {
					if _, ok := permanentByObjectID(g, fodder.ObjectID); !ok {
						t.Fatal("the reversed proposal sacrificed the artifact")
					}
				}
				return
			}
			obj, _ := requireReboundCast(t, g, cardID)
			if len(obj.SacrificedAsCostIDs) != 1 || !g.Players[game.Player1].Graveyard.Contains(fodder.CardInstanceID) {
				t.Fatalf("sacrificed as cost = %v, want the artifact sacrificed", obj.SacrificedAsCostIDs)
			}
			if !grove.Tapped || g.Players[game.Player1].Life != life-1 || obj.ManaSpentToCast != 1 {
				t.Fatalf("grove tapped=%t life=%d mana spent=%d, want the Grove to pay Thalia's tax", grove.Tapped, g.Players[game.Player1].Life, obj.ManaSpentToCast)
			}
		})
	}
}

// TestFreeCastRoutesPayWithPaymentTimeManaAbility covers the other routes
// through the shared free-cast payment: casting a card from hand without paying
// its mana cost, and casting a copy of an exiled card (CR 707.12). Each pays
// Thalia's tax from a mana ability activated while paying.
func TestFreeCastRoutesPayWithPaymentTimeManaAbility(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	spell := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:         "Free Route Spell",
			Types:        []types.Card{types.Instant},
			ManaCost:     opt.Val(cost.Mana{cost.G}),
			SpellAbility: opt.Val(game.AbilityContent{}),
		}}
	}
	for _, tc := range []struct {
		name string
		cast func(*Engine, *game.Game, [game.NumPlayers]PlayerAgent) (id.ID, bool)
		copy bool
	}{
		{name: "from hand", cast: func(e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) (id.ID, bool) {
			cardID := addCardToHand(g, game.Player1, spell())
			return cardID, e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
				game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{})
		}},
		{name: "copy of exiled card", copy: true, cast: func(e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) (id.ID, bool) {
			cardID := addCardToExile(g, game.Player1, spell())
			return cardID, e.castFreeCopyOfCard(g, game.Player1, cardID, agents, &TurnLog{})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			life := g.Players[game.Player1].Life
			agent := &paymentManaAgent{ability: "Pay 1 life"}

			cardID, ok := tc.cast(e, g, reboundKickerAgents(agent))

			if !ok || agent.activated != 1 || !grove.Tapped || g.Players[game.Player1].Life != life-1 {
				t.Fatalf("cast=%t activated=%d grove tapped=%t life=%d, want the Grove to pay Thalia's tax", ok, agent.activated, grove.Tapped, g.Players[game.Player1].Life)
			}
			obj, found := g.Stack.Peek()
			if !found || obj.Copy != tc.copy || obj.ManaSpentToCast != 1 {
				t.Fatalf("stack top = %+v, want the free cast with {1} spent", obj)
			}
			if tc.copy && !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("casting a copy moved the exiled card")
			}
		})
	}
}

// freeCastScriptAgent picks a spell whose label contains pick when first asked
// which spell to cast, prefers a kicked proposal when preferKicked is set, and
// activates the payment-time mana ability containing ability whenever offered.
type freeCastScriptAgent struct {
	pick         string
	preferKicked bool
	ability      string
	spellChoices [][]string
	castChoices  [][]string
	windows      int
	activated    int
}

func (*freeCastScriptAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *freeCastScriptAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	labels := make([]string, len(request.Options))
	for i, option := range request.Options {
		labels[i] = option.Label
	}
	pickLabel := func(want string) []int {
		for _, option := range request.Options {
			if want != "" && strings.Contains(option.Label, want) {
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	}
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Prompt == "Choose a spell to cast without paying its mana cost":
		a.spellChoices = append(a.spellChoices, labels)
		if len(a.spellChoices) == 1 {
			return pickLabel(a.pick)
		}
	case strings.Contains(request.Prompt, "without paying its mana cost"):
		a.castChoices = append(a.castChoices, labels)
		if a.preferKicked {
			return pickLabel("with kicker")
		}
	case request.Prompt == "Activate a mana ability while paying?":
		a.windows++
		selected := pickLabel(a.ability)
		if selected[0] != 0 {
			a.activated++
		}
		return selected
	}
	return request.DefaultSelection
}

func freeCastTestInstant(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         name,
		Types:        []types.Card{types.Instant},
		ManaCost:     opt.Val(cost.Mana{cost.G}),
		SpellAbility: opt.Val(game.AbilityContent{}),
	}}
}

// TestFreeCastWithoutManaPaymentOpensNoManaWindow proves the payment-time mana
// ability window opens only when the locked total includes a mana payment
// (CR 601.2g, 605.3a). With no cost increase each free route casts its spell
// for no mana, and a non-mana additional cost alone opens no window either,
// so Waterlogged Grove is never activated.
func TestFreeCastWithoutManaPaymentOpensNoManaWindow(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	sacrificeSpell := func(def *game.CardDef) *game.CardDef {
		def.AdditionalCosts = []cost.Additional{{
			Kind: cost.AdditionalSacrifice, Text: "Sacrifice an artifact", Amount: 1,
			MatchPermanentType: true, PermanentType: types.Artifact,
		}}
		return def
	}
	for _, tc := range []struct {
		name          string
		cast          func(*testing.T, *Engine, *game.Game, [game.NumPlayers]PlayerAgent) bool
		wantSacrifice bool
	}{
		{name: "rebound", cast: func(t *testing.T, e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Faithless Salvaging"))
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			castIntoReboundExile(t, e, g, cardID)
			setReboundUpkeep(g)
			e.processReboundUpkeep(g, game.Player1, agents, &TurnLog{})
			requireReboundCast(t, g, cardID)
			return true
		}},
		{name: "rebound with sacrifice cost", wantSacrifice: true, cast: func(t *testing.T, e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
			cardID := addCardToHand(g, game.Player1, sacrificeSpell(reboundTestSpell("Rebound Sacrifice Spell")))
			reboundTestArtifact(g)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			castIntoReboundExile(t, e, g, cardID)
			reboundTestArtifact(g)
			setReboundUpkeep(g)
			e.processReboundUpkeep(g, game.Player1, agents, &TurnLog{})
			requireReboundCast(t, g, cardID)
			return true
		}},
		{name: "from hand", cast: func(_ *testing.T, e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
			cardID := addCardToHand(g, game.Player1, freeCastTestInstant("Free Hand Spell"))
			return e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
				game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{})
		}},
		{name: "from hand with sacrifice cost", wantSacrifice: true, cast: func(_ *testing.T, e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
			cardID := addCardToHand(g, game.Player1, sacrificeSpell(freeCastTestInstant("Free Hand Sacrifice Spell")))
			reboundTestArtifact(g)
			return e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
				game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{})
		}},
		{name: "copy of exiled card", cast: func(_ *testing.T, e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) bool {
			cardID := addCardToExile(g, game.Player1, freeCastTestInstant("Free Copy Spell"))
			return e.castFreeCopyOfCard(g, game.Player1, cardID, agents, &TurnLog{})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			life := g.Players[game.Player1].Life
			agent := &freeCastScriptAgent{ability: "Pay 1 life"}

			if !tc.cast(t, e, g, reboundKickerAgents(agent)) {
				t.Fatal("free cast failed")
			}

			if agent.windows != 0 || grove.Tapped || g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatalf("windows=%d grove tapped=%t life=%d pool=%d, want no payment-time mana window", agent.windows, grove.Tapped, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.ManaSpentToCast != 0 || (len(obj.SacrificedAsCostIDs) == 1) != tc.wantSacrifice {
				t.Fatalf("stack top = %+v, want a free cast with sacrifice paid=%t", obj, tc.wantSacrifice)
			}
		})
	}
}

// TestFreeCastOffersOnlyFeasibleNonManaCosts resolves Electrodominance with
// Opt and Village Rites in hand and Waterlogged Grove as a payment-time mana
// source. Mana abilities cannot pay Village Rites' sacrifice, so without a
// creature it is not offered and Opt is cast by default; with a creature both
// are offered and Village Rites can be cast.
func TestFreeCastOffersOnlyFeasibleNonManaCosts(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name      string
		creature  bool
		pick      string
		wantRites bool
	}{
		{name: "no creature", pick: "Village Rites"},
		{name: "creature to sacrifice", creature: true, pick: "Village Rites", wantRites: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			optID := addCardToHand(g, game.Player1, reg.Lookup("Opt"))
			ritesID := addCardToHand(g, game.Player1, reg.Lookup("Village Rites"))
			electroID := addCardToHand(g, game.Player1, reg.Lookup("Electrodominance"))
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			if tc.creature {
				addCombatPermanent(g, game.Player1, reg.Lookup("Grizzly Bears"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.R, 2)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			if !e.applyAction(g, game.Player1, action.CastSpell(electroID, []game.Target{game.PlayerTarget(game.Player2)}, 1, nil)) {
				t.Fatal("casting Electrodominance failed")
			}
			agent := &freeCastScriptAgent{pick: tc.pick, ability: "Pay 1 life"}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

			if len(agent.spellChoices) != 1 {
				t.Fatalf("spell choices = %v, want one", agent.spellChoices)
			}
			ritesOffered := slices.ContainsFunc(agent.spellChoices[0], func(label string) bool { return strings.Contains(label, "Village Rites") })
			if ritesOffered != tc.creature {
				t.Fatalf("Village Rites offered = %t in %v, want %t", ritesOffered, agent.spellChoices[0], tc.creature)
			}
			wantID := optID
			if tc.wantRites {
				wantID = ritesID
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != wantID {
				t.Fatalf("stack top = %+v, want card %v cast", obj, wantID)
			}
			if agent.windows != 0 || grove.Tapped {
				t.Fatalf("windows=%d grove tapped=%t, want no mana window for a free total", agent.windows, grove.Tapped)
			}
		})
	}
}

// TestFreeCastReoffersAfterUnpaidProposal proves an unpaid free-cast proposal
// is reversed to the choice it came from (CR 733) rather than ending the
// opportunity. Under Thalia's tax with only Waterlogged Grove for mana, a
// kicked proposal and a spell needing {3} more are offered but cannot be paid;
// each is reversed, including the Grove activation, and the remaining proposal
// is cast with the Grove's mana.
func TestFreeCastReoffersAfterUnpaidProposal(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	t.Run("kicked option of one card", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		def := freeCastTestInstant("Free Kicker Spell")
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
			game.KickerKeyword{Cost: cost.Mana{cost.O(2)}},
		}}}
		cardID := addCardToHand(g, game.Player1, def)
		addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
		grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
		life := g.Players[game.Player1].Life
		agent := &freeCastScriptAgent{preferKicked: true, ability: "Pay 1 life"}

		if !e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
			game.Selection{}, opt.V[int]{}, false, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("free cast failed after the kicked proposal was reversed")
		}

		if len(agent.castChoices) != 2 || len(agent.castChoices[0]) != 2 || len(agent.castChoices[1]) != 1 ||
			strings.Contains(agent.castChoices[1][0], "with kicker") {
			t.Fatalf("cast choices = %v, want kicked and unkicked, then unkicked only", agent.castChoices)
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != cardID || obj.KickerPaid || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack top = %+v, want the unkicked spell paying Thalia's tax", obj)
		}
		if agent.activated != 2 || !grove.Tapped || g.Players[game.Player1].Life != life-1 {
			t.Fatalf("activations=%d grove tapped=%t life=%d, want the first activation reversed", agent.activated, grove.Tapped, g.Players[game.Player1].Life)
		}
	})
	t.Run("another spell from hand", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		costly := freeCastTestInstant("Free Costly Spell")
		costly.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
			{Label: "Pay {3}", Mana: cost.Mana{cost.O(3)}},
		}}}
		costlyID := addCardToHand(g, game.Player1, costly)
		plainID := addCardToHand(g, game.Player1, freeCastTestInstant("Free Plain Spell"))
		electroID := addCardToHand(g, game.Player1, reg.Lookup("Electrodominance"))
		g.Players[game.Player1].ManaPool.Add(mana.R, 2)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		if !e.applyAction(g, game.Player1, action.CastSpell(electroID, []game.Target{game.PlayerTarget(game.Player2)}, 1, nil)) {
			t.Fatal("casting Electrodominance failed")
		}
		addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
		grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
		life := g.Players[game.Player1].Life
		agent := &freeCastScriptAgent{pick: "Free Costly Spell", ability: "Pay 1 life"}

		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

		if len(agent.spellChoices) != 2 || len(agent.spellChoices[0]) != 2 || len(agent.spellChoices[1]) != 1 {
			t.Fatalf("spell choices = %v, want both spells, then the untried one", agent.spellChoices)
		}
		if !g.Players[game.Player1].Hand.Contains(costlyID) {
			t.Fatal("the reversed proposal left hand")
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != plainID || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack top = %+v, want the plain spell paying Thalia's tax", obj)
		}
		if agent.activated != 2 || !grove.Tapped || g.Players[game.Player1].Life != life-1 {
			t.Fatalf("activations=%d grove tapped=%t life=%d, want the first activation reversed", agent.activated, grove.Tapped, g.Players[game.Player1].Life)
		}
	})
}
