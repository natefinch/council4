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
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type authoredChoicePolicy struct {
	simPassPolicy

	choose func(PlayerObservation, game.ChoiceRequest) []int
}

func (p authoredChoicePolicy) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	return p.choose(obs, request)
}

func chooseAuthoredTargets(t *testing.T, request game.ChoiceRequest, targets ...game.Target) []int {
	t.Helper()
	for _, option := range request.Options {
		if slices.Equal(option.Targets, targets) {
			return []int{option.Index}
		}
	}
	t.Fatalf("requested targets %v not offered: %+v", targets, request)
	return nil
}

func productionChoiceCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupChecked(name)
	if err != nil || def == nil {
		t.Fatalf("production card %s unavailable: %v", name, err)
	}
	return def
}

func TestAuthoredHiveMindRealCastRetargetsAPNAPAndSimulator(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Hive Mind"))
	bolt := cards.NewDefaultRegistry().Lookup("Lightning Bolt")
	if bolt == nil {
		t.Fatal("production Lightning Bolt missing")
	}
	cardID := addCardToHand(g, game.Player2, bolt)
	g.Players[game.Player2].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyActionWithChoices(g, game.Player2,
		action.CastSpell(cardID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil), simPassPolicies(), &TurnLog{}) {
		t.Fatal("real Lightning Bolt cast failed")
	}
	setMainPhasePriority(g, game.Player3)
	policies := simPassPolicies()
	for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		policies[player] = authoredChoicePolicy{choose: func(obs PlayerObservation, request game.ChoiceRequest) []int {
			if obs.Player != player || request.Player != player || request.Kind != game.ChoiceTarget {
				t.Fatal("copy choice reached the wrong decision maker or observer")
			}
			if player == game.Player4 {
				return request.DefaultSelection
			}
			return chooseAuthoredTargets(t, request, game.PlayerTarget(player))
		}}
	}
	branch := engine.newSearchContext(g, game.Player3).Simulator().ResolvePriority(g, policies)
	log := TurnLog{}
	engine.runPriorityLoop(g, policies, &log)
	if !reflect.DeepEqual(branch.Events, g.Events) {
		t.Fatal("Simulator and scripted stack produced different events")
	}
	if len(log.Choices) != 3 {
		t.Fatalf("copy choices = %d, want 3", len(log.Choices))
	}
	for i, player := range []game.PlayerID{game.Player3, game.Player4, game.Player1} {
		if log.Choices[i].Request.Player != player || log.Choices[i].UsedFallback {
			t.Fatal("separate copy choices were not recorded in APNAP order")
		}
	}
	var copyOrder []game.PlayerID
	casts := 0
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast {
			casts++
		}
		if event.Kind == game.EventSpellCopied {
			copyOrder = append(copyOrder, event.Controller)
		}
	}
	if casts != 1 || !slices.Equal(copyOrder, []game.PlayerID{game.Player3, game.Player4, game.Player1}) {
		t.Fatal("copies were casts or were not placed on the stack in APNAP order")
	}
	for player, life := range []int{37, 34, 37, 40} {
		if g.Players[player].Life != life || branch.Players[player].Life != life {
			t.Fatal("the three separate retarget/keep decisions did not actually resolve")
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(cardID) || g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatal("copies moved the original physical card into a copier's graveyard")
	}
}

func TestAuthoredHiveMindMultipleTriggersAndDepartedPlayers(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"two minds", "departed copier", "original countered", "departed original caster"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			addCombatPermanent(g, game.Player1, authored.HiveMind.New())
			if scenario == "two minds" {
				addCombatPermanent(g, game.Player4, authored.HiveMind.New())
			}
			bolt := cards.NewDefaultRegistry().Lookup("Lightning Bolt")
			cardID := addCardToHand(g, game.Player2, bolt)
			g.Players[game.Player2].ManaPool.Add(mana.R, 1)
			setMainPhasePriority(g, game.Player2)
			if !engine.applyAction(g, game.Player2, action.CastSpell(cardID, []game.Target{game.PlayerTarget(game.Player3)}, 0, nil)) {
				t.Fatal("cast failed")
			}
			setMainPhasePriority(g, game.Player3)
			log := TurnLog{}
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &log) {
				t.Fatal("actual cast did not announce Hive Mind")
			}
			objects := g.Stack.Objects()
			if scenario == "two minds" {
				top, _ := g.Stack.Peek()
				if top.Controller != game.Player1 || len(objects) != 3 || objects[1].Controller != game.Player4 {
					t.Fatal("simultaneous Hive Mind triggers did not use APNAP")
				}
			}
			if scenario == "departed copier" {
				engine.eliminatePlayer(g, game.Player4)
			}
			if scenario == "original countered" {
				counterStackObject(g, objects[0].ID)
			}
			if scenario == "departed original caster" {
				card, _ := g.GetCardInstance(cardID)
				card.Owner = game.Player4
				objects[0].Controller = game.Player3
				engine.eliminatePlayer(g, game.Player2)
			}
			engine.runPriorityLoop(g, simPassPolicies(), &log)
			wantDamage := map[string]int{"two minds": 21, "departed copier": 9, "original countered": 0, "departed original caster": 12}[scenario]
			if g.Players[game.Player3].Life != 40-wantDamage {
				t.Fatal("copy count did not reflect triggers, departures or unavailable original")
			}
			for _, event := range g.Events {
				if scenario == "departed copier" && event.Kind == game.EventSpellCopied && event.Controller == game.Player4 {
					t.Fatal("eliminated player received a spell copy")
				}
			}
		})
	}
}

