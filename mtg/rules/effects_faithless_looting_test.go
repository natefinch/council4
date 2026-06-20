package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lootingInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
		{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
	}
}

func TestLootingDrawsBeforeChoosingCardsToDiscard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	old := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn B"}})
	addInstructionSpellToStack(g, lootingInstructions())

	agent := &orderedHandChoiceAgent{order: []string{"Drawn B", "Old"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if !slices.Contains(agent.observedHand, "Drawn B") {
		t.Fatalf("choice observation hand = %v, want newly drawn card", agent.observedHand)
	}
	if g.Players[game.Player1].Hand.Contains(old) {
		t.Fatal("chosen old card remained in hand")
	}
	if len(agent.requests) != 1 ||
		agent.requests[0].Player != game.Player1 ||
		agent.requests[0].MinChoices != 2 ||
		agent.requests[0].MaxChoices != 2 {
		t.Fatalf("choice requests = %#v, want controller exact two", agent.requests)
	}
}

func TestLootingInvalidChoiceFallsBackToExactDistinctCards(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "C"}})
	addInstructionSpellToStack(g, lootingInstructions()[1:])

	agent := &orderedHandChoiceAgent{answer: []int{0, 0}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if len(log.Choices) != 1 || !log.Choices[0].UsedFallback {
		t.Fatalf("choices = %#v, want fallback", log.Choices)
	}
	selected := log.Choices[0].Selected
	if len(selected) != 2 || selected[0] == selected[1] {
		t.Fatalf("fallback selected = %v, want exact distinct pair", selected)
	}
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("hand size = %d, want 1", g.Players[game.Player1].Hand.Size())
	}
}

func TestLootingInsufficientHandDiscardsAllAvailable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	only := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only"}})
	addInstructionSpellToStack(g, lootingInstructions()[1:])

	agent := &orderedHandChoiceAgent{order: []string{"Only"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if g.Players[game.Player1].Hand.Size() != 0 ||
		!g.Players[game.Player1].Graveyard.Contains(only) {
		t.Fatal("only available card was not discarded")
	}
	if len(log.Choices) != 1 ||
		log.Choices[0].Request.MinChoices != 1 ||
		log.Choices[0].Request.MaxChoices != 1 {
		t.Fatalf("choices = %#v, want exact one available card", log.Choices)
	}
}

func TestLootingEmitsOneSimultaneousDiscardEventPerCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addInstructionSpellToStack(g, lootingInstructions()[1:])

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	var discards []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCardDiscarded {
			discards = append(discards, event)
		}
	}

	if len(discards) != 2 {
		t.Fatalf("discard events = %#v, want 2", discards)
	}
	var simultaneousID id.ID
	for i, event := range discards {
		if event.CardID == 0 ||
			event.Player != game.Player1 ||
			event.FromZone != zone.Hand ||
			event.ToZone != zone.Graveyard {
			t.Fatalf("discard event %d = %#v", i, event)
		}
		if i == 0 {
			simultaneousID = event.SimultaneousID
		}
		if event.SimultaneousID == 0 || event.SimultaneousID != simultaneousID {
			t.Fatalf("discard events not one simultaneous batch: %#v", discards)
		}
	}
}

func TestFaithlessLootingNormalAndFlashbackEndToEnd(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Faithless Looting",
		Types:        []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.R}),
		SpellAbility: opt.Val(game.Mode{Sequence: lootingInstructions()}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{
				game.FlashbackKeyword{Cost: cost.Mana{cost.O(2), cost.R}},
			},
		}},
	}})
	for _, name := range []string{"Old A", "Old B"} {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	for _, name := range []string{"Draw 1", "Draw 2", "Draw 3", "Draw 4"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	var lands []*game.Permanent
	for range 4 {
		lands = append(lands, addBasicLandPermanent(g, game.Player1, types.Mountain))
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("normal hand cast failed")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Flashback {
		t.Fatalf("normal stack object = %+v, want non-flashback", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("normally cast Faithless Looting did not enter graveyard")
	}

	for _, land := range lands {
		land.Tapped = false
	}
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("flashback cast failed")
	}
	if obj, ok := g.Stack.Peek(); !ok || !obj.Flashback {
		t.Fatalf("flashback stack object = %+v, want marked", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) ||
		g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("flashback resolution did not exile Faithless Looting")
	}
}
