package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledOwnedSearchTargetDomains(t *testing.T) {
	t.Parallel()
	const search = "Target player searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles."
	for _, prefix := range []string{
		"Target player draws a card.",
		"Exile target creature card from a graveyard.",
		"Counter target spell.",
	} {
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()
			mode := compileOwnedSearchMode(t, prefix+" "+search+" You gain 2 life.")
			search, ok := mode.Sequence[1].Primitive.(game.Search)
			if !ok || search.Player != game.TargetPlayerReference(1) || len(mode.Targets) != 2 {
				t.Fatal("Search player did not retain its exact nonzero announced target")
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var target game.Target
			switch prefix {
			case "Target player draws a card.":
				target = game.PlayerTarget(game.Player3)
				addCardToLibrary(g, game.Player3, bearDef())
			case "Exile target creature card from a graveyard.":
				cardID := addCardToHand(g, game.Player3, bearDef())
				g.Players[game.Player3].Hand.Remove(cardID)
				g.Players[game.Player3].Graveyard.Add(cardID)
				target = game.CardTarget(cardID)
			case "Counter target spell.":
				spell := addStackSpell(g, game.Player3, "Target", []types.Card{types.Sorcery})
				target = game.StackObjectTarget(spell.ID)
			default:
				t.Fatalf("unexpected prefix %q", prefix)
			}
			forest := addCardToLibrary(g, game.Player2, repeatedSearchLandDef("Forest", types.Forest))
			g.CardInstances[forest].Owner = game.Player4
			decoy := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
			source := addInstructionSpellToStackForController(g, game.Player1, mode.Sequence,
				[]game.Target{target, game.PlayerTarget(game.Player2)})
			g.CardInstances[source].Owner = game.Player3
			log := TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g,
				[game.NumPlayers]PlayerAgent{game.Player2: &searchByNameAgent{wanted: "Forest"}}, &log)
			found := permanentForCard(g, forest)
			if found == nil || found.Controller != game.Player2 || found.Owner != game.Player4 || !found.Tapped {
				t.Fatalf("searcher/entry controller/physical owner = %#v", found)
			}
			if !g.Players[game.Player1].Library.Contains(decoy) || g.Players[game.Player1].Life != 42 {
				t.Fatal("Search followed the prefix target or changed the ability-controller rider")
			}
			for _, choice := range log.Choices {
				if choice.Request.Kind == game.ChoiceSearch && choice.Request.Player != game.Player2 {
					t.Fatal("Search was offered to the prefix actor instead of the exact Search target")
				}
			}
		})
	}
}

func TestCompiledOwnedSearchArrivalController(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t,
		"Tap target creature. Search your library for a basic land card, put it onto the battlefield tapped under target player's control, then shuffle. You gain 2 life. Untap that land.")
	search, ok := mode.Sequence[1].Primitive.(game.Search)
	if !ok || search.Player != game.ControllerReference() ||
		search.Controller != opt.Val(game.TargetPlayerReference(1)) {
		t.Fatal("arrival controller was not projected separately from the searcher")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCreaturePermanent(g, game.Player3)
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	g.CardInstances[forest].Owner = game.Player4
	addInstructionSpellToStackForController(g, game.Player1, mode.Sequence,
		[]game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(game.Player2)})
	NewEngine(nil).resolveTopOfStackWithChoices(g,
		[game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}, &TurnLog{})
	found := permanentForCard(g, forest)
	if found == nil || found.Controller != game.Player2 || found.Owner != game.Player4 || found.Tapped || !creature.Tapped {
		t.Fatal("found-card product confused owner, entry controller, searcher, or ordinary prefix target")
	}
}

func TestCompiledOwnedSearchSkippedPublication(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, "If you have no cards in hand, "+compiledOwnedLandSearch+" You gain 2 life. Untap that land.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, repeatedSearchLandDef("Forest", types.Forest))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, SourceID: g.IDGen.Next()}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}
	engine.resolveInstructionSequence(g, obj, mode.Sequence, agents, &TurnLog{})
	found := permanentForCard(g, forest)
	if found == nil || found.Tapped {
		t.Fatal("initial Search did not publish the actual arrived permanent")
	}
	found.Tapped = true
	addCardToHand(g, game.Player1, bearDef())
	engine.resolveInstructionSequence(g, obj, mode.Sequence, agents, &TurnLog{})
	if !found.Tapped || g.Players[game.Player1].Life != 44 {
		t.Fatal("skipped Search reused a previous publication or skipped its independent rider")
	}
}
