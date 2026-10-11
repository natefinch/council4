package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestCompiledEventPlayerRandomDiscardActualResult(t *testing.T) {
	for _, rider := range []bool{false, true} {
		text := "Whenever this creature deals damage to an opponent, that player discards a card at random. "
		if rider {
			text += "You gain 1 life. "
		}
		text += "If the player does, they draw a card."
		def := compileCounterTaxCard(t, "Random Player Action", text, "Creature")
		content := def.TriggeredAbilities[0].Content
		action, ok := content.Modes[0].Sequence[0].Primitive.(game.Discard)
		if !ok || !action.AtRandom || action.Player != game.EventPlayerReference() {
			t.Fatal("actor ownership changed the affected player or random discard payload")
		}
		for _, empty := range []bool{false, true} {
			t.Run(map[bool]string{false: "plain", true: "independent rider"}[rider]+map[bool]string{false: "/discard", true: "/empty"}[empty], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player3, def)
				source.Controller = game.Player1
				discarded := id.ID(0)
				if !empty {
					discarded = addCardToHand(g, game.Player2, vanillaCreature("Discarded", 1, 1))
				}
				drawn := addCardToLibrary(g, game.Player2, vanillaCreature("Drawn", 1, 1))
				addCardToLibrary(g, game.Player1, vanillaCreature("Controller trap", 1, 1))
				obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID,
					SourceCardID: source.CardInstanceID, HasTriggerEvent: true,
					TriggerEvent: game.Event{Kind: game.EventDamageDealt, SourceObjectID: source.ObjectID,
						Player: game.Player2, Controller: game.Player4, Amount: 1, DamageRecipient: game.DamageRecipientPlayer}}
				engine := NewEngine(nil)
				engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player3].Hand.Size() != 0 ||
					g.Players[game.Player4].Hand.Size() != 0 ||
					g.Players[game.Player2].Hand.Contains(drawn) == empty ||
					!empty && !g.Players[game.Player2].Graveyard.Contains(discarded) {
					t.Fatalf("THE-player result: hand=%v graveyard=%v library=%v discarded=%v drawn=%v other hands=%d/%d/%d",
						g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Graveyard.All(),
						g.Players[game.Player2].Library.All(), discarded, drawn,
						g.Players[game.Player1].Hand.Size(), g.Players[game.Player3].Hand.Size(), g.Players[game.Player4].Hand.Size())
				}
				wantLife := 40
				if rider {
					wantLife++
				}
				if g.Players[game.Player1].Life != wantLife {
					t.Fatal("independent YOU rider acquired the discard receipt")
				}
				assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-1")
				if !empty {
					g.Players[game.Player2].Hand.Remove(drawn)
					g.Players[game.Player2].Exile.Add(drawn)
				}
				engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if g.Players[game.Player2].Hand.Size() != 0 {
					t.Fatal("empty repeat reused earlier discard success")
				}
				assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-1")
			})
		}
	}
}
