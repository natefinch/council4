package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledEventControllerPaymentActualResult(t *testing.T) {
	def := compileCounterTaxCard(t, "Event Controller Payment",
		"Whenever a creature deals combat damage to one of your opponents, its controller may pay 1 life. If they do, they draw a card.", "Enchantment")
	for _, tc := range []struct {
		name                       string
		decline, forbidden, copied bool
		departed, insufficient     bool
	}{
		{name: "accepted"},
		{name: "declined", decline: true},
		{name: "forbidden", forbidden: true},
		{name: "unaffordable", insufficient: true},
		{name: "departed", departed: true},
		{name: "copied departed trigger", copied: true, departed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player3, def)
			source.Controller = game.Player1
			actor := addCombatCreaturePermanent(g, game.Player4)
			actor.Controller = game.Player2
			addCardToLibrary(g, game.Player2, vanillaCreature("Payer Draw", 1, 1))
			addCardToLibrary(g, game.Player3, vanillaCreature("Damaged Player Trap", 1, 1))
			if tc.insufficient {
				g.Players[game.Player2].Life = 0
			}
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange,
					Controller: game.Player2, AffectedPlayer: game.PlayerYou,
				})
			}
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &def.TriggeredAbilities[0], HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventDamageDealt, SourceObjectID: actor.ObjectID,
					Controller: game.Player2, Player: game.Player3, Amount: 1},
			}
			g.Stack.Push(obj)
			if tc.departed && !movePermanentToZone(g, actor, zone.Graveyard) {
				t.Fatal("event actor did not leave")
			}
			engine := NewEngine(nil)
			if tc.copied {
				addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
					[]game.Target{game.StackObjectTarget(obj.ID)})
				engine.resolveTopOfStack(g, &TurnLog{})
				var ok bool
				obj, ok = g.Stack.Peek()
				if !ok || !obj.Copy || !obj.HasTriggerEvent {
					t.Fatal("real trigger copy lost its event")
				}
			}
			payer := &libraryPaymentAgent{accept: !tc.decline}
			caster := &libraryPaymentAgent{}
			before := g.Players[game.Player2].Life
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: caster, game.Player2: payer}, &TurnLog{})
			success := !tc.decline && !tc.forbidden && !tc.insufficient
			wantDraw, wantLife := 0, before
			if success {
				wantDraw, wantLife = 1, before-1
			}
			if g.Players[game.Player2].Life != wantLife || g.Players[game.Player2].Hand.Size() != wantDraw ||
				g.Players[game.Player1].Life != 40 || g.Players[game.Player3].Hand.Size() != 0 ||
				g.Players[game.Player4].Life != 40 {
				t.Fatal("payment or draw borrowed the caster, damaged player, or physical owner")
			}
			if len(caster.prompts) != 0 {
				t.Fatal("source controller was offered the event actor's payment")
			}
			for _, request := range payer.prompts {
				if request.Player != game.Player2 || request.Prompt != "Pay resolution cost?" {
					t.Fatalf("wrong payer choice: %+v", request)
				}
			}
			assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-1")
		})
	}
}
