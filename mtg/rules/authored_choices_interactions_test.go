package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestAuthoredHiveMindCopiesProductionKickedSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addCombatPermanent(g, game.Player1, authored.HiveMind.New())
	burst := cards.NewDefaultRegistry().Lookup("Burst Lightning")
	if burst == nil {
		t.Fatal("production Burst Lightning missing")
	}
	cardID := addCardToHand(g, game.Player2, burst)
	g.Players[game.Player2].ManaPool.Add(mana.R, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 4)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastKickedSpell(cardID, []game.Target{game.PlayerTarget(game.Player3)}, 0, nil)) {
		t.Fatal("real kicked cast failed")
	}
	checkpoint := g.Clone()
	branch := engine.newSearchContext(g, game.Player2).Simulator().ResolvePriority(g, simPassPolicies())
	log := TurnLog{}
	engine.runPriorityLoop(g, simPassPolicies(), &log)
	if g.Players[game.Player3].Life != 24 || branch.Players[game.Player3].Life != 24 {
		t.Fatal("copies did not actually resolve the inherited kicked branch")
	}
	if checkpoint.Players[game.Player3].Life != 40 || checkpoint.Stack.Size() != 1 {
		t.Fatal("resolution mutated a sibling checkpoint")
	}
}

func TestAuthoredHiveMindCopiesProductionChosenMode(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addCombatPermanent(g, game.Player1, authored.HiveMind.New())
	charm := cards.NewDefaultRegistry().Lookup("Bant Charm")
	if charm == nil {
		t.Fatal("production modal spell missing")
	}
	var targets []*game.Permanent
	for player := range game.NumPlayers {
		targets = append(targets, addCombatCreaturePermanent(g, game.PlayerID(player)))
	}
	cardID := addCardToHand(g, game.Player2, charm)
	for _, color := range []mana.Color{mana.G, mana.W, mana.U} {
		g.Players[game.Player2].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastSpell(cardID, []game.Target{game.PermanentTarget(targets[1].ObjectID)}, 0, []int{1})) {
		t.Fatal("real modal cast failed")
	}
	policies := simPassPolicies()
	for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		policies[player] = authoredChoicePolicy{choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			return chooseAuthoredTargets(t, request, game.PermanentTarget(targets[player].ObjectID))
		}}
	}
	branch := engine.newSearchContext(g, game.Player2).Simulator().ResolvePriority(g, policies)
	engine.runPriorityLoop(g, policies, &TurnLog{})
	for player, permanent := range targets {
		if permanentByCardID(g, permanent.CardInstanceID) != nil ||
			!g.Players[player].Library.Contains(permanent.CardInstanceID) ||
			!branch.Players[player].Library.Contains(permanent.CardInstanceID) {
			t.Fatal("copied spell did not inherit and execute the chosen bottom-library mode")
		}
	}
}

func TestAuthoredHiveMindPartialRetargetKeepsIllegalTargetAndTargetCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	addCombatPermanent(g, game.Player1, authored.HiveMind.New())
	target := addCombatCreaturePermanent(g, game.Player2)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Two target probe", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 0, MaxTargets: 2, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer},
			},
			Sequence: []game.Instruction{{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(1), Amount: game.Fixed(3)}}},
		}.Ability()),
	}}
	cardID := addCardToHand(g, game.Player2, def)
	setMainPhasePriority(g, game.Player2)
	originalTargets := []game.Target{game.PermanentTarget(target.ObjectID), game.PlayerTarget(game.Player2)}
	if !engine.applyAction(g, game.Player2, action.CastSpell(cardID, originalTargets, 0, nil)) {
		t.Fatal("two-target cast failed")
	}
	movePermanentToZone(g, target, zone.Exile)
	addCombatCreaturePermanent(g, game.Player3)
	policies := simPassPolicies()
	policies[game.Player3] = authoredChoicePolicy{choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		for _, option := range request.Options {
			if len(option.Targets) != 0 && len(option.Targets) != 2 {
				t.Fatal("retarget offered to change the originally chosen target count")
			}
		}
		return chooseAuthoredTargets(t, request, originalTargets[0], game.PlayerTarget(game.Player3))
	}}
	branch := engine.newSearchContext(g, game.Player2).Simulator().ResolvePriority(g, policies)
	engine.runPriorityLoop(g, policies, &TurnLog{})
	if g.Players[game.Player3].Life != 37 || g.Players[game.Player2].Life != 31 ||
		!reflect.DeepEqual(branch.Events, g.Events) {
		t.Fatal("copy could not retain one illegal target while changing and resolving its other target")
	}
}

