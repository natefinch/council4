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
)

// Suspend casts follow the alternative-cost rules (CR 702.62a, 702.62d): "you
// may play it without paying its mana cost" replaces only the mana cost, while
// additional costs, cost increases and reductions still apply (CR 118.9d,
// 601.2f-h). If the card isn't cast it remains exiled. These scenarios cover
// Engine gap #3459 on the suspend route.

func suspendCastEvents(g *game.Game, cardID id.ID) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID && event.FromZone == zone.Exile {
			events = append(events, event)
		}
	}
	return events
}

// requireSuspendNotCast asserts the last time counter was removed without
// casting the card, which remains exiled and no longer suspended.
func requireSuspendNotCast(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if got := suspendCastEvents(g, cardID); len(got) != 0 {
		t.Fatalf("suspend cast events = %d, want none", len(got))
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("uncast suspended card left exile")
	}
	if got := g.SuspendedCards[cardID].TimeCounters; got != 0 {
		t.Fatalf("time counters = %d, want the last counter removed", got)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want empty after an uncast suspended card", g.Stack.Size())
	}
}

// requireSuspendCast asserts the card was cast from exile by its suspend
// ability, returning its stack object and cast event.
func requireSuspendCast(t *testing.T, g *game.Game, cardID id.ID) (*game.StackObject, game.Event) {
	t.Helper()
	events := suspendCastEvents(g, cardID)
	if len(events) != 1 {
		t.Fatalf("suspend cast events = %d, want 1", len(events))
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("suspended card is still in exile after being cast")
	}
	if _, ok := g.SuspendedCards[cardID]; ok {
		t.Fatal("suspended state remained after the cast")
	}
	for _, obj := range g.Stack.Objects() {
		if !obj.Copy && obj.SourceID == cardID {
			if !obj.Suspend || obj.SourceZone != zone.Exile {
				t.Fatalf("suspend spell suspend=%t source zone=%v, want a suspend cast from exile", obj.Suspend, obj.SourceZone)
			}
			return obj, events[0]
		}
	}
	t.Fatal("suspended spell is not on the stack")
	return nil, game.Event{}
}

func suspendAgents(agent PlayerAgent) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	return agents
}

func withSuspendSacrificeArtifact(def *game.CardDef) *game.CardDef {
	def.AdditionalCosts = []cost.Additional{{
		Kind:               cost.AdditionalSacrifice,
		Text:               "Sacrifice an artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}}
	return def
}

// TestSuspendCastPaysThaliaTax suspends a noncreature sorcery under Thalia,
// Guardian of Thraben: the free suspend cast must pay Thalia's {1} tax, so
// without that mana the card stays exiled, and with it the tax is spent.
func TestSuspendCastPaysThaliaTax(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name      string
		thalia    bool
		upkeepC   int
		wantCast  bool
		wantSpent int
	}{
		{name: "no tax control", wantCast: true},
		{name: "unpayable tax", thalia: true},
		{name: "payable tax", thalia: true, upkeepC: 1, wantCast: true, wantSpent: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
			if tc.thalia {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)

			e.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if !tc.wantCast {
				requireSuspendNotCast(t, g, cardID)
				if got := g.Players[game.Player1].ManaPool.Total(); got != tc.upkeepC {
					t.Fatalf("mana pool = %d, want %d untouched", got, tc.upkeepC)
				}
				// The card is no longer suspended, so a later upkeep does not cast it.
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
				e.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				requireSuspendNotCast(t, g, cardID)
				return
			}
			obj, event := requireSuspendCast(t, g, cardID)
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana pool = %d after suspend cast, want Thalia tax spent", got)
			}
			if obj.ManaSpentToCast != tc.wantSpent || !event.ManaSpentToCast.Exists || event.ManaSpentToCast.Val != tc.wantSpent {
				t.Fatalf("mana spent obj=%d event=%+v, want %d", obj.ManaSpentToCast, event.ManaSpentToCast, tc.wantSpent)
			}
		})
	}
}

