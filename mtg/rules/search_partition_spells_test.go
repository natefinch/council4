package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Oracle SHA-256: 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// WotC rulings retrieved 2026-10-14: Intuition 2006-07-15; Realms 2010-06-15.
// Official CR effective 2026-09-25: 201.2b, 400.7, 601.2c, 608.2b–d,
// 609.3, 614.5–6, 616.1, 701.20, 701.23a/b/d/e, 702.11c and 702.18a.

func searchPartitionEngine(t *testing.T, entry cardset.Entry) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(entry.Programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func searchPartitionTargets(entry cardset.Entry) []game.Target {
	if entry.OracleID == authored.Intuition.OracleID {
		return []game.Target{game.PlayerTarget(game.Player3)}
	}
	return nil
}

func searchPartitionReveals(g *game.Game) []game.ObjectID {
	var result []game.ObjectID
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.FromZone == zone.Library {
			result = append(result, event.CardID)
		}
	}
	return result
}

func assertSearchPartitionBatch(t *testing.T, g *game.Game, found []game.ObjectID) {
	t.Helper()
	var batch game.ObjectID
	moved := 0
	for _, event := range g.Events {
		if event.Kind != game.EventZoneChanged || event.FromZone != zone.Library ||
			!slices.Contains(found, event.CardID) {
			continue
		}
		moved++
		if len(found) > 1 {
			if event.SimultaneousID == 0 || batch != 0 && batch != event.SimultaneousID {
				t.Fatal("search partition was not one simultaneous placement")
			}
			batch = event.SimultaneousID
		}
		card, _ := g.GetCardInstance(event.CardID)
		if event.CardZoneVersion == 0 || event.CardZoneVersion != card.ZoneVersion {
			t.Fatal("placement did not create an actual arrived incarnation")
		}
	}
	if moved != len(found) {
		t.Fatalf("moved %d revealed cards, want %d", moved, len(found))
	}
}

func TestAuthoredIntuitionForcedSearchQuantity(t *testing.T) {
	t.Parallel()
	for _, librarySize := range []int{0, 1, 2, 3, 5} {
		for _, answer := range []string{"none", "too few", "correct"} {
			t.Run(fmt.Sprintf("%d/%s", librarySize, answer), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := searchPartitionEngine(t, authored.Intuition)
				for index := range librarySize {
					addCardToLibrary(g, game.Player1, vanillaCreature(fmt.Sprintf("Card %d", index), 1, 1))
				}
				before := slices.Clone(g.Players[game.Player1].Library.All())
				epoch := g.Players[game.Player1].Library.ReorderVersion()
				count := min(3, librarySize)
				searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
					if req.Kind == game.ChoiceOrder {
						return req.DefaultSelection
					}
					if req.Kind != game.ChoiceSearch || req.MinChoices != count || req.MaxChoices != count ||
						len(req.Options) != librarySize || len(obs.LibraryLookable(game.Player1)) != librarySize {
						t.Fatal("Intuition did not require as many cards as possible")
					}
					if answer == "none" {
						return nil
					}
					if answer == "too few" {
						return []int{0}
					}
					return slices.Clone(req.DefaultSelection)
				}}
				opponent := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
					if req.Player != game.Player3 || obs.Player != game.Player3 || req.MinChoices != 1 ||
						req.MaxChoices != 1 || len(req.Options) != count {
						t.Fatal("targeted opponent did not choose exactly one of the actual finds")
					}
					return []int{count - 1}
				}}
				spell := addImplementationSpellToStack(g, game.Player1, authored.Intuition.New(), searchPartitionTargets(authored.Intuition))
				log := &TurnLog{}
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, nil, opponent}, log)
				found := before[:count]
				if !slices.Equal(searchPartitionReveals(g), found) || !g.Players[game.Player1].Graveyard.Contains(spell) ||
					g.Players[game.Player1].Library.Size() != librarySize-count ||
					g.Players[game.Player1].Library.ReorderVersion() != epoch+1 {
					t.Fatalf("exact finds/shuffle failed: reveals=%v want=%v spellInGrave=%v library=%d epoch=%d", searchPartitionReveals(g), found, g.Players[game.Player1].Graveyard.Contains(spell), g.Players[game.Player1].Library.Size(), g.Players[game.Player1].Library.ReorderVersion())
				}
				for index, card := range found {
					if g.Players[game.Player1].Hand.Contains(card) != (index == count-1) ||
						g.Players[game.Player1].Graveyard.Contains(card) != (index != count-1) {
						t.Fatal("opponent's chosen card and the rest were misplaced")
					}
				}
				assertSearchPartitionBatch(t, g, found)
				if count == 0 {
					if len(log.Choices) != 0 {
						t.Fatal("empty search asked for an impossible selection")
					}
				} else {
					wantFallback := answer == "none" || answer == "too few" && count > 1
					wantChoices := 2
					if count > 2 {
						wantChoices++
					}
					if len(log.Choices) != wantChoices || log.Choices[0].UsedFallback != wantFallback ||
						log.Choices[1].UsedFallback {
						t.Fatal("invalid undersized search was not recorded as fallback, or valid opponent choice was lost")
					}
				}
			})
		}
	}
}

