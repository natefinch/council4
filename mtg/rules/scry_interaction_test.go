package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source record: the pinned Oracle ARRAY SHA256 is
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Exact pinned Oracle/revision and live Scryfall printing records retrieved
// 2026-10-22:
//   Eager Construct 39363190-7f79-4fae-bcfc-b17ec4e07bfe /
//     5b23ca61-157b-4616-bbf4-0c1298ddbbb4 (2XM 250).
//   Sea God's Revenge 8ab71a49-ce81-44f7-b940-50e7aba64df9 /
//     5acdf4e4-5933-43dc-bf8b-25f89415db5b (THS 61).
//   Devout Decree 92f8c8b1-1fd5-4e80-b8a9-51671937aaf2 /
//     d174cb01-b1cf-444c-a893-cc61ddf88b8c (FDN 571).
// The current exact-printing rulings endpoints all returned complete lists
// with has_more=false. Their complete WotC ruling texts were:
//   Eager Construct (2020-08-07): "If multiple players are instructed to scry
//   at once, such as while resolving Eager Construct's triggered ability,
//   those players each look at the top card of their library at the same time,
//   then they choose in turn order whether to put those cards on the top or
//   bottom of their library."
//   Sea God's Revenge (2013-09-15):
//   1. "You can cast Sea God's Revenge with no targets. If you do, you'll scry
//      1 when it resolves. If you choose at least one target, and all chosen
//      targets are illegal when Sea God's Revenge tries to resolve, it won't
//      resolve and none of its effects will happen."
//   2. "In multiplayer games, the three target creatures can be controlled by
//      different opponents."
//   3. "When you scry, you may put all the cards you look at back on top of
//      your library, you may put all of those cards on the bottom of your
//      library, or you may put some of those cards on top and the rest of them
//      on the bottom."
//   4. "You choose how to order cards returned to your library after scrying
//      no matter where you put them."
//   5. "You perform the actions stated on a card in sequence. For some spells
//      and abilities, that means you'll scry last. For others, that means
//      you'll scry and then perform other actions."
//   6. "Scry appears on some spells and abilities with one or more targets. If
//      all of the spell or ability's targets are illegal when it tries to
//      resolve, it won't resolve and none of its effects will happen. You
//      won't scry."
//   Devout Decree (2019-07-12): "If the target creature or planeswalker is an
//   illegal target by the time Devout Decree tries to resolve, the spell
//   doesn't resolve. You won't scry 1."
//
// Personally read the official Wizards TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// effective 2026-09-25, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Sections 101.4a-e, 105.2, 108.3, 113.7a, 115.1a, 205.1/1a-b, 307.1,
// 400.7, 401.2-4, 601.2c/f-h, 603.2/3a, 608.2b/c/e/f, 613.1b/d/e, 614.1,
// 616.1/1a-g, 701.22a-d, 707.10 and 903.9a-b were read. They govern APNAP
// decision staging and simultaneous actions, current target characteristics
// and revalidation, exact target counts and costs/timing, private library
// looks versus public counts, trigger/copy lifetime, ordered resolution,
// scry placement, and zone replacements. Specifically, CR 903.9b is a
// replacement choice when a commander would return to its owner's hand;
// CR 903.9a is a state-based action after a commander has been exiled.

const (
	eagerConstructTestOracleID = "39363190-7f79-4fae-bcfc-b17ec4e07bfe"
	seaGodsRevengeTestOracleID = "8ab71a49-ce81-44f7-b940-50e7aba64df9"
	devoutDecreeTestOracleID   = "92f8c8b1-1fd5-4e80-b8a9-51671937aaf2"
)

func scryInteractionCard(t *testing.T, oracleID string) (*Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup default-registry Oracle ID %s: %v", oracleID, err)
	}
	return NewEngineWithPrograms(nil, registry.Programs()), def
}

func fillScryInteractionLibraries(g *game.Game, size int) [game.NumPlayers][]id.ID {
	var libraries [game.NumPlayers][]id.ID
	for player := range game.NumPlayers {
		for range size {
			addCardToLibrary(g, game.PlayerID(player), vanillaCreature("Scry identity", 1, 1))
		}
		libraries[player] = g.Players[player].Library.All()
	}
	return libraries
}