// TestSuspendCastPaysAdditionalCosts proves a mandatory additional cost
// (sacrifice an artifact) is still paid on the free suspend cast, alone and
// combined with Thalia's tax, and that an unpaid total is reversed without
// sacrificing anything or spending mana (CR 733).
func TestSuspendCastPaysAdditionalCosts(t *testing.T) {
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
		{name: "artifact but tax unpayable", artifact: true, thalia: true},
		{name: "tax but no artifact", thalia: true, upkeepC: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addSuspendedCard(g, game.Player1, withSuspendSacrificeArtifact(suspendSorcery(1, cost.Mana{cost.O(1)})), 1)
			var fodder *game.Permanent
			if tc.artifact {
				fodder = reboundTestArtifact(g)
			}
			if tc.thalia {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)

			e.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if !tc.wantCast {
				requireSuspendNotCast(t, g, cardID)
				if got := g.Players[game.Player1].ManaPool.Total(); got != tc.upkeepC {
					t.Fatalf("mana pool = %d, want %d untouched", got, tc.upkeepC)
				}
				if fodder != nil {
					if _, onBattlefield := permanentByObjectID(g, fodder.ObjectID); !onBattlefield {
						t.Fatal("reversed suspend cast sacrificed the artifact")
					}
				}
				return
			}
			obj, _ := requireSuspendCast(t, g, cardID)
			if len(obj.SacrificedAsCostIDs) != 1 || !g.Players[game.Player1].Graveyard.Contains(fodder.CardInstanceID) {
				t.Fatalf("sacrificed as cost = %v, want the artifact sacrificed", obj.SacrificedAsCostIDs)
			}
			if obj.ManaSpentToCast != tc.upkeepC || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatalf("mana spent=%d pool=%d, want %d spent", obj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total(), tc.upkeepC)
			}
		})
	}
}

// TestSuspendCastOptionalKickerWithTax combines Thalia's tax with an optional
// kicker cost on the suspend cast: the kicked proposal is offered only when the
// combined mana is affordable, and the chosen total is paid.
func TestSuspendCastOptionalKickerWithTax(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name       string
		upkeepC    int
		wantKicked bool
	}{
		{name: "kicker and tax affordable", upkeepC: 2, wantKicked: true},
		{name: "only tax affordable", upkeepC: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			def := suspendSorcery(1, cost.Mana{cost.O(1)})
			def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{KeywordAbilities: []game.KeywordAbility{
				game.KickerKeyword{Cost: cost.Mana{cost.O(1)}},
			}})
			cardID := addSuspendedCard(g, game.Player1, def, 1)
			addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.upkeepC)
			agent := &reboundKickerAgent{}

			e.processSuspendUpkeep(g, game.Player1, suspendAgents(agent), &TurnLog{})

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
			obj, _ := requireSuspendCast(t, g, cardID)
			if obj.KickerPaid != tc.wantKicked || obj.ManaSpentToCast != tc.upkeepC {
				t.Fatalf("kicker paid=%t mana spent=%d, want %t/%d", obj.KickerPaid, obj.ManaSpentToCast, tc.wantKicked, tc.upkeepC)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana pool = %d, want all spent", got)
			}
		})
	}
}

// TestSuspendCastMayBeDeclined keeps the suspend cast optional (CR 702.62a
// "you may play it"): declining casts nothing, spends nothing and leaves the
// card exiled without time counters.
func TestSuspendCastMayBeDeclined(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agent := &cascadeCastChoiceAgent{decline: true}

	e.processSuspendUpkeep(g, game.Player1, suspendAgents(agent), &TurnLog{})

	if len(agent.requests) != 1 || declineOptionCount(agent.requests[0]) != 1 {
		t.Fatalf("free-cast requests = %+v, want one request offering a decline", agent.requests)
	}
	requireSuspendNotCast(t, g, cardID)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana pool = %d, want 1 untouched", got)
	}
}

// suspendTargetAgent picks the free-cast proposal targeting target.
type suspendTargetAgent struct {
	target  game.Target
	offered int
}

func (*suspendTargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *suspendTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceResolution || !strings.Contains(request.Prompt, "without paying its mana cost") {
		return request.DefaultSelection
	}
	a.offered = len(request.Options)
	for _, option := range request.Options {
		if len(option.Targets) == 1 && option.Targets[0] == a.target {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

// TestSuspendCastChoosesTargets suspends a Rift Bolt-style sorcery (Lightning
// Bolt's spell ability as a sorcery with suspend 1): the caster chooses the
// target of the free suspend cast, pays Thalia's tax, and the spell resolves
// against that target.
func TestSuspendCastChoosesTargets(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	def := *reg.Lookup("Lightning Bolt")
	def.Name = "Suspend Bolt"
	def.Types = []types.Card{types.Sorcery}
	def.StaticAbilities = append(slices.Clone(def.StaticAbilities), game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.SuspendKeyword{Cost: cost.Mana{cost.R}, TimeCounters: 1}},
	})
	e, g := newEffectiveCascadeGame()
	cardID := addSuspendedCard(g, game.Player1, &def, 1)
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agent := &suspendTargetAgent{target: game.PlayerTarget(game.Player3)}

	e.processSuspendUpkeep(g, game.Player1, suspendAgents(agent), &TurnLog{})

	obj, _ := requireSuspendCast(t, g, cardID)
	if agent.offered < 2 || len(obj.Targets) != 1 || obj.Targets[0] != agent.target || obj.ManaSpentToCast != 1 {
		t.Fatalf("offered=%d targets=%v mana spent=%d, want the chosen player targeted and Thalia's tax paid", agent.offered, obj.Targets, obj.ManaSpentToCast)
	}
	life := g.Players[game.Player3].Life
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player3].Life; got != life-3 {
		t.Fatalf("target life = %d, want %d", got, life-3)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved suspended sorcery did not go to its owner's graveyard")
	}
}

