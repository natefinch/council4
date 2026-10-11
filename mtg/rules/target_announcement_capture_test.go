package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func targetAnnouncementGrant(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Target Announcement Grant", Layout: "normal", TypeLine: "Creature",
		Power: new("3"), Toughness: new("4"),
		OracleText: `Creatures you control have "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn."`,
	})
}

type targetAnnouncementCostAgent struct {
	subject string
	target  id.ID
	choices int
}

func (*targetAnnouncementCostAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *targetAnnouncementCostAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		for _, target := range option.Targets {
			if a.target != 0 && target.Kind == game.TargetPermanent && target.PermanentID == a.target {
				a.choices++
				return []int{option.Index}
			}
		}
		if a.subject != "" && strings.Contains(option.Label, a.subject) {
			a.choices++
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

func TestTargetAnnouncementPaymentRetainsGrantedBody(t *testing.T) {
	for _, route := range []string{"activate grant", "activate recipient", "cast", "paid cast", "free cast"} {
		t.Run(route, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.PriorityPlayer = game.Player2
			grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
			grant.Controller = game.Player2
			recipient := addCombatPermanent(g, game.Player1, vanillaCreature("Original Recipient", 2, 3))
			recipient.Controller = game.Player2
			observed := addCardToLibrary(g, game.Player2, vanillaCreature("Original Observation", 1, 1))
			costAgent := &targetAnnouncementCostAgent{subject: "Target Announcement Grant"}
			wantSource := recipient
			if route == "activate recipient" {
				costAgent.subject = "Original Recipient"
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player2: costAgent}
			if strings.HasPrefix(route, "activate") {
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Announcement Actuator", Layout: "normal", TypeLine: "Artifact",
					OracleText: "Sacrifice a creature: Target creature gets +1/+1 until end of turn.",
				})
				actuator := addCombatPermanent(g, game.Player2, def)
				if !engine.applyActivateAbilityWithChoices(g, game.Player2, action.ActivateAbilityAction{
					SourceID: actuator.ObjectID, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
				}, agents, &TurnLog{}) {
					t.Fatal("legal target-then-sacrifice activation failed")
				}
			} else {
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Announcement Spell", Layout: "normal", TypeLine: "Instant", ManaCost: "{0}",
					OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nTarget creature gets +1/+1 until end of turn.",
				})
				cardID := addCardToHand(g, game.Player2, def)
				var cast bool
				if route == "cast" {
					cast = engine.applyCastSpellWithChoices(g, game.Player2, action.CastSpellAction{
						CardID: cardID, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
					}, agents, &TurnLog{})
				} else {
					// Resolution casts choose the first legal target: the grant
					// itself, which is also the creature selected for payment.
					wantSource = grant
					if route == "paid cast" {
						cast = engine.castPaidTargetedSpell(g, game.Player2, cardID, zone.Hand, agents, &TurnLog{})
					} else {
						cast = engine.castFreeTargetedSpell(g, game.Player2, cardID, zone.Hand, false, agents, &TurnLog{})
					}
				}
				if !cast {
					t.Fatal("legal target-then-sacrifice cast failed")
				}
			}
			if costAgent.choices != 1 || g.Stack.Size() != 1 {
				t.Fatal("actual payment did not select exactly one subject and publish one action")
			}
			actionObject, _ := g.Stack.Peek()
			var targetEvents []game.Event
			for _, event := range g.Events {
				if event.Kind == game.EventObjectBecameTarget {
					targetEvents = append(targetEvents, event)
				}
			}
			if len(targetEvents) != 1 || targetEvents[0].StackObjectID != actionObject.ID ||
				targetEvents[0].PermanentID != wantSource.ObjectID ||
				!targetEvents[0].TriggeredAbilitiesCaptured || len(targetEvents[0].TriggeredAbilities) != 1 {
				t.Fatal("target announcement did not retain one body with the exact eventual action identity")
			}
			captured := targetEvents[0].TriggeredAbilities[0]
			if captured.Controller != game.Player2 || captured.SourceID != wantSource.ObjectID ||
				captured.SourceCardID != wantSource.CardInstanceID {
				t.Fatal("captured trigger rebound to the granting source or physical owner")
			}
			recipient.Controller = game.Player3
			addCombatPermanent(g, game.Player2, vanillaCreature(g.CardInstances[wantSource.CardInstanceID].Def.Name, 2, 3))
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
				t.Fatal("successful payment removed or rebound the event-time trigger")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Hand.Contains(observed) {
				t.Fatal("retained body did not consume the original ability controller's exact observation")
			}
		})
	}
}

