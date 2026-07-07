package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// nakiaMonarchTriggerDef builds Nakia, Wakandan Operative's monarch trigger:
// "Whenever your commander enters, you become the monarch." The enters trigger's
// subject is restricted to a commander you control (SubjectSelection.MatchCommander
// under TriggerControllerYou).
func nakiaMonarchTriggerDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Nakia",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventPermanentEnteredBattlefield,
					Controller:       game.TriggerControllerYou,
					SubjectSelection: game.Selection{MatchCommander: true},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.BecomeMonarch{Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

// enterBattlefieldFromHand drives a card from a player's hand onto the
// battlefield through the real engine entry path so it emits a genuine
// EventPermanentEnteredBattlefield event (rather than being appended directly).
func enterBattlefieldFromHand(t *testing.T, engine *Engine, g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, controller, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("entering card instance not found")
	}
	g.Players[controller].Hand.Remove(cardID)
	permanent, ok := createCardPermanentFaceWithOptions(engine, g, card, controller, zone.Hand, game.FaceFront, nil, permanentCreationOptions{}, [game.NumPlayers]PlayerAgent{}, nil)
	if !ok {
		t.Fatal("card did not enter the battlefield")
	}
	return permanent
}

// TestNakiaCommanderEntersBecomeMonarch proves the real enters trigger: when a
// commander you control enters the battlefield, Nakia's trigger fires and you
// become the monarch. The commander enters through the genuine engine entry
// path and the trigger is collected by the ordinary trigger machinery.
func TestNakiaCommanderEntersBecomeMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, nakiaMonarchTriggerDef())

	commander := enterBattlefieldFromHand(t, engine, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Commander",
		Types: []types.Card{types.Creature},
	}})
	g.CommanderIDs[commander.CardInstanceID] = true

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commander-enters trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("controller did not become the monarch when their commander entered")
	}
}

// TestNakiaNonCommanderEnterDoesNotBecomeMonarch proves the MatchCommander
// subject filter: a non-commander permanent you control entering does not make
// you the monarch.
func TestNakiaNonCommanderEnterDoesNotBecomeMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, nakiaMonarchTriggerDef())

	// A creature you control, but not a commander.
	enterBattlefieldFromHand(t, engine, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ordinary Creature",
		Types: []types.Card{types.Creature},
	}})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for a non-commander entering")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("controller became the monarch when a non-commander entered")
	}
}

// TestNakiaOpponentCommanderEnterDoesNotBecomeMonarch proves the
// TriggerControllerYou restriction on "your commander": an opponent's commander
// entering does not make Nakia's controller the monarch.
func TestNakiaOpponentCommanderEnterDoesNotBecomeMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, nakiaMonarchTriggerDef())

	opponentCommander := enterBattlefieldFromHand(t, engine, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Commander",
		Types: []types.Card{types.Creature},
	}})
	g.CommanderIDs[opponentCommander.CardInstanceID] = true

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for an opponent's commander entering")
	}
	if g.Players[game.Player1].IsMonarch {
		t.Fatal("controller became the monarch when an opponent's commander entered")
	}
}

// nakiaCounterAbilityDef builds Nakia's activated ability: "{2}, {T}: Put two
// +1/+1 counters on target creature or Vehicle. Activate only as a sorcery." The
// target is the creature-or-Vehicle union (Selection.AnyOf) and the effect adds
// two +1/+1 counters. The mana cost is omitted so the test exercises targeting
// and resolution without mana setup; the tap cost and sorcery-speed timing are
// kept faithful.
func nakiaCounterAbilityDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Nakia",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "Put two +1/+1 counters on target creature or Vehicle. Activate only as a sorcery.",
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Timing:          game.SorceryOnly,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{AnyOf: []game.Selection{
						{RequiredTypesAny: []types.Card{types.Creature}},
						{SubtypesAny: []types.Sub{types.Vehicle}},
					}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.AddCounter{
						Amount:      game.Fixed(2),
						Object:      game.TargetPermanentReference(0),
						CounterKind: counter.PlusOnePlusOne,
					},
				}},
			}.Ability(),
		}},
	}}
}

// addVehiclePermanent adds an artifact permanent with the Vehicle subtype.
func addVehiclePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Vehicle",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Vehicle},
	}})
}

