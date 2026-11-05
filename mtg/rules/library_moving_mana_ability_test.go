package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3690. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Millikin
// fa4dffda-6f04-4d0b-829d-28a1a5794dee {2} Artifact Creature — Construct 0/1
// "{T}, Mill a card: Add {C}. (Activate only as an instant. To mill a card,
// put the top card of your library into your graveyard.)"; Deranged Assistant
// 9e1b0034-bd2c-4412-8d81-db3306c8814f {1}{U} Creature — Human Wizard 1/1, the
// same ability; Phyrexian Altar 8d02b297-97c4-4379-9862-0a462400f66f
// "Sacrifice a creature: Add one mana of any color."; Blightsteel Colossus
// e80772e2-8623-4094-81a2-70828b2b151c "… If Blightsteel Colossus would be put
// into a graveyard from anywhere, reveal Blightsteel Colossus and shuffle it
// into its owner's library instead."; Sol Ring, Waterlogged Grove, Nurturing
// Peatland, Silent Clearing, Thalia, Guardian of Thraben and the cast and
// activation routes as reviewed in priority_cast_mana_window_test.go and
// activation_mana_window_test.go. Millikin ruling 2026-08-05: "Millikin's
// ability is not a mana ability. Players can respond to it, and it can't be
// activated in the middle of paying for a cost." The Hobbit Update Bulletin
// and Comprehensive Rules effective 2026-09-25, 605.1a: an activated ability
// is a mana ability only if, among other criteria, it doesn't move cards to or
// from a library as part of its cost or effect, "even if they make mana",
// ignoring replacement effects other than self-replacement effects, so
// Phyrexian Altar's ability stays a mana ability when the Colossus it
// sacrifices is shuffled into a library; 605.3a a mana ability is activated
// whenever its controller has priority or is asked to pay a cost, and 605.3b
// it doesn't use the stack; 602.2 an activated ability that isn't a mana
// ability is put on the stack and its costs paid as it is activated; 601.2g
// only mana abilities may be activated while paying; 733.1 "Players may not
// reverse actions that moved cards … from a library to any zone other than the
// stack" (the Colossus's shuffle).

// libraryMovingManaProducers are the registered cards whose mana-producing
// ability mills as a cost, so isn't a mana ability (CR 605.1a).
var libraryMovingManaProducers = []string{"Millikin", "Deranged Assistant"}

const (
	millProducerMana = "Mill a card: Add {C}."
	altarMana        = "Sacrifice a creature: Add one mana of any color."
)

// TestLibraryMovingManaProducerUsesTheStack activates Millikin's and Deranged
// Assistant's ability with priority during the opponent's turn. It isn't a mana
// ability (CR 605.1a): its mill cost is paid as it is activated (CR 602.2b),
// but it is put on the stack, where players may respond, and makes no mana
// until it resolves. Once it resolves, the {C} pays for Sol Ring, which the
// payment planner can't pay for with the untapped producer alone.
func TestLibraryMovingManaProducerUsesTheStack(t *testing.T) {
	t.Parallel()
	for _, name := range libraryMovingManaProducers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, registryCardDef(t, name))
			below := addCardToLibraryNamed(g, game.Player1, "Below")
			top := addCardToLibraryNamed(g, game.Player1, "Top Card")
			solRing := addCardToHand(g, game.Player1, registryCardDef(t, "Sol Ring"))
			giveMainPhasePriority(g, game.Player1)
			if actionsContain(e.legalActions(g, game.Player1), action.CastSpell(solRing, nil, 0, nil)) {
				t.Fatal("Sol Ring offered with only the untapped producer to pay for it")
			}
			giveMainPhasePriority(g, game.Player2)
			g.Turn.PriorityPlayer = game.Player1
			act := sourceActivation(t, e, g, source.ObjectID)
			activate, _ := act.ActivateAbilityPayload()
			if isManaAbilityActivation(g, game.Player1, activate) {
				t.Fatalf("%s's ability classified as a mana ability", name)
			}

			if !e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(&priorityCastWindowAgent{}), &TurnLog{}) {
				t.Fatalf("%s's ability wasn't activated with priority on the opponent's turn", name)
			}

			p := g.Players[game.Player1]
			obj, ok := g.Stack.Peek()
			if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != source.ObjectID {
				t.Fatalf("stack top = %+v, want %s's ability on the stack", obj, name)
			}
			if !source.Tapped || !p.Graveyard.Contains(top) || p.ManaPool.Total() != 0 {
				t.Fatalf("tapped=%t top milled=%t pool=%d, want the costs paid and no mana before resolution",
					source.Tapped, p.Graveyard.Contains(top), p.ManaPool.Total())
			}
			for _, entry := range g.Events {
				if entry.Kind == game.EventManaProduced {
					t.Fatal("mana was added before the ability resolved")
				}
			}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(&priorityCastWindowAgent{}), &TurnLog{})

			if g.Stack.Size() != 0 || p.ManaPool.Amount(mana.C) != 1 || !p.Library.Contains(below) {
				t.Fatalf("stack=%d pool C=%d, want the resolved ability to add {C} and mill nothing more", g.Stack.Size(), p.ManaPool.Amount(mana.C))
			}
			giveMainPhasePriority(g, game.Player1)
			if !actionsContain(e.legalActions(g, game.Player1), action.CastSpell(solRing, nil, 0, nil)) {
				t.Fatal("Sol Ring not offered with the resolved ability's {C} in pool")
			}
		})
	}
}

