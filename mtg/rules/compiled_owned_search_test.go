package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const compiledOwnedLandSearch = "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle."

func compileOwnedSearchMode(t *testing.T, text string) game.Mode {
	t.Helper()
	defs := legacyCardDefs(t, carddata.ScryfallCard{
		Name: "Owned Search", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
	})

	return defs[0].SpellAbility.Val.Modes[0]
}

func TestCompiledOwnedSearchPositions(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		compiledOwnedLandSearch,
		"Draw a card. " + compiledOwnedLandSearch,
		compiledOwnedLandSearch + " Scry 1.",
		"Draw a card. " + compiledOwnedLandSearch + " Scry 1.",
		"You gain 2 life. " + compiledOwnedLandSearch + " Create a 1/1 white Soldier creature token.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			mode := compileOwnedSearchMode(t, text)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, bearDef())
			forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
			addCardToLibrary(g, game.Player1, bearDef())
			g.CardInstances[forest].Owner = game.Player2
			source := addInstructionSpellToStack(g, mode.Sequence)
			g.CardInstances[source].Owner = game.Player3
			log := TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g,
				[game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}, &log)
			found := permanentForCard(g, forest)
			if found == nil || !found.Tapped || found.Controller != game.Player1 || found.Owner != game.Player2 {
				t.Fatalf("actual searched permanent = %#v", found)
			}
			if !g.Players[game.Player3].Graveyard.Contains(source) {
				t.Fatal("resolving spell's physical owner was lost")
			}
			for _, choice := range log.Choices {
				if choice.Request.Player != game.Player1 {
					t.Fatalf("resolving choice actor = %v", choice.Request.Player)
				}
			}
		})
	}
}

func TestCompiledOwnedSearchGroupTruth(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t,
		"If you have no cards in hand, search your library for a basic land card, reveal it, put it into your hand, then shuffle, then draw a card. "+
			"If you have no cards in hand, you gain 2 life.")
	if len(mode.Sequence) != 3 || mode.Sequence[0].PublishCondition == "" ||
		mode.Sequence[1].ConditionGate != mode.Sequence[0].PublishCondition || !mode.Sequence[2].Condition.Exists {
		t.Fatal("collapsed Search did not preserve group capture and independent reevaluation")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	bear := addCardToLibrary(g, game.Player1, bearDef())
	addInstructionSpellToStack(g, mode.Sequence)
	NewEngine(nil).resolveTopOfStackWithChoices(g,
		[game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(forest) || !g.Players[game.Player1].Hand.Contains(bear) ||
		g.Players[game.Player1].Life != 40 {
		t.Fatal("group draw reevaluated after Search, or independent life clause reused the old truth")
	}
}

func TestCompiledOwnedSearchActualProduct(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, search                                   string
		decline, empty, skipped, protected, redirected bool
	}{
		{name: "found", search: compiledOwnedLandSearch},
		{name: "declined", search: compiledOwnedLandSearch, decline: true},
		{name: "empty", search: compiledOwnedLandSearch, empty: true},
		{name: "skipped", search: "If you have no cards in hand, " + compiledOwnedLandSearch, skipped: true},
		{name: "protected entry", search: "Search your library for a creature card, put it onto the battlefield tapped, then shuffle.", protected: true},
		{name: "redirected entry", search: compiledOwnedLandSearch, redirected: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			noun := "land"
			if tt.protected {
				noun = "creature"
			}
			mode := compileOwnedSearchMode(t, "Tap target creature. "+tt.search+" You gain 2 life. Untap that "+noun+". Untap target land.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			decoy := addCreaturePermanent(g, game.Player2)
			land := addBasicLandPermanent(g, game.Player2, types.Island)
			land.Tapped = true
			var cardID id.ID
			wanted := "Forest"
			if !tt.empty {
				def := repeatedSearchLandDef(wanted, types.Forest)
				if tt.protected {
					def, wanted = bearDef(), "Bear"
					grafdiggersCagePermanent(g, game.Player2)
				}
				cardID = addCardToLibrary(g, game.Player1, def)
			}
			if tt.skipped {
				addCardToHand(g, game.Player1, bearDef())
			}
			if tt.decline {
				wanted = ""
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: wanted}}
			if tt.redirected {
				addCombatPermanent(g, game.Player2, oppositionAgentDef())
				agents[game.Player2] = &searchByNameAgent{wanted: wanted}
			}
			addInstructionSpellToStackForController(g, game.Player1, mode.Sequence, []game.Target{
				game.PermanentTarget(decoy.ObjectID), game.PermanentTarget(land.ObjectID),
			})
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)
			if !decoy.Tapped || land.Tapped {
				t.Fatal("compression confused ordinary nonzero target slots with the actual search product")
			}
			found := permanentForCard(g, cardID)
			wantFound := !tt.decline && !tt.empty && !tt.skipped && !tt.protected && !tt.redirected
			if (found != nil) != wantFound || found != nil && found.Tapped {
				t.Fatalf("actual product = %#v, want found=%v and untapped", found, wantFound)
			}
			if g.Players[game.Player1].Life != 42 {
				t.Fatal("failed or skipped Search swallowed the independent rider")
			}
			if tt.skipped && len(log.Choices) != 0 {
				t.Fatal("skipped Search made a choice")
			}
		})
	}
}

