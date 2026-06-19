package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/opt"
)

// optionalMayAgent answers the optional "Apply optional effect?" may-choice
// according to accept and passes on every action.
type optionalMayAgent struct {
	accept bool
}

func (optionalMayAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a optionalMayAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay && a.accept {
		return []int{1}
	}
	return []int{0}
}

// TestOptionalGainLifeDeclineSkips verifies that declining an Optional GainLife
// instruction leaves the controller's life total unchanged.
func TestOptionalGainLifeDeclineSkips(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional:  true,
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the gain)", got, before)
	}
}

// TestOptionalGainLifeAcceptPerforms verifies that accepting an Optional
// GainLife instruction adds the life to the controller's total.
func TestOptionalGainLifeAcceptPerforms(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional:  true,
		Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+3 {
		t.Fatalf("life = %d, want %d (accepting must gain 3)", got, before+3)
	}
}

// ifYouDoMultiGateInstructions builds the "you may X. If you do, Y and Z" flow
// where a single optional instruction publishes its result and two trailing
// instructions are both gated on that result having succeeded. The optional
// gains 1 life; the gated effects draw a card and gain 5 more life, so both the
// life total and hand reveal whether each gated effect ran.
func ifYouDoMultiGateInstructions() []game.Instruction {
	gate := opt.Val(game.InstructionResultGate{Key: game.ResultKey("if-you-do"), Succeeded: game.TriTrue})
	return []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: gate,
		},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(5)},
			ResultGate: gate,
		},
	}
}

// TestOptionalIfYouDoMultiGateDeclineSkipsAllGatedEffects verifies that
// declining the optional "you may X" instruction skips every effect gated on its
// result ("If you do, Y and Z"): neither gated effect runs because the optional
// did not succeed.
func TestOptionalIfYouDoMultiGateDeclineSkipsAllGatedEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, ifYouDoMultiGateInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the optional and both gated effects)", got, before)
	}
	if g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("declining the optional must skip the gated draw")
	}
}

// TestOptionalIfYouDoMultiGateAcceptPerformsAllGatedEffects verifies that
// accepting the optional "you may X" instruction performs every effect gated on
// its result: both the gated draw and the gated life gain happen.
func TestOptionalIfYouDoMultiGateAcceptPerformsAllGatedEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, ifYouDoMultiGateInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+6 {
		t.Fatalf("life = %d, want %d (accepting gains 1 + gated 5)", got, before+6)
	}
	if !g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("accepting the optional must perform the gated draw")
	}
}
