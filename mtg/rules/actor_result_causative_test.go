package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledControllerCausativeResultRetainsChooserAndRecipient(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Causative Actor", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever a card is put into an opponent's graveyard from anywhere, you may have that player lose 2 life. If you do, you gain 2 life.",
	})
	for _, tc := range []struct {
		name                        string
		accept, forbidden, departed bool
		copied                      bool
	}{
		{name: "accepted", accept: true},
		{name: "declined"},
		{name: "accepted action prevented", accept: true, forbidden: true},
		{name: "departed source", accept: true, departed: true},
		{name: "copied departed source", accept: true, copied: true, departed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player3, def)
			source.Controller = game.Player1
			card := addCardToGraveyard(g, game.Player2, evidenceCard("Event card", 1))
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &def.TriggeredAbilities[0], HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventZoneChanged, CardID: card, Player: game.Player2,
					Controller: game.Player4, FromZone: zone.Library, ToZone: zone.Graveyard},
			}
			g.Stack.Push(obj)
			engine := NewEngine(nil)
			chooser := game.Player1
			if tc.copied {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3,
					Targets: []game.Target{game.StackObjectTarget(obj.ID)}},
					game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, nil)
				var ok bool
				obj, ok = g.Stack.Peek()
				if !ok || !obj.Copy || obj.Controller != game.Player3 {
					t.Fatal("real copy lost its resolving chooser")
				}
				chooser = game.Player3
			}
			if tc.departed && !movePermanentToZone(g, source, zone.Exile) {
				t.Fatal("source did not depart")
			}
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange,
					Controller: game.Player2, AffectedPlayer: game.PlayerYou,
				})
			}
			controller := &scopedMayAgent{accept: []bool{tc.accept}}
			recipientTrap := &scopedMayAgent{accept: []bool{!tc.accept}}
			agents := [game.NumPlayers]PlayerAgent{game.Player2: recipientTrap}
			agents[chooser] = controller
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			success := tc.accept && !tc.forbidden
			for player, state := range g.Players {
				want := 40
				if success && game.PlayerID(player) == chooser {
					want = 42
				}
				if success && game.PlayerID(player) == game.Player2 {
					want = 38
				}
				if state.Life != want {
					t.Fatalf("player %d life=%d want=%d", player, state.Life, want)
				}
			}
			assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-2")
		})
	}
}
