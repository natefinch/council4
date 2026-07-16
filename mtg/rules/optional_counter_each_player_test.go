package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const testOptionalCounterLink game.LinkedKey = "test-optional-counter"

func optionalCounterGoadInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.OptionalCounterForEachPlayer{
			Players:       game.AllPlayersReference(),
			Selection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Amount:        game.Fixed(2),
			CounterKind:   counter.PlusOnePlusOne,
			PublishLinked: testOptionalCounterLink,
		}},
		{Primitive: game.Goad{
			Group:         game.LinkedObjectsGroup(testOptionalCounterLink),
			ConsumeLinked: true,
		}},
	}
}

type optionalCounterAgent struct {
	player    game.PlayerID
	accept    bool
	permanent int
	order     *[]game.PlayerID
	requests  int
	onChoose  func()
}

func (*optionalCounterAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *optionalCounterAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests++
	switch request.Kind {
	case game.ChoiceMay:
		if a.order != nil {
			*a.order = append(*a.order, a.player)
		}
		if a.accept {
			return []int{1}
		}
		return []int{0}
	case game.ChoicePayment:
		if a.onChoose != nil {
			a.onChoose()
		}
		return []int{a.permanent}
	default:
		return request.DefaultSelection
	}
}

func TestOptionalCounterForEachPlayerUsesAPNAPAndMemberChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player2
	engine := NewEngine(nil)
	p1a := addCombatCreaturePermanent(g, game.Player1)
	p1b := addCombatCreaturePermanent(g, game.Player1)
	p2 := addCombatCreaturePermanent(g, game.Player2)
	p2.Controller = game.Player3
	p3 := addCombatCreaturePermanent(g, game.Player3)
	var order []game.PlayerID
	p2Agent := &optionalCounterAgent{player: game.Player2, accept: true, order: &order}
	p4Agent := &optionalCounterAgent{player: game.Player4, accept: true, order: &order}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &optionalCounterAgent{player: game.Player1, accept: true, permanent: 1, order: &order},
		game.Player2: p2Agent,
		game.Player3: &optionalCounterAgent{player: game.Player3, accept: true, permanent: 1, order: &order},
		game.Player4: p4Agent,
	}
	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !slices.Equal(order, []game.PlayerID{game.Player3, game.Player1}) {
		t.Fatalf("offer order = %v, want Player3 then Player1; players without creatures must not be asked", order)
	}
	if p2Agent.requests != 0 || p4Agent.requests != 0 {
		t.Fatalf("players without creatures received requests: Player2=%d Player4=%d", p2Agent.requests, p4Agent.requests)
	}
	if got := p1a.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("Player1 first creature counters = %d, want 0", got)
	}
	for name, permanent := range map[string]*game.Permanent{"Player1 second": p1b, "Player3 own": p3} {
		if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
			t.Fatalf("%s creature counters = %d, want 2", name, got)
		}
		if _, ok := permanent.Goaded[game.Player1]; !ok {
			t.Fatalf("%s creature was not goaded by source controller", name)
		}
	}
	if got := p2.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("control-changed creature counters = %d, want 0 after Player3 chose its other creature", got)
	}
}

func TestOptionalCounterForEachPlayerUsesActingPlayerForReplacements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	replacer := addReplacementPermanent(t, g, game.Player2, anyCounterDoublingReplacementCardDef())
	chosen := addCombatCreaturePermanent(g, game.Player2)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &optionalCounterAgent{player: game.Player2, accept: true, permanent: 1},
	}
	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := chosen.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("chosen counters = %d, want 4 from Player2's replacement", got)
	}
	if got := replacer.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("replacement creature counters = %d, want 0", got)
	}
	if _, ok := chosen.Goaded[game.Player1]; !ok {
		t.Fatal("replacement-modified placement was not linked and goaded")
	}
}

func TestOptionalCounterDoesNotLinkWhenReplacementPreventsPlacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	preventer := &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Prevention",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.AnyCounterPlacementReplacement(
				"If you would put two counters on a permanent, put no counters on it instead.",
				0,
				-2,
				game.TriggerControllerYou,
			),
		},
	}}
	addReplacementPermanent(t, g, game.Player2, preventer)
	creature := addCombatCreaturePermanent(g, game.Player2)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &optionalCounterAgent{player: game.Player2, accept: true},
	}
	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("prevented counters = %d, want 0", got)
	}
	if _, ok := creature.Goaded[game.Player1]; ok {
		t.Fatal("creature with prevented counter placement was goaded")
	}
	if len(g.LinkedObjects) != 0 {
		t.Fatalf("prevented placement was linked: %#v", g.LinkedObjects)
	}
}

func TestOptionalCounterLinkedGoadDoesNotLeakAcrossResolutions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanent(g, game.Player2)
	accept := &optionalCounterAgent{player: game.Player2, accept: true}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: accept}

	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	delete(creature.Goaded, game.Player1)
	accept.accept = false
	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if _, ok := creature.Goaded[game.Player1]; ok {
		t.Fatal("declined second resolution re-goaded stale linked creature")
	}
	if len(g.LinkedObjects) != 0 {
		t.Fatalf("linked objects remain after consuming goad: %#v", g.LinkedObjects)
	}
}

func TestOptionalCounterDoesNotLinkCreatureThatLeavesDuringChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanent(g, game.Player2)
	agent := &optionalCounterAgent{player: game.Player2, accept: true}
	agent.onChoose = func() {
		movePermanentToZone(g, creature, zone.Graveyard)
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	addInstructionSpellToStack(g, optionalCounterGoadInstructions())
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("departed creature counters = %d, want 0", got)
	}
	if len(g.LinkedObjects) != 0 {
		t.Fatalf("departed creature was linked: %#v", g.LinkedObjects)
	}
}