func TestAuthoredIntuitionCastAndTargetLegality(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"opponent", "self", "no target", "too many targets", "hexproof", "shroud", "departed", "no legal opponents", "becomes illegal", "leaves before resolution"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := searchPartitionEngine(t, authored.Intuition)
			for index := range 5 {
				addCardToLibrary(g, game.Player1, vanillaCreature(fmt.Sprintf("Candidate %d", index), 1, 1))
			}
			before := slices.Clone(g.Players[game.Player1].Library.All())
			epoch := g.Players[game.Player1].Library.ReorderVersion()
			spell := addCardToHand(g, game.Player1, authored.Intuition.New())
			targets := []game.Target{game.PlayerTarget(game.Player3)}
			if scenario == "self" {
				targets[0] = game.PlayerTarget(game.Player1)
			}
			if scenario == "no target" {
				targets = nil
			}
			if scenario == "too many targets" {
				targets = append(targets, game.PlayerTarget(game.Player2))
			}
			if scenario == "hexproof" || scenario == "no legal opponents" {
				applyPlayerRuleEffect(engine, g, game.Player3, game.RuleEffectPlayerHexproof)
			}
			if scenario == "no legal opponents" {
				applyPlayerRuleEffect(engine, g, game.Player2, game.RuleEffectPlayerHexproof)
				applyPlayerRuleEffect(engine, g, game.Player4, game.RuleEffectPlayerHexproof)
			}
			if scenario == "shroud" {
				applyPlayerRuleEffect(engine, g, game.Player3, game.RuleEffectPlayerShroud)
			}
			if scenario == "departed" {
				engine.eliminatePlayer(g, game.Player3)
			}
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			setMainPhasePriority(g, game.Player1)
			cast := engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil))
			wantCast := scenario == "opponent" || scenario == "becomes illegal" || scenario == "leaves before resolution"
			if cast != wantCast {
				t.Fatalf("cast = %v, want %v", cast, wantCast)
			}
			if !cast {
				if !g.Players[game.Player1].Hand.Contains(spell) || !g.Stack.IsEmpty() ||
					!slices.Equal(before, g.Players[game.Player1].Library.All()) {
					t.Fatal("illegal announcement consumed the card or changed the library")
				}
				return
			}
			if scenario == "becomes illegal" {
				applyPlayerRuleEffect(engine, g, game.Player3, game.RuleEffectPlayerHexproof)
			}
			if scenario == "leaves before resolution" {
				engine.eliminatePlayer(g, game.Player3)
			}
			log := &TurnLog{}
			opponent := &controlScript{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{&controlScript{}, nil, opponent}, log)
			if !g.Stack.IsEmpty() || !g.Players[game.Player1].Graveyard.Contains(spell) {
				t.Fatal("cast instant did not leave the stack for the graveyard")
			}
			if scenario != "opponent" {
				if len(log.Choices) != 0 || len(searchPartitionReveals(g)) != 0 ||
					!slices.Equal(before, g.Players[game.Player1].Library.All()) ||
					g.Players[game.Player1].Library.ReorderVersion() != epoch {
					t.Fatalf("all-illegal target allowed searching/shuffling: choices=%d reveals=%v before=%v after=%v epoch=%d", len(log.Choices), searchPartitionReveals(g), before, g.Players[game.Player1].Library.All(), g.Players[game.Player1].Library.ReorderVersion())
				}
			} else if len(opponent.choices) != 1 || len(searchPartitionReveals(g)) != 3 ||
				g.Players[game.Player1].Hand.Size() != 1 ||
				g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("real cast did not pay and resolve through the targeted opponent")
			}
		})
	}
}