func TestAuthoredHiveMindInheritsModesXAndActualCostSubjects(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addCombatPermanent(g, game.Player1, authored.HiveMind.New())
	paid := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Modal cost probe", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalSacrifice, Amount: 1, MatchPermanentType: true, PermanentType: types.Creature, SubjectKey: "paid",
		}},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1, MaxModes: 1,
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(99)}}}},
				{
					Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
					Sequence: []game.Instruction{
						{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})}},
						{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
							Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.PaidCostReference("paid", game.PaidCostSacrifice, types.Creature)),
								ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
							})})},
					},
				},
			},
		}),
	}}
	paidDef, _ := g.GetCardInstance(paid.CardInstanceID)
	paidDef.Def = vanillaCreature("Red cost", 5, 5)
	paidDef.Def.Colors = []color.Color{color.Red}
	cardID := addCardToHand(g, game.Player2, def)
	g.Players[game.Player2].ManaPool.Add(mana.R, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastSpell(cardID, []game.Target{game.PlayerTarget(game.Player2)}, 3, []int{1})) {
		t.Fatal("real modal X cast with sacrifice cost failed")
	}
	original, _ := g.Stack.Peek()
	if len(original.PaidCostSubjects) != 1 || original.PaidCostSubjects[0].Snapshot.ObjectID != paid.ObjectID ||
		!g.Players[game.Player2].Graveyard.Contains(paid.CardInstanceID) {
		t.Fatal("cast did not actually pay and capture the sacrifice")
	}
	policies := simPassPolicies()
	for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		policies[player] = authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
			return chooseAuthoredTargets(t, req, game.PlayerTarget(player))
		}}
	}
	branch := engine.newSearchContext(g, game.Player2).Simulator().ResolvePriority(g, policies)
	log := TurnLog{}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, policies, &log)
	engine.resolveTopOfStackWithChoices(g, policies, &log)
	for _, object := range g.Stack.Objects() {
		if object.XValue != 3 || !slices.Equal(object.ChosenModes, []int{1}) ||
			!reflect.DeepEqual(object.PaidCostSubjects, original.PaidCostSubjects) {
			t.Fatal("a Hive Mind copy lost the mode, X or frozen payment facts")
		}
	}
	engine.runPriorityLoop(g, policies, &log)
	for player := range game.NumPlayers {
		if g.Players[player].Life != 39 || branch.Players[player].Life != 39 {
			t.Fatal("inherited mode/X/paid-cost predicate did not execute for every copier")
		}
	}
	if !reflect.DeepEqual(branch.Events, g.Events) || original.PaidCostSubjects[0].Snapshot.Colors[0] != color.Red {
		t.Fatal("copy/Simulator state aliased payment facts or diverged")
	}
}

func TestAuthoredGildedDrakeRealEntryAndSimulator(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	target := addCombatCreaturePermanent(g, game.Player2)
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Gilded Drake"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	policies := simPassPolicies()
	policies[game.Player1] = authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		return chooseAuthoredTargets(t, req, game.PermanentTarget(target.ObjectID))
	}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), policies, &TurnLog{}) {
		t.Fatal("Drake cast failed")
	}
	branch := engine.newSearchContext(g, game.Player1).Simulator().ResolvePriority(g, policies)
	if permanentByCardID(g, cardID) != nil || effectiveController(g, target) != game.Player2 {
		t.Fatal("Simulator mutated the original checkpoint")
	}
	log := TurnLog{}
	engine.runPriorityLoop(g, policies, &log)
	for _, state := range []*game.Game{g, branch} {
		drake := permanentByCardID(state, cardID)
		victim := permanentByCardID(state, target.CardInstanceID)
		if drake == nil || victim == nil || effectiveController(state, drake) != game.Player2 ||
			effectiveController(state, victim) != game.Player1 || !drake.SummoningSick || !victim.SummoningSick {
			t.Fatal("actual entry trigger did not perform both control changes")
		}
	}
	if len(log.Choices) != 1 || log.Choices[0].UsedFallback || !reflect.DeepEqual(branch.Events, g.Events) {
		t.Fatal("trigger target was not a recorded scripted decision or clone diverged")
	}
	engine.eliminatePlayer(g, game.Player1)
	if effectiveController(g, target) != game.Player2 || permanentByCardID(branch, cardID) == nil {
		t.Fatal("control grant did not end on recipient departure or leaked into clone")
	}
}

