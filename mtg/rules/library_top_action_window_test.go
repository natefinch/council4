package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3625 (sub-issue of #3562). Default-registry cards
// checked against the pinned corpus 6da9ced0…a1b4 (oracle_id): Courser of
// Kruphix 46779609-4fa7-4fd2-b5b4-7d4d749339e6 {1}{G}{G} "Play with the top
// card of your library revealed. You may play lands from the top of your
// library. Landfall — …"; Future Sight 93a0e938-b916-47ec-9803-369e885e253c;
// Bolas's Citadel 2bd111bb-ce02-414c-b5b7-e0e037d8d96b; Thriving Bluff
// 91fceb34-0f2d-4392-be27-00dcd765637f "This land enters tapped. As it enters,
// choose a color other than red. …"; Millikin
// fa4dffda-6f04-4d0b-829d-28a1a5794dee "{T}, Mill a card: Add {C}. …" and
// Deranged Assistant 9e1b0034-bd2c-4412-8d81-db3306c8814f, the same ability,
// not mana abilities (605.1a; Millikin ruling 2026-08-05, see
// library_moving_mana_ability_test.go); Phyrexian Altar
// 8d02b297-97c4-4379-9862-0a462400f66f "Sacrifice a creature: Add one mana of
// any color.", still a mana ability when the test creature Homing Golem's own
// replacement effect puts the sacrificed Golem on top of its owner's library
// (605.1a ignores replacement effects other than the ability's own); Mana
// Confluence d0ee5bdc-2b69-4b73-9a20-ffcc18783b29 "{T}, Pay 1 life: Add one
// mana of any color."; Sinister Concoction ffc3beec-6485-428d-a251-fa0bc21263a5 {B} "{B}, Pay 1
// life, Mill a card, Discard a card, Sacrifice this enchantment: Destroy target
// creature."; Saw It Coming 90edaf33-d0ab-47e0-8f6a-6fba38286e6e (Foretell
// {1}{U}); Slickshot Show-Off 74c7a029-59aa-4511-8e4c-f8f167cda4e2 (Plot
// {1}{R}); Willbender 0aae277e-e58e-4115-b5fd-0459451e17ec (Morph {1}{U}).
// Rulings: Courser of Kruphix 2021-03-19 "If the top card of your library
// changes while you're casting a spell, playing a land, or activating an
// ability, you don't reveal the new top card until you finish doing so. This
// means that if you play a land from the top of your library, you can't look
// at the next one until you've handled any replacement effects that modify
// that land entering the battlefield …"; Future Sight 2019-06-14, the same;
// Sinister Concoction 2016-04-08 "You may pay Sinister Concoction's costs in
// any order. …". Comprehensive Rules (2025-09-19 text): 401.5 "… The same is
// true with relation to an ability being activated. If the top card of the
// player's library changes while a player is taking a special action (see
// rule 116, "Special Actions"), the new card won't be revealed and can't be
// looked at until the player has finished taking that special action."; 116.2a
// playing a land, 116.2b turning a face-down creature face up, 116.2f suspend,
// 116.2h foretell and 116.2k plot are special actions; 602.2b activation
// follows 601.2b–i; 605.3a–b a mana ability may be activated while paying and
// resolves immediately.

// libraryTopActionGame gives Player1 main-phase priority with permission
// (Future Sight, Courser of Kruphix or Bolas's Citadel) on the battlefield and
// a known card under top on their library, returning that card's ID and top's.
func libraryTopActionGame(t *testing.T, permission string, top *game.CardDef) (*game.Game, *Engine, id.ID, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, registryCardDef(t, permission))
	next := addCardToLibraryNamed(g, game.Player1, "Next Card")
	var topID id.ID
	if top != nil {
		topID = addCardToLibrary(g, game.Player1, top)
	} else {
		topID = addCardToLibraryNamed(g, game.Player1, "Top Card")
	}
	giveMainPhasePriority(g, game.Player1)
	return g, NewEngine(nil), topID, next
}

// assertTopShown requires every observer to be shown want on top of Player1's
// library after the action finished: revealed under Future Sight or Courser,
// or privately to Player1 alone under Bolas's Citadel.
func assertTopShown(t *testing.T, g *game.Game, permission string, want id.ID) {
	t.Helper()
	assertNoLibraryTopWindow(t, g)
	owner, opponent := NewObservation(g, game.Player1), NewObservation(g, game.Player2)
	if permission == "Bolas's Citadel" {
		if card, ok := owner.LibraryTopLookable(game.Player1); !ok || card.CardInstanceID != want {
			t.Fatalf("Player1 can't look at %v on top after the action", want)
		}
		if topCardShown(opponent, game.Player1) != 0 {
			t.Fatal("Citadel's private look reached the opponent")
		}
		return
	}
	for _, o := range []PlayerObservation{owner, opponent} {
		if card, ok := o.LibraryTopRevealed(game.Player1); !ok || card.CardInstanceID != want {
			t.Fatalf("viewer %d isn't shown %v on top after the action", o.viewer, want)
		}
	}
}