func castScryInteractionSpell(t *testing.T, g *game.Game, engine *Engine, owner game.PlayerID, def *game.CardDef, targets []game.Target) id.ID {
	t.Helper()
	cardID := addCardToHand(g, owner, def)
	setMainPhasePriority(g, owner)
	if !engine.applyAction(g, owner, action.CastSpell(cardID, targets, 0, nil)) {
		t.Fatalf("%s was not cast from hand with its announced targets", def.Name)
	}
	return cardID
}

func scryInteractionPermanent(g *game.Game, name string, owner, controller game.PlayerID, cardTypes []types.Card, colors []color.Color) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: slices.Clone(cardTypes), Colors: slices.Clone(colors),
	}}
	permanent := addCombatPermanent(g, controller, def)
	permanent.Owner = owner
	card, _ := g.GetCardInstance(permanent.CardInstanceID)
	card.Owner = owner
	return permanent
}

func resolveWithScryChoice(t *testing.T, engine *Engine, g *game.Game, player game.PlayerID, top id.ID, answer int) *TurnLog {
	return resolveWithScryChoiceAfter(t, engine, g, player, top, answer, nil)
}

func resolveWithScryChoiceAfter(t *testing.T, engine *Engine, g *game.Game, player game.PlayerID, top id.ID, answer int, beforeChoice func()) *TurnLog {
	t.Helper()
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind != game.ChoiceScry || request.Player != player ||
			!request.Subject.Exists || request.Subject.Val.CardID != top {
			t.Fatalf("scry request = %+v, want the private top card for player %v", request, player)
		}
		if beforeChoice != nil {
			beforeChoice()
		}
		return []int{answer}
	}}
	log := &TurnLog{}
	var agents [game.NumPlayers]PlayerAgent
	agents[player] = agent
	engine.resolveTopOfStackWithChoices(g, agents, log)
	assertPrivateScryLog(t, log)
	return log
}

