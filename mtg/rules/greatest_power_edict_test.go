package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Crackling Doom (KTK a66e5673-e34b-46e8-a0e4-55f3ee20f99a) and
// Shadowgrange Archfiend (SOI de4e445f-4d54-44a8-a5f5-ef02b3d57251) Oracle text
// from the pinned corpus; https://api.scryfall.com/cards/<id>/rulings returned
// four 2014-09-20 Crackling Doom rulings (no targets; sacrifice independent of
// damage; APNAP choices then one simultaneous sacrifice; a player chooses among
// their own tied creatures; later players know earlier choices) and the
// 2021-11-19 Shadowgrange tie ruling. CR 101.4/101.4b (APNAP choices visible to
// later players), 107.1b (negative power compares as itself), 608.2h
// (last-known power) and 701.21a (sacrifice) apply.
// These fixtures prove the shared greatest-power edict; neither card is admitted.

var edictCreatureSelection = game.Selection{RequiredTypes: []types.Card{types.Creature}}

func greatestPowerEdict() game.SacrificePermanents {
	return game.SacrificePermanents{
		PlayerGroup: game.OpponentsReference(),
		Amount:      game.Fixed(1),
		Selection:   edictCreatureSelection,
		Extremum:    game.PermanentChoiceGreatestPower,
	}
}

func cracklingDoomFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Crackling Doom",
		ManaCost: opt.Val(cost.Mana{cost.R, cost.W, cost.B}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Damage{
				Amount:    game.Fixed(2),
				Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
			}},
			{Primitive: greatestPowerEdict()},
		}}.Ability()),
	}}
}

type edictChoiceRecorder struct {
	simPassPolicy

	g        *game.Game
	choice   []int
	requests []game.ChoiceRequest
	// mustRemain is checked at choice time so a later chooser proves earlier
	// choices were not yet carried out.
	mustRemain id.ID
	t          *testing.T
}

func (a *edictChoiceRecorder) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if a.mustRemain != 0 && !permanentStillPresent(a.g, a.mustRemain) {
		a.t.Fatal("an earlier edict choice was sacrificed before every player chose")
	}
	return append([]int(nil), a.choice...)
}

func castCracklingDoom(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, cracklingDoomFixture())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Crackling Doom cast failed")
	}
}

func sacrificeBatchIDs(g *game.Game, permanents ...*game.Permanent) []id.ID {
	var ids []id.ID
	for _, permanent := range permanents {
		for _, event := range g.Events {
			if event.Kind == game.EventPermanentSacrificed && event.PermanentID == permanent.ObjectID {
				ids = append(ids, event.SimultaneousID)
			}
		}
	}
	return ids
}

func TestCracklingDoomSacrificesEachOpponentsOwnGreatestPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	casterCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	tieA := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	tieB := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	smallP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	bigP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 6)
	tieP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 6)
	smallP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)

	castCracklingDoom(t, g, engine)
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{1}}
	p3 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}, mustRemain: tieB.ObjectID}
	agents := simPassPolicies()
	agents[game.Player2], agents[game.Player3] = p2, p3
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if g.Players[player].Life != 38 {
			t.Fatalf("%v life = %d, want 38", player, g.Players[player].Life)
		}
	}
	if g.Players[game.Player1].Life != 40 || !permanentStillPresent(g, casterCreature.ObjectID) {
		t.Fatal("Crackling Doom affected its caster")
	}
	if len(p2.requests) != 1 || len(p2.requests[0].Options) != 2 {
		t.Fatalf("Player2 choices = %#v, want one choice between the two tied creatures", p2.requests)
	}
	if len(p3.requests) != 1 || len(p3.requests[0].Options) != 2 {
		t.Fatalf("Player3 choices = %#v, want one choice between their own tied creatures", p3.requests)
	}
	if !permanentStillPresent(g, tieA.ObjectID) || permanentStillPresent(g, tieB.ObjectID) ||
		!permanentStillPresent(g, smallP2.ObjectID) {
		t.Fatal("Player2 did not sacrifice exactly the chosen tied greatest-power creature")
	}
	if permanentStillPresent(g, bigP3.ObjectID) || !permanentStillPresent(g, tieP3.ObjectID) ||
		!permanentStillPresent(g, smallP3.ObjectID) {
		t.Fatal("Player3 did not sacrifice exactly their chosen greatest-power creature")
	}
	batches := sacrificeBatchIDs(g, tieB, bigP3)
	if len(batches) != 2 || batches[0] == 0 || batches[0] != batches[1] {
		t.Fatalf("sacrifice batches = %v, want one simultaneous batch", batches)
	}
}