func TestAuthoredIntuitionPrivateSearchAndDuplicateNames(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := searchPartitionEngine(t, authored.Intuition)
	secret := addCardToLibrary(g, game.Player1, vanillaCreature("Unrevealed secret", 1, 1))
	other := addCardToLibrary(g, game.Player1, vanillaCreature("Other", 1, 1))
	first := addCardToLibrary(g, game.Player1, vanillaCreature("Same", 1, 1))
	second := addCardToLibrary(g, game.Player1, vanillaCreature("Same", 1, 1))
	addCardToHand(g, game.Player1, vanillaCreature("Private hand", 1, 1))
	searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind == game.ChoiceOrder {
			return req.DefaultSelection
		}
		if req.Kind != game.ChoiceSearch || len(req.Options) != 4 ||
			len(obs.LibraryLookable(game.Player1)) != 4 ||
			len(NewObservation(g, game.Player2).LibraryLookable(game.Player1)) != 0 ||
			len(NewObservation(g, game.Player3).LibraryLookable(game.Player1)) != 0 {
			t.Fatal("private library search was not permission-scoped")
		}
		return []int{2, 0, 1}
	}}
	opponent := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if len(obs.LibraryLookable(game.Player1)) != 0 || len(obs.HandOf(game.Player1)) != 0 ||
			len(g.SearchControlScopes) != 0 || len(req.Options) != 3 ||
			!slices.Equal(searchPartitionReveals(g), []game.ObjectID{other, second, first}) {
			t.Fatal("opponent chose before revelation or inherited private search visibility")
		}
		for _, option := range req.Options {
			if !option.Card.Exists || option.Card.Val.CardID == secret {
				t.Fatal("partition exposed an unrevealed card or hid revealed characteristics")
			}
		}
		return []int{1}
	}}
	addImplementationSpellToStack(g, game.Player1, authored.Intuition.New(), searchPartitionTargets(authored.Intuition))
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, nil, opponent}, log)
	if !g.Players[game.Player1].Hand.Contains(second) || !g.Players[game.Player1].Library.Contains(secret) ||
		!g.Players[game.Player1].Graveyard.Contains(first) || !g.Players[game.Player1].Graveyard.Contains(other) ||
		len(log.Choices) != 3 || log.Choices[0].UsedFallback || log.Choices[1].UsedFallback ||
		len(NewObservation(g, game.Player1).LibraryLookable(game.Player1)) != 0 {
		t.Fatal("duplicate-name unrestricted search, reordered choices or private-scope cleanup failed")
	}
}