func TestEagerConstructDefaultRegistryPaidCastAndAPNAP(t *testing.T) {
	engine, def := scryInteractionCard(t, eagerConstructTestOracleID)
	if def.Name != "Eager Construct" || authored.EagerConstruct.OracleID != eagerConstructTestOracleID {
		t.Fatalf("default registry identity = %q / %s", def.Name, authored.EagerConstruct.OracleID)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	libraries := fillScryInteractionLibraries(g, 2)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	if engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) ||
		!g.Players[game.Player1].Hand.Contains(cardID) || !g.Stack.IsEmpty() {
		t.Fatal("Eager Construct cast outside sorcery timing")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("paid {2} Eager Construct cast was rejected or failed to spend its cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, cardID)
	if permanent == nil || !slices.Equal(def.Types, []types.Card{types.Artifact, types.Creature}) ||
		permanent.Controller != game.Player1 {
		t.Fatal("Eager Construct did not resolve as its exact artifact creature")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the actual Eager Construct entry did not trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.Controller != game.Player1 ||
		trigger.SourceID != permanent.ObjectID {
		t.Fatal("Eager Construct trigger lost its source or controller")
	}

	g.Turn.ActivePlayer = game.Player3
	apnap := []game.PlayerID{game.Player3, game.Player4, game.Player1, game.Player2}
	var offered, placed []game.PlayerID
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		for player := range game.NumPlayers {
			if !slices.Equal(g.Players[player].Library.All(), libraries[player]) {
				t.Fatal("a library placement happened before all Eager Construct choices")
			}
		}
		if countEventsOfKind(g.Events, game.EventScry) != 0 {
			t.Fatal("Eager Construct published a scry before all placements")
		}
		switch request.Kind {
		case game.ChoiceMay:
			if len(placed) != 0 || len(offered) >= len(apnap) ||
				request.Player != apnap[len(offered)] || request.Subject.Exists {
				t.Fatalf("optional participation request was not private-free APNAP: %+v", request)
			}
			offered = append(offered, request.Player)
			return []int{1}
		case game.ChoiceScry:
			if len(offered) != len(apnap) || len(placed) >= len(apnap) ||
				request.Player != apnap[len(placed)] || !request.Subject.Exists ||
				request.Subject.Val.CardID != libraries[request.Player][0] {
				t.Fatalf("Eager Construct did not stage private scries in APNAP order: %+v", request)
			}
			if len(request.ScryDecisions) != len(apnap) {
				t.Fatal("public scry decisions omitted participants")
			}
			for index, decision := range request.ScryDecisions {
				if decision.Player != apnap[index] || !decision.Accepted ||
					decision.PlacementChosen != (index < len(placed)) ||
					decision.TopCount != 0 ||
					decision.BottomCount != boolInt(index < len(placed)) {
					t.Fatal("later scry chooser did not receive only prior public participation/count data")
				}
			}
			for _, option := range request.Options {
				if option.Card.Exists || option.Label == "Scry identity" {
					t.Fatal("another player's hidden identity appeared in a scry request")
				}
			}
			request.ScryDecisions[0].Accepted = false
			placed = append(placed, request.Player)
			return []int{1}
		default:
			t.Fatalf("unexpected Eager Construct decision kind %v", request.Kind)
			return nil
		}
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	if !slices.Equal(offered, apnap) || !slices.Equal(placed, apnap) ||
		countEventsOfKind(g.Events, game.EventScry) != game.NumPlayers {
		t.Fatalf("offers=%v placements=%v scry events=%d", offered, placed, countEventsOfKind(g.Events, game.EventScry))
	}
	var batch game.ObjectID
	for player := range game.NumPlayers {
		want := []id.ID{libraries[player][1], libraries[player][0]}
		if !slices.Equal(g.Players[player].Library.All(), want) {
			t.Fatalf("player %d library = %v, want %v", player, g.Players[player].Library.All(), want)
		}
	}
	for _, event := range g.Events {
		if event.Kind != game.EventScry {
			continue
		}
		if batch == 0 {
			batch = event.SimultaneousID
		}
		if batch == 0 || batch != event.SimultaneousID {
			t.Fatal("Eager Construct scries were not one simultaneous event cohort")
		}
	}
	assertPrivateScryLog(t, log)
}

func TestEagerConstructEmptyLibrariesAndDepartedPlayers(t *testing.T) {
	engine, def := scryInteractionCard(t, eagerConstructTestOracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fillScryInteractionLibraries(g, 0)
	permanent := enterEagerConstructForTest(t, g, engine, def, game.Player1)
	g.Players[game.Player3].Eliminated = true
	var offered []game.PlayerID
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind != game.ChoiceMay || request.Subject.Exists {
			t.Fatalf("empty libraries should offer only private-free participation choices: %+v", request)
		}
		offered = append(offered, request.Player)
		return []int{1}
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	if !permanentStillPresent(g, permanent.ObjectID) ||
		!slices.Equal(offered, []game.PlayerID{game.Player1, game.Player2, game.Player4}) ||
		countEventsOfKind(g.Events, game.EventScry) != 3 {
		t.Fatalf("empty/departed participant behavior: offers=%v scries=%d", offered, countEventsOfKind(g.Events, game.EventScry))
	}
	for _, choice := range log.Choices {
		if choice.Request.Kind == game.ChoiceScry {
			t.Fatal("an empty library prompted for a nonexistent card")
		}
	}
}

func enterEagerConstructForTest(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, controller game.PlayerID) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, controller, def)
	resolveInstruction(engine, g, &game.StackObject{SourceCardID: cardID, Controller: controller}, game.PutOnBattlefield{
		Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
		Recipient: opt.Val(game.ControllerReference()),
	}, &TurnLog{})
	permanent := permanentByCardID(g, cardID)
	if permanent == nil || !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("actual Eager Construct entry failed to put its trigger on the stack")
	}
	return permanent
}

func TestEagerConstructSourceDepartureCopyCloneAndPrivateReplay(t *testing.T) {
	engine, def := scryInteractionCard(t, eagerConstructTestOracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fillScryInteractionLibraries(g, 2)
	permanent := enterEagerConstructForTest(t, g, engine, def, game.Player1)
	addEffectSpellToStack(g, game.Player1, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(permanent.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, permanent.CardInstanceID) != nil {
		t.Fatal("Eager Construct source did not leave before its trigger resolved")
	}
	trigger, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(trigger.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 2 {
		t.Fatal("Engine did not copy Eager Construct's pending triggered ability")
	}
	clone := g.Clone()
	for resolution := range 2 {
		answer := 1
		if resolution == 1 {
			answer = 0
		}
		beforeGame, beforeClone := g.Clone(), clone.Clone()
		var libraryAtChoice []id.ID
		log := &TurnLog{}
		agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind == game.ChoiceMay {
				if request.Player == game.Player2 {
					return []int{1}
				}
				return []int{0}
			}
			libraryAtChoice = g.Players[game.Player2].Library.All()
			if request.Kind != game.ChoiceScry || request.Player != game.Player2 ||
				len(libraryAtChoice) == 0 || !request.Subject.Exists ||
				request.Subject.Val.CardID != libraryAtChoice[0] {
				t.Fatalf("copied/original Eager Construct trigger used the wrong private chooser: request=%+v library=%v", request, libraryAtChoice)
			}
			return []int{answer}
		}}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
		if countEventsOfKind(g.Events, game.EventScry) != countEventsOfKind(beforeGame.Events, game.EventScry)+1 ||
			countEventsOfKind(clone.Events, game.EventScry) != countEventsOfKind(beforeClone.Events, game.EventScry) {
			t.Fatal("a source-departed trigger failed or mutated its independent clone")
		}
		for player := range game.NumPlayers {
			if !slices.Equal(clone.Players[player].Library.All(), beforeClone.Players[player].Library.All()) {
				t.Fatal("resolving a copied trigger in one branch changed the other branch's library")
			}
		}
		wantLibrary := slices.Clone(libraryAtChoice)
		if answer == 1 {
			slices.Reverse(wantLibrary)
		}
		if !slices.Equal(g.Players[game.Player2].Library.All(), wantLibrary) {
			t.Fatal("copied/original trigger did not apply its own placement choice")
		}
		replay := beforeClone.Clone()
		replayLog := &TurnLog{}
		engine.resolveTopOfStackWithChoices(replay, replayAgentsFromTurnLog(log), replayLog)
		if !reflect.DeepEqual(log, replayLog) {
			t.Fatal("copied-trigger replay changed recorded private selections")
		}
		for player := range game.NumPlayers {
			if !slices.Equal(g.Players[player].Library.All(), replay.Players[player].Library.All()) {
				t.Fatal("cloned copied-trigger replay did not reproduce only its chosen library")
			}
		}
		clone = replay
		assertPrivateScryLog(t, log)
	}
	if !g.Stack.IsEmpty() || !clone.Stack.IsEmpty() {
		t.Fatal("both copied and original Eager Construct triggers did not resolve")
	}
	if !slices.Equal(g.Players[game.Player2].Library.All(), clone.Players[game.Player2].Library.All()) {
		t.Fatal("clone/replay did not preserve independent and reproducible scry choices")
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func replayAgentsFromTurnLog(log *TurnLog) [game.NumPlayers]PlayerAgent {
	var choices [game.NumPlayers][][]int
	for _, choice := range log.Choices {
		player := choice.Request.Player
		choices[player] = append(choices[player], slices.Clone(choice.Selected))
	}
	var agents [game.NumPlayers]PlayerAgent
	for player := range game.NumPlayers {
		agents[player] = &choiceOnlyAgent{choices: choices[player]}
	}
	return agents
}

func TestSeaGodsRevengeDefaultRegistryPaidCastTimingAndTargetCounts(t *testing.T) {
	engine, def := scryInteractionCard(t, seaGodsRevengeTestOracleID)
	if def.Name != "Sea God's Revenge" || authored.SeaGodsRevenge.OracleID != seaGodsRevengeTestOracleID {
		t.Fatalf("default registry identity = %q / %s", def.Name, authored.SeaGodsRevenge.OracleID)
	}
	for _, count := range []int{0, 1, 3} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			top := addCardToLibrary(g, game.Player1, vanillaCreature("Private Sea scry", 1, 1))
			targets := make([]game.Target, 0, count)
			for index, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
				if index >= count {
					break
				}
				permanent := scryInteractionPermanent(g, "Opponent creature", player, player,
					[]types.Card{types.Creature}, nil)
				targets = append(targets, game.PermanentTarget(permanent.ObjectID))
			}
			cardID := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 5)
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			setMainPhasePriority(g, game.Player1)
			if count == 1 {
				g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
				if engine.applyAction(g, game.Player1, action.CastSpell(cardID, targets, 0, nil)) {
					t.Fatal("Sea God's Revenge cast outside sorcery timing")
				}
				setMainPhasePriority(g, game.Player1)
			}
			if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, targets, 0, nil)) ||
				g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("paid {5}{U} Sea God's Revenge cast with this target count failed")
			}
			log := resolveWithScryChoice(t, engine, g, game.Player1, top, 1)
			if countEventsOfKind(g.Events, game.EventScry) != 1 ||
				!g.Players[game.Player1].Graveyard.Contains(cardID) {
				t.Fatal("zero-to-three-target resolution omitted its scry or normal spell destination")
			}
			if count == 0 && len(log.Choices) != 1 {
				t.Fatal("zero-target spell did not resolve its independent scry")
			}
			if count != 0 && (len(g.Players[game.Player1].Library.All()) != 1 ||
				g.Players[game.Player1].Library.All()[0] != top) {
				t.Fatal("resolved Sea God's Revenge did not scry after returning targets")
			}
		})
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := scryInteractionPermanent(g, "One creature", game.Player2, game.Player2,
		[]types.Card{types.Creature}, nil)
	duplicateSpell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	duplicate := []game.Target{game.PermanentTarget(permanent.ObjectID), game.PermanentTarget(permanent.ObjectID)}
	if engine.applyAction(g, game.Player1, action.CastSpell(duplicateSpell, duplicate, 0, nil)) ||
		!g.Players[game.Player1].Hand.Contains(duplicateSpell) || !g.Stack.IsEmpty() {
		t.Fatal("Sea God's Revenge accepted duplicate targets for one target phrase")
	}
}

