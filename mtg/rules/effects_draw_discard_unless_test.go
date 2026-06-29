package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// drawDiscardUnlessAgent answers the discard-unless-type choices: pathChoice is
// returned for the menu picking the exempt branch versus the full discard, and
// cardOrder names the cards to discard by name for the card-selection prompts.
type drawDiscardUnlessAgent struct {
	pathChoice []int
	cardOrder  []string
	pathAsked  bool
}

func (*drawDiscardUnlessAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Action{}
}

func (a *drawDiscardUnlessAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	hasCardOptions := false
	for _, option := range request.Options {
		if option.Card.Exists {
			hasCardOptions = true
			break
		}
	}
	if !hasCardOptions && a.pathChoice != nil && !a.pathAsked {
		a.pathAsked = true
		return append([]int(nil), a.pathChoice...)
	}
	selected := make([]int, 0, len(a.cardOrder))
	for _, name := range a.cardOrder {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.Name == name {
				selected = append(selected, option.Index)
				break
			}
		}
	}
	return selected
}

func drawDiscardUnlessInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
		{Primitive: game.DiscardUnlessType{
			Player:      game.ControllerReference(),
			Amount:      2,
			ExemptTypes: []types.Card{types.Creature},
		}},
	}
}

// TestDiscardUnlessTypeExemptDiscardsSingleMatching proves that when the player
// chooses the exempt branch they discard exactly one matching card, satisfying
// the effect without the full discard (Thirst for Identity's creature waiver).
func TestDiscardUnlessTypeExemptDiscardsSingleMatching(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	for _, name := range []string{"Top 1", "Top 2", "Top 3"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}})
	}
	addInstructionSpellToStack(g, drawDiscardUnlessInstructions())

	agent := &drawDiscardUnlessAgent{pathChoice: []int{0}, cardOrder: []string{"Bear"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Drew 3 (hand 1 -> 4), then discarded 1 creature (hand 4 -> 3).
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("hand size = %d, want 3", got)
	}
}

// TestDiscardUnlessTypeFullDiscardWhenChosen proves that choosing the full
// branch discards the fixed count even when an exempt card is held.
func TestDiscardUnlessTypeFullDiscardWhenChosen(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	for _, name := range []string{"Top 1", "Top 2", "Top 3"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}})
	}
	addInstructionSpellToStack(g, drawDiscardUnlessInstructions())

	// Choose the full-discard branch, then discard two of the drawn sorceries.
	agent := &drawDiscardUnlessAgent{pathChoice: []int{1}, cardOrder: []string{"Top 1", "Top 2"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Drew 3 (hand 1 -> 4), then discarded 2 (hand 4 -> 2).
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2", got)
	}
}

// TestDiscardUnlessTypeNoExemptDiscardsFull proves that without any exempt card
// the player discards the full fixed count.
func TestDiscardUnlessTypeNoExemptDiscardsFull(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, name := range []string{"Top 1", "Top 2", "Top 3"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}})
	}
	addInstructionSpellToStack(g, drawDiscardUnlessInstructions())

	agent := &drawDiscardUnlessAgent{cardOrder: []string{"Top 1", "Top 2"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Drew 3 (hand 0 -> 3), then discarded 2 (hand 3 -> 1).
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1", got)
	}
}
