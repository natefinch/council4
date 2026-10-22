package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Sources: pinned Oracle ARRAY SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// ONE 235 printing 56bfdd21-4d46-4ec2-ba25-f16b6993f1a8.
// https://api.scryfall.com/cards/56bfdd21-4d46-4ec2-ba25-f16b6993f1a8/rulings
// fetched 2026-10-22: complete response, one WotC ruling (2023-02-04).
// Verified official CR effective 2026-09-25, high-review-rules.txt SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca:
// 101.4a-b/608.2c-f/701.22a-d require controller scry first, all opponent may
// choices, APNAP private placement choices with earlier public counts, then
// simultaneous placement. 401.2/4 keep identities/order private; 113.7a/603.3a/6a
// preserve the triggered controller after source departure; 601.2f-h require
// complete payment; 707.2/10 preserve the copied ability, not prior scry choices.
const myrCustodianTestOracleID = "43923f24-7e95-4934-9a5c-8d9a5e06ddea"

func TestMyrCustodianDefaultRegistryAdmission(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(myrCustodianTestOracleID)
	if err != nil || def.Name != "Myr Custodian" || registry.Lookup("Myr Custodian") != def {
		t.Fatalf("new Myr Custodian exact/default identity was not admitted: %v", err)
	}
}

func myrCustodianScenario(t *testing.T, size int) (*game.Game, *Engine, *game.CardDef, [game.NumPlayers][]id.ID) {
	t.Helper()
	registry, err := cards.NewProgramRegistry([]cardset.Entry{authored.MyrCustodian})
	if err != nil {
		t.Fatal(err)
	}
	def, err := registry.LookupOracleID(myrCustodianTestOracleID)
	if err != nil || def.Name != "Myr Custodian" {
		t.Fatalf("checked exact Myr Custodian lookup = %v, %v", def, err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	var libraries [game.NumPlayers][]id.ID
	for player := range game.NumPlayers {
		libraries[player] = make([]id.ID, size)
		for index := size - 1; index >= 0; index-- {
			libraries[player][index] = addCardToLibrary(g, game.PlayerID(player),
				vanillaCreature(fmt.Sprintf("Scry secret player %d card %d", player, index), 2, 2))
		}
	}
	return g, NewEngineWithPrograms(nil, registry.Programs()), def, libraries
}

func enterMyrCustodian(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, controller game.PlayerID) *game.Permanent {
	t.Helper()
	card := addCardToHand(g, game.Player1, def)
	resolveInstruction(engine, g, &game.StackObject{SourceCardID: card, Controller: controller}, game.PutOnBattlefield{
		Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
		Recipient: opt.Val(game.ControllerReference()),
	}, &TurnLog{})
	permanent := permanentByCardID(g, card)
	if permanent == nil || effectiveController(g, permanent) != controller {
		t.Fatal("Engine entry did not give Myr Custodian to the specified controller")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("real Myr Custodian entry did not trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.Controller != controller ||
		trigger.SourceID != permanent.ObjectID || len(trigger.Targets) != 0 {
		t.Fatal("Myr Custodian trigger has wrong controller/source or targets")
	}
	return permanent
}

func TestMyrCustodianPaidCastAndPrivateScry(t *testing.T) {
	g, engine, def, libraries := myrCustodianScenario(t, 3)
	card := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	cast := action.CastSpell(card, nil, 0, nil)
	if engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Myr Custodian cast without paying {3}")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	if engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Myr Custodian cast outside sorcery timing")
	}
	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) ||
		!engine.applyAction(g, game.Player1, cast) {
		t.Fatal("printed nontargeted cast did not pay and enter the stack")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Myr Custodian did not spend its full printed cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, card)
	if permanent == nil || countEventsOfKind(g.Events, game.EventScry) != 0 ||
		!engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Myr Custodian did not enter before its scry trigger")
	}
	requests := 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		requests++
		switch requests {
		case 1:
			if request.Kind != game.ChoiceScry || request.Player != game.Player1 || len(request.Options) != 2 ||
				request.Options[0].Card.Val.CardID != libraries[game.Player1][0] {
				t.Fatal("controller did not privately scry the actual top two first")
			}
			return nil
		case 2:
			if request.Kind != game.ChoiceOrder || request.Player != game.Player1 {
				t.Fatal("controller could not order the two cards kept on top")
			}
			return []int{1, 0}
		case 3, 4, 5:
			if request.Kind != game.ChoiceMay || request.Player != game.PlayerID(requests-2) ||
				!slices.Equal(g.Players[game.Player1].Library.All(), []id.ID{
					libraries[0][1], libraries[0][0], libraries[0][2],
				}) {
				t.Fatal("opponents were not offered scry after controller finished")
			}
			if request.Player == game.Player3 {
				return []int{0}
			}
			return []int{1}
		case 6, 7:
			player := game.Player2
			if requests == 7 {
				player = game.Player4
				if len(request.ScryDecisions) != 3 || !request.ScryDecisions[0].PlacementChosen ||
					request.ScryDecisions[0].BottomCount != 1 || request.ScryDecisions[1].Accepted {
					t.Fatal("later opponent lost earlier public scry/decline decisions")
				}
			}
			if request.Kind != game.ChoiceScry || request.Player != player ||
				!request.Subject.Exists || request.Subject.Val.CardID != libraries[player][0] ||
				!slices.Equal(g.Players[game.Player2].Library.All(), libraries[game.Player2]) ||
				!slices.Equal(g.Players[game.Player4].Library.All(), libraries[game.Player4]) {
				t.Fatal("opponents' scry was not staged privately before either placement")
			}
			return []int{1}
		default:
			t.Fatal("unexpected extra Myr Custodian choice")
			return nil
		}
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	if requests != 7 || !g.Stack.IsEmpty() || countEventsOfKind(g.Events, game.EventScry) != 3 {
		t.Fatal("Myr Custodian skipped or duplicated printed scry behavior")
	}
	for _, player := range []game.PlayerID{game.Player2, game.Player4} {
		if !slices.Equal(g.Players[player].Library.All(), []id.ID{libraries[player][1], libraries[player][2], libraries[player][0]}) {
			t.Fatal("accepting opponent did not put its chosen card on bottom")
		}
	}
	if !slices.Equal(g.Players[game.Player3].Library.All(), libraries[game.Player3]) {
		t.Fatal("declining opponent's library changed")
	}
	assertPrivateScryLog(t, log)
}