func setSorcerySpeedPriority(g *game.Game, playerID game.PlayerID) {
	g.Turn.ActivePlayer = playerID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = playerID
}

// nakiaActivateTargetIDs collects, from the real legal-action enumeration, the
// permanent IDs Nakia's counter ability is offered as targets.
func nakiaActivateTargetIDs(engine *Engine, g *game.Game, playerID game.PlayerID, nakiaID id.ID) map[id.ID]bool {
	targets := map[id.ID]bool{}
	for _, act := range engine.legalActions(g, playerID) {
		if act.Kind != action.ActionActivateAbility {
			continue
		}
		payload, ok := act.ActivateAbilityPayload()
		if !ok || payload.SourceID != nakiaID {
			continue
		}
		for _, tgt := range payload.Targets {
			if tgt.Kind == game.TargetPermanent {
				targets[tgt.PermanentID] = true
			}
		}
	}
	return targets
}

// TestNakiaCounterAbilityEnumeratesCreatureAndVehicle proves the real target
// enumeration for the counter ability admits a creature and a Vehicle (the
// creature-or-Vehicle union) while excluding an ineligible permanent.
func TestNakiaCounterAbilityEnumeratesCreatureAndVehicle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	nakia := addCombatPermanent(g, game.Player1, nakiaCounterAbilityDef())
	creature := addCreaturePermanent(g, game.Player1)
	vehicle := addVehiclePermanent(g, game.Player1)
	enchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
	setSorcerySpeedPriority(g, game.Player1)

	targets := nakiaActivateTargetIDs(engine, g, game.Player1, nakia.ObjectID)
	if !targets[creature.ObjectID] {
		t.Fatal("creature was not offered as a target for the counter ability")
	}
	if !targets[vehicle.ObjectID] {
		t.Fatal("Vehicle was not offered as a target for the counter ability")
	}
	if targets[enchantment.ObjectID] {
		t.Fatal("an ineligible enchantment was offered as a target")
	}
}

// TestNakiaCounterAbilityPlacesTwoCountersOnCreature drives the activation and
// resolution through the real engine path and asserts two +1/+1 counters land
// on the target creature.
func TestNakiaCounterAbilityPlacesTwoCountersOnCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	nakia := addCombatPermanent(g, game.Player1, nakiaCounterAbilityDef())
	creature := addCreaturePermanent(g, game.Player1)
	setSorcerySpeedPriority(g, game.Player1)

	act := action.ActivateAbility(nakia.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the counter ability on a creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters on creature = %d, want 2", got)
	}
}

// TestNakiaCounterAbilityPlacesTwoCountersOnVehicle proves the union target's
// Vehicle alternative resolves: two +1/+1 counters land on the target Vehicle.
func TestNakiaCounterAbilityPlacesTwoCountersOnVehicle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	nakia := addCombatPermanent(g, game.Player1, nakiaCounterAbilityDef())
	vehicle := addVehiclePermanent(g, game.Player1)
	setSorcerySpeedPriority(g, game.Player1)

	act := action.ActivateAbility(nakia.ObjectID, 0, []game.Target{game.PermanentTarget(vehicle.ObjectID)}, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("activating the counter ability on a Vehicle failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := vehicle.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters on Vehicle = %d, want 2", got)
	}
}

// TestNakiaCounterAbilitySorcerySpeedOnly proves the sorcery-speed restriction:
// the ability is activatable at sorcery speed but not at instant speed.
func TestNakiaCounterAbilitySorcerySpeedOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	nakia := addCombatPermanent(g, game.Player1, nakiaCounterAbilityDef())
	creature := addCreaturePermanent(g, game.Player1)

	// Instant-speed window (opponent's upkeep): the sorcery-only ability must be
	// rejected.
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	g.Turn.PriorityPlayer = game.Player1
	instantAct := action.ActivateAbility(nakia.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
	if engine.applyAction(g, game.Player1, instantAct) {
		t.Fatal("sorcery-only ability was activated at instant speed")
	}

	// Sorcery speed: the same activation is legal.
	setSorcerySpeedPriority(g, game.Player1)
	sorceryAct := action.ActivateAbility(nakia.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
	if !engine.applyAction(g, game.Player1, sorceryAct) {
		t.Fatal("sorcery-only ability was not activatable at sorcery speed")
	}
}