// TestSuspendedCreatureWithAdditionalCostGainsHaste proves a creature spell
// cast by suspend pays its additional cost and still gains suspend haste, and
// that an unpayable additional cost leaves it exiled.
func TestSuspendedCreatureWithAdditionalCostGainsHaste(t *testing.T) {
	for _, tc := range []struct {
		name     string
		artifact bool
	}{
		{name: "additional cost paid", artifact: true},
		{name: "additional cost unpayable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addSuspendedCard(g, game.Player1, withSuspendSacrificeArtifact(suspendCreature(1, cost.Mana{cost.O(1)})), 1)
			if tc.artifact {
				reboundTestArtifact(g)
			}

			e.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if !tc.artifact {
				requireSuspendNotCast(t, g, cardID)
				return
			}
			obj, _ := requireSuspendCast(t, g, cardID)
			if len(obj.SacrificedAsCostIDs) != 1 {
				t.Fatalf("sacrificed as cost = %v, want the artifact", obj.SacrificedAsCostIDs)
			}
			e.resolveTopOfStack(g, &TurnLog{})
			permanent := permanentByCardID(g, cardID)
			if permanent == nil {
				t.Fatal("suspended creature did not enter")
			}
			if !canAttackWith(g, permanent, game.Player1) {
				t.Fatal("suspended creature cannot attack despite suspend haste")
			}
		})
	}
}

// TestSuspendCastPaysWithPaymentTimeManaAbility uses Waterlogged Grove, whose
// only mana ability costs life and is never activated automatically, as the
// sole source for Thalia's tax: the suspend proposal is offered and the caster
// activates the ability while paying (CR 601.2g-h). The cast runs in a clone
// without changing the original game.
func TestSuspendCastPaysWithPaymentTimeManaAbility(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	setReboundUpkeep(g)

	clone := g.Clone()
	agent := &paymentManaAgent{ability: "Pay 1 life"}
	e.processSuspendUpkeep(clone, game.Player1, suspendAgents(agent), &TurnLog{})

	obj, event := requireSuspendCast(t, clone, cardID)
	grove := reboundPermanentNamed(t, clone, "Waterlogged Grove")
	if agent.activated != 1 || !grove.Tapped || clone.Players[game.Player1].Life != g.Players[game.Player1].Life-1 {
		t.Fatalf("activated=%d grove tapped=%t life=%d, want the Grove paid for Thalia's tax", agent.activated, grove.Tapped, clone.Players[game.Player1].Life)
	}
	if obj.ManaSpentToCast != 1 || event.ManaSpentToCast.Val != 1 || clone.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana spent obj=%d event=%+v pool=%d, want the Grove's mana spent on the tax", obj.ManaSpentToCast, event.ManaSpentToCast, clone.Players[game.Player1].ManaPool.Total())
	}
	if g.SuspendedCards[cardID].TimeCounters != 1 || !g.Players[game.Player1].Exile.Contains(cardID) ||
		reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || g.Stack.Size() != 0 {
		t.Fatal("casting the suspended card in the clone changed the original game")
	}
}

// TestSuspendCastRollsBackPaymentTimeManaAbilities: with two Thalias the {2}
// tax exceeds Waterlogged Grove's mana, so the proposal is offered but the
// unpaid cast is reversed along with the reversible mana activation (CR 733),
// leaving the card exiled.
func TestSuspendCastRollsBackPaymentTimeManaAbilities(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
	grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	setReboundUpkeep(g)
	life := g.Players[game.Player1].Life
	agent := &paymentManaAgent{ability: "Pay 1 life"}

	e.processSuspendUpkeep(g, game.Player1, suspendAgents(agent), &TurnLog{})

	if agent.windows != 1 || agent.activated != 1 {
		t.Fatalf("payment windows=%d activations=%d, want 1/1", agent.windows, agent.activated)
	}
	requireSuspendNotCast(t, g, cardID)
	if grove.Tapped || g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("grove tapped=%t life=%d pool=%d, want no mana activation left behind", grove.Tapped, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total())
	}
}
