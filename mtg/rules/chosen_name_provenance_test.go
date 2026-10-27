package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

const chosenNameRuntimeSearch = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."

var chosenNameRuntimeCompositions = []struct {
	text     string
	searches int
	gain     int
}{
	{"Choose a land you control. You gain 2 life. " + chosenNameRuntimeSearch, 1, 2},
	{"You gain 2 life. Choose a land you control. " + chosenNameRuntimeSearch, 1, 2},
	{"Choose a land you control. " + chosenNameRuntimeSearch + " You gain 2 life.", 1, 2},
	{"Choose a land you control. You gain 2 life. " + chosenNameRuntimeSearch + " You gain 3 life. " + chosenNameRuntimeSearch, 2, 5},
}

type chosenNameAgent struct {
	name     string
	count    int
	requests []game.ChoiceRequest
}

func TestCompiledChosenNameCanoptekPaymentRuntime(t *testing.T) {
	const text = "When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then " + chosenNameRuntimeSearch
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Chosen payment runtime", Layout: "normal", TypeLine: "Artifact Creature", OracleText: text,
		Power: new("2"), Toughness: new("1"),
	})
	for _, tc := range []struct {
		name   string
		pay    bool
		absent bool
		mana   int
		search bool
	}{
		{name: "paid", pay: true, mana: 3, search: true},
		{name: "declined", mana: 3},
		{name: "insufficient", pay: true, mana: 2},
		{name: "absent source", pay: true, absent: true, mana: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			land := addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
			if tc.mana < 3 {
				land.Tapped = true
			}
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.mana)
			if tc.absent {
				movePermanentToZone(g, source, zone.Graveyard)
			}
			pay := 0
			if tc.pay {
				pay = 1
			}
			agent := &choiceOnlyAgent{choices: [][]int{{pay}, {0}, {0, 1}}}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
			log := &TurnLog{}
			NewEngine(nil).resolveTriggeredAbilityBodyWithChoices(g, obj, def, &def.TriggeredAbilities[0],
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, log)
			wantLibrary, wantMana := 2, tc.mana
			if tc.search {
				wantLibrary, wantMana = 0, 0
			}
			if g.Players[game.Player1].Library.Size() != wantLibrary ||
				g.Players[game.Player1].ManaPool.Total() != wantMana {
				t.Fatalf("payment library/mana=%d/%d, want %d/%d",
					g.Players[game.Player1].Library.Size(), g.Players[game.Player1].ManaPool.Total(), wantLibrary, wantMana)
			}
			_, present := permanentByObjectID(g, source.ObjectID)
			if present == (tc.search || tc.absent) {
				t.Fatal("source sacrifice used the wrong source or outcome")
			}
		})
	}
}

func (*chosenNameAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *chosenNameAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	count := a.count
	if request.Prompt == "Choose a permanent" {
		count = 1
	}
	var selected []int
	for _, option := range request.Options {
		if len(selected) < count && option.Label == a.name {
			selected = append(selected, option.Index)
		}
	}
	return selected
}

func chosenNameRuntimeDef(t *testing.T, text string) *game.CardDef {
	t.Helper()
	return legacyCard(t, carddata.ScryfallCard{
		Name: "Chosen name runtime", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	})
}