func TestCracklingDoomFollowsEffectivePowerAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	printedBest := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	pumped := addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	castCracklingDoom(t, g, engine)
	// The pump arrives after casting, so only resolution-time power decides.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: pumped.ObjectID,
		Layer: game.LayerPowerToughnessModify, PowerDelta: 2,
		Duration: game.DurationUntilEndOfTurn, CreatedTurn: g.Turn.TurnNumber,
	})
	p2 := &edictChoiceRecorder{g: g, t: t}
	agents := simPassPolicies()
	agents[game.Player2] = p2
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentStillPresent(g, pumped.ObjectID) || !permanentStillPresent(g, printedBest.ObjectID) {
		t.Fatal("edict ignored the pump that made a different creature greatest")
	}
	if len(p2.requests) != 0 {
		t.Fatal("a single greatest-power creature required a choice")
	}
}

func TestGreatestPowerEdictProtectedGreatestSacrificesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	protected := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	lesser := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCantBeSacrificed, Controller: game.Player2,
		AffectedObjectID: protected.ObjectID,
	})
	edict := greatestPowerEdict()
	edict.PlayerGroup = game.PlayerGroupReference{}
	edict.Player = game.TargetPlayerReference(0)
	edict.Fallback = game.SacrificeFallback{Kind: game.SacrificeFallbackLoseLife, Amount: game.Fixed(3)}
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceCardID: g.IDGen.Next(),
		Targets: []game.Target{game.PlayerTarget(game.Player2)}, TargetCounts: []int{1},
	}
	content := game.Mode{
		Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
		Sequence: []game.Instruction{{Primitive: edict}},
	}.Ability()
	engine.resolveAbilityContentWithChoices(g, obj, content, simPassPolicies(), &TurnLog{})

	if !permanentStillPresent(g, protected.ObjectID) || !permanentStillPresent(g, lesser.ObjectID) {
		t.Fatal("an unsacrificeable greatest creature let a lesser creature be sacrificed")
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("Player2 life = %d, want the who-can't fallback to apply", g.Players[game.Player2].Life)
	}
}

func TestShadowgrangeArchfiendGainsGreatestLastKnownSacrificedPower(t *testing.T) {
	reference := game.ProgramReference{ID: "test/shadowgrange-archfiend-enters", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			sacrificed, err := ctx.Execute(greatestPowerEdict())
			if err != nil {
				return err
			}
			greatest := 0
			for _, object := range sacrificed.Objects {
				if object.Power.Exists && object.Power.Val > greatest {
					greatest = object.Power.Val
				}
			}
			if greatest == 0 {
				return nil
			}
			_, err = ctx.Execute(game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(greatest)})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	archfiend := &game.CardDef{CardFace: game.CardFace{
		Name:      "Shadowgrange Archfiend",
		ManaCost:  opt.Val(cost.Mana{cost.O(6), cost.B}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Demon},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 4}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Program: reference}.Ability(),
		}},
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, catalog)
	p2Best := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	p2Lesser := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	p3Best := addCombatCreaturePermanentWithPower(g, game.Player3, 4)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: p2Best.ObjectID,
		Layer: game.LayerPowerToughnessModify, PowerDelta: 3,
		Duration: game.DurationUntilEndOfTurn, CreatedTurn: g.Turn.TurnNumber,
	})
	cardID := addCardToHand(g, game.Player1, archfiend)
	g.Players[game.Player1].ManaPool.Add(mana.B, 7)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Shadowgrange Archfiend cast failed")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Shadowgrange Archfiend's enters trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	if permanentStillPresent(g, p2Best.ObjectID) || !permanentStillPresent(g, p2Lesser.ObjectID) ||
		permanentStillPresent(g, p3Best.ObjectID) {
		t.Fatal("each opponent did not sacrifice their own greatest-power creature")
	}
	if g.Players[game.Player1].Life != 46 {
		t.Fatalf("controller life = %d, want 46 from the pumped creature's last-known power", g.Players[game.Player1].Life)
	}
}

