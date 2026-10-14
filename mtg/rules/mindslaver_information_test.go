package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func hasCardName(cards []CardView, name string) bool {
	return slices.ContainsFunc(cards, func(card CardView) bool { return card.Name == name })
}

func TestMindslaverObservationsAndObserverPermissions(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	own := addCardToHand(g, game.Player1, vanillaCreature("Controller hand", 1, 1))
	controlled := addCardToHand(g, game.Player2, vanillaCreature("Controlled hand", 2, 2))
	addCardToHand(g, game.Player3, vanillaCreature("Uninvolved hand", 3, 3))
	addCardToLibraryNamed(g, game.Player2, "Controlled unknown library")
	addCardToLibraryNamed(g, game.Player3, "Uninvolved unknown library")
	hidden := addFaceDownPermanent(g, game.Player2, vanillaCreature("Controlled morph", 4, 4), game.FaceDownMorph)
	uninvolved := addFaceDownPermanent(g, game.Player3, vanillaCreature("Uninvolved morph", 5, 5), game.FaceDownMorph)
	foretold := addCardToHand(g, game.Player2, vanillaCreature("Controlled foretell", 6, 6))
	g.Players[game.Player2].Hand.Remove(foretold)
	g.Players[game.Player2].Exile.Add(foretold)
	g.Players[game.Player2].Exile.SetFaceDown(foretold, true)
	g.ForetoldCards[foretold] = g.Turn.TurnNumber
	unknownExile := addCardToHand(g, game.Player2, vanillaCreature("Unlookable face-down exile", 7, 7))
	g.Players[game.Player2].Hand.Remove(unknownExile)
	g.Players[game.Player2].Exile.Add(unknownExile)
	g.Players[game.Player2].Exile.SetFaceDown(unknownExile, true)
	stackCard := addCardToHand(g, game.Player2, vanillaCreature("Controlled morph spell", 8, 8))
	g.Players[game.Player2].Hand.Remove(stackCard)
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: stackCard,
		Controller: game.Player2, FaceDown: true})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Private library top permission", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectLookAtTopCardAnyTime, AffectedPlayer: game.PlayerYou,
		}}}},
	}})
	controller, victim, other := NewObservation(g, game.Player1), NewObservation(g, game.Player2), NewObservation(g, game.Player3)
	decision := decisionObservation(g, game.Player2, game.DecisionRules)
	if decision.Player != game.Player2 || decision.DecisionMaker != game.Player1 ||
		!hasCardName(decision.Hand(), "Controlled hand") || !hasCardName(decision.HandOf(game.Player1), "Controller hand") ||
		len(decision.HandOf(game.Player3)) != 0 || len(victim.HandOf(game.Player1)) != 0 ||
		!hasCardName(controller.HandOf(game.Player2), "Controlled hand") || len(other.HandOf(game.Player2)) != 0 {
		t.Fatal("hand permissions were not directed, or exposed uninvolved information")
	}
	if len(controller.LibraryLookable(game.Player2)) != 0 || len(victim.LibraryLookable(game.Player2)) != 0 ||
		len(controller.LibraryLookable(game.Player3)) != 0 {
		t.Fatal("turn control exposed an unknown library")
	}
	if top, visible := controller.LibraryTopLookable(game.Player2); !visible || top.Name != "Controlled unknown library" {
		t.Fatal("controller did not inherit the victim's private top permission")
	}
	if _, visible := other.LibraryTopLookable(game.Player2); visible {
		t.Fatal("private top permission leaked to uninvolved player")
	}
	if len(findPermanentView(t, controller, hidden.ObjectID).FaceDownCards) != 1 ||
		len(findPermanentView(t, controller, uninvolved.ObjectID).FaceDownCards) != 0 ||
		len(findPermanentView(t, other, hidden.ObjectID).FaceDownCards) != 0 ||
		len(findPermanentView(t, victim, hidden.ObjectID).FaceDownCards) != 1 {
		t.Fatal("face-down permanent permission leaked or was not inherited")
	}
	if !hasCardName(controller.Exile(game.Player2), "Controlled foretell") ||
		hasCardName(controller.Exile(game.Player2), "Unlookable face-down exile") ||
		hasCardName(other.Exile(game.Player2), "Controlled foretell") {
		t.Fatal("face-down exile visibility exceeded the victim's actual permission")
	}
	if !controller.Stack()[0].FaceDownCard.Exists || controller.Stack()[0].Name != "" ||
		controller.Stack()[0].FaceDownCard.Val.Name != "Controlled morph spell" ||
		other.Stack()[0].FaceDownCard.Exists {
		t.Fatal("face-down stack private information was not permission filtered")
	}
	// Action observers keep their own seat and share precisely CR 723.4 info.
	var notified [game.NumPlayers]bool
	agents := [game.NumPlayers]PlayerAgent{}
	for seat := range agents {
		viewer := game.PlayerID(seat)
		agents[seat] = &controlScript{observed: func(actor game.PlayerID, _ action.Action, obs PlayerObservation) {
			notified[viewer] = true
			if actor != game.Player2 || obs.Player != viewer || obs.DecisionMaker != viewer {
				t.Fatal("observer notification exchanged seats")
			}
			if viewer == game.Player1 && !hasCardName(obs.HandOf(game.Player2), "Controlled hand") {
				t.Fatal("controller observer lost victim information")
			}
			if viewer == game.Player3 && len(obs.HandOf(game.Player2)) != 0 {
				t.Fatal("uninvolved observer saw victim's hand")
			}
		}}
	}
	e.notifyActionObservers(g, agents, game.Player2, action.Pass())
	if !notified[game.Player1] || !notified[game.Player3] || notified[game.Player2] {
		t.Fatal("wrong observers were notified")
	}
	// Read-only snapshots cannot change either player's hand.
	views := decision.Hand()
	views[0].Name = "Tampered"
	if !g.Players[game.Player2].Hand.Contains(controlled) || !g.Players[game.Player1].Hand.Contains(own) ||
		!hasCardName(decision.Hand(), "Controlled hand") {
		t.Fatal("decision observation aliased game data")
	}
	e.advanceToNextTurn(g)
	if len(NewObservation(g, game.Player1).HandOf(game.Player2)) != 0 ||
		len(findPermanentView(t, NewObservation(g, game.Player1), hidden.ObjectID).FaceDownCards) != 0 {
		t.Fatal("private visibility persisted after control expired")
	}
}

