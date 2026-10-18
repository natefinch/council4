package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: OTJ 220 printing ee29b59c-d57c-4a03-bae7-e9dfa57d6bb1;
// https://api.scryfall.com/cards/ee29b59c-d57c-4a03-bae7-e9dfa57d6bb1/rulings
// returned six WotC rulings dated 2024-04-12. CR 2026-09-25 from the verified
// official copy governs independent abilities (113.7a), optional resolving
// payment (118.12), library order/privacy (401), and announcement-time crime
// (700.13). The short-library ruling requires the sole card to go to hand.
const marchesaDealerOfDeathTestOracleID = "37449c93-4d32-4f1a-982f-8f3416214368"

type marchesaChoicePolicy struct {
	simPassPolicy

	pay       bool
	digChoice int
	requests  []game.ChoiceRequest
}

func (p *marchesaChoicePolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.requests = append(p.requests, request)
	switch request.Kind {
	case game.ChoiceMay:
		if p.pay {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceDig:
		return []int{p.digChoice}
	default:
		return request.DefaultSelection
	}
}

func marchesaScenario(t *testing.T, libraryCount int, paymentMana bool) (*game.Game, *Engine, *game.Permanent, id.ID, []id.ID) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	marchesaDef, err := registry.LookupOracleID(marchesaDealerOfDeathTestOracleID)
	if err != nil {
		t.Fatalf("exact Marchesa default-registry lookup failed: %v", err)
	}
	if marchesaDef.Name != "Marchesa, Dealer of Death" {
		t.Fatalf("exact Marchesa lookup returned %q", marchesaDef.Name)
	}
	boltDef := registry.Lookup("Lightning Bolt")
	if boltDef == nil {
		t.Fatal("production Lightning Bolt is missing")
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	marchesa := addCombatPermanent(g, game.Player1, marchesaDef)
	var library []id.ID
	for i := range libraryCount {
		library = append(library, addCardToLibrary(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: fmt.Sprintf("Marchesa library card %d", i)},
		}))
	}
	boltID := addCardToHand(g, game.Player1, boltDef)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	if paymentMana {
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	}

	return g, NewEngine(nil), marchesa, boltID, library
}

func castMarchesaCrimeSpell(t *testing.T, g *game.Game, engine *Engine, boltID id.ID, target game.PlayerID) {
	t.Helper()
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		boltID, []game.Target{game.PlayerTarget(target)}, 0, nil,
	)) {
		t.Fatal("Lightning Bolt cast failed")
	}
}

func putMarchesaCrimeTrigger(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 1 {
		t.Fatal("opponent-targeted cast did not commit exactly one crime")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Marchesa's crime trigger was not put on the stack")
	}
}

func TestMarchesaDealerOfDeathCrimeTriggerPaysAndResolvesAfterSourceLeaves(t *testing.T) {
	g, engine, marchesa, boltID, library := marchesaScenario(t, 2, true)
	castMarchesaCrimeSpell(t, g, engine, boltID, game.Player2)
	putMarchesaCrimeTrigger(t, g, engine)

	trigger, ok := g.Stack.Peek()
	if !ok || trigger.SourceID != marchesa.ObjectID {
		t.Fatal("crime trigger was not above the spell that committed the crime")
	}

	addEffectSpellToStack(g, game.Player2, game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(marchesa.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, stillOnBattlefield := permanentByObjectID(g, marchesa.ObjectID); stillOnBattlefield {
		t.Fatal("setup effect did not remove Marchesa")
	}
	if !g.Players[game.Player1].Graveyard.Contains(marchesa.CardInstanceID) {
		t.Fatal("Marchesa did not move to its owner's graveyard")
	}

	agent := &marchesaChoicePolicy{pay: true, digChoice: 1}
	policies := simPassPolicies()
	policies[game.Player1] = agent
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})

	chosen, unchosen := library[0], library[1]
	if !g.Players[game.Player1].Hand.Contains(chosen) {
		t.Fatal("chosen looked-at card did not go to hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(unchosen) {
		t.Fatal("unchosen looked-at card did not go to the graveyard")
	}
	if g.Players[game.Player1].Library.Contains(chosen) || g.Players[game.Player1].Library.Contains(unchosen) {
		t.Fatal("a looked-at card remained in the library")
	}
	if len(agent.requests) != 2 ||
		agent.requests[0].Kind != game.ChoiceMay || agent.requests[1].Kind != game.ChoiceDig {
		t.Fatalf("Marchesa choices = %#v, want payment then a private dig choice", agent.requests)
	}
	for _, request := range agent.requests {
		if request.Player != game.Player1 {
			t.Fatalf("choice was offered to %v, want Marchesa's controller", request.Player)
		}
	}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && (event.CardID == chosen || event.CardID == unchosen) {
			t.Fatal("looking at cards for Marchesa publicly revealed them")
		}
	}
	remaining, ok := g.Stack.Peek()
	if !ok || remaining.Kind != game.StackSpell || len(g.Stack.Objects()) != 1 {
		t.Fatal("resolving Marchesa's trigger also resolved the spell that committed the crime")
	}
}