func TestSeaGodsRevengeSimultaneousOwnerRelativeReturns(t *testing.T) {
	engine, def := scryInteractionCard(t, seaGodsRevengeTestOracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCardToLibrary(g, game.Player1, vanillaCreature("Private Sea scry", 1, 1))
	first := scryInteractionPermanent(g, "Opponent creature one", game.Player3, game.Player2,
		[]types.Card{types.Creature}, nil)
	second := scryInteractionPermanent(g, "Opponent creature two", game.Player1, game.Player3,
		[]types.Card{types.Creature}, nil)
	third := scryInteractionPermanent(g, "Opponent commander", game.Player4, game.Player4,
		[]types.Card{types.Creature}, nil)
	trackCommanderID(g, game.Player4, third.CardInstanceID)
	targets := []game.Target{
		game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID),
		game.PermanentTarget(third.ObjectID),
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	castScryInteractionSpell(t, g, engine, game.Player1, def, targets)
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Sea God's Revenge did not pay {5}{U}")
	}
	var choiceKinds []game.ChoiceKind
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		choiceKinds = append(choiceKinds, request.Kind)
		if request.Kind == game.ChoiceMay {
			if request.Player != game.Player4 || len(request.Options) != 2 {
				t.Fatalf("commander's owner did not receive its replacement choice: %+v", request)
			}
			return []int{1}
		}
		if request.Kind != game.ChoiceScry || request.Player != game.Player1 ||
			!request.Subject.Exists || request.Subject.Val.CardID != top {
			t.Fatalf("Sea God's Revenge scry happened before or lacked the expected card: %+v", request)
		}
		for _, target := range []*game.Permanent{first, second, third} {
			if permanentStillPresent(g, target.ObjectID) {
				t.Fatal("a target had not moved before the following scry")
			}
		}
		return []int{1}
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	if !slices.Equal(choiceKinds, []game.ChoiceKind{game.ChoiceMay, game.ChoiceScry}) ||
		!g.Players[game.Player3].Hand.Contains(first.CardInstanceID) ||
		!g.Players[game.Player1].Hand.Contains(second.CardInstanceID) ||
		!g.Players[game.Player4].CommandZone.Contains(third.CardInstanceID) ||
		countEventsOfKind(g.Events, game.EventScry) != 1 {
		t.Fatal("simultaneous owner-relative return or commander replacement was incorrect")
	}
	var batch game.ObjectID
	for _, target := range []*game.Permanent{first, second, third} {
		found := false
		for _, event := range g.Events {
			if event.Kind != game.EventZoneChanged || event.CardID != target.CardInstanceID ||
				event.FromZone != zone.Battlefield {
				continue
			}
			found = true
			if batch == 0 {
				batch = event.SimultaneousID
			}
			if batch == 0 || batch != event.SimultaneousID {
				t.Fatal("Sea God's Revenge did not move all legal targets in one simultaneous batch")
			}
		}
		if !found {
			t.Fatal("Sea God's Revenge omitted a legal target's zone-change event")
		}
	}
	assertPrivateScryLog(t, log)
}

