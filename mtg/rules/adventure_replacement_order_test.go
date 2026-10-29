package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Adventure resolution exile competes with other graveyard replacements
// (#3359).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     614.15 and 616.1a (only self-replacement effects are applied first),
//     616.1/616.1e-f (the affected object's controller chooses one applicable
//     replacement, then repeats with the modified event), 702.34a ("If the
//     flashback cost was paid, exile this card instead of putting it anywhere
//     else any time it would leave the stack") and 715.3d ("Instead of putting
//     a spell that was cast as an Adventure into its owner's graveyard as it
//     resolves, its controller exiles it. For as long as that card remains
//     exiled, that player may play it.").
//   - Young Red Dragon // Bathe in Gold, oracle_id
//     2f865ae9-9328-4f6f-924f-d5e0a65aaa96, Scryfall rulings 2022-06-10 for
//     d0b9865a-be87-48fd-a325-be6aca8a31e9: the controller exiles a resolving
//     Adventure instead of putting it into the graveyard; a card exiled "for
//     any other reason than by exiling itself while resolving" gives no
//     permission.
//   - Dauthi Voidwalker, oracle_id f1c2dbe2-fbe0-4058-bdf1-91d1b1832786
//     (rulings 2021-06-18 for 3573b9a2-7911-475c-8ae7-25bd0dbb7159): "If a card
//     would be put into an opponent's graveyard from anywhere, instead exile it
//     with a void counter on it."
//   - Torrential Gearhulk, oracle_id a6395447-677d-4c39-8eda-2d57e527c94e
//     (rulings 2016-09-20 for 92872fea-2fb4-4154-815e-cad9d966e60f): "If that
//     spell would be put into your graveyard, exile it instead." The rider is
//     reproduced with the shared CastForFree ExileOnResolution option because
//     the registered riders only target instant cards.
//
// All three replacements apply to the same stack-to-graveyard event, none is a
// self-replacement, so the resolving spell's controller chooses which applies.
// Only the Adventure replacement grants the CR 715.3d permission.

func adventureReplacementRegistry(t *testing.T) (*cards.Registry, *game.CardDef, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	dragon := registry.Lookup("Young Red Dragon")
	dauthi := registry.Lookup("Dauthi Voidwalker")
	if dragon == nil || dauthi == nil || dragon.Layout != game.LayoutAdventure ||
		!dragon.Alternate.Exists || dragon.Alternate.Val.Name != "Bathe in Gold" {
		t.Fatal("registered Young Red Dragon // Bathe in Gold or Dauthi Voidwalker missing")
	}
	return registry, dragon, dauthi
}

// resolveDauthiVoidwalker puts registered Dauthi Voidwalker onto the
// battlefield under controller's control with its replacement effect.
func resolveDauthiVoidwalker(t *testing.T, g *game.Game, engine *Engine, controller game.PlayerID, dauthi *game.CardDef) {
	t.Helper()
	addImplementationSpellToStack(g, controller, dauthi, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("Dauthi Voidwalker replacement effects = %d, want 1", len(g.ReplacementEffects))
	}
}

// castBatheInGold has Player1 cast registered Bathe in Gold from hand.
func castBatheInGold(t *testing.T, g *game.Game, engine *Engine, dragon *game.CardDef) (id.ID, *game.StackObject) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, dragon)
	for range 2 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("Player1 could not cast Bathe in Gold")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FaceAlternate {
		t.Fatalf("stack object = %+v, want Bathe in Gold", obj)
	}
	return cardID, obj
}

// replacementOrderAgents makes chooser pick the Adventure replacement or the
// other offered replacement, recording each replacement request it receives.
func replacementOrderAgents(t *testing.T, chooser game.PlayerID, chooseAdventure bool, requests *[]game.ChoiceRequest) [game.NumPlayers]PlayerAgent {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		agents[player] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceReplacement {
				return request.DefaultSelection
			}
			if player != chooser {
				t.Fatalf("%v was asked to order replacements, want %v", player, chooser)
			}
			*requests = append(*requests, request)
			for _, option := range request.Options {
				if (option.Label == adventureReplacementLabel) == chooseAdventure {
					return []int{option.Index}
				}
			}
			t.Fatalf("replacement options %+v lack the requested choice", request.Options)
			return nil
		}}
	}
	return agents
}

