package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func TestResolutionPaymentPlansSideEffectingManaBeforeDraw(t *testing.T) {
	for _, decline := range []bool{false, true} {
		t.Run(map[bool]string{false: "pay", true: "decline"}[decline], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			addInstructionSpellToStack(g, []game.Instruction{
				{Primitive: game.Pay{Payment: game.ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.G})}}, PublishResult: "paid"},
				{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
					ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue})},
			})
			windows := 0
			agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceMay && decline {
					return []int{0}
				}
				if req.Prompt == "Activate a mana ability while paying?" {
					windows++
					if grove.Tapped || g.Players[game.Player1].Life != 40 {
						t.Fatal("planning applied the Grove's tap/life cost before the payment choice")
					}
				}
				return req.DefaultSelection
			}}
			engine.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
			if decline {
				if windows != 0 || grove.Tapped || g.Players[game.Player1].Life != 40 || g.Players[game.Player1].Hand.Contains(drawn) {
					t.Fatal("declined payment activated Grove or drew a card")
				}
			} else if windows != 1 || !grove.Tapped || g.Players[game.Player1].Life != 39 || !g.Players[game.Player1].Hand.Contains(drawn) {
				t.Fatal("planned Grove did not pay the resolution cost and enable its result")
			}
		})
	}
}

// TestResolutionPaymentEnergyAdditionalCost proves a "you may pay {E}{E}. If you
// do, ..." resolution payment spends the controller's energy counters and gates
// the consequence on the payment, backing the Kaladesh energy cycle's attack and
// enter riders (Thriving Rats, Aether Swooper, and the rest).
func TestResolutionPaymentEnergyAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].EnergyCounters = 3
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:          "Pay {E}{E}?",
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalEnergy, Amount: 2, Text: "Pay {E}{E}"}},
			}},
			PublishResult: "paid",
		},
		{
			Primitive:  game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Accepted: game.TriTrue, Succeeded: game.TriTrue}),
		},
	})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if got := g.Players[game.Player1].EnergyCounters; got != 1 {
		t.Fatalf("energy = %d, want 1 after paying {E}{E}", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want energy payment branch to draw", got)
	}
}
