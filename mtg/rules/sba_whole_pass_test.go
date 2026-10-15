package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestSBAWholePassSeparatesNewlyIllegalAura(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Initially legal Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&game.TargetSpec{
			Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		})},
	}})
	if !attachPermanent(g, aura, creature) {
		t.Fatal("could not attach the legal Aura")
	}
	creature.MarkedDamage = 2

	NewEngine(nil).applyStateBasedActions(g)

	batches := sbaDepartureBatches(g)
	if batches[creature.ObjectID] == 0 || batches[aura.ObjectID] == 0 || batches[creature.ObjectID] == batches[aura.ObjectID] {
		t.Fatalf("creature/Aura batches = %d/%d, want distinct nonzero passes", batches[creature.ObjectID], batches[aura.ObjectID])
	}
}

func TestSBAWholePassCombinesAlreadyApplicableCategories(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	first.MarkedDamage = 2
	second.MarkedDamage = 3
	aura := addPermanentForSBA(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Already illegal Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
	}})

	NewEngine(nil).applyStateBasedActions(g)

	batches := sbaDepartureBatches(g)
	if batches[first.ObjectID] == 0 || batches[first.ObjectID] != batches[second.ObjectID] || batches[first.ObjectID] != batches[aura.ObjectID] {
		t.Fatalf("independent deaths/Aura batches = %d/%d/%d, want one pass",
			batches[first.ObjectID], batches[second.ObjectID], batches[aura.ObjectID])
	}
}

func TestSBAWholePassShieldDoesNotReplaceRulesDestruction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	first.MarkedDamage, second.MarkedDamage = 2, 2
	first.Counters.Add(counter.Shield, 1)
	NewEngine(nil).applyStateBasedActions(g)
	batches := sbaDepartureBatches(g)
	if batches[first.ObjectID] == 0 || batches[first.ObjectID] != batches[second.ObjectID] {
		t.Fatalf("lethal destruction batches = %d/%d, want one pass; shield only replaces effect destruction",
			batches[first.ObjectID], batches[second.ObjectID])
	}
}

func TestSBAWholePassLegendChoiceSeesDyingMembers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Twins", Types: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
	first := addCombatPermanent(g, game.Player1, def)
	second := addCombatPermanent(g, game.Player1, def)
	first.MarkedDamage = 2
	choices := 0
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Choose one legendary permanent to keep." {
			choices++
			if len(request.Options) != 2 || len(g.Battlefield) != 2 {
				t.Fatal("legend choice did not see the pre-pass battlefield")
			}
			return []int{0}
		}
		return request.DefaultSelection
	}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	batches := sbaDepartureBatches(g)
	if choices != 1 || batches[first.ObjectID] == 0 || batches[first.ObjectID] != batches[second.ObjectID] || len(g.Battlefield) != 0 {
		t.Fatalf("legend/lethal result = choices %d, batches %d/%d, permanents %d",
			choices, batches[first.ObjectID], batches[second.ObjectID], len(g.Battlefield))
	}
}

func TestSBAWholePassAnthemDepartureCreatesLaterCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anthem := addAnthemPermanent(g, game.Player1)
	follower := addCombatCreaturePermanentWithPower(g, game.Player1, 0)
	anthem.MarkedDamage = 2
	NewEngine(nil).applyStateBasedActions(g)
	batches := sbaDepartureBatches(g)
	if batches[anthem.ObjectID] == 0 || batches[follower.ObjectID] == 0 || batches[anthem.ObjectID] == batches[follower.ObjectID] {
		t.Fatal("anthem-dependent zero toughness was applied in the original pass")
	}
	anthemLKI, ok := lastKnownObject(g, anthem.ObjectID)
	if !ok || !anthemLKI.Toughness.Exists || anthemLKI.Toughness.Val != 2 {
		t.Fatalf("anthem pre-pass LKI = %+v", anthemLKI)
	}
	followerLKI, ok := lastKnownObject(g, follower.ObjectID)
	if !ok || !followerLKI.Toughness.Exists || followerLKI.Toughness.Val != 0 {
		t.Fatalf("later follower LKI = %+v", followerLKI)
	}
}

func TestSBAWholePassCounterCancellationPreservesLKIAndActualRetention(t *testing.T) {
	g, engine, skull := skullbriarGame(t)
	skull.Counters.Add(counter.PlusOnePlusOne, 1)
	skull.Counters.Add(counter.MinusOneMinusOne, 1)
	skull.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	snapshot, found := lastKnownObject(g, skull.ObjectID)
	if !found || snapshot.Counters.Get(counter.PlusOnePlusOne) != 1 || snapshot.Counters.Get(counter.MinusOneMinusOne) != 1 {
		t.Fatal("departure LKI did not preserve pre-pass opposing counters")
	}
	retained := g.CardCounters[skull.CardInstanceID]
	if retained.Get(counter.PlusOnePlusOne) != 0 || retained.Get(counter.MinusOneMinusOne) != 0 {
		t.Fatalf("canceled counters traveled with Skullbriar: %v", retained.All())
	}
}

func TestSBAWholePassDuplicateMovementAppliesReplacementOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Twins", Types: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary},
		Power: opt.Val(game.PT{Value: 0}), Toughness: opt.Val(game.PT{Value: 0}),
	}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Twins", Types: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	for _, destination := range []zone.Type{zone.Exile, zone.Library} {
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard,
			ReplaceToZone: destination,
		})
	}
	replacements := 0
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Choose one legendary permanent to keep." {
			return []int{1}
		}
		if request.Kind == game.ChoiceReplacement {
			replacements++
		}
		return request.DefaultSelection
	}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	if replacements != 1 || !g.Players[game.Player1].Exile.Contains(first.CardInstanceID) || !permanentStillPresent(g, second.ObjectID) {
		t.Fatalf("duplicate movement replacement count = %d, want one redirected departure", replacements)
	}
}

