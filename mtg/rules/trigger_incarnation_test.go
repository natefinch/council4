package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestTriggerGrantIncarnationLimits(t *testing.T) {
	for _, scenario := range []string{
		"same grant", "source reentry", "recipient reentry", "concurrent grants",
		"slot shift", "clone", "clone independent", "next turn", "old capture new source", "queued copy", "copied grant",
	} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
			grant.Controller = game.Player2
			recipient := addCombatPermanent(g, game.Player1, vanillaCreature("Recipient", 2, 3))
			recipient.Controller = game.Player2
			actuator := addCombatPermanent(g, game.Player3, compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Identity Actuator", Layout: "normal", TypeLine: "Artifact",
				OracleText: "{0}: Target creature gets +1/+1 until end of turn.",
			}))
			actuator.Controller = game.Player2
			g.Turn.PriorityPlayer = game.Player2
			target := func(want int) {
				t.Helper()
				if !engine.applyActivateAbilityWithChoices(g, game.Player2, action.ActivateAbilityAction{
					SourceID: actuator.ObjectID, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
				}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
					t.Fatal("legal targeting failed")
				}

				engine.putTriggeredAbilitiesOnStack(g)
				got := 0
				for _, obj := range g.Stack.Objects() {
					if obj.Kind == game.StackTriggeredAbility {
						got++
						if obj.InlineTrigger.MaxTriggersPerTurn > 0 && obj.TriggerOrigin().RecipientID != recipient.ObjectID ||
							obj.Controller != game.Player2 || obj.SourceID != recipient.ObjectID {
							t.Fatal("captured identity/controller lost")
						}
					}
				}
				if got != want {
					t.Fatalf("queued triggers=%d, want %d", got, want)
				}
				for g.Stack.Size() > 0 {
					g.Stack.Pop()
				}
			}
			if scenario == "concurrent grants" || scenario == "slot shift" || scenario == "copied grant" {
				second := addCombatPermanent(g, game.Player3, targetAnnouncementGrant(t))
				second.Controller = game.Player2
				if scenario == "copied grant" {
					_, ok := createTokenPermanent(g, game.Player2, copyCardDef(g.CardInstances[grant.CardInstanceID].Def))
					if !ok {
						t.Fatal("native copied grant token creation failed")
					}
					movePermanentToZone(g, second, zone.Exile)
				}
				target(2)
				target(2)
				if scenario == "slot shift" && !sacrificePermanent(g, grant) {
					t.Fatal("earlier flattened grant departure failed")
				}
				target(0)
				return
			}
			target(1)
			if scenario == "clone independent" {
				cloned := g.Clone()
				target(1)
				target(0)
				g = cloned
				target(1)
				target(0)
				return
			}
			target(1)
			switch scenario {
			case "source reentry", "recipient reentry":
				departed := grant
				if scenario == "recipient reentry" {
					departed = recipient
				}
				if !sacrificePermanent(g, departed) {
					t.Fatal("real departure failed")
				}
				card := g.CardInstances[departed.CardInstanceID]
				addEffectSpellToStack(g, game.Player2, game.PutOnBattlefield{
					Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
					Recipient: opt.Val(game.ControllerReference()),
				}, []game.Target{currentCardTarget(t, g, card.ID)})
				engine.resolveTopOfStack(g, &TurnLog{})
				returned := permanentForCard(g, card.ID)
				if returned == nil || returned.ObjectID == departed.ObjectID || returned.Owner != departed.Owner {
					t.Fatal("exact physical card did not reenter with new occurrence")
				}
				if scenario == "recipient reentry" {
					recipient = returned
				}
				target(1)
				target(1)
				target(0)
			case "clone":
				clone := g.Clone()
				g = clone
				target(0)
			case "next turn":
				engine.advanceToNextTurn(g)
				g.Turn.PriorityPlayer = game.Player2
				target(1)
				target(1)
				target(0)
			case "old capture new source":
				if !engine.applyActivateAbilityWithChoices(g, game.Player2, action.ActivateAbilityAction{
					SourceID: actuator.ObjectID, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
				}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
					t.Fatal("capture failed")
				}
				if !sacrificePermanent(g, grant) {
					t.Fatal("source departure failed")
				}
				replacement := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
				replacement.Controller = game.Player2
				if engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("old exhausted captured origin borrowed new source allowance")
				}
				g.Stack.Pop()
				target(1)
			case "queued copy":
				engine.advanceToNextTurn(g)
				g.Turn.PriorityPlayer = game.Player2
				if !engine.applyActivateAbilityWithChoices(g, game.Player2, action.ActivateAbilityAction{
					SourceID: actuator.ObjectID, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
				}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) || !engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("new-turn trigger failed")
				}
				obj, _ := g.Stack.Peek()
				copied := game.NewStackObjectCopy(obj, g.IDGen.Next())
				if copied.TriggerOrigin() != obj.TriggerOrigin() ||
					copied.TriggerEvent.TriggeredAbilities[0].TriggerOrigin() != obj.TriggerOrigin() {
					t.Fatal("copy lost immutable captured grant")
				}
				for g.Stack.Size() > 0 {
					g.Stack.Pop()
				}
				target(1)
				target(0)
			default:
				before := g.Clone()
				for range 10 {
					g.BeginStaticSourceFrame()
					permanentEffectiveAbilities(g, recipient)
					g.EndStaticSourceFrame()
				}
				if !reflect.DeepEqual(before.Clone(), g.Clone()) {
					t.Fatal("pure reads changed runtime usage state")
				}
				target(0)
			}
			if g.Players[game.Player2].Graveyard.Contains(grant.CardInstanceID) &&
				grant.Owner != game.Player2 {
				t.Fatal("physical owner/controller conflated")
			}
		})
	}
}
