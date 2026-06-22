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

// reflexiveWhenYouDoInstructions builds the runtime shape produced by the
// reflexive "you may X. When you do, Y" flow: an optional action publishes its
// result and a single trailing effect is gated on that result having
// succeeded. This is the same published/gated envelope the "If you do, Y" rider
// lowers to, so the reflexive trigger gates correctly at runtime. The optional
// gains 2 life; the gated effect draws a card, so both the life total and hand
// reveal whether each step ran.
func reflexiveWhenYouDoInstructions() []game.Instruction {
	gate := opt.Val(game.InstructionResultGate{Key: game.ResultKey("if-you-do"), Succeeded: game.TriTrue})
	return []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: gate,
		},
	}
}

// TestReflexiveWhenYouDoDeclineSkipsGatedEffect verifies that declining the
// optional action of a reflexive "When you do" flow skips the gated trailing
// effect: the controller neither gains life nor draws.
func TestReflexiveWhenYouDoDeclineSkipsGatedEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, reflexiveWhenYouDoInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the optional and gated draw)", got, before)
	}
	if g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("declining the optional must skip the gated draw")
	}
}

// TestReflexiveWhenYouDoAcceptPerformsGatedEffect verifies that accepting the
// optional action of a reflexive "When you do" flow performs the gated trailing
// effect: the controller gains life and draws the gated card.
func TestReflexiveWhenYouDoAcceptPerformsGatedEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, reflexiveWhenYouDoInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+2 {
		t.Fatalf("life = %d, want %d (accepting must gain 2)", got, before+2)
	}
	if !g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("accepting the optional must perform the gated draw")
	}
}

// payLifeIfYouDoInstructions builds the runtime shape produced by lowering "You
// may pay N life. If you do, draw a card.": paying life is losing that much life
// (CR 119.1b), so the optional life loss publishes its result and the gated draw
// runs only when the controller chooses to pay. The optional loses 2 life; the
// gated effect draws a card, so both the life total and hand reveal whether each
// step ran.
func payLifeIfYouDoInstructions() []game.Instruction {
	gate := opt.Val(game.InstructionResultGate{Key: game.ResultKey("if-you-do"), Succeeded: game.TriTrue})
	return []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: gate,
		},
	}
}

// TestOptionalPayLifeIfYouDoDeclineSkips verifies that declining the optional
// "you may pay N life" instruction loses no life and skips the gated draw.
func TestOptionalPayLifeIfYouDoDeclineSkips(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, payLifeIfYouDoInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must pay no life)", got, before)
	}
	if g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("declining the optional pay-life must skip the gated draw")
	}
}

// TestOptionalPayLifeIfYouDoAcceptPerforms verifies that accepting the optional
// "you may pay N life" instruction loses N life and performs the gated draw.
func TestOptionalPayLifeIfYouDoAcceptPerforms(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, payLifeIfYouDoInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before-2 {
		t.Fatalf("life = %d, want %d (accepting must pay 2 life)", got, before-2)
	}
	if !g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("accepting the optional pay-life must perform the gated draw")
	}
}

// ifYouDoElseInstructions builds the "you may X. If you do, Y. Otherwise/If you
// don't, Z." else-branch flow: an optional instruction publishes its result, Y
// is gated on that result having succeeded (TriTrue), and the else effect Z is
// gated on the exact complement — the result having failed (TriFalse) — so
// exactly one of Y/Z runs. The optional gains 1 life; Y draws a card; Z gains 5
// life, so the life total and hand reveal which branch ran.
func ifYouDoElseInstructions() []game.Instruction {
	key := game.ResultKey("if-you-do")
	return []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			PublishResult: key,
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: key, Succeeded: game.TriTrue}),
		},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(5)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: key, Succeeded: game.TriFalse}),
		},
	}
}

// TestOptionalIfYouDoElseAcceptRunsThenBranch verifies that accepting the
// optional "you may X" runs the "if you do" branch Y and skips the else branch
// Z: the controller gains 1 (optional) and draws (Y), but does not gain the 5
// life from Z.
func TestOptionalIfYouDoElseAcceptRunsThenBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, ifYouDoElseInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+1 {
		t.Fatalf("life = %d, want %d (accepting gains 1 and runs Y, not the else +5)", got, before+1)
	}
	if !g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("accepting must run the if-you-do branch (the gated draw)")
	}
}

// TestOptionalIfYouDoElseDeclineRunsElseBranch verifies that declining the
// optional "you may X" runs the else branch Z and skips the "if you do" branch
// Y: the controller does not gain the optional life and does not draw (Y), but
// gains the 5 life from the else branch Z.
func TestOptionalIfYouDoElseDeclineRunsElseBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	gatedDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, ifYouDoElseInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != before+5 {
		t.Fatalf("life = %d, want %d (declining skips Y and runs the else +5)", got, before+5)
	}
	if g.Players[game.Player1].Hand.Contains(gatedDraw) {
		t.Fatal("declining must skip the if-you-do branch (the gated draw)")
	}
}