// TestLibraryTopChangedByLandPlayStaysHidden plays Thriving Bluff off the top
// of Player1's library under each play-from-top permission. Playing a land is
// a special action (CR 116.2a); the card below becomes the new top as the land
// leaves the library, and stays hidden from everyone while Thriving Bluff's
// "As it enters, choose a color" choice is made, being revealed (or lookable)
// once the land play finishes (CR 401.5).
func TestLibraryTopChangedByLandPlayStaysHidden(t *testing.T) {
	t.Parallel()
	for _, permission := range []string{"Courser of Kruphix", "Future Sight", "Bolas's Citadel"} {
		t.Run(permission, func(t *testing.T) {
			t.Parallel()
			g, e, bluff, next := libraryTopActionGame(t, permission, registryCardDef(t, "Thriving Bluff"))
			agent := &libraryTopWindowAgent{owner: game.Player1, start: bluff}

			if !e.applyActionWithChoices(g, game.Player1, action.PlayLandFaceFromZone(bluff, zone.Library, game.FaceFront), agent.agents(), &TurnLog{}) {
				t.Fatal("Thriving Bluff wasn't played from the library")
			}

			agent.assertHiddenWhileChanged(t)
			if _, ok := findPermanentByCardID(g, bluff); !ok || g.Turn.LandsPlayedThisTurn != 1 {
				t.Fatal("Thriving Bluff isn't on the battlefield as the turn's land play")
			}
			assertTopShown(t, g, permission, next)
		})
	}
}

// sourceActivation returns the legal activation of source's ability for player.
func sourceActivation(t *testing.T, e *Engine, g *game.Game, source id.ID) action.Action {
	t.Helper()
	for _, act := range e.legalActions(g, game.Player1) {
		if activate, ok := act.ActivateAbilityPayload(); ok && activate.SourceID == source {
			return act
		}
	}
	t.Fatalf("no legal activation of %v", source)
	return action.Action{}
}

