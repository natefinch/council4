package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// These tests pin down the safety property behind ADR 0002 (smart priority
// skips empty priority passes): priority is offered exactly when a legal
// instant-speed response exists, and is skipped only when no player can act.
// If the legal-action generation that smart priority relies on ever stops
// surfacing a real response, these tests fail.

// instantNoTarget is a {G} instant with no targets, so it is castable purely on
// timing and mana without needing a creature on the board.
func instantNoTarget(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Instant},
	}}
}

// sorceryNoTarget is the sorcery-speed counterpart of instantNoTarget.
func sorceryNoTarget(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Sorcery},
	}}
}

// stackSpell puts a bare spell on the stack controlled by a player, modeling
// something an opponent might respond to.
func (s *scenario) stackSpell(controller game.PlayerID) *scenario {
	s.g.Stack.Push(&game.StackObject{
		ID:         s.g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: controller,
	})
	return s
}

// stackTrigger puts a triggered ability on the stack controlled by a player.
func (s *scenario) stackTrigger(controller game.PlayerID) *scenario {
	s.g.Stack.Push(&game.StackObject{
		ID:         s.g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		Controller: controller,
	})
	return s
}

// mana gives a player one mana of the given color.
func (s *scenario) mana(player game.PlayerID, color mana.Color, amount int) *scenario {
	s.g.Players[player].ManaPool.Add(color, amount)
	return s
}

func castActionFor(actions []action.Action, cardID id.ID) bool {
	for _, act := range actions {
		if act.Kind != action.ActionCastSpell {
			continue
		}
		payload, ok := act.CastSpellPayload()
		if ok && payload.CardID == cardID {
			return true
		}
	}
	return false
}

func anyNonPassAction(actions []action.Action) bool {
	for _, act := range actions {
		if act.Kind != action.ActionPass {
			return true
		}
	}
	return false
}

// TestPriorityOfferedWhenInstantResponseAvailable proves a player with a
// castable instant and the mana to pay for it is offered the cast while a spell
// sits on the stack — priority must not be skipped here.
func TestPriorityOfferedWhenInstantResponseAvailable(t *testing.T) {
	s := newScenario(t)
	s.stackSpell(game.Player1)
	s.g.Turn.PriorityPlayer = game.Player2
	instantID := s.hand(game.Player2, instantNoTarget("Reflex"))
	s.mana(game.Player2, mana.G, 1)

	legal := s.legalActions(game.Player2)

	if !castActionFor(legal, instantID) {
		t.Fatalf("a player holding a payable instant must be offered its cast in response; legal = %+v", legal)
	}
}

// TestPriorityNotOfferedWithoutMana proves no cast is offered when the instant
// cannot be paid for: there is genuinely no legal response, so skipping it hides
// nothing.
func TestPriorityNotOfferedWithoutMana(t *testing.T) {
	s := newScenario(t)
	s.stackSpell(game.Player1)
	s.g.Turn.PriorityPlayer = game.Player2
	instantID := s.hand(game.Player2, instantNoTarget("Reflex"))

	legal := s.legalActions(game.Player2)

	if castActionFor(legal, instantID) {
		t.Fatalf("an unpayable instant must not be offered as a response; legal = %+v", legal)
	}
}

// TestPriorityNotOfferedForSorceryResponse proves a sorcery is not castable at
// instant speed in response to a spell, even with mana available.
func TestPriorityNotOfferedForSorceryResponse(t *testing.T) {
	s := newScenario(t)
	s.stackSpell(game.Player1)
	s.g.Turn.PriorityPlayer = game.Player2
	sorceryID := s.hand(game.Player2, sorceryNoTarget("Slowpoke"))
	s.mana(game.Player2, mana.G, 1)

	legal := s.legalActions(game.Player2)

	if castActionFor(legal, sorceryID) {
		t.Fatalf("a sorcery must not be castable in response to a spell; legal = %+v", legal)
	}
}

// TestPriorityOfferedInResponseToTriggeredAbility proves the response window is
// granted over a triggered ability on the stack, not only over spells.
func TestPriorityOfferedInResponseToTriggeredAbility(t *testing.T) {
	s := newScenario(t)
	s.stackTrigger(game.Player1)
	s.g.Turn.PriorityPlayer = game.Player2
	instantID := s.hand(game.Player2, instantNoTarget("Reflex"))
	s.mana(game.Player2, mana.G, 1)

	legal := s.legalActions(game.Player2)

	if !castActionFor(legal, instantID) {
		t.Fatalf("a player must be able to respond to a triggered ability with an instant; legal = %+v", legal)
	}
}

// offerRecordingAgent records every legal-action set it is offered and always
// passes, so a priority loop can be inspected for which responses were exposed.
type offerRecordingAgent struct {
	offered [][]action.Action
}

func (a *offerRecordingAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	snapshot := make([]action.Action, len(legal))
	copy(snapshot, legal)
	a.offered = append(a.offered, snapshot)
	return actionBuild.pass()
}

func (a *offerRecordingAgent) sawNonPass() bool {
	return slices.ContainsFunc(a.offered, anyNonPassAction)
}

// TestPriorityLoopGrantsResponseWindowWithInstant runs the full priority loop
// and proves the loop itself exposes the instant to its holder before the stack
// resolves — the response window is not optimized away.
func TestPriorityLoopGrantsResponseWindowWithInstant(t *testing.T) {
	s := newScenario(t)
	s.stackSpell(game.Player1)
	instantID := s.hand(game.Player2, instantNoTarget("Reflex"))
	s.mana(game.Player2, mana.G, 1)

	responder := &offerRecordingAgent{}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: responder}

	s.engine().runPriorityLoop(s.game(), agents, &TurnLog{})

	if !s.game().Stack.IsEmpty() {
		t.Fatal("stack should resolve once every player passes")
	}
	sawInstant := false
	for _, legal := range responder.offered {
		if castActionFor(legal, instantID) {
			sawInstant = true
			break
		}
	}
	if !sawInstant {
		t.Fatalf("the responder was never offered its instant during the priority loop; offers = %+v", responder.offered)
	}
}

// TestPriorityLoopSkipsWhenNoResponseExists runs the full priority loop with no
// player able to respond and proves no non-pass action is ever offered, yet the
// stack still resolves — priority is correctly skipped without hiding anything.
func TestPriorityLoopSkipsWhenNoResponseExists(t *testing.T) {
	s := newScenario(t)
	s.stackSpell(game.Player1)

	var recorders [game.NumPlayers]*offerRecordingAgent
	var agents [game.NumPlayers]PlayerAgent
	for i := range recorders {
		recorders[i] = &offerRecordingAgent{}
		agents[i] = recorders[i]
	}

	s.engine().runPriorityLoop(s.game(), agents, &TurnLog{})

	if !s.game().Stack.IsEmpty() {
		t.Fatal("stack should resolve once every player passes")
	}
	for i, recorder := range recorders {
		if recorder.sawNonPass() {
			t.Fatalf("player %d was offered a response when none was legal; offers = %+v", i, recorder.offered)
		}
	}
}