func requireTwoWayReplacementChoice(t *testing.T, requests []game.ChoiceRequest, chooser game.PlayerID, other string) {
	t.Helper()
	if len(requests) != 1 || requests[0].Player != chooser || len(requests[0].Options) != 2 {
		t.Fatalf("replacement requests = %+v, want one two-way choice for %v", requests, chooser)
	}
	labels := []string{requests[0].Options[0].Label, requests[0].Options[1].Label}
	if labels[0] != adventureReplacementLabel || (other != "" && labels[1] != other) {
		t.Fatalf("replacement labels = %q, want Adventure then %q", labels, other)
	}
}

func controlsTreasure(g *game.Game, player game.PlayerID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.Controller == player && permanent.TokenDef != nil && permanent.TokenDef.Name == "Treasure" {
			return true
		}
	}
	return false
}

// assertAdventureOutcome checks the card's exile state and the CR 715.3d
// permission, then funds both relevant players and checks real castability.
func assertAdventureOutcome(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, beneficiary game.PlayerID, permission bool, voidCounters int) {
	t.Helper()
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Bathe in Gold did not end in its owner's exile")
	}
	if got := g.ExileCounterCount(cardID, counter.Void); got != voidCounters {
		t.Fatalf("void counters = %d, want %d", got, voidCounters)
	}
	adventure, ok := g.AdventureCards[cardID]
	if ok != permission || (permission && adventure.Player != beneficiary) {
		t.Fatalf("Adventure permission = %+v (%v), want %v for %v", adventure, ok, permission, beneficiary)
	}
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		for range 4 {
			addBasicLandPermanent(g, player, types.Mountain)
		}
		assertAdventureCastable(t, g, engine, player, cardID, permission && player == beneficiary)
	}
}

func TestAdventureExileCompetesWithDauthiVoidwalker(t *testing.T) {
	t.Parallel()
	for _, chooseAdventure := range []bool{true, false} {
		name := "Dauthi first"
		if chooseAdventure {
			name = "Adventure first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			registry, dragon, dauthi := adventureReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			cardID, _ := castBatheInGold(t, g, engine, dragon)
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, replacementOrderAgents(t, game.Player1, chooseAdventure, &requests), &TurnLog{})
			requireTwoWayReplacementChoice(t, requests, game.Player1, g.ReplacementEffects[0].Description)
			if !controlsTreasure(g, game.Player1) {
				t.Fatal("Bathe in Gold did not create its Treasure")
			}
			// Whichever applies first, the other no longer applies: the card is
			// no longer going to a graveyard (CR 616.1f).
			voidCounters := 1
			if chooseAdventure {
				voidCounters = 0
			}
			assertAdventureOutcome(t, g, engine, cardID, game.Player1, chooseAdventure, voidCounters)
		})
	}
}

func TestStolenAdventureControllerOrdersGraveyardReplacements(t *testing.T) {
	t.Parallel()
	for _, chooseAdventure := range []bool{true, false} {
		name := "Dauthi first"
		if chooseAdventure {
			name = "Adventure first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			registry, dragon, dauthi := adventureReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			resolveDauthiVoidwalker(t, g, engine, game.Player3, dauthi)
			cardID, obj := castBatheInGold(t, g, engine, dragon)
			stealStackSpell(t, g, engine, obj.ID, game.Player2)
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, replacementOrderAgents(t, game.Player2, chooseAdventure, &requests), &TurnLog{})
			requireTwoWayReplacementChoice(t, requests, game.Player2, g.ReplacementEffects[0].Description)
			if !controlsTreasure(g, game.Player2) || controlsTreasure(g, game.Player1) {
				t.Fatal("stolen Bathe in Gold did not create its controller's Treasure")
			}
			voidCounters := 1
			if chooseAdventure {
				voidCounters = 0
			}
			assertAdventureOutcome(t, g, engine, cardID, game.Player2, chooseAdventure, voidCounters)
		})
	}
}

