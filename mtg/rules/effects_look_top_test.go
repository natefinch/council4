package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type lookChoiceAgent struct {
	requests []game.ChoiceRequest
}

func (*lookChoiceAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	return legal[0]
}

func (a *lookChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	return nil
}

func TestLookAtLibraryTopIsPrivateAndDoesNotReveal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Elf",
		Types: []types.Card{types.Creature},
	}})
	addInstructionSpellToStack(g, []game.Instruction{{
		Primitive: game.LookAtLibraryTop{
			Player:        game.ControllerReference(),
			PublishLinked: "looked",
		},
	}})
	controller := &lookChoiceAgent{}
	opponent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = controller
	agents[game.Player2] = opponent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(controller.requests) != 1 {
		t.Fatalf("controller requests = %#v, want one private look", controller.requests)
	}

	request := controller.requests[0]
	if !request.Subject.Exists || request.Subject.Val.CardID != topID {
		t.Fatalf("look subject = %#v, want exact top card %d", request.Subject, topID)
	}
	if len(opponent.requests) != 0 {
		t.Fatalf("opponent requests = %#v, want none", opponent.requests)
	}
	if !g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("looked-at card left the library")
	}
	if eventRevealedCard(g, topID, 0) {
		t.Fatal("private look emitted a public reveal event")
	}
}

func TestLookedAtMatchingCreatureCanBeRevealedAndMovedByExactIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Secret Elf",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}})
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: "looked",
			},
		},
		{
			Primitive: game.Reveal{
				Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"},
			},
			CardCondition: opt.Val(game.CardCondition{
				Card:              game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"},
				Types:             []types.Card{types.Creature},
				ChosenSubtypeFrom: game.EntryTypeChoiceKey,
			}),
			Optional:      true,
			PublishResult: "revealed",
		},
		{
			Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"},
				FromZone:    zone.Library,
				Destination: zone.Hand,
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       "revealed",
				Accepted:  game.TriTrue,
				Succeeded: game.TriTrue,
			}),
		},
	})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}

	obj.ResolutionChoices = map[string]game.ResolutionChoiceResult{
		string(game.EntryTypeChoiceKey): {
			Kind:    game.ResolutionChoiceSubtype,
			Subtype: types.Elf,
		},
	}
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(topID) {
		t.Fatal("matching looked-at card was not moved to hand")
	}
	if g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("matching looked-at card remained in library")
	}
	if !eventRevealedCard(g, topID, obj.ID) {
		t.Fatal("moving the looked-at card did not emit its public reveal event")
	}
}

func TestTriggeredAbilityCapturesSourceEntryChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, []game.Instruction{{
		Primitive: game.LookAtLibraryTop{
			Player:        game.ControllerReference(),
			PublishLinked: "looked",
		},
	}}, nil)
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {
			Kind:    game.ResolutionChoiceSubtype,
			Subtype: types.Elf,
		},
	}
	emitBeginningOfStepEvent(g, game.StepUpkeep)

	NewEngine(nil).putTriggeredAbilitiesOnStack(g)

	trigger, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("upkeep trigger was not put on stack")
	}
	choice, ok := trigger.ResolutionChoices[string(game.EntryTypeChoiceKey)]
	if !ok || choice.Kind != game.ResolutionChoiceSubtype || choice.Subtype != types.Elf {
		t.Fatalf("captured choice = %#v ok = %v, want Elf subtype", choice, ok)
	}
}

// chosenTypeLibraryTopSequence is the exact Herald's Horn upkeep instruction
// sequence the cardgen backend lowers: privately look at the top card, then
// reveal and move it to hand only when it is a creature card of the chosen
// subtype and the controller chooses to.
func chosenTypeLibraryTopSequence() []game.Instruction {
	looked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"}
	return []game.Instruction{
		{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "looked"}},
		{
			Primitive: game.Reveal{Card: looked},
			CardCondition: opt.Val(game.CardCondition{
				Card:              looked,
				Types:             []types.Card{types.Creature},
				ChosenSubtypeFrom: game.EntryTypeChoiceKey,
			}),
			Optional:      true,
			PublishResult: "revealed",
		},
		{
			Primitive:  game.MoveCard{Card: looked, FromZone: zone.Library, Destination: zone.Hand},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "revealed", Accepted: game.TriTrue, Succeeded: game.TriTrue}),
		},
	}
}

func TestLookedAtNonMatchingCardIsLeftOnTop(t *testing.T) {
	cases := map[string]*game.CardDef{
		"wrong subtype": {CardFace: game.CardFace{
			Name:     "Secret Goblin",
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Goblin},
		}},
		"non-creature": {CardFace: game.CardFace{
			Name:  "Secret Relic",
			Types: []types.Card{types.Artifact},
		}},
	}
	for name, def := range cases {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			topID := addCardToLibrary(g, game.Player1, def)
			addInstructionSpellToStack(g, chosenTypeLibraryTopSequence())
			obj, ok := g.Stack.Peek()
			if !ok {
				t.Fatal("spell was not put on stack")
			}
			obj.ResolutionChoices = map[string]game.ResolutionChoiceResult{
				string(game.EntryTypeChoiceKey): {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
			}
			agent := &lookChoiceAgent{}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = agent

			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			if !g.Players[game.Player1].Library.Contains(topID) {
				t.Fatal("non-matching looked-at card left the library")
			}
			if g.Players[game.Player1].Hand.Contains(topID) {
				t.Fatal("non-matching looked-at card was moved to hand")
			}
			if eventRevealedCard(g, topID, obj.ID) {
				t.Fatal("non-matching looked-at card was publicly revealed")
			}
		})
	}
}

func TestLookAtLibraryTopEmptyLibraryIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addInstructionSpellToStack(g, chosenTypeLibraryTopSequence())
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}
	obj.ResolutionChoices = map[string]game.ResolutionChoiceResult{
		string(game.EntryTypeChoiceKey): {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Size() != 0 {
		t.Fatalf("library size = %d, want empty after no-op look", g.Players[game.Player1].Library.Size())
	}
	if len(agent.requests) != 0 {
		t.Fatalf("requests = %#v, want no private look on an empty library", agent.requests)
	}
}
