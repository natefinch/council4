package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

type controllerObservationAgent struct {
	targets   []game.Target
	accept    []bool
	requests  []game.ChoiceRequest
	may       int
	afterLook func()
	afterMay  func()
}

func (*controllerObservationAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (agent *controllerObservationAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	agent.requests = append(agent.requests, request)
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			if slices.Equal(option.Targets, agent.targets) {
				return []int{option.Index}
			}
		}
		panic("controller observation target set was not offered")
	}
	if request.Kind == game.ChoiceMay {
		accepted := agent.may < len(agent.accept) && agent.accept[agent.may]
		agent.may++
		if agent.afterMay != nil {
			agent.afterMay()
		}
		if accepted {
			return []int{1}
		}

		return []int{0}
	}
	if request.Kind == game.ChoiceResolution && agent.afterLook != nil {
		agent.afterLook()
	}
	return nil
}

func TestCompiledControllerObservationInterveningRider(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Observation With Rider", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "At the beginning of your upkeep, you may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.",
	})
	sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
	for _, look := range []bool{false, true} {
		for _, move := range []bool{false, true} {
			t.Run(fmt.Sprintf("look=%v/move=%v", look, move), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, def)
				below := addCardToLibrary(g, game.Player2, vanillaCreature("Below", 1, 1))
				observed := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 2))
				observer := &controllerObservationAgent{targets: []game.Target{game.PlayerTarget(game.Player2)},
					accept: []bool{look, move}}
				owner := &controllerObservationAgent{accept: []bool{!look, !move}}
				agents := [game.NumPlayers]PlayerAgent{game.Player1: observer, game.Player2: owner}
				engine := NewEngine(nil)
				emitEvent(g, beginningStepEvent(game.Player1, game.StepUpkeep))
				if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
					t.Fatal("natural observation-with-rider trigger missing")
				}
				obj, _ := g.Stack.Peek()
				var offeredReceipt game.InstructionResolutionResult
				observer.afterMay = func() {
					if observer.may == 2 || observer.may == 1 && sequence[0].PublishResult == "" {
						offeredReceipt = obj.ResolutionResults[string(sequence[2].ResultGate.Val.Key)]
					}
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				wantMay, wantTop := 1, observed
				if look {
					wantMay++
					if move {
						wantTop = below
					}
				}
				top, _ := g.Players[game.Player2].Library.Top()
				if observer.may != wantMay || top != wantTop || g.Players[game.Player1].Life != 41 ||
					len(owner.requests) != 0 || sequence[0].PublishResult == "" || sequence[1].PublishResult != "" ||
					sequence[2].ResultGate.Val.Key != sequence[0].PublishResult {
					t.Fatalf("look=%v choices=%d want=%d library-owner-choices=%d top=%d want=%d life=%d publishers=%q/%q move-gate=%q offered-receipt=%+v",
						look, observer.may, wantMay, owner.may, top, wantTop, g.Players[game.Player1].Life,
						sequence[0].PublishResult, sequence[1].PublishResult, sequence[2].ResultGate.Val.Key, offeredReceipt)
				}
			})
		}
	}
}
func compiledControllerObservation(t *testing.T, prefix string) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Precognition", Layout: "normal", TypeLine: "Enchantment", ManaCost: "{4}{U}",
		OracleText: "At the beginning of your upkeep, " + prefix +
			"you may look at the top card of target opponent's library. If you do, you may put that card on the bottom of that player's library.",
	})
}