func TestSeaGodsRevengeRevalidatesEachTargetAndSkipsScryWhenAllIllegal(t *testing.T) {
	t.Run("changed control is partial", func(t *testing.T) {
		engine, def := scryInteractionCard(t, seaGodsRevengeTestOracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		top := addCardToLibrary(g, game.Player1, vanillaCreature("Private Sea scry", 1, 1))
		changed := scryInteractionPermanent(g, "Control changed", game.Player2, game.Player2,
			[]types.Card{types.Creature}, nil)
		legal := scryInteractionPermanent(g, "Still legal", game.Player3, game.Player3,
			[]types.Card{types.Creature}, nil)
		targets := []game.Target{game.PermanentTarget(changed.ObjectID), game.PermanentTarget(legal.ObjectID)}
		g.Players[game.Player1].ManaPool.Add(mana.C, 5)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		castScryInteractionSpell(t, g, engine, game.Player1, def, targets)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), AffectedObjectID: changed.ObjectID, Layer: game.LayerControl,
			NewController: opt.Val(game.Player1),
		})
		resolveWithScryChoiceAfter(t, engine, g, game.Player1, top, 0, func() {
			if effectiveController(g, changed) != game.Player1 ||
				permanentStillPresent(g, legal.ObjectID) ||
				!g.Players[game.Player3].Hand.Contains(legal.CardInstanceID) {
				t.Fatal("Sea God's Revenge scried before returning the still-legal target")
			}
		})
		if effectiveController(g, changed) != game.Player1 ||
			!permanentStillPresent(g, changed.ObjectID) ||
			permanentStillPresent(g, legal.ObjectID) ||
			!g.Players[game.Player3].Hand.Contains(legal.CardInstanceID) ||
			countEventsOfKind(g.Events, game.EventScry) != 1 {
			t.Fatal("control revalidation did not preserve the changed target and resolve the legal target")
		}
	})

	t.Run("changed type and new incarnation are all illegal", func(t *testing.T) {
		engine, def := scryInteractionCard(t, seaGodsRevengeTestOracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addCardToLibrary(g, game.Player1, vanillaCreature("Private Sea scry", 1, 1))
		changedType := scryInteractionPermanent(g, "Type changed", game.Player2, game.Player2,
			[]types.Card{types.Creature}, nil)
		reincarnating := scryInteractionPermanent(g, "New incarnation", game.Player3, game.Player3,
			[]types.Card{types.Creature}, nil)
		g.Players[game.Player1].ManaPool.Add(mana.C, 5)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		castScryInteractionSpell(t, g, engine, game.Player1, def, []game.Target{
			game.PermanentTarget(changedType.ObjectID), game.PermanentTarget(reincarnating.ObjectID),
		})
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), AffectedObjectID: changedType.ObjectID, Layer: game.LayerType,
			RemoveTypes: []types.Card{types.Creature},
		})
		if !movePermanentToZone(g, reincarnating, zone.Hand) {
			t.Fatal("setup could not move the targeted permanent to a new zone")
		}
		resolveInstruction(engine, g, &game.StackObject{
			SourceCardID: reincarnating.CardInstanceID, Controller: game.Player3,
		}, game.PutOnBattlefield{
			Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
			Recipient: opt.Val(game.ControllerReference()),
		}, &TurnLog{})
		newIncarnation := permanentByCardID(g, reincarnating.CardInstanceID)
		if newIncarnation == nil || newIncarnation.ObjectID == reincarnating.ObjectID {
			t.Fatal("setup did not create a new battlefield incarnation")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if countEventsOfKind(g.Events, game.EventScry) != 0 ||
			!permanentStillPresent(g, changedType.ObjectID) ||
			!permanentStillPresent(g, newIncarnation.ObjectID) {
			t.Fatal("all-illegal Sea God's Revenge resolved on a changed type or new incarnation")
		}
	})
}