func TestCompiledChosenNameIndependentRidersRuntime(t *testing.T) {
	for _, tc := range chosenNameRuntimeCompositions {
		for _, name := range []string{"Forest", "Island"} {
			for _, count := range []int{0, 1, 2} {
				t.Run(fmt.Sprintf("%s/%s/%d", tc.text, name, count), func(t *testing.T) {
					def := chosenNameRuntimeDef(t, tc.text)
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					g.Players[game.Player1].Life = 20
					// Physical ownership differs from the chooser's controller.
					chosen := addCombatPermanent(g, game.Player2, basicNamedLand(name))
					chosen.Controller = game.Player1
					addCombatPermanent(g, game.Player3, basicNamedLand(name))
					other := "Island"
					if name == other {
						other = "Forest"
					}
					addCombatPermanent(g, game.Player1, basicNamedLand(other))
					for range 2 * tc.searches {
						addCardToLibrary(g, game.Player1, basicNamedLand(name))
					}
					addCardToLibrary(g, game.Player1, basicNamedLand(other))
					addCardToLibrary(g, game.Player2, basicNamedLand(name))
					agent := &chosenNameAgent{name: name, count: count}
					source := addCardToLibrary(g, game.Player4, def)
					g.Players[game.Player4].Library.Remove(source)
					obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
						SourceID: source, SourceCardID: source, Controller: game.Player1}
					g.Stack.Push(obj)
					NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
					if g.Players[game.Player1].Life != 20+tc.gain ||
						g.Players[game.Player1].Library.Size() != 2*tc.searches+1-count*tc.searches ||
						g.Players[game.Player2].Library.Size() != 1 {
						t.Fatalf("life/library/ownerLibrary=%d/%d/%d", g.Players[game.Player1].Life,
							g.Players[game.Player1].Library.Size(), g.Players[game.Player2].Library.Size())
					}
					arrivals := 0
					for _, permanent := range g.Battlefield {
						if permanent.Tapped && permanentEffectiveName(g, permanent) == name {
							arrivals++
							if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 {
								t.Fatal("arrival borrowed the chosen permanent's owner/controller")
							}
						}
					}
					if arrivals != count*tc.searches || len(agent.requests) != 1+tc.searches {
						t.Fatalf("arrivals/prompts=%d/%d, want %d/%d", arrivals, len(agent.requests), count*tc.searches, 1+tc.searches)
					}
					for _, request := range agent.requests {
						if request.Player != game.Player1 {
							t.Fatal("choice offered to physical owner instead of controller")
						}
					}
				})
			}
		}
	}
}