func TestCompiledControllerObservationNaturalChoicesAndVisibility(t *testing.T) {
	for _, prefix := range []string{"", "Target player gains 1 life. "} {
		def := compiledControllerObservation(t, prefix)
		for _, look := range []bool{false, true} {
			for _, move := range []bool{false, true} {
				for _, empty := range []bool{false, true} {
					t.Run(fmt.Sprintf("prefix=%q/look=%v/move=%v/empty=%v", prefix, look, move, empty), func(t *testing.T) {
						g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
						source := addCombatPermanent(g, game.Player3, def)
						source.Controller = game.Player1
						decoy := addCardToLibrary(g, game.Player2, vanillaCreature("Below", 1, 1))
						observed := addCardToLibrary(g, game.Player2, vanillaCreature("Private top", 2, 2))
						foreign := addCardToLibrary(g, game.Player1, vanillaCreature("Observer library", 3, 3))
						if empty {
							g.Players[game.Player2].Library.Remove(observed)
							g.Players[game.Player2].Library.Remove(decoy)
						}
						targets := []game.Target{game.PlayerTarget(game.Player2)}
						if prefix != "" {
							targets = append([]game.Target{game.PlayerTarget(game.Player3)}, targets...)
						}
						observer := &controllerObservationAgent{targets: targets, accept: []bool{look, move}}
						owner, physicalOwner := &controllerObservationAgent{accept: []bool{!look, !move}}, &controllerObservationAgent{}
						agents := [game.NumPlayers]PlayerAgent{game.Player1: observer, game.Player2: owner, game.Player3: physicalOwner}
						engine := NewEngine(nil)
						emitEvent(g, beginningStepEvent(game.Player1, game.StepUpkeep))
						if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
							t.Fatal("natural upkeep failed to queue observation")
						}
						obj, ok := g.Stack.Peek()
						if !ok || obj.Controller != game.Player1 || !slices.Equal(obj.Targets, targets) {
							t.Fatal("natural target announcement lost controller or nonzero library slot")
						}
						engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
						wantMay := 1
						if look && !empty {
							wantMay++
						}
						if observer.may != wantMay || len(owner.requests) != 0 || len(physicalOwner.requests) != 0 {
							t.Fatal("optional look or independent move was offered to a foreign player")
						}
						looks := 0
						for _, request := range observer.requests {
							if request.Player != game.Player1 {
								t.Fatal("observer request used the library or source owner")
							}
							if request.Kind == game.ChoiceResolution {
								looks++
								if !request.Subject.Exists || request.Subject.Val.CardID != observed {
									t.Fatal("private look exposed an unrelated card")
								}
							}
						}
						wantLooks := 0
						if look && !empty {
							wantLooks = 1
						}
						top, exists := g.Players[game.Player2].Library.Top()
						wantTop := observed
						if look && move {
							wantTop = decoy
						}
						if looks != wantLooks || exists == empty || !empty && top != wantTop ||
							!g.Players[game.Player1].Library.Contains(foreign) ||
							eventRevealedCard(g, observed, obj.ID) {
							t.Fatal("look/move declined, empty, private visibility or actual owner identity changed")
						}
						if len(obj.LocalLinkedProducts) != 0 {
							t.Fatal("observation identity escaped its local body")
						}
						wantOwnerLife := 40
						if prefix != "" {
							wantOwnerLife++
						}
						if g.Players[game.Player3].Life != wantOwnerLife {
							t.Fatal("independent prefix became gated by an observation choice")
						}
					})
				}
			}
		}
	}
}