func TestAuthoredIntuitionRetargetedCopySearchesCopiersLibrary(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	programs := append(slices.Clone(authored.Intuition.Programs), authored.Twincast.Programs...)
	catalog, err := cardprogram.NewCatalog(programs)
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	for index := range 4 {
		addCardToLibrary(g, game.Player1, vanillaCreature(fmt.Sprintf("Original %d", index), 1, 1))
		addCardToLibrary(g, game.Player3, vanillaCreature(fmt.Sprintf("Copier %d", index), 1, 1))
	}
	originalCards := slices.Clone(g.Players[game.Player1].Library.All())
	copyCards := slices.Clone(g.Players[game.Player3].Library.All())
	source := addCardToHand(g, game.Player1, authored.Intuition.New())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(source, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("original Intuition cast failed")
	}
	original, _ := g.Stack.Peek()
	copySpell := addCardToHand(g, game.Player3, authored.Twincast.New())
	g.Players[game.Player3].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player3
	if !engine.applyAction(g, game.Player3, action.CastSpell(copySpell, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil)) {
		t.Fatal("Twincast cast failed")
	}
	copier := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		for _, option := range req.Options {
			if len(option.Targets) == 1 && option.Targets[0].PlayerID == game.Player4 {
				return []int{option.Index}
			}
		}
		t.Fatal("copy did not offer a legal replacement opponent")
		return nil
	}}
	copyLog := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{nil, nil, copier}, copyLog)
	copied, _ := g.Stack.Peek()
	if !copied.Copy || copied.Controller != game.Player3 || copied.Targets[0].PlayerID != game.Player4 ||
		original.Targets[0].PlayerID != game.Player2 || len(copyLog.Choices) != 1 || copyLog.Choices[0].UsedFallback {
		t.Fatal("copy lost controller/retargeting or changed the original's target")
	}
	opponent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Player != game.Player4 || len(req.Options) != 3 {
			t.Fatal("retargeted copy asked the original opponent")
		}
		return []int{2}
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{nil, nil, &controlScript{}, opponent}, &TurnLog{})
	if !g.Players[game.Player3].Hand.Contains(copyCards[2]) ||
		!g.Players[game.Player3].Graveyard.Contains(copyCards[0]) ||
		!g.Players[game.Player3].Graveyard.Contains(copyCards[1]) ||
		!slices.Equal(originalCards, g.Players[game.Player1].Library.All()) || g.Stack.Size() != 1 {
		t.Fatal("retargeted copy searched the original library or misplaced its own finds")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{&controlScript{}, &controlScript{}}, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player1].Library.Size() != 1 ||
		g.Players[game.Player3].Library.Size() != 1 || !g.Stack.IsEmpty() {
		t.Fatal("original and copy did not finish independent searches")
	}
}

func TestAuthoredRealmsUnchartedDifferentNamesAndFewerFinds(t *testing.T) {
	t.Parallel()
	for count := range 5 {
		t.Run(fmt.Sprintf("%d lands", count), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := searchPartitionEngine(t, authored.RealmsUncharted)
			nonland := addCardToLibrary(g, game.Player1, vanillaCreature("Not a land", 1, 1))
			var lands []game.ObjectID
			for _, name := range []string{"Forest", "Island", "Mountain", "Plains"} {
				lands = append(lands, addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}))
			}
			duplicate := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Plains", Types: []types.Card{types.Land}}})
			epoch := g.Players[game.Player1].Library.ReorderVersion()
			picks := 0
			chosenNames := make(map[string]bool)
			searcher := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceOrder {
					return req.DefaultSelection
				}
				if req.Kind == game.ChoicePlayer {
					if len(g.SearchControlScopes) != 0 || len(obs.LibraryLookable(game.Player1)) != 0 ||
						len(searchPartitionReveals(g)) != count || len(req.Options) != 3 {
						t.Fatal("opponent was chosen before revelation or private search scope persisted")
					}
					for _, option := range req.Options {
						if len(option.Targets) != 0 {
							t.Fatal("Realms treated choosing an opponent as targeting")
						}
						if option.Label == "Player 3" {
							return []int{option.Index}
						}
					}
					t.Fatal("opponent choices omitted Player3")
					return nil
				}
				if req.Kind != game.ChoiceSearch || req.MinChoices != 0 || req.MaxChoices != 1 ||
					len(obs.LibraryLookable(game.Player1)) != 6 ||
					len(NewObservation(g, game.Player3).LibraryLookable(game.Player1)) != 0 {
					t.Fatal("qualified distinct-name search did not permit fail-to-find or hid nonmatches from the searcher")
				}
				for _, option := range req.Options {
					if !option.Card.Exists || !slices.Contains(option.Card.Val.Types, types.Land) ||
						option.Card.Val.CardID == nonland || chosenNames[option.Label] {
						t.Fatal("search offered nonland or duplicate selected name")
					}
				}
				if picks == count {
					return nil
				}
				chosenNames[req.Options[0].Label] = true
				picks++
				return []int{0}
			}}
			partition := min(2, count)
			var graveyard []game.ObjectID
			opponent := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				if req.Player != game.Player3 || req.MinChoices != partition || req.MaxChoices != partition ||
					len(req.Options) != count || len(obs.LibraryLookable(game.Player1)) != 0 {
					t.Fatal("non-targeted opponent did not choose as many as possible from the revealed lands")
				}
				var selected []int
				for _, option := range req.Options[len(req.Options)-partition:] {
					selected = append(selected, option.Index)
					graveyard = append(graveyard, option.Card.Val.CardID)
				}
				return selected
			}}
			applyPlayerRuleEffect(engine, g, game.Player3, game.RuleEffectPlayerHexproof)
			applyPlayerRuleEffect(engine, g, game.Player3, game.RuleEffectPlayerShroud)
			spell := addCardToHand(g, game.Player1, authored.RealmsUncharted.New())
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			setMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
				t.Fatal("Realms could not be cast without targets")
			}
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, nil, opponent}, log)
			found := searchPartitionReveals(g)
			if len(found) != count || !g.Players[game.Player1].Library.Contains(nonland) ||
				!g.Players[game.Player1].Library.Contains(lands[3]) ||
				g.Players[game.Player1].Library.Size() != 6-count ||
				g.Players[game.Player1].Hand.Size() != count-partition ||
				!g.Players[game.Player1].Graveyard.Contains(spell) ||
				g.Players[game.Player1].Library.ReorderVersion() != epoch+1 ||
				(count > 0 && !slices.Contains(found, duplicate)) {
				t.Fatalf("qualified search/partition failed: found=%v duplicate=%v nonlandInLibrary=%v originalPlainsInLibrary=%v library=%d hand=%d spellInGrave=%v epoch=%d", found, duplicate, g.Players[game.Player1].Library.Contains(nonland), g.Players[game.Player1].Library.Contains(lands[3]), g.Players[game.Player1].Library.Size(), g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Graveyard.Contains(spell), g.Players[game.Player1].Library.ReorderVersion())
			}
			for _, card := range found {
				if g.Players[game.Player1].Graveyard.Contains(card) != slices.Contains(graveyard, card) ||
					g.Players[game.Player1].Hand.Contains(card) == slices.Contains(graveyard, card) {
					t.Fatal("opponent's two choices were not placed correctly")
				}
			}
			assertSearchPartitionBatch(t, g, found)
			for _, decision := range log.Choices {
				if decision.UsedFallback {
					t.Fatal("a legal search stop or opponent choice was replaced by fallback")
				}
			}
			if len(opponent.choices) != min(1, count) {
				t.Fatal("empty search asked opponent an impossible choice")
			}
		})
	}
}