func TestSeaGodsRevengeZeroTargetsScry(t *testing.T) {
	engine, def := scryInteractionCard(t, seaGodsRevengeTestOracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	top := addCardToLibrary(g, game.Player1, vanillaCreature("Private zero-target scry", 1, 1))
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cardID := castScryInteractionSpell(t, g, engine, game.Player1, def, nil)
	resolveWithScryChoice(t, engine, g, game.Player1, top, 1)
	if countEventsOfKind(g.Events, game.EventScry) != 1 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("zero-target Sea God's Revenge did not resolve and scry")
	}
}

func TestDevoutDecreeDefaultRegistryPaidCastAndCharacteristicUnions(t *testing.T) {
	engine, def := scryInteractionCard(t, devoutDecreeTestOracleID)
	if def.Name != "Devout Decree" || authored.DevoutDecree.OracleID != devoutDecreeTestOracleID {
		t.Fatalf("default registry identity = %q / %s", def.Name, authored.DevoutDecree.OracleID)
	}
	tests := []struct {
		name      string
		cardTypes []types.Card
		colors    []color.Color
		legal     bool
	}{
		{"black creature", []types.Card{types.Creature}, []color.Color{color.Black}, true},
		{"red creature", []types.Card{types.Creature}, []color.Color{color.Red}, true},
		{"black planeswalker", []types.Card{types.Planeswalker}, []color.Color{color.Black}, true},
		{"red planeswalker", []types.Card{types.Planeswalker}, []color.Color{color.Red}, true},
		{"black red creature", []types.Card{types.Creature}, []color.Color{color.Black, color.Red}, true},
		{"blue creature", []types.Card{types.Creature}, []color.Color{color.Blue}, false},
		{"black artifact", []types.Card{types.Artifact}, []color.Color{color.Black}, false},
		{"white planeswalker", []types.Card{types.Planeswalker}, []color.Color{color.White}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			top := addCardToLibrary(g, game.Player1, vanillaCreature("Private decree scry", 1, 1))
			target := scryInteractionPermanent(g, test.name, game.Player2, game.Player2, test.cardTypes, test.colors)
			cardID := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			setMainPhasePriority(g, game.Player1)
			cast := engine.applyAction(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil))
			if cast != test.legal {
				t.Fatalf("cast legal=%v, want %v for %s", cast, test.legal, test.name)
			}
			if !test.legal {
				if !g.Players[game.Player1].Hand.Contains(cardID) || !g.Stack.IsEmpty() {
					t.Fatal("an out-of-union target was consumed")
				}
				return
			}
			if g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("paid {1}{W} Devout Decree did not spend its cost")
			}
			resolveWithScryChoice(t, engine, g, game.Player1, top, 1)
			if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) ||
				countEventsOfKind(g.Events, game.EventScry) != 1 {
				t.Fatal("Devout Decree failed its legal type/color-union exile followed by scry")
			}
		})
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	legalTarget := scryInteractionPermanent(g, "Legal timing target", game.Player2, game.Player2,
		[]types.Card{types.Creature}, []color.Color{color.Black})
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	if engine.applyAction(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.PermanentTarget(legalTarget.ObjectID)}, 0, nil)) {
		t.Fatal("Devout Decree cast outside sorcery timing")
	}
	if !g.Players[game.Player1].Hand.Contains(cardID) || !g.Stack.IsEmpty() {
		t.Fatal("out-of-timing Devout Decree consumed the card")
	}
}