func TestCompiledControllerObservationCopiesAndCurrentCardIdentity(t *testing.T) {
	def := compiledControllerObservation(t, "")
	for _, mutation := range []string{"none", "new top", "departed card", "reentered card", "replaced departure", "departed source", "copied departed source"} {
		t.Run(mutation, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player3, def)
			source.Controller = game.Player1
			below := addCardToLibrary(g, game.Player2, vanillaCreature("Below", 1, 1))
			observed := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 2))
			version := g.CardInstances[observed].ZoneVersion
			observer := &controllerObservationAgent{
				targets: []game.Target{game.PlayerTarget(game.Player2)}, accept: []bool{true, true},
			}

			libraryOwner, sourceOwner := &controllerObservationAgent{}, &controllerObservationAgent{}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: observer, game.Player2: libraryOwner, game.Player3: sourceOwner}
			engine := NewEngine(nil)
			emitEvent(g, beginningStepEvent(game.Player1, game.StepUpkeep))
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
				t.Fatal("actual observation trigger missing")
			}
			obj, _ := g.Stack.Peek()
			chooser := game.Player1
			if mutation == "copied departed source" {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player4,
					Targets: []game.Target{game.StackObjectTarget(obj.ID)}},
					game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, nil)
				obj, _ = g.Stack.Peek()
				if !obj.Copy || obj.Controller != game.Player4 || !obj.HasTriggerEvent {
					t.Fatal("real trigger copy lost its resolving observer")
				}
				chooser = game.Player4
				agents[game.Player1], agents[game.Player4] = &controllerObservationAgent{}, observer
			}
			if mutation == "departed source" || mutation == "copied departed source" {
				if !movePermanentToZone(g, source, zone.Exile) {
					t.Fatal("source failed to leave")
				}
			}
			var newTop id.ID
			observer.afterLook = func() {
				if mutation == "new top" {
					newTop = addCardToLibrary(g, game.Player2, vanillaCreature("New top", 4, 4))
				}
				if mutation == "departed card" || mutation == "reentered card" || mutation == "replaced departure" {
					if mutation == "replaced departure" {
						g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
							ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
							MatchFromZone: true, FromZone: zone.Library,
							MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Graveyard,
						})
					}
					if !moveCardBetweenZones(g, game.Player2, observed, zone.Library, zone.Hand) {
						t.Fatal("observed card departure failed")
					}
					if mutation == "reentered card" && !moveCardBetweenZones(g, game.Player2, observed, zone.Hand, zone.Library) {
						t.Fatal("observed card reentry failed")
					}
				}
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			top, _ := g.Players[game.Player2].Library.Top()
			wantTop := below
			if mutation == "new top" {
				wantTop = newTop
			}
			if mutation == "reentered card" {
				wantTop = observed
			}
			if top != wantTop || observer.may != 2 || len(libraryOwner.requests) != 0 || len(sourceOwner.requests) != 0 ||
				eventRevealedCard(g, observed, obj.ID) {
				t.Fatal("controller observation moved a replacement top, stale incarnation or foreign library")
			}
			for _, request := range observer.requests {
				if request.Kind != game.ChoiceTarget && request.Player != chooser {
					t.Fatal("copied or departed source changed the deciding controller")
				}
			}
			if mutation == "reentered card" && g.CardInstances[observed].ZoneVersion == version {
				t.Fatal("reentry did not establish the stale-reference control")
			}
		})
	}
}

func TestCompiledControllerObservationSkippedStateDoesNotBorrowReceipt(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Conditional Observer", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "If you control an artifact, you may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.",
	})
	for _, artifact := range []bool{false, true} {
		t.Run(fmt.Sprintf("artifact=%v", artifact), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if artifact {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Artifact", Types: []types.Card{types.Artifact},
				}})
			}
			below := addCardToLibrary(g, game.Player2, vanillaCreature("Below", 1, 1))
			observed := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 2))
			card := addCardToHand(g, game.Player1, def)
			setMainPhasePriority(g, game.Player1)
			engine := NewEngine(nil)
			if !engine.applyAction(g, game.Player1, action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
				t.Fatal("normal conditional observation cast failed")
			}
			obj, _ := g.Stack.Peek()
			key := string(def.SpellAbility.Val.Modes[0].Sequence[0].PublishResult)
			obj.ResolutionResults = map[string]game.InstructionResolutionResult{key: {Succeeded: true, Amount: 91}}
			observer := &controllerObservationAgent{accept: []bool{true, true}}
			owner := &controllerObservationAgent{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: observer, game.Player2: owner}, &TurnLog{})
			wantMay, wantTop := 0, observed
			if artifact {
				wantMay, wantTop = 2, below
			}
			top, _ := g.Players[game.Player2].Library.Top()
			if observer.may != wantMay || top != wantTop || len(owner.requests) != 0 ||
				g.Players[game.Player1].Life != 41 || obj.ResolutionResults[key].Amount != 91 {
				t.Fatal("skipped observation borrowed enclosing or unconditional Gain success")
			}
		})
	}
}