func TestCompiledChosenNameNoAvailablePermanentDoesNotSearch(t *testing.T) {
	for _, tc := range chosenNameRuntimeCompositions {
		t.Run(tc.text, func(t *testing.T) {
			def := chosenNameRuntimeDef(t, tc.text)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player2, basicNamedLand("Forest"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			agent := &chosenNameAgent{name: "Forest", count: 2}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveInstructionSequence(g, obj, def.SpellAbility.Val.Modes[0].Sequence,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if len(agent.requests) != 0 || g.Players[game.Player1].Library.Size() != 1 ||
				obj.ResolutionChoices[string(game.ResolutionChosenPermanentChoiceKey)].CardName != "" {
				t.Fatal("missing choice borrowed an opponent's permanent or manufactured a name")
			}
		})
	}
}

func TestCompiledChosenNameCapturedNameSurvivesDepartureAndReentry(t *testing.T) {
	def := chosenNameRuntimeDef(t, "Choose a land you control. You gain 2 life. "+chosenNameRuntimeSearch)
	sequence := def.SpellAbility.Val.Modes[0].Sequence
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
	addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	obj := &game.StackObject{Controller: game.Player1}
	agent := &chosenNameAgent{name: "Forest", count: 1}
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	// Execute the original compiled choice, then inspect its captured-value
	// protocol independently of public sentence assembly.
	engine.resolveInstructionSequence(g, obj, sequence[:1], agents, &TurnLog{})
	choice, ok := linkedResolutionChoice(obj, string(game.ResolutionChosenPermanentChoiceKey))
	if !ok || choice.PermanentID != land.ObjectID || choice.CardName != "Forest" {
		t.Fatalf("real choice=%+v available=%v", choice, ok)
	}
	movePermanentToZone(g, land, zone.Graveyard)
	card := g.CardInstances[land.CardInstanceID]
	card.Def = basicNamedLand("Island")
	reentered, entered := newEffectResolver(engine, g, obj, agents, &TurnLog{}).putResolvedCardOnBattlefieldValue(
		card, zone.Graveyard, game.Player1, nil, permanentCreationOptions{})
	if !entered || reentered.ObjectID == choice.PermanentID || reentered.CardInstanceID != land.CardInstanceID {
		t.Fatal("departure reused the original incarnation")
	}
	engine.resolveInstructionSequence(g, obj, sequence[1:], agents, &TurnLog{})
	if g.Players[game.Player1].Library.Size() != 1 {
		t.Fatal("captured name was lost after departure")
	}
	for cardID := range g.Players[game.Player1].Library.Range {
		if g.CardInstances[cardID].Def.Name != "Island" {
			t.Fatal("search used the new/current name instead of captured Forest")
		}
	}
}

func TestCompiledChosenNameRealStackCopyAndGameClone(t *testing.T) {
	def := chosenNameRuntimeDef(t, "Choose a land you control. You gain 2 life. "+chosenNameRuntimeSearch)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
	addCombatPermanent(g, game.Player1, basicNamedLand("Island"))
	addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	source := addCardToLibrary(g, game.Player1, def)
	g.Players[game.Player1].Library.Remove(source)
	original := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
		SourceID: source, SourceCardID: source, Controller: game.Player1}
	g.Stack.Push(original)
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	stackCopy, ok := g.Stack.Peek()
	if !ok || !stackCopy.Copy || stackCopy.ID == original.ID {
		t.Fatal("engine did not create a real stack copy")
	}
	cloned := g.Clone()
	for _, name := range []string{"Island", "Forest"} {
		agent := &chosenNameAgent{name: name, count: 1}
		engine.resolveTopOfStackWithChoices(cloned, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
		if len(agent.requests) != 2 {
			t.Fatal("copy/original did not make its own permanent and search choices")
		}
	}
	if cloned.Players[game.Player1].Library.Size() != 0 || g.Players[game.Player1].Library.Size() != 2 ||
		g.Stack.Size() != 2 || len(original.ResolutionChoices) != 0 || len(stackCopy.ResolutionChoices) != 0 {
		t.Fatal("copy/clone borrowed a choice or mutated the original game's resolution state")
	}
}

func TestCompiledChosenNameSearchRequiresExactChoiceKindAndKey(t *testing.T) {
	def := chosenNameRuntimeDef(t, "Choose a land you control. You gain 2 life. "+chosenNameRuntimeSearch)
	sequence := def.SpellAbility.Val.Modes[0].Sequence
	search, ok := sequence[len(sequence)-1].Primitive.(game.Search)
	if !ok {
		t.Fatal("missing original search instruction")
	}
	for _, tc := range []struct {
		name   string
		key    string
		choice game.ResolutionChoiceResult
	}{
		{name: "missing"},
		{name: "foreign key", key: "another-choice", choice: game.ResolutionChoiceResult{Kind: game.ResolutionChoicePermanent, CardName: "Forest"}},
		{name: "card kind", key: string(game.ResolutionChosenPermanentChoiceKey), choice: game.ResolutionChoiceResult{Kind: game.ResolutionChoiceCard, CardName: "Forest"}},
		{name: "missing name", key: string(game.ResolutionChosenPermanentChoiceKey), choice: game.ResolutionChoiceResult{Kind: game.ResolutionChoicePermanent}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			obj := &game.StackObject{Controller: game.Player1}
			if tc.key != "" {
				rememberResolutionChoice(obj, tc.key, tc.choice)
			}
			agent := &chosenNameAgent{name: "Forest", count: 1}
			receipt := handleSearch(newEffectResolver(NewEngine(nil), g, obj,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{}), search)
			if receipt.succeeded || g.Players[game.Player1].Library.Size() != 1 || len(agent.requests) != 0 {
				t.Fatal("wrong/missing choice kind or key was consumed")
			}
		})
	}
}

func TestChosenNameExistingProtocolMultipleConsumersAcrossIndependentYou(t *testing.T) {
	def := chosenNameRuntimeDef(t, "Choose a land you control. You gain 2 life. "+chosenNameRuntimeSearch)
	original := def.SpellAbility.Val.Modes[0].Sequence
	if len(original) != 3 {
		t.Fatal("changed original choice/gain/search sequence")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
	for range 4 {
		addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	}
	addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	obj := &game.StackObject{Controller: game.Player1}
	agent := &chosenNameAgent{name: "Forest", count: 2}
	// This is a runtime-protocol witness, not public multiple-Search admission.
	sequence := append([]game.Instruction(nil), original...)
	sequence = append(sequence, original[1], original[2])
	NewEngine(nil).resolveInstructionSequence(g, obj, sequence,
		[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if g.Players[game.Player1].Library.Size() != 1 || len(agent.requests) != 3 {
		t.Fatal("compatible consumers lost the real choice across independent YOU")
	}
	for cardID := range g.Players[game.Player1].Library.Range {
		if g.CardInstances[cardID].Def.Name != "Island" {
			t.Fatal("consumer changed the captured name filter")
		}
	}
}