func TestDevoutDecreeRevalidatesTargetAndExilesThenScries(t *testing.T) {
	engine, def := scryInteractionCard(t, devoutDecreeTestOracleID)
	t.Run("zone replacement", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		top := addCardToLibrary(g, game.Player1, vanillaCreature("Private decree scry", 1, 1))
		target := scryInteractionPermanent(g, "Replacement target", game.Player2, game.Player2,
			[]types.Card{types.Creature}, []color.Color{color.Red})
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Battlefield,
			MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Graveyard,
		})
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		castScryInteractionSpell(t, g, engine, game.Player1, def,
			[]game.Target{game.PermanentTarget(target.ObjectID)})
		resolveWithScryChoiceAfter(t, engine, g, game.Player1, top, 1, func() {
			if permanentStillPresent(g, target.ObjectID) ||
				!g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
				t.Fatal("Devout Decree scried before the replacement-aware exile completed")
			}
		})
		if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
			g.Players[game.Player2].Exile.Contains(target.CardInstanceID) ||
			countEventsOfKind(g.Events, game.EventScry) != 1 {
			t.Fatal("replacement-aware exile did not apply its actual destination before Devout Decree's scry")
		}
	})

	t.Run("commander exile is followed by its state action", func(t *testing.T) {
		engine, def := scryInteractionCard(t, devoutDecreeTestOracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		top := addCardToLibrary(g, game.Player1, vanillaCreature("Private commander scry", 1, 1))
		target := scryInteractionPermanent(g, "Black commander", game.Player2, game.Player2,
			[]types.Card{types.Creature}, []color.Color{color.Black})
		trackCommanderID(g, game.Player2, target.CardInstanceID)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		castScryInteractionSpell(t, g, engine, game.Player1, def,
			[]game.Target{game.PermanentTarget(target.ObjectID)})
		scryAgent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceScry || request.Player != game.Player1 ||
				!request.Subject.Exists || request.Subject.Val.CardID != top {
				t.Fatalf("commander was moved before the required scry: %+v", request)
			}
			if permanentStillPresent(g, target.ObjectID) ||
				!g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
				t.Fatal("Devout Decree scried before its exile action")
			}
			return []int{1}
		}}
		log := &TurnLog{}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: scryAgent}, log)
		if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) ||
			countEventsOfKind(g.Events, game.EventScry) != 1 {
			t.Fatal("Devout Decree did not exile first and scry before state-based commander handling")
		}
		owner := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceMay || request.Player != game.Player2 {
				t.Fatalf("commander replacement was not offered to its owner: %+v", request)
			}
			return []int{1}
		}}
		engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: owner}, log)
		engine.applyStateBasedActions(g)
		if !g.Players[game.Player2].CommandZone.Contains(target.CardInstanceID) ||
			g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
			t.Fatal("owner's commander-to-command-zone choice was not applied after resolution")
		}
	})
}