func TestMarchesaDealerOfDeathDeclinesOrCannotPay(t *testing.T) {
	for _, test := range []struct {
		name          string
		paymentMana   bool
		wantMayChoice bool
	}{
		{name: "declined", paymentMana: true, wantMayChoice: true},
		{name: "unaffordable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, engine, _, boltID, library := marchesaScenario(t, 2, test.paymentMana)
			castMarchesaCrimeSpell(t, g, engine, boltID, game.Player2)
			putMarchesaCrimeTrigger(t, g, engine)

			agent := &marchesaChoicePolicy{}
			policies := simPassPolicies()
			policies[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})

			if g.Players[game.Player1].Library.Size() != 2 {
				t.Fatal("declining or being unable to pay changed the library")
			}
			for _, cardID := range library {
				if !g.Players[game.Player1].Library.Contains(cardID) ||
					g.Players[game.Player1].Hand.Contains(cardID) ||
					g.Players[game.Player1].Graveyard.Contains(cardID) {
					t.Fatalf("unpaid trigger moved library card %d", cardID)
				}
			}
			wantRequests := 0
			if test.wantMayChoice {
				wantRequests = 1
			}
			if len(agent.requests) != wantRequests ||
				(test.wantMayChoice && agent.requests[0].Kind != game.ChoiceMay) {
				t.Fatalf("choices = %#v, want %d payment choices and no dig", agent.requests, wantRequests)
			}
		})
	}
}

func TestMarchesaDealerOfDeathOneCardLibrary(t *testing.T) {
	g, engine, _, boltID, library := marchesaScenario(t, 1, true)
	castMarchesaCrimeSpell(t, g, engine, boltID, game.Player2)
	putMarchesaCrimeTrigger(t, g, engine)

	agent := &marchesaChoicePolicy{pay: true}
	policies := simPassPolicies()
	policies[game.Player1] = agent
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})

	cardID := library[0]
	if !g.Players[game.Player1].Hand.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("the only card in the library was not put into hand")
	}
	if g.Players[game.Player1].Library.Size() != 0 {
		t.Fatal("the only card remained in the library")
	}
	if len(agent.requests) != 2 || agent.requests[1].Kind != game.ChoiceDig ||
		len(agent.requests[1].Options) != 1 {
		t.Fatalf("choices = %#v, want payment and one-card dig choices", agent.requests)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.CardID == cardID {
			t.Fatal("looking at the only card publicly revealed it")
		}
	}
}

func TestMarchesaDealerOfDeathDoesNotTriggerForCrimeAgainstSelf(t *testing.T) {
	g, engine, _, boltID, _ := marchesaScenario(t, 0, false)
	castMarchesaCrimeSpell(t, g, engine, boltID, game.Player1)

	if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 0 {
		t.Fatal("a spell targeting its controller committed a crime")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Marchesa triggered for a spell that targeted only its controller")
	}
}

func TestMarchesaDealerOfDeathTriggersForActivatedAndTriggeredCrimes(t *testing.T) {
	for _, test := range []struct {
		name      string
		triggered bool
	}{
		{name: "activated"},
		{name: "triggered", triggered: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, engine, marchesa, _, library := marchesaScenario(t, 2, true)
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Crime target", 3, 3))
			targets := []game.TargetSpec{{
				Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1,
				Selection: opt.Val(game.Selection{
					RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent,
				}),
			}}
			instructions := []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}}
			if test.triggered {
				addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
					Event: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.TriggerPlayerYou,
				}, instructions, targets)
				g.Turn.ActivePlayer = game.Player1
				emitBeginningOfStepEvent(g, game.StepUpkeep)
				if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 0 {
					t.Fatal("crime occurred before the targeted trigger was put on the stack")
				}
				if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
					t.Fatal("targeted upkeep ability was not put on the stack")
				}
			} else {
				source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Crime activation fixture", Types: []types.Card{types.Artifact},
					ActivatedAbilities: []game.ActivatedAbility{{
						ZoneOfFunction: zone.Battlefield,
						Content:        game.Mode{Targets: targets, Sequence: instructions}.Ability(),
					}},
				}})
				setMainPhasePriority(g, game.Player1)
				if !engine.applyAction(g, game.Player1, action.ActivateAbility(
					source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
				)) {
					t.Fatal("targeted activated ability was not announced")
				}
			}
			putMarchesaCrimeTrigger(t, g, engine)
			top, ok := g.Stack.Peek()
			if !ok || top.SourceID != marchesa.ObjectID || target.Tapped {
				t.Fatal("Marchesa did not trigger above the unresolved crime ability")
			}
			agent := &marchesaChoicePolicy{pay: true}
			policies := simPassPolicies()
			policies[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
			if g.Players[game.Player1].Library.Size() != 0 ||
				!g.Players[game.Player1].Hand.Contains(library[1]) ||
				!g.Players[game.Player1].Graveyard.Contains(library[0]) ||
				target.Tapped || g.Stack.Size() != 1 {
				t.Fatal("Marchesa's paid trigger did not resolve independently before the crime ability")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !target.Tapped || countEventsOfKind(g.Events, game.EventCrimeCommitted) != 1 {
				t.Fatal("crime ability did not resolve or incorrectly committed another crime at resolution")
			}
		})
	}
}
