package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledEventPlayerFixedLifeNegativeActualResult(t *testing.T) {
	testCompiledEventPlayerFixedLifeActualResult(t, false)
}

func TestCompiledEventPlayerFixedLifePositiveActualResult(t *testing.T) {
	testCompiledEventPlayerFixedLifeActualResult(t, true)
}

func testCompiledEventPlayerFixedLifeActualResult(t *testing.T, positive bool) {
	t.Helper()
	text := "Whenever an opponent draws a card, that player may pay 2 life. If the player doesn't, you create a Treasure token."
	event := game.Event{Kind: game.EventCardDrawn, Player: game.Player2, Controller: game.Player4}
	gate := game.TriFalse
	if positive {
		text = "Whenever a player casts a spell, that player may pay 2 life. If the player does, they draw a card."
		event = game.Event{Kind: game.EventSpellCast, Controller: game.Player2, Player: game.Player4}
		gate = game.TriTrue
	}
	def := compileCounterTaxCard(t, "Player Cost Result", text, "Enchantment")
	content := def.TriggeredAbilities[0].Content
	sequence := content.Modes[0].Sequence
	payment, ok := sequence[0].Primitive.(game.Pay)
	if !ok || !payment.Payment.Payer.Exists || payment.Payment.Payer.Val != game.EventPlayerReference() ||
		len(payment.Payment.AdditionalCosts) != 1 || payment.Payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
		payment.Payment.AdditionalCosts[0].Amount != 2 || payment.Payment.ManaCost.Exists {
		t.Fatal("fixed life action changed its payer, cost or currency")
	}
	key := sequence[0].PublishResult
	if key == "" || !sequence[1].ResultGate.Exists ||
		sequence[1].ResultGate.Val.Key != key || sequence[1].ResultGate.Val.Succeeded != gate {
		t.Fatal("printed antecedent does not consume the exact actual Pay outcome")
	}
	for _, tc := range []struct {
		name                         string
		accept, forbidden, protected bool
		departed, copied             bool
		life                         int
	}{
		{name: "paid", accept: true, life: 40},
		{name: "declined", life: 40},
		{name: "unaffordable", accept: true, life: 1},
		{name: "forbidden", accept: true, life: 40, forbidden: true},
		{name: "paid protected observer", accept: true, life: 40, protected: true},
		{name: "declined protected observer", life: 40, protected: true},
		{name: "paid departed source", accept: true, life: 40, departed: true},
		{name: "declined copied departed source", life: 40, departed: true, copied: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player3, def)
			source.Controller = game.Player1
			g.Players[game.Player2].Life = tc.life
			g.Players[game.Player2].ManaPool.Add(mana.C, 10)
			drawn := addCardToLibrary(g, game.Player2, vanillaCreature("Player draw", 1, 1))
			addCardToLibrary(g, game.Player1, vanillaCreature("Controller trap", 1, 1))
			addCardToLibrary(g, game.Player3, vanillaCreature("Owner trap", 1, 1))
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange,
					Controller: game.Player2, AffectedPlayer: game.PlayerYou,
				})
			}
			obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, HasTriggerEvent: true,
				TriggerEvent: event}
			payer := &libraryPaymentAgent{accept: tc.accept}
			caster := &libraryPaymentAgent{accept: !tc.accept}
			probed := withResultProbes(content, []resultProbe{
				{key: key, accepted: game.TriTrue, life: 1},
				{key: key, succeeded: game.TriTrue, life: 2},
			})
			trigger := def.TriggeredAbilities[0]
			trigger.Content = probed
			obj.ID = g.IDGen.Next()
			obj.Kind = game.StackTriggeredAbility
			obj.InlineTrigger = &trigger
			g.Stack.Push(obj)
			engine := NewEngine(nil)
			chooser := game.Player1
			if tc.copied {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3,
					Targets: []game.Target{game.StackObjectTarget(obj.ID)}},
					game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, nil)
				var ok bool
				obj, ok = g.Stack.Peek()
				if !ok || !obj.Copy || !obj.HasTriggerEvent {
					t.Fatal("real copy lost its event-player payment")
				}
				chooser = game.Player3
			}
			if tc.departed && !movePermanentToZone(g, source, zone.Exile) {
				t.Fatal("payment source did not depart")
			}
			if tc.protected {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange,
					Controller: chooser, AffectedPlayer: game.PlayerYou,
				})
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player2: payer}
			agents[chooser] = caster
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			paid := tc.accept && tc.life >= 2 && !tc.forbidden
			wantLife, wantTokens, wantCasterLife := tc.life, 1, 40
			if paid {
				wantLife, wantTokens, wantCasterLife = tc.life-2, 0, 43
			}
			if positive {
				wantTokens = 0
			}
			if tc.protected {
				wantCasterLife = 40
			}
			tokens := 0
			for _, permanent := range g.Battlefield {
				if permanent.Token {
					if permanent.Controller != chooser {
						t.Fatal("negative branch created the wrong player's token")
					}
					tokens++
				}
			}
			if g.Players[game.Player2].Life != wantLife || g.Players[chooser].Life != wantCasterLife ||
				tokens != wantTokens ||
				g.Players[game.Player4].Life != 40 || g.Players[game.Player2].ManaPool.Total() != 10 {
				t.Fatalf("paid=%v payer life=%d want=%d observer=%d want=%d tokens=%d want=%d mana=%d",
					paid, g.Players[game.Player2].Life, wantLife, g.Players[chooser].Life,
					wantCasterLife, tokens, wantTokens, g.Players[game.Player2].ManaPool.Total())
			}
			if g.Players[game.Player2].Hand.Contains(drawn) != (positive && paid) ||
				g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player3].Hand.Size() != 0 ||
				g.Players[game.Player4].Hand.Size() != 0 {
				t.Fatal("positive branch drew for a foreign actor or an uncompleted payment")
			}
			for player, state := range g.Players {
				if game.PlayerID(player) != chooser && game.PlayerID(player) != game.Player2 && state.Life != 40 {
					t.Fatal("source owner or unrelated controller lost or gained life")
				}
			}
			if len(caster.prompts) != 0 {
				t.Fatal("controller was asked to pay event player's life cost")
			}
			for _, request := range payer.prompts {
				if request.Player != game.Player2 || request.Prompt != "Pay resolution cost?" {
					t.Fatal("wrong fixed-life payment decision")
				}
			}
			assertResultKeysCleaned(t, obj, string(key))
		})
	}
}