func TestSBAWholePassSagaSacrificeHasActualControllerAndIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	saga := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(1)}))
	saga.Controller = game.Player2
	saga.Counters.Add(counter.Lore, 1)
	NewEngine(nil).applyStateBasedActions(g)
	card, _ := g.GetCardInstance(saga.CardInstanceID)
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentSacrificed && event.PermanentID == saga.ObjectID {
			count++
			if event.Controller != game.Player2 || event.Player != game.Player2 || event.CardZoneVersion != card.ZoneVersion ||
				event.ToZone != zone.Graveyard || event.SimultaneousID == 0 {
				t.Fatalf("Saga sacrifice = %+v, want actual actor, arrival and batch", event)
			}
		}
	}
	if count != 1 {
		t.Fatalf("Saga sacrifice events = %d, want one", count)
	}
}

func TestSBAWholePassProhibitedSagaSacrificeDoesNotHideOtherActions(t *testing.T) {
	for _, lethal := range []bool{false, true} {
		t.Run(map[bool]string{false: "protected", true: "independent destruction"}[lethal], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := sagaCard([]game.ChapterAbility{sagaDrawChapter(1)})
			def.Types = append(def.Types, types.Creature)
			def.Power, def.Toughness = opt.Val(game.PT{Value: 2}), opt.Val(game.PT{Value: 2})
			saga := addCombatPermanent(g, game.Player1, def)
			saga.Counters.Add(counter.Lore, 1)
			g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
				Kind: game.RuleEffectCantBeSacrificed, Controller: game.Player1,
				SourceObjectID: saga.ObjectID, AffectedSource: true,
			})
			if lethal {
				saga.MarkedDamage = 2
			}
			NewEngine(nil).applyStateBasedActions(g)
			if permanentStillPresent(g, saga.ObjectID) == lethal {
				t.Fatal("prohibited sacrifice suppressed or invented an independent departure")
			}
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentSacrificed && event.PermanentID == saga.ObjectID {
					t.Fatal("prohibited Saga sacrifice emitted a success fact")
				}
			}
		})
	}
}

func TestSBAWholePassRequiredChoicesAreAPNAPAcrossCategories(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	addLegendaryPermanent(g, game.Player1, "Twins")
	addLegendaryPermanent(g, game.Player1, "Twins")
	commander := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.CommanderIDs[commander.CardInstanceID] = true
	if !movePermanentToZone(g, commander, zone.Graveyard) {
		t.Fatal("could not prepare the commander graveyard arrival")
	}
	var choosers []game.PlayerID
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Choose one legendary permanent to keep." || request.Player == game.Player2 {
			choosers = append(choosers, request.Player)
		}
		return request.DefaultSelection
	}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: policy, game.Player2: policy}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.applyStateBasedActions(g)
	if !reflect.DeepEqual(choosers, []game.PlayerID{game.Player1, game.Player2}) {
		t.Fatalf("cross-category chooser order = %v, want Player1 then Player2", choosers)
	}
}

func TestSBAWholePassStabilizesBeforePriorityAndCleanup(t *testing.T) {
	for _, cleanup := range []bool{false, true} {
		t.Run(map[bool]string{false: "priority", true: "cleanup"}[cleanup], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			power := 2
			if cleanup {
				power = 0
			}
			host := addCombatCreaturePermanentWithPower(g, game.Player1, power)
			if cleanup {
				host.TemporaryToughnessModifier = 2
			} else {
				host.MarkedDamage = 2
			}
			aura := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Attached Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
				StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&game.TargetSpec{
					Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})},
			}})
			if !attachPermanent(g, aura, host) {
				t.Fatal("could not attach Aura")
			}
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event: game.EventPermanentDied, Controller: game.TriggerControllerYou, OneOrMore: true,
			}, []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}, nil)
			policies := simPassPolicies()
			if cleanup {
				engine := NewEngine(nil)
				engine.setReplacementChoiceContext(g, policies, &TurnLog{})
				engine.runEndingPhaseWithLog(g, policies, &TurnLog{})
				g.ClearChoiceContext()
				if g.Players[game.Player1].Life != 42 {
					t.Fatalf("cleanup one-or-more trigger life = %d, want 42", g.Players[game.Player1].Life)
				}
				return
			}
			setMainPhasePriority(g, game.Player1)
			first := NewEngine(nil).Simulator().ResolvePriority(g, policies)
			second := NewEngine(nil).Simulator().ResolvePriority(g, policies)
			if first.Players[game.Player1].Life != 42 || second.Players[game.Player1].Life != 42 || g.Players[game.Player1].Life != 40 {
				t.Fatal("separate SBA passes did not resolve twice without mutating the original")
			}
			if !reflect.DeepEqual(first.Events, second.Events) {
				t.Fatal("equal Simulator choices did not reproduce SBA events")
			}
			snapshot := first.LastKnownInformation[host.ObjectID]
			snapshot.Counters.Add(counter.Charge, 3)
			first.LastKnownInformation[host.ObjectID] = snapshot
			other := second.LastKnownInformation[host.ObjectID]
			if other.Counters.Get(counter.Charge) != 0 {
				t.Fatal("SBA LKI leaked between clones")
			}
		})
	}
}

func sbaDepartureBatches(g *game.Game) map[id.ID]id.ID {
	batches := make(map[id.ID]id.ID)
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield && event.ToZone == zone.Graveyard {
			batches[event.PermanentID] = event.SimultaneousID
		}
	}
	return batches
}
