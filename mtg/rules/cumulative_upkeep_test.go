package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRegisteredCumulativeUpkeepPlansWholeTreasurePayment(t *testing.T) {
	for _, scenario := range []string{"pay", "decline", "unpayable"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, registryCardDef(t, "Mystic Remora"))
			source.Counters.Add(counter.Age, 1)
			count := 2
			if scenario == "unpayable" {
				count = 1
			}
			for range count {
				if _, ok := createTokenPermanent(g, game.Player1, recklessLackeyTreasure(t)); !ok {
					t.Fatal("Treasure creation failed")
				}
			}
			emitBeginningOfStepEvent(g, game.StepUpkeep)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("registered cumulative upkeep did not trigger")
			}
			consent, windows := 0, 0
			agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceMay {
					consent++
					if req.Prompt != "Pay {2}?" || source.Counters.Get(counter.Age) != 2 {
						t.Fatal("upkeep payment was not materialized after adding the age counter")
					}
					if scenario == "decline" {
						return []int{0}
					}
				}
				if req.Prompt == "Activate a mana ability while paying?" {
					windows++
					for _, option := range req.Options {
						if !strings.Contains(option.Label, "Treasure") {
							t.Fatal("partial upkeep payment offered")
						}
					}
				}
				return req.DefaultSelection
			}}
			engine.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
			_, remains := permanentByObjectID(g, source.ObjectID)
			if source.Counters.Get(counter.Age) != 2 || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("upkeep counter or mana receipt was incorrect")
			}
			if scenario == "pay" {
				if !remains || windows != 2 || consent != 1 || countTokenPermanentsNamed(g, "Treasure") != 0 {
					t.Fatal("whole cumulative cost was not paid with both Treasures")
				}
			} else if remains || windows != 0 || countTokenPermanentsNamed(g, "Treasure") != count ||
				consent != map[bool]int{true: 1}[scenario == "decline"] {
				t.Fatal("unpaid upkeep made a partial payment or did not sacrifice its source")
			}
		})
	}
}

func TestCumulativeUpkeepKeepsRepeatedPaymentChoicesDistinct(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCumulativeUpkeepPermanent(g, game.Player1, cost.Mana{cost.PhyrexianMana(mana.U)})
	source.Counters.Add(counter.Age, 1)
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	choices := 0
	agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Prompt == "Choose Phyrexian payment" {
			choices++
			want := "Pay with mana"
			if choices == 1 {
				want = "Pay 2 life"
			}
			for _, option := range req.Options {
				if option.Label == want {
					return []int{option.Index}
				}
			}
			t.Fatalf("independent upkeep choice %d unavailable: %+v", choices, req.Options)
		}
		return req.DefaultSelection
	}}
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cumulative upkeep did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
	if choices != 2 || !island.Tapped || g.Players[game.Player1].Life != 38 ||
		g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("repeated cost choices were merged or the whole payment did not apply")
	}
}

func TestCumulativeUpkeepAddsAgeThenPaysMultipliedExactMana(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		initialAge int
		lands      int
		wantAge    int
		wantPrompt string
	}{
		{name: "first upkeep", lands: 2, wantAge: 1, wantPrompt: "Pay {1}{U}?"},
		{name: "second upkeep", initialAge: 1, lands: 4, wantAge: 2, wantPrompt: "Pay {2}{U}{U}?"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCumulativeUpkeepPermanent(g, game.Player1, cost.Mana{cost.O(1), cost.U})
			source.Counters.Add(counter.Age, test.initialAge)
			lands := make([]*game.Permanent, test.lands)
			for i := range lands {
				lands[i] = addBasicLandPermanent(g, game.Player1, types.Island)
			}

			emitBeginningOfStepEvent(g, game.StepUpkeep)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("cumulative upkeep trigger was not put on the stack")
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if got := source.Counters.Get(counter.Age); got != test.wantAge {
				t.Fatalf("age counters = %d; want %d", got, test.wantAge)
			}
			if _, ok := g.PermanentByID(source.ObjectID); !ok {
				t.Fatal("paid cumulative upkeep permanent was sacrificed")
			}
			tapped := 0
			for _, land := range lands {
				if land.Tapped {
					tapped++
				}
			}
			if tapped != test.lands {
				t.Fatalf("tapped lands = %d; want %d", tapped, test.lands)
			}
			if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != test.wantPrompt {
				t.Fatalf("choices = %+v; want prompt %q", log.Choices, test.wantPrompt)
			}
		})
	}
}

func TestCumulativeUpkeepSacrificesOnDeclineOrInability(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		landSubtype types.Sub
		choices     [][]int
		wantChoices int
	}{
		{name: "controller declines", landSubtype: types.Island, choices: [][]int{{0}}, wantChoices: 1},
		{name: "controller cannot pay exact color", landSubtype: types.Forest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCumulativeUpkeepPermanent(g, game.Player1, cost.Mana{cost.U})
			addBasicLandPermanent(g, game.Player1, test.landSubtype)

			emitBeginningOfStepEvent(g, game.StepUpkeep)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("cumulative upkeep trigger was not put on the stack")
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: test.choices},
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if got := source.Counters.Get(counter.Age); got != 1 {
				t.Fatalf("age counters = %d; want 1 before sacrifice", got)
			}
			if _, ok := g.PermanentByID(source.ObjectID); ok {
				t.Fatal("unpaid cumulative upkeep permanent remains on battlefield")
			}
			if got := g.Players[game.Player1].Graveyard.Size(); got != 1 {
				t.Fatalf("graveyard size = %d; want 1", got)
			}
			if len(log.Choices) != test.wantChoices {
				t.Fatalf("choices = %+v; want %d", log.Choices, test.wantChoices)
			}
			assertEvent(t, g.Events, game.EventPermanentSacrificed, func(event game.Event) bool {
				return event.PermanentID == source.ObjectID
			})
			assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
				return event.PermanentID == source.ObjectID &&
					event.FromZone == zone.Battlefield &&
					event.ToZone == zone.Graveyard
			})
		})
	}
}

func TestCumulativeUpkeepKeepsObjectIdentityThroughBlink(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCumulativeUpkeepPermanent(g, game.Player1, cost.Mana{cost.U})

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cumulative upkeep trigger was not put on the stack")
	}
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("failed to exile cumulative upkeep source")
	}
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok {
		t.Fatal("failed to return cumulative upkeep source")
	}
	if returned.ObjectID == source.ObjectID {
		t.Fatal("blink preserved object identity")
	}

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if got := returned.Counters.Get(counter.Age); got != 0 {
		t.Fatalf("returned permanent age counters = %d; want 0", got)
	}
	if _, ok := g.PermanentByID(returned.ObjectID); !ok {
		t.Fatal("returned permanent was sacrificed by old object's trigger")
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v; want no payment for missing source object", log.Choices)
	}
}

func TestCumulativeUpkeepUsesNormalAPNAPTriggerOrdering(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	addCumulativeUpkeepPermanent(g, game.Player1, cost.Mana{cost.U})
	addTriggeredPermanent(g, game.Player3, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, nil, nil)

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("upkeep triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d; want 2", got)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack unexpectedly empty")
	}
	if top.Controller != game.Player3 {
		t.Fatalf("top trigger controller = %v; want Player3 after APNAP placement", top.Controller)
	}
}

func addCumulativeUpkeepPermanent(g *game.Game, controller game.PlayerID, manaCost cost.Mana) *game.Permanent {
	ability := game.CumulativeUpkeepTriggeredAbility(manaCost)
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Cumulative Upkeep Permanent",
		Types:              []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{ability},
	}}
	return addCombatPermanent(g, controller, def)
}
