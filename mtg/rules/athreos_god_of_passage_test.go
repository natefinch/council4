package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const athreosTargetPaidKey = game.ResultKey("target-player-paid")

func resolveAthreosReturn(t *testing.T, accept bool, targetLife int) (*game.Game, *game.StackObject, game.CardReference) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Life = targetLife
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Returned Creature",
		Types: []types.Card{types.Creature},
	}})
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("event card missing")
	}
	if card.ZoneVersion == 0 {
		card.ZoneVersion = 1
	}
	eventCard := game.CardReference{Kind: game.CardReferenceEvent}
	obj := &game.StackObject{
		Kind:       game.StackTriggeredAbility,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
		TriggerEvent: game.Event{
			Kind:            game.EventPermanentDied,
			CardID:          cardID,
			CardZoneVersion: card.ZoneVersion,
			FromZone:        zone.Battlefield,
			ToZone:          zone.Graveyard,
		},
		HasTriggerEvent: true,
	}
	sequence := []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt: "Pay 3 life?",
				Payer:  opt.Val(game.TargetPlayerReference(0)),
				AdditionalCosts: []cost.Additional{{
					Kind:   cost.AdditionalPayLife,
					Text:   "3 life",
					Amount: 3,
				}},
			}},
			PublishResult: athreosTargetPaidKey,
		},
		{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       athreosTargetPaidKey,
				Succeeded: game.TriFalse,
			}),
		},
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: optionalMayAgent{accept: accept},
	}
	resolver := newEffectResolver(NewEngine(nil), g, obj, agents, &TurnLog{})
	for i := range sequence {
		resolver.resolveInstruction(&sequence[i])
	}
	return g, obj, eventCard
}

func TestAthreosTargetOpponentPaysLifeToStopReturn(t *testing.T) {
	g, _, _ := resolveAthreosReturn(t, true, 40)
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("target opponent life = %d, want 37", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player1].Graveyard.Size() != 1 {
		t.Fatal("event card moved despite the target opponent paying")
	}
}

func TestAthreosDeclinedOrUnpayablePaymentReturnsEventCard(t *testing.T) {
	for _, test := range []struct {
		name       string
		accept     bool
		targetLife int
	}{
		{name: "declined", accept: false, targetLife: 40},
		{name: "unpayable", accept: true, targetLife: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, obj, _ := resolveAthreosReturn(t, test.accept, test.targetLife)
			if g.Players[game.Player2].Life != test.targetLife {
				t.Fatalf("target opponent life = %d, want %d", g.Players[game.Player2].Life, test.targetLife)
			}
			if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player1].Graveyard.Size() != 0 {
				t.Fatalf(
					"event card was not returned after a failed payment: results=%#v hand=%d graveyard=%d",
					obj.ResolutionResults,
					g.Players[game.Player1].Hand.Size(),
					g.Players[game.Player1].Graveyard.Size(),
				)
			}
		})
	}
}
