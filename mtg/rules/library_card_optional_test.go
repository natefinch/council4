package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledObservedCardOptionalRevealAndMoveUsesOneDecision(t *testing.T) {
	for _, subject := range []struct {
		noun string
		face game.CardFace
	}{
		{"land card", game.CardFace{Types: []types.Card{types.Land}}},
		{"snow card", game.CardFace{Types: []types.Card{types.Land}, Supertypes: []types.Super{types.Snow}}},
		{"Zombie card", game.CardFace{Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Zombie}}},
		{"instant or sorcery card", game.CardFace{Types: []types.Card{types.Instant}}},
	} {
		t.Run(subject.noun, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Observed Optional Group", Layout: "normal", TypeLine: "Artifact",
				OracleText: "{T}: Look at the top card of your library. If it's a " + subject.noun +
					", you may reveal it and put it into your hand. You gain 1 life.",
			})
			for _, outcome := range []string{"accept", "decline", "empty", "wrong card"} {
				t.Run(outcome, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					observed := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: subject.face})
					decoy := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: subject.face})
					if outcome == "empty" {
						g.Players[game.Player1].Library.Remove(observed)
					}
					if outcome == "wrong card" {
						g.CardInstances[observed].Def = &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}}
					}
					agent := &scopedMayAgent{accept: []bool{outcome == "accept"}}
					obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ActivatedAbilities[0].Content,
						[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
					wantPrompts := 0
					if outcome == "accept" || outcome == "decline" {
						wantPrompts = 1
					}
					accepted := outcome == "accept"
					if agent.next != wantPrompts || g.Players[game.Player1].Life != 41 ||
						g.Players[game.Player1].Hand.Contains(observed) != accepted ||
						eventRevealedCard(g, observed, obj.ID) != accepted ||
						!g.Players[game.Player3].Library.Contains(decoy) {
						t.Fatalf("outcome=%s prompts=%d, want %d; hand=%v reveal=%v life=%d",
							outcome, agent.next, wantPrompts, g.Players[game.Player1].Hand.Contains(observed),
							eventRevealedCard(g, observed, obj.ID), g.Players[game.Player1].Life)
					}
				})
			}
		})
	}
}

func TestOptionalLibraryObservationDoesNotGateIndependentRider(t *testing.T) {
	for _, accepted := range []bool{false, true} {
		for _, empty := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			if !empty {
				addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			}
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player3)},
			}
			var agents [game.NumPlayers]PlayerAgent
			if !accepted {
				agents[game.Player1] = &declineChoiceAgent{}
			}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.Reveal{
					Player: game.TargetPlayerReference(0), Amount: game.Fixed(1), PublishLinked: "observed",
				},
				Optional: true,
			}, agents, &TurnLog{})
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})}),
			}, agents, &TurnLog{})
			engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
			}, agents, &TurnLog{})
			want := 22
			if accepted && !empty {
				want = 26
			}
			if actual := g.Players[game.Player1].Life; actual != want {
				t.Fatalf("accepted=%v empty=%v: life=%d, want %d", accepted, empty, actual, want)
			}
		}
	}
}
