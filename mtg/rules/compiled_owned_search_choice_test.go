package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

const compiledOwnedNameSearch = "Choose a land you control. Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."

func TestCompiledOwnedSearchNameChoice(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, "You gain 2 life. "+compiledOwnedNameSearch+" You gain 1 life.")
	choose, ok := mode.Sequence[1].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("choice primitive=%T, want Choose", mode.Sequence[1].Primitive)
	}
	search, ok := mode.Sequence[2].Primitive.(game.Search)
	if !ok {
		t.Fatalf("search primitive=%T, want Search", mode.Sequence[2].Primitive)
	}
	if choose.PublishChoice != game.ResolutionChosenPermanentChoiceKey ||
		search.Spec.NameFromChoice != choose.PublishChoice || search.PublishLinked != "" {
		t.Fatal("name correlation substituted linked-object publication for the existing choice protocol")
	}
	sequence := []game.Instruction{mode.Sequence[1], mode.Sequence[0], mode.Sequence[2], mode.Sequence[3]}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
	addCombatPermanent(g, game.Player2, basicNamedLand("Island"))
	forestA := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	forestB := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	island := addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
	addInstructionSpellToStack(g, sequence)
	clone := g.Clone()
	for _, state := range []*game.Game{g, clone} {
		NewEngine(nil).resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0, 1}}},
		}, &TurnLog{})
		for _, cardID := range []id.ID{forestA, forestB} {
			found := permanentForCard(state, cardID)
			if found == nil || !found.Tapped {
				t.Fatal("the actual declared land name did not filter the searched cards")
			}
		}
		if !state.Players[game.Player1].Library.Contains(island) || state.Players[game.Player1].Life != 43 {
			t.Fatal("name correlation or independent controller riders changed")
		}
	}
}

func TestCompiledOwnedSearchNameChoiceUnavailable(t *testing.T) {
	t.Parallel()
	mode := compileOwnedSearchMode(t, compiledOwnedNameSearch)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	addInstructionSpellToStack(g, mode.Sequence)
	clone := g.Clone()
	for _, state := range []*game.Game{g, clone} {
		log := TurnLog{}
		NewEngine(nil).resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{}, &log)
		if !state.Players[game.Player1].Library.Contains(forest) || len(log.Choices) != 0 {
			t.Fatal("an unavailable copied choice substituted a land name or made a Search choice")
		}
	}
}

func TestCompiledCanoptekOwnedSearchPayment(t *testing.T) {
	t.Parallel()
	const oracle = "Wraith Form — This creature can't be blocked.\nTransdimensional Scout — When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	defs := legacyCardDefs(t, carddata.ScryfallCard{
		Name: "Canoptek Wraith", Layout: "normal", TypeLine: "Artifact Creature — Wraith",
		OracleText: oracle, ManaCost: "{3}", Power: new("2"), Toughness: new("1"),
	})

	for _, tt := range []struct {
		name string
		mana int
		pay  bool
	}{
		{"paid", 3, true},
		{"declined", 3, false},
		{"unavailable", 2, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, defs[0])
			addCombatPermanent(g, game.Player1, basicNamedLand("Forest"))
			a := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			b := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
			g.Players[game.Player1].ManaPool.Add(mana.C, tt.mana)
			dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 2, true)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("actual compiled combat trigger was not queued")
			}
			choices := [][]int{{0}}
			if tt.pay {
				choices = [][]int{{1}, {0}, {0, 1}}
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: choices},
			}, &log)
			_, present := permanentByObjectID(g, source.ObjectID)
			if present == tt.pay {
				t.Fatal("compiled payment did not retain exact source-sacrifice outcome")
			}
			for _, cardID := range []id.ID{a, b} {
				found := permanentForCard(g, cardID)
				if (found != nil) != tt.pay || found != nil && !found.Tapped {
					t.Fatal("paid/declined Search did not retain its actual choice and payment gate")
				}
			}
			if !tt.pay {
				for _, choice := range log.Choices {
					if choice.Request.Kind == game.ChoiceSearch {
						t.Fatal("declined or unavailable payment reused a stale name choice")
					}
				}
			}
			wantMana := tt.mana
			if tt.pay {
				wantMana -= 3
			}
			if g.Players[game.Player1].ManaPool.Total() != wantMana {
				t.Fatal("payment spent mana outside the accepted source-sacrifice envelope")
			}
		})
	}
}