type ownedSearchReentryAgent struct {
	g          *game.Game
	engine     *Engine
	cardID     id.ID
	reentry    *game.Permanent
	skipSecond bool
	searches   int
}

func (*ownedSearchReentryAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *ownedSearchReentryAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceSearch {
		a.searches++
		if a.skipSecond && a.searches == 2 {
			if found := permanentForCard(a.g, a.cardID); found != nil {
				found.Tapped = true
			}
			return nil
		}
		for _, option := range request.Options {
			if option.Label == a.g.CardInstances[a.cardID].Def.Name {
				return []int{option.Index}
			}
		}
		return nil
	}
	if a.reentry == nil {
		original := permanentForCard(a.g, a.cardID)
		if original != nil {
			obj := &game.StackObject{Controller: game.Player1, SourceID: original.ObjectID, SourceCardID: a.cardID}
			r := &effectResolver{engine: a.engine, game: a.g, obj: obj, log: &TurnLog{}}
			handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile})
			handlePutOnBattlefield(r, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})})
			a.reentry = permanentForCard(a.g, a.cardID)
			if a.reentry != nil {
				a.reentry.Tapped = true
			}
		}
	}
	return nil
}

func TestCompiledOwnedSearchDoesNotFollowReentry(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, compiledOwnedLandSearch+" Scry 1. Untap that land.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	addCardToLibrary(g, game.Player1, bearDef())
	agent := &ownedSearchReentryAgent{g: g, engine: engine, cardID: forest}
	addInstructionSpellToStack(g, mode.Sequence)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if agent.reentry == nil || !agent.reentry.Tapped {
		t.Fatal("linked untap followed the searched card into a later incarnation")
	}
}

func TestCompiledOwnedSearchIndependentProducts(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, compiledOwnedLandSearch+" You gain 1 life. Untap that land. "+
		compiledOwnedLandSearch+" You gain 1 life. Untap that land.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	island := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Island", types.Island))
	addInstructionSpellToStack(g, mode.Sequence)
	NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, cardID := range []id.ID{forest, island} {
		found := permanentForCard(g, cardID)
		if found == nil || found.Tapped {
			t.Fatal("independent Search products did not retain their own consumers")
		}
	}
	if g.Players[game.Player1].Life != 42 {
		t.Fatal("independent units lost their riders")
	}
}

func TestCompiledOwnedSearchFailedSecondProductIsIsolated(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, compiledOwnedLandSearch+" You gain 1 life. Untap that land. "+
		compiledOwnedLandSearch+" You gain 1 life. Untap that land.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Island", types.Island))
	agent := &ownedSearchReentryAgent{g: g, engine: engine, cardID: forest, skipSecond: true}
	addInstructionSpellToStack(g, mode.Sequence)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	first := permanentForCard(g, forest)
	if first == nil || !first.Tapped || agent.searches != 2 {
		t.Fatalf("failed second Search product: first=%#v search choices=%d", first, agent.searches)
	}
}