func TestMindslaverLibrarySearchControlScopesDoNotFlattenOppositionAgent(t *testing.T) {
	for _, opposed := range []bool{false, true} {
		name := "turn controller"
		if opposed {
			name = "scoped Opposition controller"
		}
		t.Run(name, func(t *testing.T) {
			g, e, _ := controlledMindslaverGame(t)
			if opposed {
				addCombatPermanent(g, game.Player3, oppositionAgentDef())
			}
			nonmatch := addCardToLibrary(g, game.Player2, permanentWitness(types.Enchantment))
			first := addCardToLibrary(g, game.Player2, vanillaCreature("First match", 1, 1))
			second := addCardToLibrary(g, game.Player2, vanillaCreature("Second match", 2, 2))
			addCardToHand(g, game.Player1, vanillaCreature("Controller private", 1, 1))
			addCardToHand(g, game.Player2, vanillaCreature("Victim private", 1, 1))
			addCardToHand(g, game.Player4, vanillaCreature("Uninvolved private", 1, 1))
			turnChooser, searchChooser := &controlScript{}, &controlScript{}
			decider := game.Player1
			if opposed {
				decider = game.Player3
			}
			chooser := turnChooser
			if opposed {
				chooser = searchChooser
			}
			calls := 0
			chooser.choice = func(obs PlayerObservation, req game.ChoiceRequest) []int {
				calls++
				if req.Player != game.Player2 || obs.Player != game.Player2 || obs.DecisionMaker != decider ||
					len(obs.LibraryLookable(game.Player2)) != 3 ||
					!slices.ContainsFunc(obs.LibraryLookable(game.Player2), func(view CardView) bool { return view.CardInstanceID == nonmatch }) ||
					len(obs.HandOf(game.Player4)) != 0 || len(obs.LibraryLookable(game.Player4)) != 0 ||
					!hasCardName(obs.Hand(), "Victim private") {
					t.Fatal("scoped search leaked unrelated information or hid a nonmatching library card")
				}
				if opposed && len(obs.HandOf(game.Player1)) != 0 {
					t.Fatal("Opposition Agent inherited Mindslaver controller's private hand")
				}
				if len(NewObservation(g, game.Player2).LibraryLookable(game.Player2)) != 3 {
					t.Fatal("the searcher lost their own information permission")
				}
				for _, option := range req.Options {
					if option.Label == "Second match" {
						return []int{option.Index}
					}
				}
				t.Fatal("second matching card was not offered")
				return nil
			}
			addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
				SourceZone: zone.Library, Destination: zone.Hand,
				Filter: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			}), nil)
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{turnChooser, &controlScript{}, searchChooser}, &TurnLog{})
			if calls != 1 || !g.Players[game.Player2].Library.Contains(first) ||
				len(g.SearchControlScopes) != 0 || g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 ||
				len(NewObservation(g, game.Player1).LibraryLookable(game.Player2)) != 0 {
				t.Fatal("search control escaped its scope or did not return to turn control")
			}
			if opposed {
				if !g.Players[game.Player2].Exile.Contains(second) || g.Players[game.Player2].Hand.Contains(second) {
					t.Fatal("Mindslaver bypassed Opposition Agent's found-card replacement")
				}
				effect, exists := playFromExilePermissionFor(g, second)
				if !exists || effect.Controller != game.Player3 || len(turnChooser.choices) != 0 {
					t.Fatal("search beneficiary or scoped decision maker was flattened")
				}
			} else if !g.Players[game.Player2].Hand.Contains(second) || len(searchChooser.choices) != 0 {
				t.Fatal("ordinary search did not use turn controller")
			}
		})
	}
}

