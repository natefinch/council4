package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestOptionalEnteredObjectAcrossInterveningActions(t *testing.T) {
	for _, body := range []struct {
		name, text      string
		groupLife, life int
	}{
		{"group life", "you may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste until end of turn.", 1, 0},
		{"expanded group", "you may return target creature card from your graveyard to the battlefield and add {R}{G}. It gains haste until end of turn.", 0, 0},
		{"unconditional life", "you may return target creature card from your graveyard to the battlefield and gain 1 life. You gain 2 life. It gains haste until end of turn.", 1, 2},
		{"expanded unconditional", "you may return target creature card from your graveyard to the battlefield and gain 1 life. Add {R}{G}. It gains haste until end of turn.", 1, 0},
	} {
		t.Run(body.name, func(t *testing.T) {
			for _, state := range []struct {
				name                           string
				accept, failed, falseCondition bool
			}{
				{"accepted", true, false, false},
				{"declined", false, false, false},
				{"failed", true, true, false},
				{"false", false, false, true},
			} {
				t.Run(state.name, func(t *testing.T) {
					content := compiledScopedResultContent(t, "Tap target artifact. If you have no cards in hand, "+body.text)
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
					cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
					target := currentCardTarget(t, g, cardID)
					if state.failed {
						g.Players[game.Player1].Graveyard.Remove(cardID)
					}
					if state.falseCondition {
						addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
					}
					agent := &scopedMayAgent{accept: []bool{state.accept}}
					agents := [game.NumPlayers]PlayerAgent{}
					agents[game.Player1] = agent
					obj := &game.StackObject{SourceID: artifact.ObjectID, Controller: game.Player1,
						Targets: []game.Target{game.PermanentTarget(artifact.ObjectID), target}}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
					entered, exists := reanimatedPermanent(g, cardID)
					wantEntered := state.accept && !state.failed && !state.falseCondition
					if exists != wantEntered || exists && !hasKeyword(g, entered, game.Haste) {
						t.Fatal("rider did not consume the actual earlier entered-object publication")
					}
					wantLife := 40 + body.life
					if state.accept && !state.falseCondition {
						wantLife += body.groupLife
					}
					wantChoices := 1
					if state.falseCondition {
						wantChoices = 0
					}
					if g.Players[game.Player1].Life != wantLife || agent.next != wantChoices ||
						!artifact.Tapped || hasKeyword(g, artifact, game.Haste) {
						t.Fatalf("life=%d choices=%d, want %d/%d; actor/target isolation lost",
							g.Players[game.Player1].Life, agent.next, wantLife, wantChoices)
					}
				})
			}
		})
	}
}