func TestAuthoredRealmsUnchartedLibraryExhaustionAndLivingOpponent(t *testing.T) {
	t.Parallel()
	for count := range 4 {
		t.Run(fmt.Sprintf("%d available", count), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := searchPartitionEngine(t, authored.RealmsUncharted)
			for index := range count {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: fmt.Sprintf("Land %d", index), Types: []types.Card{types.Land},
				}})
			}
			addImplementationSpellToStack(g, game.Player1, authored.RealmsUncharted.New(), nil)
			// There was no announced target; the living opponent is chosen later.
			engine.eliminatePlayer(g, game.Player2)
			engine.eliminatePlayer(g, game.Player3)
			opponent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Player != game.Player4 || len(req.Options) != count || req.MinChoices != min(2, count) {
					t.Fatal("Realms did not use the still-living opponent or the actual available lands")
				}
				return fallbackChoice(req)
			}}
			searcher := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoicePlayer && (len(req.Options) != 1 || req.Options[0].Label != "Player 4") {
					t.Fatal("departed players remained in the non-targeted opponent choice")
				}
				return fallbackChoice(req)
			}}
			epoch := g.Players[game.Player1].Library.ReorderVersion()
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, nil, nil, opponent}, &TurnLog{})
			if g.Players[game.Player1].Library.Size() != 0 || len(searchPartitionReveals(g)) != count ||
				g.Players[game.Player1].Hand.Size() != max(0, count-2) ||
				g.Players[game.Player1].Library.ReorderVersion() != epoch+1 ||
				len(opponent.choices) != min(1, count) {
				t.Fatal("library exhaustion failed to place all actual finds and shuffle, including zero")
			}
		})
	}
}