func TestMindslaverSearchContextPinsPermittedTopAndOnlyKnownHands(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Private top", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectLookAtTopCardAnyTime, AffectedPlayer: game.PlayerYou,
		}}}},
	}})
	hands := make(map[game.PlayerID]id.ID)
	for player := game.Player1; player <= game.Player4; player++ {
		hands[player] = addCardToHand(g, player, vanillaCreature("Hand", 1, 1))
		for range 12 {
			addCardToLibraryNamed(g, player, "Library")
		}

	}
	top, _ := g.Players[game.Player2].Library.Top()
	changed := false
	ctx := e.newSearchContext(g, game.Player2)
	for range 12 {
		world := ctx.Determinize()
		if got, _ := world.Players[game.Player2].Library.Top(); got != top ||
			!world.Players[game.Player2].Hand.Contains(hands[game.Player2]) ||
			!world.Players[game.Player1].Hand.Contains(hands[game.Player1]) {
			t.Fatal("search discarded information the decision maker actually knows")
		}
		changed = changed || !world.Players[game.Player3].Hand.Contains(hands[game.Player3])
	}
	if !changed {
		t.Fatal("uninvolved opponent's actual hand was retained in every sample")
	}
}

func TestMindslaverSearchWorldSamplesUninvolvedFaceDownIdentities(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	known := addFaceDownPermanent(g, game.Player2, vanillaCreature("Known controlled morph", 2, 2), game.FaceDownMorph)
	unknown := addFaceDownPermanent(g, game.Player3, vanillaCreature("Unknown foreign morph", 3, 3), game.FaceDownMorph)
	commander := addFaceDownPermanent(g, game.Player3, vanillaCreature("Public commander", 4, 4), game.FaceDownMorph)
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player3].CommanderInstanceID = commander.CardInstanceID
	libraryCommander := addCardToLibraryNamed(g, game.Player3, "Public library commander")
	g.CommanderIDs[libraryCommander] = true
	for range 8 {
		addCardToLibraryNamed(g, game.Player2, "Controlled library")
		addCardToLibraryNamed(g, game.Player3, "Foreign library")
	}
	ctx := e.newSearchContext(g, game.Player2)
	sampled := false
	for range 10 {
		world := ctx.Determinize()
		if world.CardInstances[known.CardInstanceID].Def.Name != "Known controlled morph" {
			t.Fatal("search changed a visible face-down identity")
		}
		if world.CardInstances[commander.CardInstanceID].Def.Name != "Public commander" ||
			world.CardInstances[libraryCommander].Def.Name != "Public library commander" {
			t.Fatal("search changed a publicly designated commander identity")
		}
		sampled = sampled || world.CardInstances[unknown.CardInstanceID].Def.Name != "Unknown foreign morph"
	}
	if !sampled || g.CardInstances[unknown.CardInstanceID].Def.Name != "Unknown foreign morph" ||
		g.CardInstances[known.CardInstanceID].Def.Name != "Known controlled morph" {
		t.Fatal("search exposed the uninvolved true face-down identity or mutated live definitions")
	}
}