func TestAdventureExileCompetesWithCastExileRider(t *testing.T) {
	t.Parallel()
	for _, chooseAdventure := range []bool{true, false} {
		name := "rider first"
		if chooseAdventure {
			name = "Adventure first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, dragon, _ := adventureReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToGraveyard(g, game.Player1, dragon)
			sourceID := addEffectSpellToStack(g, game.Player1, game.CastForFree{
				Player:            game.ControllerReference(),
				Card:              game.CardReference{Kind: game.CardReferenceTarget},
				Zone:              zone.Graveyard,
				ExileOnResolution: true,
			}, []game.Target{currentCardTarget(t, g, cardID)})
			source, _ := g.GetCardInstance(sourceID)
			source.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
			}}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				for _, option := range request.Options {
					if request.Kind == game.ChoiceResolution && option.Label == "Bathe in Gold" {
						return []int{option.Index}
					}
				}
				return request.DefaultSelection
			}}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			obj, ok := g.Stack.Peek()
			if !ok || obj.SourceID != cardID || obj.Face != game.FaceAlternate || !obj.ExileInsteadOfGraveyard {
				t.Fatalf("free cast = %+v, want Bathe in Gold with the exile rider", obj)
			}
			// The rider is stack-object state, so a clone replays the same choice.
			replay := g.Clone()
			var requests, replayRequests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, replacementOrderAgents(t, game.Player1, chooseAdventure, &requests), &TurnLog{})
			engine.resolveTopOfStackWithChoices(replay, replacementOrderAgents(t, game.Player1, chooseAdventure, &replayRequests), &TurnLog{})
			requireTwoWayReplacementChoice(t, requests, game.Player1, "Exile instead of graveyard")
			requireTwoWayReplacementChoice(t, replayRequests, game.Player1, "Exile instead of graveyard")
			if g.AdventureCards[cardID] != replay.AdventureCards[cardID] || len(g.Events) != len(replay.Events) ||
				!replay.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("cloned free cast resolved differently from the original")
			}
			if !controlsTreasure(g, game.Player1) {
				t.Fatal("Bathe in Gold did not create its Treasure")
			}
			assertAdventureOutcome(t, g, engine, cardID, game.Player1, chooseAdventure, 0)
		})
	}
}

func TestAdventureReplacementDefaultKeepsPermissionWithoutAgent(t *testing.T) {
	t.Parallel()
	registry, dragon, dauthi := adventureReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
	cardID, _ := castBatheInGold(t, g, engine, dragon)
	engine.resolveTopOfStack(g, &TurnLog{})
	decisions := g.ReplacementDecisions
	if len(decisions) == 0 || !decisions[len(decisions)-1].UsedFallback ||
		decisions[len(decisions)-1].Options[0] != adventureReplacementLabel {
		t.Fatalf("replacement decisions = %+v, want an Adventure-first fallback", decisions)
	}
	assertAdventureOutcome(t, g, engine, cardID, game.Player1, true, 0)
}

func TestFlashbackExileCompetesWithDauthiVoidwalker(t *testing.T) {
	t.Parallel()
	for _, chooseFlashback := range []bool{true, false} {
		name := "Dauthi first"
		if chooseFlashback {
			name = "Flashback first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			registry, _, dauthi := adventureReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Flashback Draw", Types: []types.Card{types.Sorcery},
				ManaCost: opt.Val(cost.Mana{cost.O(5)}),
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}}}.Ability()),
				StaticAbilities: []game.StaticAbility{{
					KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.R}}},
				}},
			}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			addBasicLandPermanent(g, game.Player1, types.Mountain)
			giveMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
				t.Fatal("flashback cast failed")
			}
			if obj, ok := g.Stack.Peek(); !ok || !obj.Flashback {
				t.Fatalf("stack object = %+v, want a flashback spell", obj)
			}
			var requests []game.ChoiceRequest
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoiceReplacement {
					return request.DefaultSelection
				}
				requests = append(requests, request)
				for _, option := range request.Options {
					if (option.Label == "Flashback (CR 702.34a)") == chooseFlashback {
						return []int{option.Index}
					}
				}
				t.Fatalf("replacement options %+v lack the requested choice", request.Options)
				return nil
			}}}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if len(requests) != 1 || requests[0].Player != game.Player1 || len(requests[0].Options) != 2 {
				t.Fatalf("replacement requests = %+v, want one two-way choice for Player1", requests)
			}
			voidCounters := 1
			if chooseFlashback {
				voidCounters = 0
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) || g.ExileCounterCount(cardID, counter.Void) != voidCounters {
				t.Fatalf("flashback card exiled = %v with %d void counters, want %d",
					g.Players[game.Player1].Exile.Contains(cardID), g.ExileCounterCount(cardID, counter.Void), voidCounters)
			}
		})
	}
}
