package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type libraryPaymentAgent struct {
	accept   bool
	prompts  []game.ChoiceRequest
	afterMay func()
}

func (*libraryPaymentAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *libraryPaymentAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		a.prompts = append(a.prompts, request)
		if a.afterMay != nil {
			a.afterMay()
		}
		if a.accept {
			return []int{1}
		}
	}
	return []int{0}
}

func TestObservedCardLifePaymentUsesActualControllerPayment(t *testing.T) {
	for _, tc := range []struct {
		name      string
		life      int
		accept    bool
		forbidden bool
		empty     bool
		land      bool
		change    bool
		prompts   int
		accepted  bool
		succeeded bool
		wantLife  int
	}{
		{name: "paid", life: 20, accept: true, prompts: 1, accepted: true, succeeded: true, wantLife: 18},
		{name: "exact life", life: 2, accept: true, prompts: 1, accepted: true, succeeded: true, wantLife: 0},
		{name: "insufficient", life: 1, accept: true, wantLife: 1},
		{name: "forbidden", life: 20, accept: true, forbidden: true, wantLife: 20},
		{name: "declined", life: 20, prompts: 1, wantLife: 20},
		{name: "empty library", life: 20, accept: true, empty: true, wantLife: 20},
		{name: "wrong card type", life: 20, accept: true, land: true, wantLife: 20},
		{name: "accepted but payment failed", life: 20, accept: true, change: true, prompts: 1, accepted: true, wantLife: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = tc.life
			g.Players[game.Player3].Life = 30
			decoy := addCardToLibrary(g, game.Player1, vanillaCreature("Decoy", 1, 1))
			cardID := addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			if tc.empty {
				g.Players[game.Player3].Library.Remove(cardID)
			}
			if tc.land {
				g.CardInstances[cardID].Def.Types = []types.Card{types.Land}
			}
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange,
					Controller: game.Player1, AffectedPlayer: game.PlayerYou,
				})
			}
			controller, owner := &libraryPaymentAgent{accept: tc.accept}, &libraryPaymentAgent{accept: true}
			if tc.change {
				controller.afterMay = func() { g.Players[game.Player1].Life = 1 }
			}
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
			}
			resolver := newEffectResolver(NewEngine(nil), g, obj,
				[game.NumPlayers]PlayerAgent{game.Player1: controller, game.Player3: owner}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.LookAtLibraryTop{
				Player: game.TargetPlayerReference(1), PublishLinked: "observed",
			}})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.Pay{Payment: game.ResolutionPayment{
					Payer:           opt.Val(game.ControllerReference()),
					AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}},
				}},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
				})}),
				PublishResult: "paid",
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.MoveCard{
					Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
					FromZone: zone.Library, Destination: zone.Graveyard,
				},
				ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue}),
			})
			if g.Players[game.Player1].Life != tc.wantLife || g.Players[game.Player3].Life != 30 {
				t.Fatalf("payment used wrong payer or amount: controller=%d owner=%d",
					g.Players[game.Player1].Life, g.Players[game.Player3].Life)
			}
			if len(controller.prompts) != tc.prompts || len(owner.prompts) != 0 {
				t.Fatalf("payment prompts: controller=%d owner=%d, want %d/0",
					len(controller.prompts), len(owner.prompts), tc.prompts)
			}
			result := obj.ResolutionResults["paid"]
			if result.Accepted != tc.accepted || result.Succeeded != tc.succeeded {
				t.Fatalf("payment receipt=%#v, want accepted=%v succeeded=%v", result, tc.accepted, tc.succeeded)
			}
			if g.Players[game.Player3].Graveyard.Contains(cardID) != tc.succeeded ||
				g.Players[game.Player1].Graveyard.Contains(cardID) ||
				!g.Players[game.Player1].Library.Contains(decoy) {
				t.Fatal("payment moved a decoy, guessed a destination owner, or treated acceptance as success")
			}
		})
	}
}