// windowOptionsAgent pays as its priorityCastWindowAgent does, recording the
// label of every mana ability offered in a payment-time mana window.
type windowOptionsAgent struct {
	priorityCastWindowAgent
	offered []string
}

func (a *windowOptionsAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		for _, option := range request.Options {
			a.offered = append(a.offered, option.Label)
		}
	}
	return a.priorityCastWindowAgent.ChooseChoice(obs, request)
}

// requireNotOffered checks no offered window option names label.
func requireNotOffered(t *testing.T, offered []string, label string) {
	t.Helper()
	for _, offered := range offered {
		if strings.Contains(offered, label) {
			t.Fatalf("%q offered while paying (CR 601.2g), want only mana abilities", offered)
		}
	}
}

// TestLibraryMovingManaProducerNotActivatedWhilePaying casts with priority on
// each cast route, with a locked two-mana total and Waterlogged Grove, and
// with Millikin or Deranged Assistant untapped. Only mana abilities may be
// activated while paying (CR 601.2g, 605.3a), so the producer can't add the
// second mana and the cast isn't offered. With Nurturing Peatland too it is,
// the producer is never offered in the payment-time window, and the lands pay.
func TestLibraryMovingManaProducerNotActivatedWhilePaying(t *testing.T) {
	t.Parallel()
	for _, name := range libraryMovingManaProducers {
		for _, route := range priorityCastRoutes() {
			t.Run(name+"/"+route.name, func(t *testing.T) {
				t.Parallel()
				g, e, cardID := route.build(t)
				producer := addCombatPermanent(g, game.Player1, registryCardDef(t, name))
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
				library := g.Players[game.Player1].Library.All()
				if _, offered := findPriorityCast(e, g, route, cardID); offered {
					t.Fatalf("cast offered with only Waterlogged Grove and %s to pay two mana", name)
				}
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
				act, offered := findPriorityCast(e, g, route, cardID)
				if !offered {
					t.Fatal("cast payable with both lands was not offered")
				}
				agent := &windowOptionsAgent{priorityCastWindowAgent: priorityCastWindowAgent{limit: -1, colors: route.colors}}

				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

				if outcome != actionApplied || agent.activated != 2 {
					t.Fatalf("outcome = %v, activated = %d; want the lands to pay the cast", outcome, agent.activated)
				}
				route.done(t, g, cardID)
				requireNotOffered(t, agent.offered, millProducerMana)
				if producer.Tapped || len(agent.offered) == 0 {
					t.Fatalf("%s tapped = %t over %d offered options, want it unused", name, producer.Tapped, len(agent.offered))
				}
				if got := g.Players[game.Player1].Library.Size(); route.name != "library top with Future Sight" && got != len(library) {
					t.Fatalf("library size = %d, want %d: nothing milled", got, len(library))
				}
			})
		}
	}
}

// TestLibraryMovingManaProducerNotActivatedWhileActivating activates an
// ability from each zone a source functions from whose cost one land can't pay
// alone, with Millikin or Deranged Assistant untapped too. The producer's
// ability isn't a mana ability (CR 605.1a), so it can't be activated while
// paying the activation cost (CR 602.2b, 601.2g) and the activation isn't
// offered until the last land is added; the producer is then never offered in
// the payment-time window.
func TestLibraryMovingManaProducerNotActivatedWhileActivating(t *testing.T) {
	t.Parallel()
	for _, name := range libraryMovingManaProducers {
		for _, route := range activationWindowRoutes() {
			t.Run(name+"/"+route.name, func(t *testing.T) {
				t.Parallel()
				g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
				source := route.build(t, g)
				producer := addCombatPermanent(g, game.Player1, registryCardDef(t, name))
				for _, land := range route.lands[:len(route.lands)-1] {
					addCombatPermanent(g, game.Player1, registryCardDef(t, land))
				}
				if _, ok := findRouteActivation(e, g, route, source); ok {
					t.Fatalf("activation offered with %s standing in for the last land", name)
				}
				addCombatPermanent(g, game.Player1, registryCardDef(t, route.lands[len(route.lands)-1]))
				act, ok := findRouteActivation(e, g, route, source)
				if !ok {
					t.Fatal("activation payable with the lands wasn't offered")
				}
				agent := &activationOptionsAgent{activationWindowAgent: newActivationWindowAgent(route.order, top)}

				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

				if outcome != actionApplied {
					t.Fatalf("outcome = %v, want the ability activated", outcome)
				}
				route.done(t, g, source)
				requireNotOffered(t, agent.offered, millProducerMana)
				if producer.Tapped || len(agent.offered) == 0 {
					t.Fatalf("%s tapped = %t over %d offered options, want it unused", name, producer.Tapped, len(agent.offered))
				}
				assertTopShown(t, g, "Future Sight", top)
			})
		}
	}
}