func TestTargetAnnouncementRejectedPaymentPublishesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
	grant.Controller = game.Player2
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Unpaid Announcement", Layout: "normal", TypeLine: "Instant", ManaCost: "{2}",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nTarget creature gets +1/+1 until end of turn.",
	})
	cardID := addCardToHand(g, game.Player2, def)
	beforeEvents := len(g.Events)
	if engine.castPaidTargetedSpell(g, game.Player2, cardID, zone.Hand, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("unpayable proposal succeeded")
	}
	if len(g.Events) != beforeEvents || g.Stack.Size() != 0 || len(g.TriggeredAbilitiesThisTurn) != 0 ||
		!g.Players[game.Player2].Hand.Contains(cardID) || permanentForCard(g, grant.CardInstanceID) == nil {
		t.Fatal("unpaid proposal published a target, consumed a limit/cost, or failed rollback")
	}
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player3,
		Targets: []game.Target{game.PermanentTarget(grant.ObjectID)},
	}
	snapshots := prepareTargetEvents(g, obj)
	if len(snapshots) != 1 || len(snapshots[0].TriggeredAbilities) != 1 ||
		g.Stack.Size() != 0 || len(g.Events) != beforeEvents || len(g.TriggeredAbilitiesThisTurn) != 0 {
		t.Fatal("unpublished preparation changed action/event/trigger-limit state")
	}
}

func TestTargetAnnouncementRejectedActionConsumesNoGrant(t *testing.T) {
	grantDef := targetAnnouncementGrant(t)
	abilityDef := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Rejected Announcement Actuator", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Sacrifice a creature: Target creature gets +1/+1 until end of turn.",
	})
	spellDef := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Rejected Announcement Spell", Layout: "normal", TypeLine: "Instant", ManaCost: "{0}",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nTarget creature gets +1/+1 until end of turn.",
	})
	for _, route := range []string{"activate", "cast"} {
		for _, refusal := range []string{"wrong target type", "missing target object", "foreign actor"} {
			t.Run(route+"/"+refusal, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				grant := addCombatPermanent(g, game.Player4, grantDef)
				grant.Controller = game.Player2
				recipient := addCombatPermanent(g, game.Player1, vanillaCreature("Recipient", 2, 3))
				recipient.Controller = game.Player2
				actuator := addCombatPermanent(g, game.Player2, abilityDef)
				cardID := addCardToHand(g, game.Player2, spellDef)
				observed := addCardToLibrary(g, game.Player2, vanillaCreature("Observation", 1, 1))
				costAgent := &targetAnnouncementCostAgent{subject: "Target Announcement Grant"}
				agents := [game.NumPlayers]PlayerAgent{game.Player2: costAgent}
				player := game.Player2
				target := game.PermanentTarget(recipient.ObjectID)
				switch refusal {
				case "wrong target type":
					target = game.PermanentTarget(actuator.ObjectID)
				case "missing target object":
					target = game.PermanentTarget(g.IDGen.Next())
				case "foreign actor":
					player = game.Player3
				default:
					t.Fatalf("unknown rejection case %q", refusal)
				}
				g.Turn.PriorityPlayer = player
				apply := func(player game.PlayerID, target game.Target) bool {
					if route == "cast" {
						return engine.applyCastSpellWithChoices(g, player, action.CastSpellAction{
							CardID: cardID, Targets: []game.Target{target},
						}, agents, &TurnLog{})
					}
					return engine.applyActivateAbilityWithChoices(g, player, action.ActivateAbilityAction{
						SourceID: actuator.ObjectID, Targets: []game.Target{target},
					}, agents, &TurnLog{})
				}
				beforeEvents := len(g.Events)
				if apply(player, target) {
					t.Fatal("invalid target/actor action succeeded")
				}
				if len(g.Events) != beforeEvents || g.Stack.Size() != 0 || len(g.TriggeredAbilitiesThisTurn) != 0 ||
					costAgent.choices != 0 || permanentForCard(g, grant.CardInstanceID) != grant ||
					!g.Players[game.Player2].Hand.Contains(cardID) || !g.Players[game.Player2].Library.Contains(observed) {
					t.Fatal("rejected action published a target, consumed a grant/cost, or changed card zones")
				}
				g.Turn.PriorityPlayer = game.Player2
				if !apply(game.Player2, game.PermanentTarget(recipient.ObjectID)) || costAgent.choices != 1 ||
					!engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
					t.Fatal("rejected action prevented the next valid payment and granted trigger")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player2].Hand.Contains(observed) {
					t.Fatal("valid retry did not resolve the original controller's exact observation")
				}
			})
		}
	}
}

func targetAnnouncementCount(g *game.Game, stackID id.ID) int {
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventObjectBecameTarget && event.StackObjectID == stackID {
			count++
		}
	}
	return count
}