func TestAuthoredGiftsUngivenOppositionAgentPrivateViewAndPermissions(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	registry := cards.NewDefaultRegistry()
	agent := addCombatPermanent(g, game.Player3, registry.Lookup("Opposition Agent"))
	addCardToHand(g, game.Player1, vanillaCreature("Searcher's private hand", 2, 2))
	addCardToHand(g, game.Player3, vanillaCreature("Controller's private hand", 2, 2))
	var found []game.ObjectID
	for _, name := range []string{"One", "Two", "Three", "Four"} {
		found = append(found, addCardToLibrary(g, game.Player1, vanillaCreature(name, 2, 2)))
	}
	addImplementationSpellToStack(g, game.Player1, authored.GiftsUngiven.New(), []game.Target{game.PlayerTarget(game.Player2)})
	setMainPhasePriority(g, game.Player1)
	policies := simPassPolicies()
	policies[game.Player1] = authoredChoicePolicy{choose: func(PlayerObservation, game.ChoiceRequest) []int {
		t.Fatal("searcher made a decision while Opposition Agent controlled its search")
		return nil
	}}
	policies[game.Player2] = authoredChoicePolicy{choose: func(PlayerObservation, game.ChoiceRequest) []int {
		t.Fatal("targeted opponent was asked to partition cards that were already exiled")
		return nil
	}}
	policies[game.Player3] = authoredChoicePolicy{choose: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Player != game.Player1 || obs.DecisionMaker != game.Player3 ||
			request.Kind != game.ChoiceSearch || obs.Player != game.Player1 ||
			len(obs.Hand()) != 1 || obs.Hand()[0].Name != "Searcher's private hand" ||
			!request.Options[0].Card.Exists || request.Options[0].Card.Val.Name == "" {
			t.Fatal("controlled search did not separate decision maker from the searcher's private information view")
		}
		return []int{request.Options[0].Index}
	}}
	branch := engine.newSearchContext(g, game.Player1).Simulator().ResolvePriority(g, policies)
	log := TurnLog{}
	engine.runPriorityLoop(g, policies, &log)
	if len(log.Choices) != 4 || !reflect.DeepEqual(branch.Events, g.Events) {
		t.Fatal("controlled search choices were not isolated, deterministic and recorded")
	}
	for _, cardID := range found {
		permission, ok := playFromExilePermissionFor(g, cardID)
		if !g.Players[game.Player1].Exile.Contains(cardID) || !ok ||
			permission.Controller != game.Player3 || !permission.SpendAnyMana ||
			!slices.Contains(foreignExileCastableCards(g, game.Player3), cardID) {
			t.Fatal("each found card did not receive its actual exile destination and lasting permission")
		}
	}
	movePermanentToZone(g, agent, zone.Graveyard)
	if !slices.Contains(foreignExileCastableCards(g, game.Player3), found[0]) {
		t.Fatal("Opposition Agent leaving ended a permanent exile permission")
	}
}

func TestAuthoredGiftsUngivenPlacementReplacementsAndSimultaneousShuffle(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"graveyard redirect", "replacement shuffle"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			if scenario == "graveyard redirect" {
				addImplementationSpellToStack(g, game.Player3, cards.NewDefaultRegistry().Lookup("Dauthi Voidwalker"), nil)
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			defs := []*game.CardDef{vanillaCreature("One", 2, 2), vanillaCreature("Two", 2, 2),
				vanillaCreature("Three", 2, 2), vanillaCreature("Four", 2, 2)}
			if scenario == "replacement shuffle" {
				defs[0] = selfLibraryReplacementCardDef()
			}
			ids := make([]game.ObjectID, 4)
			for i, def := range defs {
				ids[i] = addCardToLibrary(g, game.Player1, def)
			}
			addImplementationSpellToStack(g, game.Player1, authored.GiftsUngiven.New(), []game.Target{game.PlayerTarget(game.Player2)})
			setMainPhasePriority(g, game.Player1)
			beforeEpoch := g.Players[game.Player1].Library.ReorderVersion()
			policies := simPassPolicies()
			policies[game.Player1] = authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				for _, cardID := range ids {
					for _, option := range req.Options {
						if option.Card.Exists && option.Card.Val.CardID == cardID {
							return []int{option.Index}
						}
					}
				}
				t.Fatal("search choice did not offer the actual library")
				return nil
			}}
			policies[game.Player2] = authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				return []int{0, 1}
			}}
			branch := engine.newSearchContext(g, game.Player1).Simulator().ResolvePriority(g, policies)
			engine.runPriorityLoop(g, policies, &TurnLog{})
			if !reflect.DeepEqual(branch.Events, g.Events) || !g.Players[game.Player1].Hand.Contains(ids[2]) ||
				!g.Players[game.Player1].Hand.Contains(ids[3]) {
				t.Fatal("placement replacement hid a different found card or corrupted the simultaneous hand partition")
			}
			if scenario == "graveyard redirect" {
				for _, cardID := range ids[:2] {
					if !g.Players[game.Player1].Exile.Contains(cardID) || g.ExileCounterCount(cardID, counter.Void) != 1 {
						t.Fatal("selected graveyard partition bypassed the actual Dauthi replacement")
					}
				}
			} else if !g.Players[game.Player1].Library.Contains(ids[0]) ||
				!g.Players[game.Player1].Graveyard.Contains(ids[1]) ||
				g.Players[game.Player1].Library.ReorderVersion() != beforeEpoch+2 {
				t.Fatal("self-shuffle replacement or final Gifts shuffle did not occur")
			}
		})
	}
}

func TestAuthoredGildedDrakeExchangeUsesLayersAndFrozenAbilityController(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	def := authored.GildedDrake.New()
	drake := addCombatPermanent(g, game.Player1, def)
	target := addCombatCreaturePermanent(g, game.Player2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: drake.ObjectID,
		Layer: game.LayerControl, NewController: opt.Val(game.Player3), Duration: game.DurationPermanent,
	})
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: drake.ObjectID,
		SourceCardID: drake.CardInstanceID, Controller: game.Player1, InlineTrigger: &def.TriggeredAbilities[0],
		Targets: []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if effectiveController(g, drake) != game.Player2 || effectiveController(g, target) != game.Player3 {
		t.Fatal("exchange used owners, base controllers or the ability controller instead of both current layer-2 controllers")
	}
	engine.eliminatePlayer(g, game.Player3)
	if effectiveController(g, target) != game.Player2 || permanentByCardID(g, target.CardInstanceID) == nil {
		t.Fatal("departing recipient did not release the stolen, foreign-owned creature")
	}
}