func TestAuthoredGiftsUngivenRealCastHiddenChoicesAndSimulator(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	ids := make(map[string]game.ObjectID)
	for _, name := range []string{"One", "Two", "Three", "Four", "Not found"} {
		ids[name] = addCardToLibrary(g, game.Player1, vanillaCreature(name, 2, 2))
	}
	addCardToHand(g, game.Player1, vanillaCreature("Searcher's secret", 2, 2))
	addCardToHand(g, game.Player2, vanillaCreature("Opponent's secret", 2, 2))
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Gifts Ungiven"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Gifts cast failed")
	}
	beforeShuffle := g.Players[game.Player1].Library.ReorderVersion()
	policies := simPassPolicies()
	policies[game.Player1] = authoredChoicePolicy{choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if obs.Player != game.Player1 || req.Player != game.Player1 || req.Kind != game.ChoiceSearch ||
			len(obs.Hand()) != 1 || obs.Hand()[0].Name != "Searcher's secret" || req.MinChoices != 0 || req.MaxChoices != 1 {
			t.Fatal("library search used the wrong information view or decision bounds")
		}
		if _, exposed := obs.LibraryTopRevealed(game.Player1); exposed {
			t.Fatal("search exposed private library order through observation")
		}
		for _, name := range []string{"One", "Two", "Three", "Four"} {
			for _, option := range req.Options {
				if option.Label == name {
					if !option.Card.Exists || option.Card.Val.CardID != ids[name] {
						t.Fatal("search options did not carry the privately visible card")
					}
					return []int{option.Index}
				}
			}
		}
		return nil
	}}
	policies[game.Player2] = authoredChoicePolicy{choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if obs.Player != game.Player2 || req.Player != game.Player2 || req.Kind != game.ChoiceZoneSelection ||
			len(obs.Hand()) != 1 || obs.Hand()[0].Name != "Opponent's secret" || req.MinChoices != 2 || req.MaxChoices != 2 ||
			len(req.Options) != 4 || len(obs.Graveyard(game.Player1)) != 0 || obs.PlayerState(game.Player1).LibrarySize != 5 {
			t.Fatal("opponent partition did not occur at resolution before placement with its own hidden view")
		}
		for i, name := range []string{"One", "Two", "Three", "Four"} {
			if !req.Options[i].Card.Exists || req.Options[i].Card.Val.CardID != ids[name] {
				t.Fatal("partition offered anything other than the exact revealed set")
			}
		}
		return []int{1, 3}
	}}
	branch := engine.newSearchContext(g, game.Player1).Simulator().ResolvePriority(g, policies)
	if g.Players[game.Player1].Library.Size() != 5 || g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatal("Gifts Simulator mutated its original library/stack checkpoint")
	}
	log := TurnLog{}
	engine.runPriorityLoop(g, policies, &log)
	if !reflect.DeepEqual(branch.Events, g.Events) || len(log.Choices) != 5 {
		t.Fatal("scripted production body and Simulator did not agree")
	}
	for _, decision := range log.Choices {
		if decision.UsedFallback {
			t.Fatal("a success-shaped fallback replaced a scripted Gifts choice")
		}
	}
	var reveals []game.ObjectID
	var batch game.ObjectID
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			reveals = append(reveals, event.CardID)
		}
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Library {
			if len(reveals) != 4 || event.SimultaneousID == 0 || batch != 0 && batch != event.SimultaneousID {
				t.Fatal("partition was not one simultaneous placement after the exact reveal")
			}
			batch = event.SimultaneousID
		}
	}
	if !slices.Equal(reveals, []game.ObjectID{ids["One"], ids["Two"], ids["Three"], ids["Four"]}) ||
		!g.Players[game.Player1].Hand.Contains(ids["One"]) || !g.Players[game.Player1].Hand.Contains(ids["Three"]) ||
		!g.Players[game.Player1].Graveyard.Contains(ids["Two"]) || !g.Players[game.Player1].Graveyard.Contains(ids["Four"]) ||
		!g.Players[game.Player1].Library.Contains(ids["Not found"]) || g.Players[game.Player1].Library.ReorderVersion() != beforeShuffle+1 {
		t.Fatalf("Gifts reveal/partition/shuffle: reveals=%v hand=%v graveyard=%v library=%v epoch=%d before=%d",
			reveals, g.Players[game.Player1].Hand.All(), g.Players[game.Player1].Graveyard.All(),
			g.Players[game.Player1].Library.All(), g.Players[game.Player1].Library.ReorderVersion(), beforeShuffle)
	}
}