func TestAuthoredSearchPartitionZoneReplacementsAndDeferredShuffle(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.Intuition, authored.RealmsUncharted} {
		for _, destination := range []zone.Type{zone.Exile, zone.Library} {
			t.Run(entry.Name+"/"+destination.String(), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := searchPartitionEngine(t, entry)
				var library []game.ObjectID
				for index := range 5 {
					library = append(library, addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
						Name: fmt.Sprintf("Land %d", index), Types: []types.Card{types.Land},
					}}))
				}
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Library, MatchToZone: true, ToZone: zone.Graveyard,
					ReplaceToZone: destination, ShuffleIntoLibrary: destination == zone.Library,
				})
				epoch := g.Players[game.Player1].Library.ReorderVersion()
				searcher := &controlScript{}
				opponent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
					if entry.OracleID == authored.Intuition.OracleID {
						return []int{1}
					}
					return []int{0, 2}
				}}
				addImplementationSpellToStack(g, game.Player1, entry.New(), searchPartitionTargets(entry))
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, opponent, opponent}, &TurnLog{})
				found := searchPartitionReveals(g)
				wantFound, wantHand := 4, 2
				if entry.OracleID == authored.Intuition.OracleID {
					wantFound, wantHand = 3, 1
				}
				if len(found) != wantFound || g.Players[game.Player1].Hand.Size() != wantHand {
					t.Fatal("replacement shuffle invalidated other simultaneous hand placements")
				}
				for _, card := range found {
					if g.Players[game.Player1].Hand.Contains(card) {
						continue
					}
					if destination == zone.Library && !g.Players[game.Player1].Library.Contains(card) ||
						destination == zone.Exile && !g.Players[game.Player1].Exile.Contains(card) ||
						g.Players[game.Player1].Graveyard.Contains(card) {
						t.Fatal("placement ignored a destination replacement")
					}
				}
				wantEpoch := epoch + 1
				if destination == zone.Library {
					wantEpoch++
				}
				if g.Players[game.Player1].Library.ReorderVersion() != wantEpoch {
					t.Fatalf("replacement shuffles were not deferred: epoch=%d want=%d", g.Players[game.Player1].Library.ReorderVersion(), wantEpoch)
				}
				assertSearchPartitionBatch(t, g, found)
				for _, card := range library {
					locations := 0
					for _, container := range []zone.Zone{g.Players[game.Player1].Library, g.Players[game.Player1].Hand, g.Players[game.Player1].Exile} {
						if container.Contains(card) {
							locations++
						}
					}
					if locations != 1 {
						t.Fatal("a replacement lost or duplicated a physical card")
					}
				}
			})
		}
	}
}

func TestAuthoredSearchPartitionOppositionAgent(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.Intuition, authored.RealmsUncharted} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := searchPartitionEngine(t, entry)
			addCombatPermanent(g, game.Player4, oppositionAgentDef())
			for index := range 5 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: fmt.Sprintf("Land %d", index), Types: []types.Card{types.Land},
				}})
			}
			ordinary := &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int {
				t.Fatal("Opposition Agent did not own the search, or a vanished find was offered for partition")
				return nil
			}}
			epoch := g.Players[game.Player1].Library.ReorderVersion()
			controller := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				if obs.Player != game.Player1 || obs.DecisionMaker != game.Player4 ||
					len(obs.LibraryLookable(game.Player1)) != 5 ||
					len(NewObservation(g, game.Player3).LibraryLookable(game.Player1)) != 0 {
					t.Fatal("controlled search did not scope private disclosure to its authorized decider")
				}
				return slices.Clone(req.DefaultSelection)
			}}
			addImplementationSpellToStack(g, game.Player1, entry.New(), searchPartitionTargets(entry))
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{ordinary, ordinary, ordinary, controller}, log)
			want := 4
			if entry.OracleID == authored.Intuition.OracleID {
				want = 3
			}
			if g.Players[game.Player1].Exile.Size() != want || g.Players[game.Player1].Hand.Size() != 0 ||
				g.Players[game.Player1].Library.Size() != 5-want ||
				g.Players[game.Player1].Library.ReorderVersion() != epoch+1 ||
				len(g.SearchControlScopes) != 0 || len(NewObservation(g, game.Player4).LibraryLookable(game.Player1)) != 0 {
				t.Fatalf("opposition search failed: exile=%d hand=%d library=%d epoch=%d scopes=%d visible=%d", g.Players[game.Player1].Exile.Size(), g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Library.Size(), g.Players[game.Player1].Library.ReorderVersion(), len(g.SearchControlScopes), len(NewObservation(g, game.Player4).LibraryLookable(game.Player1)))
			}
			for _, card := range g.Players[game.Player1].Exile.All() {
				permission, exists := playFromExilePermissionFor(g, card)
				if !exists || permission.Controller != game.Player4 {
					t.Fatal("search-only replacement lost its independent play beneficiary")
				}
			}
			for _, decision := range log.Choices {
				if decision.Request.Kind != game.ChoiceSearch || decision.DecisionMaker != game.Player4 || decision.UsedFallback {
					t.Fatal("controlled search escaped its scope or lost a recorded valid answer")
				}
			}
		})
	}
}