// TestLibraryTopChangedByActivationStaysHidden activates abilities that
// change Player1's revealed top card: Phyrexian Altar's mana ability
// sacrificing Homing Golem, whose own replacement effect puts it on top of the
// library (still a mana ability, CR 605.1a), with a color chosen as it
// resolves (CR 605.3b); and abilities whose costs mill the top card, Millikin's
// and Deranged Assistant's (not mana abilities, so put on the stack, CR 602.2)
// and Sinister Concoction's. The new top stays hidden from everyone through the
// rest of the activation, a mana ability's resolution included, and is shown
// as soon as the activation finishes (CR 401.5, 602.2b).
func TestLibraryTopChangedByActivationStaysHidden(t *testing.T) {
	t.Parallel()
	type activationRoute struct {
		name   string
		source func(t *testing.T) *game.CardDef
		// changedChoice and unchangedChoice require a choice made after, or
		// before, the top card changed.
		changedChoice, unchangedChoice bool
		prepare                        func(t *testing.T, g *game.Game)
		// finished checks the completed activation and returns the card then
		// on top.
		finished func(t *testing.T, g *game.Game, source *game.Permanent, top, next id.ID) id.ID
	}
	milled := func(t *testing.T, g *game.Game, top, next id.ID) id.ID {
		t.Helper()
		if !g.Players[game.Player1].Graveyard.Contains(top) {
			t.Fatal("the activation's cost didn't mill the top card")
		}
		return next
	}
	routes := []activationRoute{
		{
			name:          "Phyrexian Altar mana ability resolving with a choice",
			source:        func(t *testing.T) *game.CardDef { return registryCardDef(t, "Phyrexian Altar") },
			changedChoice: true,
			prepare: func(t *testing.T, g *game.Game) {
				addCombatPermanent(g, game.Player1, homingGolemDef())
			},
			finished: func(t *testing.T, g *game.Game, _ *game.Permanent, top, _ id.ID) id.ID {
				t.Helper()
				library := g.Players[game.Player1].Library.All()
				if len(library) != 3 || g.CardInstances[library[0]].Def.Name != homingGolem || library[1] != top {
					t.Fatalf("library = %v, want Homing Golem put on top over %v", library, top)
				}
				if g.Stack.Size() != 0 || g.Players[game.Player1].ManaPool.Total() != 1 {
					t.Fatal("Phyrexian Altar's mana ability didn't add one mana without using the stack")
				}
				return library[0]
			},
		},
		{
			name:            "Sinister Concoction",
			source:          func(t *testing.T) *game.CardDef { return registryCardDef(t, "Sinister Concoction") },
			unchangedChoice: true,
			prepare: func(t *testing.T, g *game.Game) {
				addCombatPermanent(g, game.Player2, vanillaCreature("Victim", 2, 2))
				addCardToHand(g, game.Player1, vanillaCreature("First Spare", 1, 1))
				addCardToHand(g, game.Player1, vanillaCreature("Second Spare", 1, 1))
				g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			},
			finished: func(t *testing.T, g *game.Game, _ *game.Permanent, top, next id.ID) id.ID {
				t.Helper()
				if obj, ok := g.Stack.Peek(); !ok || obj.Kind != game.StackActivatedAbility || g.Players[game.Player1].Life != 19 {
					t.Fatalf("stack top = %+v, want Sinister Concoction's ability activated for 1 life", obj)
				}
				return milled(t, g, top, next)
			},
		},
	}
	for _, name := range libraryMovingManaProducers {
		routes = append(routes, activationRoute{
			name:   name + "'s ability",
			source: func(t *testing.T) *game.CardDef { return registryCardDef(t, name) },
			finished: func(t *testing.T, g *game.Game, source *game.Permanent, top, next id.ID) id.ID {
				t.Helper()
				obj, ok := g.Stack.Peek()
				if !source.Tapped || !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != source.ObjectID ||
					g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatalf("stack top = %+v, want %s tapped and its ability on the stack, not yet adding {C}", obj, name)
				}
				return milled(t, g, top, next)
			},
		})
	}
	for _, permission := range []string{"Future Sight", "Bolas's Citadel"} {
		for _, route := range routes {
			t.Run(permission+"/"+route.name, func(t *testing.T) {
				t.Parallel()
				g, e, top, next := libraryTopActionGame(t, permission, nil)
				g.Players[game.Player1].Life = 20
				source := addCombatPermanent(g, game.Player1, route.source(t))
				if route.prepare != nil {
					route.prepare(t, g)
				}
				agent := &libraryTopWindowAgent{owner: game.Player1, start: top}

				if !e.applyActionWithChoices(g, game.Player1, sourceActivation(t, e, g, source.ObjectID), agent.agents(), &TurnLog{}) {
					t.Fatal("the activation failed")
				}

				newTop := route.finished(t, g, source, top, next)
				if route.changedChoice {
					agent.assertHiddenWhileChanged(t)
				} else if len(agent.leaks) != 0 {
					t.Fatalf("the changed top card was shown during activation: %v", agent.leaks)
				}
				if route.unchangedChoice && agent.unchanged == 0 {
					t.Fatal("no choice was made before the top card changed")
				}
				if agent.unchangedVisible != agent.unchanged {
					t.Fatalf("the unchanged top was shown at %d of %d choices, want all", agent.unchangedVisible, agent.unchanged)
				}
				assertTopShown(t, g, permission, newTop)
			})
		}
	}
}

// nestedManaWindowAgent pays a madness cast's {2} in its payment-time mana
// window by activating Phyrexian Altar, sacrificing Homing Golem, which goes
// on top of the library over the revealed top card, then Mana Confluence,
// whose color choice is made while Confluence's own window is nested inside
// the cast's.
type nestedManaWindowAgent struct {
	*libraryTopWindowAgent
	activated   []string
	colorChoice bool
	depth       int
}