func TestMindslaverOppositionMultiZoneSearchAndGraveyardOnly(t *testing.T) {
	for _, library := range []bool{false, true} {
		t.Run(map[bool]string{false: "graveyard only", true: "both zones"}[library], func(t *testing.T) {
			g, e, _ := controlledMindslaverGame(t)
			addCombatPermanent(g, game.Player3, oppositionAgentDef())
			found := addCardToGraveyard(g, game.Player2, finaleCreatureDef("Found graveyard creature", 2))
			addCardToLibrary(g, game.Player2, finaleLandDef("Hidden library land"))
			turnChoices, searchChoices := 0, 0
			turnController := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				turnChoices++
				if obs.DecisionMaker != game.Player1 || req.Player != game.Player2 {
					t.Fatal("zone choice was stolen before the library search")
				}
				if req.Kind == game.ChoiceZoneSelection {
					if library {
						return []int{0, 1}
					}
					return []int{1}
				}
				return fallbackChoice(req)
			}}
			oppositionController := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				searchChoices++
				if !library || req.Kind != game.ChoiceSearch || obs.DecisionMaker != game.Player3 ||
					len(obs.LibraryLookable(game.Player2)) != 1 {
					t.Fatal("Opposition Agent bypassed or extended its actual library-search scope")
				}
				return []int{0}
			}}
			addInstructionSpellToStackForController(g, game.Player2, finaleSearchShuffleInstructions(), nil)
			obj, _ := g.Stack.Peek()
			obj.XValue = 5
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{turnController, &controlScript{}, oppositionController}, &TurnLog{})
			if library {
				effect, exists := playFromExilePermissionFor(g, found)
				if !g.Players[game.Player2].Exile.Contains(found) || !exists || effect.Controller != game.Player3 ||
					turnChoices != 1 || searchChoices != 1 {
					t.Fatal("multi-zone library search did not exile the graveyard find for Opposition Agent")
				}
			} else if permanentByCardID(g, found) == nil || permanentByCardID(g, found).Controller != game.Player2 ||
				turnChoices != 2 || searchChoices != 0 {
				t.Fatal("graveyard-only search was incorrectly controlled or exiled by Opposition Agent")
			}
			if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 || len(g.SearchControlScopes) != 0 {
				t.Fatal("multi-zone search failed to restore turn control")
			}
		})
	}
}