func TestMyrCustodianControllerAndAPNAP(t *testing.T) {
	for _, controller := range []game.PlayerID{game.Player2, game.Player3} {
		t.Run(fmt.Sprint(controller), func(t *testing.T) {
			g, engine, def, _ := myrCustodianScenario(t, 2)
			g.Turn.ActivePlayer = game.Player4
			enterMyrCustodian(t, g, engine, def, controller)
			var expected []game.PlayerID
			for _, player := range []game.PlayerID{game.Player4, game.Player1, game.Player2, game.Player3} {
				if player != controller {
					expected = append(expected, player)
				}
			}
			var offered []game.PlayerID
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceMay {
					offered = append(offered, request.Player)
					return []int{0}
				}
				if request.Player != controller {
					t.Fatal("source owner or active player scried instead of triggered controller")
				}
				if request.Kind == game.ChoiceOrder {
					return []int{0, 1}
				}
				return nil
			}}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
			if !slices.Equal(offered, expected) || countEventsOfKind(g.Events, game.EventScry) != 1 {
				t.Fatalf("opponent offers = %v, want APNAP %v", offered, expected)
			}
		})
	}
}

func TestMyrCustodianSourceDepartureCopiedTriggerAndClone(t *testing.T) {
	g, engine, def, libraries := myrCustodianScenario(t, 3)
	permanent := enterMyrCustodian(t, g, engine, def, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(permanent.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, permanent.CardInstanceID) != nil {
		t.Fatal("setup failed to remove Myr Custodian before trigger resolution")
	}
	trigger, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(trigger.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Stack.Objects()) != 2 {
		t.Fatal("Engine did not create a real copy of the pending triggered ability")
	}
	clone := g.Clone()
	for branch, state := range []*game.Game{g, clone} {
		for resolution := range 2 {
			before := countEventsOfKind(state.Events, game.EventScry)
			controller := game.Player1
			if resolution == 1 {
				controller = game.Player2
			}
			log := &TurnLog{}
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceMay {
					if request.Player == controller {
						t.Fatal("copy/original controller was treated as its own opponent")
					}
					return []int{0}
				}
				if request.Player != controller || request.Kind != game.ChoiceScry {
					t.Fatal("copied trigger used the original source controller or retained earlier choices")
				}
				return []int{branch}
			}}
			engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
			if countEventsOfKind(state.Events, game.EventScry) != before+1 {
				t.Fatal("independent copied/original source-departed trigger failed to scry")
			}
			want := []id.ID{libraries[controller][1-branch], libraries[controller][2], libraries[controller][branch]}
			if !slices.Equal(state.Players[controller].Library.All(), want) {
				t.Fatal("copied/original scry did not honor its own branch's actual private choice")
			}
			for _, decision := range log.Choices {
				if decision.UsedFallback {
					t.Fatal("copied/original trigger ignored a valid resolving choice")
				}
			}
			assertPrivateScryLog(t, log)
		}
		if branch == 0 {
			for player := range game.NumPlayers {
				if !slices.Equal(clone.Players[player].Library.All(), libraries[player]) {
					t.Fatal("resolving Myr Custodian modified its pre-resolution clone")
				}
			}
		}
	}
	if !g.Stack.IsEmpty() || !clone.Stack.IsEmpty() {
		t.Fatal("copied Myr Custodian trigger did not resolve in both branches")
	}
	clone.Players[game.Player2].Library.Remove(libraries[game.Player2][0])
	if !g.Players[game.Player2].Library.Contains(libraries[game.Player2][0]) {
		t.Fatal("Myr Custodian clone shared its mutable library")
	}
}

func TestMyrCustodianEmptyLibrariesAndDepartedOpponent(t *testing.T) {
	g, engine, def, _ := myrCustodianScenario(t, 0)
	enterMyrCustodian(t, g, engine, def, game.Player1)
	g.Players[game.Player3].Eliminated = true
	var offered []game.PlayerID
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind != game.ChoiceMay {
			t.Fatal("empty library offered a nonexistent card")
		}
		offered = append(offered, request.Player)
		return []int{1}
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	if !slices.Equal(offered, []game.PlayerID{game.Player2, game.Player4}) ||
		countEventsOfKind(g.Events, game.EventScry) != 3 {
		t.Fatal("empty-library scries or departed-opponent exclusion are incorrect")
	}
	assertPrivateScryLog(t, log)
}