func TestAuthoredSearchPartitionCloneIsolation(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.Intuition, authored.RealmsUncharted} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := searchPartitionEngine(t, entry)
			for index := range 5 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: fmt.Sprintf("Land %d", index), Types: []types.Card{types.Land},
				}})
			}
			addImplementationSpellToStack(g, game.Player1, entry.New(), searchPartitionTargets(entry))
			before := slices.Clone(g.Players[game.Player1].Library.All())
			epoch := g.Players[game.Player1].Library.ReorderVersion()
			clone := g.Clone()
			searcher := &controlScript{}
			first := &controlScript{}
			last := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				var selected []int
				for _, option := range req.Options[len(req.Options)-req.MinChoices:] {
					selected = append(selected, option.Index)
				}
				return selected
			}}
			engine.resolveTopOfStackWithChoices(clone, [game.NumPlayers]PlayerAgent{searcher, first, first}, &TurnLog{})
			if !slices.Equal(before, g.Players[game.Player1].Library.All()) || g.Stack.Size() != 1 ||
				g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player1].Library.ReorderVersion() != epoch {
				t.Fatalf("clone changed original: before=%v after=%v stack=%d hand=%d epoch=%d", before, g.Players[game.Player1].Library.All(), g.Stack.Size(), g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Library.ReorderVersion())
			}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{searcher, last, last}, &TurnLog{})
			if slices.Equal(g.Players[game.Player1].Hand.All(), clone.Players[game.Player1].Hand.All()) ||
				!g.Stack.IsEmpty() || !clone.Stack.IsEmpty() {
				t.Fatal("independent opponent choices did not produce independent clone outcomes")
			}
		})
	}
}

func TestAuthoredSearchPartitionFoundHandlesRespectIncarnations(t *testing.T) {
	t.Parallel()
	for _, stale := range []string{"moved", "shuffled"} {
		t.Run(stale, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for index := range 3 {
				addCardToLibrary(g, game.Player1, vanillaCreature(fmt.Sprintf("Card %d", index), 1, 1))
			}
			ctx := &programContext{
				engine: searchPartitionEngine(t, authored.Intuition), game: g,
				obj: &game.StackObject{Controller: game.Player1}, active: true, log: &TurnLog{},
			}
			search := game.Search{Player: game.ControllerReference(), Amount: game.Fixed(3),
				Spec: game.SearchSpec{SourceZone: zone.Library, Reveal: true, RevealOnly: true, FailToFindPolicy: game.SearchMustFindIfAvailable}}
			found, err := ctx.Search(search)
			if err != nil || len(found.Cards) != 3 {
				t.Fatalf("exact reveal-only search: %v", err)
			}
			first := found.Cards[0]
			if stale == "moved" {
				if !moveCardBetweenZones(g, game.Player1, first.Identity.CardID, zone.Library, zone.Graveyard) ||
					!moveCardBetweenZones(g, game.Player1, first.Identity.CardID, zone.Graveyard, zone.Library) {
					t.Fatal("incarnation fixture failed")
				}
			} else if _, err := ctx.Execute(game.ShuffleLibrary{Player: game.ControllerReference()}); err != nil {
				t.Fatal(err)
			}
			moved, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Hand})
			want := 2
			if stale == "shuffled" {
				want = 0
			}
			if err != nil || len(moved) != want || !g.Players[game.Player1].Library.Contains(first.Identity.CardID) {
				t.Fatal("found handles followed a moved card back into the library or survived a shuffle")
			}
			again, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Graveyard})
			if err != nil || len(again) != 0 {
				t.Fatal("used handles followed already-moved card incarnations")
			}
			ctx.active = false
			if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Hand}); err == nil {
				t.Fatal("expired search invocation retained mutable authority")
			}
		})
	}
}