func modifyPowerUntilEndOfTurn(g *game.Game, permanent *game.Permanent, delta int) {
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: permanent.ObjectID,
		Layer: game.LayerPowerToughnessModify, PowerDelta: delta,
		Duration: game.DurationUntilEndOfTurn, CreatedTurn: g.Turn.TurnNumber,
	})
}

func TestCracklingDoomComparesSignedNegativePower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Player2: a -1/3 and a 0/2; only the 0/2 has the greatest power.
	negativeP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	zeroP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	modifyPowerUntilEndOfTurn(g, negativeP2, -4)
	modifyPowerUntilEndOfTurn(g, zeroP2, -2)
	// Player3: a -2/1 and a -1/1; every power is negative, -1 is still greatest.
	lowerP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	higherP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	modifyPowerUntilEndOfTurn(g, lowerP3, -3)
	modifyPowerUntilEndOfTurn(g, higherP3, -2)

	castCracklingDoom(t, g, engine)
	p2 := &edictChoiceRecorder{g: g, t: t}
	p3 := &edictChoiceRecorder{g: g, t: t}
	agents := simPassPolicies()
	agents[game.Player2], agents[game.Player3] = p2, p3
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(p2.requests) != 0 || len(p3.requests) != 0 {
		t.Fatal("a single signed greatest-power creature was offered alongside a lesser one")
	}
	if !permanentStillPresent(g, negativeP2.ObjectID) || permanentStillPresent(g, zeroP2.ObjectID) {
		t.Fatal("Player2 did not sacrifice the 0-power creature over the -1-power one")
	}
	if !permanentStillPresent(g, lowerP3.ObjectID) || permanentStillPresent(g, higherP3.ObjectID) {
		t.Fatal("Player3 did not sacrifice the -1-power creature over the -2-power one")
	}
}

func TestCracklingDoomLaterPlayersSeeEarlierSacrificeChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	chosenP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	forcedP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	chosenP4 := addCombatCreaturePermanentWithPower(g, game.Player4, 5)
	addCombatCreaturePermanentWithPower(g, game.Player4, 5)

	castCracklingDoom(t, g, engine)
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{1}}
	p4 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}}
	agents := simPassPolicies()
	agents[game.Player2], agents[game.Player4] = p2, p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(p2.requests) != 1 || len(p2.requests[0].SacrificeDecisions) != 0 {
		t.Fatalf("Player2 requests = %#v, want one request with no earlier decisions", p2.requests)
	}
	if len(p4.requests) != 1 {
		t.Fatalf("Player4 requests = %d, want 1", len(p4.requests))
	}
	got := p4.requests[0].SacrificeDecisions
	want := []game.SacrificeDecision{
		{Player: game.Player2, Permanents: []id.ID{chosenP2.ObjectID}},
		{Player: game.Player3, Permanents: []id.ID{forcedP3.ObjectID}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Player4 saw earlier decisions %#v, want %#v", got, want)
	}
	if permanentStillPresent(g, chosenP2.ObjectID) || permanentStillPresent(g, forcedP3.ObjectID) ||
		permanentStillPresent(g, chosenP4.ObjectID) {
		t.Fatal("the chosen greatest-power creatures were not sacrificed")
	}
}