func (a *nestedManaWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.check(obs, request.Prompt)
	if selected, ok := sacrificeNamed(request, homingGolem); ok {
		return selected
	}
	if request.Prompt == "Choose a color" {
		a.colorChoice = true
		a.depth = len(obs.g.LibraryTopWindows)
	}
	if request.Prompt == "Activate a mana ability while paying?" && len(a.activated) < 2 {
		want := []string{"Sacrifice a creature", "Pay 1 life"}[len(a.activated)]
		for _, option := range request.Options {
			if strings.Contains(option.Label, want) {
				a.activated = append(a.activated, want)
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// TestNestedManaAbilityKeepsOuterLibraryTopSnapshot casts Sol Ring for its
// madness cost under Future Sight, paying with Phyrexian Altar and then Mana
// Confluence in the payment-time mana window (CR 605.3a). The Altar's
// sacrifice of Homing Golem puts the Golem on top of the library (its own
// replacement effect, so the Altar's ability is still a mana ability, CR
// 605.1a). Confluence's activation opens a window nested in the cast's after
// the top already changed; the cast's snapshot governs, so the new top stays
// hidden through Confluence's color choice and is revealed once the spell
// becomes cast (CR 401.5, 601.2i).
func TestNestedManaAbilityKeepsOuterLibraryTopSnapshot(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	giveMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].Life = 20
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Future Sight"))
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
	golem := addCombatPermanent(g, game.Player1, homingGolemDef()).CardInstanceID
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Mana Confluence"))
	addCardToLibraryNamed(g, game.Player1, "Next Card")
	start := addCardToLibraryNamed(g, game.Player1, "Revealed Card")
	solRing := addCardToExile(g, game.Player1, registryCardDef(t, "Sol Ring"))
	agent := &nestedManaWindowAgent{libraryTopWindowAgent: &libraryTopWindowAgent{owner: game.Player1, start: start}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	if !NewEngine(nil).castMadnessSpellWithChoices(g, game.Player1, g.CardInstances[solRing], cost.Mana{cost.O(2)}, agents, &TurnLog{}) {
		t.Fatal("Sol Ring wasn't cast for its madness cost")
	}

	if top, _ := g.Players[game.Player1].Library.Top(); len(agent.activated) != 2 || top != golem {
		t.Fatalf("activated %v, want Phyrexian Altar (putting Homing Golem on top) then Mana Confluence", agent.activated)
	}
	if !agent.colorChoice || agent.depth != 2 {
		t.Fatalf("Confluence's color choice made = %t with %d windows open, want it inside the cast's window", agent.colorChoice, agent.depth)
	}
	agent.assertHiddenWhileChanged(t)
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != solRing {
		t.Fatalf("stack top = %+v, want Sol Ring cast", obj)
	}
	assertTopShown(t, g, "Future Sight", golem)
}

// TestUnchangedLibraryTopStaysShownThroughSpecialActions takes each special
// action that doesn't change a library, and an activation that doesn't either,
// under Future Sight and Bolas's Citadel. The top card stays shown at every
// choice made during the action and afterwards, and the action's window is
// closed whether it succeeds or fails (CR 401.5 withholds only a new top).
func TestUnchangedLibraryTopStaysShownThroughSpecialActions(t *testing.T) {
	t.Parallel()
	routes := []struct {
		name string
		// take sets up the action and returns it, and whether it should succeed.
		take func(t *testing.T, g *game.Game) (action.Action, bool)
	}{
		{name: "foretell", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			g.Players[game.Player1].ManaPool.Add(mana.U, 2)
			return action.ForetellCard(addCardToHand(g, game.Player1, registryCardDef(t, "Saw It Coming"))), true
		}},
		{name: "foretell without mana", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			return action.ForetellCard(addCardToHand(g, game.Player1, registryCardDef(t, "Saw It Coming"))), false
		}},
		{name: "plot", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			g.Players[game.Player1].ManaPool.Add(mana.R, 2)
			return action.PlotCard(addCardToHand(g, game.Player1, registryCardDef(t, "Slickshot Show-Off"))), true
		}},
		{name: "suspend", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			return action.SuspendCard(addCardToHand(g, game.Player1, suspendSorcery(3, cost.Mana{cost.G}))), true
		}},
		{name: "turn face up", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			willbender := addCombatPermanent(g, game.Player1, registryCardDef(t, "Willbender"))
			willbender.FaceDown, willbender.FaceDownKind, willbender.FaceDownFace = true, game.FaceDownMorph, game.FaceFront
			g.Players[game.Player1].ManaPool.Add(mana.U, 2)
			return action.TurnFaceUp(willbender.ObjectID), true
		}},
		{name: "land play from hand", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			return action.PlayLand(addCardToHand(g, game.Player1, registryCardDef(t, "Thriving Bluff"))), true
		}},
		{name: "land play with none left", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			g.Turn.LandsPlayedThisTurn = 1
			return action.PlayLand(addCardToHand(g, game.Player1, registryCardDef(t, "Thriving Bluff"))), false
		}},
		{name: "Sol Ring's mana ability", take: func(t *testing.T, g *game.Game) (action.Action, bool) {
			solRing := addCombatPermanent(g, game.Player1, registryCardDef(t, "Sol Ring"))
			return action.ActivateAbility(solRing.ObjectID, 0, nil, 0), true
		}},
	}
	for _, permission := range []string{"Future Sight", "Bolas's Citadel"} {
		for _, route := range routes {
			t.Run(permission+"/"+route.name, func(t *testing.T) {
				t.Parallel()
				g, e, top, _ := libraryTopActionGame(t, permission, nil)
				act, succeeds := route.take(t, g)
				agent := &libraryTopWindowAgent{owner: game.Player1, start: top}

				if got := e.applyActionWithChoices(g, game.Player1, act, agent.agents(), &TurnLog{}); got != succeeds {
					t.Fatalf("applied = %t, want %t", got, succeeds)
				}

				if agent.changed != 0 || agent.unchangedVisible != agent.unchanged {
					t.Fatalf("choices: %d unchanged (%d shown), %d changed; want the unchanged top shown at each",
						agent.unchanged, agent.unchangedVisible, agent.changed)
				}
				assertTopShown(t, g, permission, top)
			})
		}
	}
}