func TestDevoutDecreeIllegalTargetOmitsScry(t *testing.T) {
	for _, change := range []string{"lost qualifying color", "lost creature type"} {
		t.Run(change, func(t *testing.T) {
			engine, def := scryInteractionCard(t, devoutDecreeTestOracleID)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, vanillaCreature("Private decree scry", 1, 1))
			target := scryInteractionPermanent(g, "Changing target", game.Player2, game.Player2,
				[]types.Card{types.Creature}, []color.Color{color.Black})
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			g.Players[game.Player1].ManaPool.Add(mana.W, 1)
			castScryInteractionSpell(t, g, engine, game.Player1, def,
				[]game.Target{game.PermanentTarget(target.ObjectID)})
			effect := game.ContinuousEffect{ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID}
			switch change {
			case "lost qualifying color":
				effect.Layer = game.LayerColor
				effect.SetColorless = true
			case "lost creature type":
				effect.Layer = game.LayerType
				effect.RemoveTypes = []types.Card{types.Creature}
			}
			g.ContinuousEffects = append(g.ContinuousEffects, effect)
			engine.resolveTopOfStack(g, &TurnLog{})
			if countEventsOfKind(g.Events, game.EventScry) != 0 ||
				!permanentStillPresent(g, target.ObjectID) {
				t.Fatal("Devout Decree resolved after its sole target ceased to qualify")
			}
		})
	}
}