// activationOptionsAgent pays as its activationWindowAgent does, recording
// the label of every mana ability offered in a payment-time mana window.
type activationOptionsAgent struct {
	*activationWindowAgent
	offered []string
}

func (a *activationOptionsAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		for _, option := range request.Options {
			a.offered = append(a.offered, option.Label)
		}
	}
	return a.activationWindowAgent.ChooseChoice(obs, request)
}

// addAltarAndColossus gives player Phyrexian Altar and control of
// Blightsteel Colossus, owned by the opponent, whose library is empty,
// returning the Colossus's card. Sacrificing the Colossus to the Altar is a
// mana ability, the Colossus's replacement effect not being a self-replacement
// effect of the Altar's (CR 605.1a), but it reveals the Colossus and shuffles
// it into its owner's library, an action that can't be reversed (CR 733.1).
// Shuffling it into an otherwise empty library has one outcome, so the shuffle
// repeats identically when an earlier mana ability is reversed around it.
func addAltarAndColossus(t *testing.T, g *game.Game, player game.PlayerID) id.ID {
	t.Helper()
	owner := game.Player2
	if player == owner {
		owner = game.Player1
	}
	addCombatPermanent(g, player, registryCardDef(t, "Phyrexian Altar"))
	colossus := addCombatPermanent(g, player, registryCardDef(t, "Blightsteel Colossus"))
	colossus.Owner = owner
	g.CardInstances[colossus.CardInstanceID].Owner = owner
	if size := g.Players[owner].Library.Size(); size != 0 {
		t.Fatalf("Colossus owner's library has %d cards, want it empty", size)
	}
	return colossus.CardInstanceID
}

// homingGolem is the name of homingGolemDef's creature.
const homingGolem = "Homing Golem"

// homingGolemDef is a test artifact creature whose own replacement effect
// puts it on top of its owner's library: "If this creature would be put into
// a graveyard from anywhere, put it on top of its owner's library instead."
// Sacrificing it to a mana ability such as Phyrexian Altar's changes the top
// of the library deterministically; the ability is still a mana ability, the
// replacement effect not being the ability's own (CR 605.1a).
func homingGolemDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      homingGolem,
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbility{{
			Text: "If this creature would be put into a graveyard from anywhere, put it on top of its owner's library instead.",
			Replacement: game.ReplacementEffect{
				MatchEvent:    game.EventZoneChanged,
				MatchToZone:   true,
				ToZone:        zone.Graveyard,
				ReplaceToZone: zone.Library,
				Duration:      game.DurationPermanent,
			},
		}},
	}}
}

// sacrificeNamed answers a sacrifice choice, such as Phyrexian Altar's, with
// the permanent named name.
func sacrificeNamed(request game.ChoiceRequest, name string) ([]int, bool) {
	if request.Kind != game.ChoicePayment || !strings.HasPrefix(request.Prompt, "Sacrifice") {
		return nil, false
	}
	for _, option := range request.Options {
		if strings.Contains(option.Label, name) {
			return []int{option.Index}, true
		}
	}
	return nil, false
}

// requireColossusShuffled checks that the sacrifice of colossus to a mana
// ability was kept: the Colossus is in its owner's library rather than on the
// battlefield or in a graveyard.
func requireColossusShuffled(t *testing.T, g *game.Game, colossus id.ID) {
	t.Helper()
	owner := g.Players[g.CardInstances[colossus].Owner]
	if _, ok := findPermanentByCardID(g, colossus); ok || !owner.Library.Contains(colossus) || owner.Graveyard.Contains(colossus) {
		t.Fatal("Blightsteel Colossus wasn't sacrificed and shuffled into its owner's library")
	}
}

// windowAgent answers priority and choices.
type windowAgent interface {
	PlayerAgent
	ChoiceAgent
}

// sacrificingAgent answers as next does, except that when next activates a
// mana ability with a "Sacrifice a creature" cost while paying, such as
// Phyrexian Altar's, it sacrifices the creature named victim.
type sacrificingAgent struct {
	next       windowAgent
	victim     string
	pending    bool
	sacrifices int
}

func (a *sacrificingAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	return a.next.ChooseAction(obs, legal)
}

func (a *sacrificingAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if a.pending {
		if selected, ok := sacrificeNamed(request, a.victim); ok {
			a.pending = false
			return selected
		}
	}
	selected := a.next.ChooseChoice(obs, request)
	if request.Prompt == "Activate a mana ability while paying?" && len(selected) == 1 {
		for _, option := range request.Options {
			if option.Index == selected[0] && strings.Contains(option.Label, "Sacrifice a creature") {
				a.pending = true
				a.sacrifices++
			}
		}
	}
	return selected
}
