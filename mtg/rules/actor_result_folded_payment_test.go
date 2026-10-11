package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func TestCompiledFoldedPlayerPaymentActualResult(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		event      game.Event
		cost       int
	}{
		{"draw", "Whenever a player casts a spell, that player may pay {2}. If the player does, they draw a card.",
			game.Event{Kind: game.EventSpellCast, Controller: game.Player2, Player: game.Player4}, 2},
		{"treasure", "Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token.",
			game.Event{Kind: game.EventCardDrawn, Player: game.Player2, Controller: game.Player4}, 2},
		{"attached untap", "At the beginning of the upkeep of enchanted creature's controller, that player may pay {4}. If the player does, untap the creature.",
			game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2, Controller: game.Player4}, 4},
	} {
		def := compileCounterTaxCard(t, "Folded Player Cost", tc.text, "Enchantment")
		content := def.TriggeredAbilities[0].Content
		sequence := content.Modes[0].Sequence
		payment, ok := sequence[0].Primitive.(game.Pay)
		if !ok || !payment.Payment.Payer.Exists ||
			payment.Payment.Payer.Val != game.EventPlayerReference() || sequence[0].PublishResult == "" {
			t.Fatal("folded payment lost its exact event-player receipt")
		}
		key := sequence[0].PublishResult
		for _, choice := range []struct {
			name   string
			accept bool
			mana   int
		}{
			{"paid", true, 10},
			{"declined", false, 10},
			{"unaffordable", true, 0},
		} {
			t.Run(tc.name+"/"+choice.name, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player3, def)
				source.Controller = game.Player1
				attached := addCombatPermanent(g, game.Player4, vanillaCreature("Attached object", 2, 2))
				attached.Controller, attached.Tapped = game.Player2, true
				source.AttachedTo = opt.Val(attached.ObjectID)
				drawn := addCardToLibrary(g, game.Player2, vanillaCreature("Paid draw", 1, 1))
				addCardToLibrary(g, game.Player1, vanillaCreature("Foreign draw", 1, 1))
				g.Players[game.Player2].ManaPool.Add(mana.C, choice.mana)
				g.Players[game.Player1].ManaPool.Add(mana.C, 20)
				obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID,
					SourceCardID: source.CardInstanceID, HasTriggerEvent: true, TriggerEvent: tc.event}
				payer := &libraryPaymentAgent{accept: choice.accept}
				foreign := &libraryPaymentAgent{accept: !choice.accept}
				trigger := def.TriggeredAbilities[0]
				trigger.Content = withResultProbes(content, []resultProbe{
					{key: key, accepted: game.TriTrue, life: 1},
					{key: key, succeeded: game.TriTrue, life: 2},
				})
				obj.ID, obj.Kind, obj.InlineTrigger = g.IDGen.Next(), game.StackTriggeredAbility, &trigger
				g.Stack.Push(obj)
				NewEngine(nil).resolveTopOfStackWithChoices(g,
					[game.NumPlayers]PlayerAgent{game.Player1: foreign, game.Player2: payer}, &TurnLog{})
				paid := choice.accept && choice.mana >= tc.cost
				wantMana, wantLife := choice.mana, 40
				if paid {
					wantMana, wantLife = choice.mana-tc.cost, 43
				}
				tokens := 0
				for _, permanent := range g.Battlefield {
					if permanent.Token {
						if permanent.Controller != game.Player1 {
							t.Fatal("foreign token recipient")
						}
						tokens++
					}
				}
				wantTokens := 0
				if tc.name == "treasure" && !paid {
					wantTokens = 1
				}
				if g.Players[game.Player2].ManaPool.Total() != wantMana ||
					g.Players[game.Player1].ManaPool.Total() != 20 || g.Players[game.Player1].Life != wantLife ||
					g.Players[game.Player2].Life != 40 || g.Players[game.Player3].Life != 40 || g.Players[game.Player4].Life != 40 ||
					g.Players[game.Player2].Hand.Contains(drawn) != (tc.name == "draw" && paid) ||
					g.Players[game.Player1].Hand.Size() != 0 || tokens != wantTokens ||
					attached.Tapped != (tc.name != "attached untap" || !paid) || len(foreign.prompts) != 0 {
					t.Fatal("folded actor, payment receipt or consequence domain changed")
				}
			})
		}
	}
}
