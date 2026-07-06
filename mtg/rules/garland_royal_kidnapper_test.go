package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// garlandGainControlEffect mirrors the continuous effect the cardgen lowering
// emits for Garland, Royal Kidnapper's second ability ("Whenever an opponent
// becomes the monarch, gain control of target creature that player controls for
// as long as they're the monarch."): a layer-2 control change whose expiry binds
// to the triggering event's player through ExpiresForRef=EventPlayerReference.
func garlandGainControlEffect() game.ContinuousEffect {
	return game.ContinuousEffect{
		Layer:         game.LayerControl,
		NewController: opt.Val(game.Player1),
		ExpiresForRef: opt.Val(game.EventPlayerReference()),
	}
}

// garlandBecameMonarchTrigger builds the resolving triggered-ability stack
// object for Garland's second ability, controlled by controller and capturing
// the become-monarch event for eventPlayer.
func garlandBecameMonarchTrigger(g *game.Game, controller, eventPlayer game.PlayerID, source, target *game.Permanent) *game.StackObject {
	return &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      controller,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:       game.EventBecameMonarch,
			Controller: eventPlayer,
			Player:     eventPlayer,
		},
		Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
	}
}

// TestGarlandGainsControlWhileOpponentHoldsCrownAndRevertsWhenLost proves
// Garland's second ability: when an opponent becomes the monarch, Garland's
// controller gains control of a creature that opponent controls, keeps it while
// that opponent stays the monarch, and loses it when the crown moves on.
func TestGarlandGainsControlWhileOpponentHoldsCrownAndRevertsWhenLost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	stolen := makeCreaturePermanent(g, game.Player2, "Kidnapped Beast")

	// Player2, an opponent, takes the crown.
	if !setMonarch(g, game.Player2) {
		t.Fatal("Player2 could not become the monarch")
	}

	obj := garlandBecameMonarchTrigger(g, game.Player1, game.Player2, source, stolen)
	if !applyTypedContinuousEffects(g, obj, stolen, []game.ContinuousEffect{garlandGainControlEffect()}, game.DurationForAsLongAsPlayerIsMonarch) {
		t.Fatal("applyTypedContinuousEffects returned false for the gain-control effect")
	}

	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller after gaining control = %v, want Player1", got)
	}

	// The monarch's player rides ExpiresFor, resolved from the trigger event.
	var found bool
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer != game.LayerControl {
			continue
		}
		found = true
		if effect.ExpiresForRef.Exists {
			t.Fatal("ExpiresForRef should be resolved to a concrete ExpiresFor at application time")
		}
		if effect.ExpiresFor != game.Player2 {
			t.Fatalf("ExpiresFor = %v, want the monarch Player2", effect.ExpiresFor)
		}
	}
	if !found {
		t.Fatal("no layer-2 control effect was created")
	}

	// While Player2 keeps the crown the control effect persists.
	expireConditionalControlDurations(g)
	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller while Player2 still monarch = %v, want Player1", got)
	}

	// A third player takes the crown: Player2 is no longer the monarch, so the
	// duration ends and control reverts to the creature's owner.
	if !setMonarch(g, game.Player3) {
		t.Fatal("Player3 could not become the monarch")
	}
	expireConditionalControlDurations(g)
	if got := effectiveController(g, stolen); got != game.Player2 {
		t.Fatalf("controller after Player2 lost the crown = %v, want reverted to Player2", got)
	}
}

// garlandGainControlSourceDef mirrors the cardgen lowering of Garland, Royal
// Kidnapper's second ability as a standalone source permanent so tests can drive
// the real trigger -> target -> resolve pipeline: a "whenever an opponent becomes
// the monarch" triggered ability whose target is a creature that player controls
// and whose effect gains control of it for as long as they're the monarch.
func garlandGainControlSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Garland, Royal Kidnapper",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 4}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventBecameMonarch,
					Player: game.TriggerPlayerOpponent,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature that player controls",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ControlledByEventPlayer: true}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
							ExpiresForRef: opt.Val(game.EventPlayerReference()),
						}},
						Duration: game.DurationForAsLongAsPlayerIsMonarch,
					},
				}},
			}.Ability(),
		}},
	}}
}

// TestGarlandGainControlEndToEndStealsMonarchsCreatureAndReverts drives the real
// engine pipeline: with Garland in play, an opponent becoming the monarch fires
// Garland's second ability, targets a creature that opponent controls, gains
// control of it, and reverts control when that opponent loses the crown. This
// exercises the production trigger -> target-selection -> resolve path (no
// hand-built stack object), so it fails if target selection cannot resolve the
// triggering event player.
func TestGarlandGainControlEndToEndStealsMonarchsCreatureAndReverts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, garlandGainControlSourceDef())
	stolen := makeCreaturePermanent(g, game.Player2, "Kidnapped Beast")

	if !setMonarch(g, game.Player2) {
		t.Fatal("Player2 could not become the monarch")
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Garland's become-monarch ability did not fire or found no legal target")
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no triggered ability reached the stack")
	}
	if len(top.Targets) != 1 || top.Targets[0].Kind != game.TargetPermanent || top.Targets[0].PermanentID != stolen.ObjectID {
		t.Fatalf("triggered ability targets = %+v, want the monarch's creature %v", top.Targets, stolen.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller after resolving gain-control = %v, want Garland's controller Player1", got)
	}

	// While Player2 keeps the crown the control effect persists.
	engine.applyStateBasedActions(g)
	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller while Player2 still monarch = %v, want Player1", got)
	}

	// A different player takes the crown: the duration ends and control reverts.
	if !setMonarch(g, game.Player3) {
		t.Fatal("Player3 could not become the monarch")
	}
	engine.applyStateBasedActions(g)
	if got := effectiveController(g, stolen); got != game.Player2 {
		t.Fatalf("controller after Player2 lost the crown = %v, want reverted to owner Player2", got)
	}
}

// TestGarlandGainControlTargetRestrictedToPlayerWhoBecameMonarch proves the
// second ability's target restriction ("target creature that player controls")
// through the real engine: only a creature controlled by the opponent who became
// the monarch is a legal target. It runs two production scenarios — a mixed board
// where a bystander's creature must be skipped in favour of the monarch's, and a
// board where the only creature belongs to a non-monarch, which must yield no
// legal target so the ability is removed from the stack (CR 603.3d).
func TestGarlandGainControlTargetRestrictedToPlayerWhoBecameMonarch(t *testing.T) {
	// Mixed board: the bystander's creature is enumerated first, so a broken
	// restriction would steal it; the fixed restriction targets only the monarch's.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, garlandGainControlSourceDef())
	bystanderCreature := makeCreaturePermanent(g, game.Player3, "Bystander's Creature")
	monarchsCreature := makeCreaturePermanent(g, game.Player2, "Monarch's Creature")

	if !setMonarch(g, game.Player2) {
		t.Fatal("Player2 could not become the monarch")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Garland's become-monarch ability did not fire or found no legal target")
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no triggered ability reached the stack")
	}
	if len(top.Targets) != 1 || top.Targets[0].PermanentID != monarchsCreature.ObjectID {
		t.Fatalf("triggered ability targets = %+v, want only the monarch's creature %v", top.Targets, monarchsCreature.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, monarchsCreature); got != game.Player1 {
		t.Fatalf("monarch's creature controller = %v, want stolen by Player1", got)
	}
	if got := effectiveController(g, bystanderCreature); got != game.Player3 {
		t.Fatalf("bystander's creature controller = %v, want untouched Player3", got)
	}

	// Only a non-monarch's creature exists: the ability has a target but no legal
	// target, so it is removed and gains control of nothing.
	g2 := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine2 := NewEngine(nil)
	addCombatPermanent(g2, game.Player1, garlandGainControlSourceDef())
	onlyBystander := makeCreaturePermanent(g2, game.Player3, "Lone Bystander")
	if !setMonarch(g2, game.Player2) {
		t.Fatal("Player2 could not become the monarch")
	}
	if engine2.putTriggeredAbilitiesOnStack(g2) {
		t.Fatal("ability was placed on the stack despite no creature controlled by the monarch")
	}
	if got := effectiveController(g2, onlyBystander); got != game.Player3 {
		t.Fatalf("a non-monarch's creature controller = %v, want unchanged Player3", got)
	}
}

// garlandControlNotOwnStatic mirrors the lowered shape of Garland's third
// ability ("Creatures you control but don't own get +2/+2 and can't be
// sacrificed."): a +2/+2 anthem scoped to the controller's creatures they don't
// own, plus a can't-be-sacrificed rule effect over the same group.
func garlandControlNotOwnStatic(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Garland, Royal Kidnapper",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerPowerToughnessModify,
				Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, OwnerNotController: true}),
				PowerDelta:     2,
				ToughnessDelta: 2,
			}},
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeSacrificed,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection:  game.Selection{OwnerNotController: true},
			}},
		}},
	}})
}

// controlNotOwnCreature adds a creature owned by owner but controlled by
// controller, base 2/2.
func controlNotOwnCreature(g *game.Game, owner, controller game.PlayerID) *game.Permanent {
	permanent := makeCreaturePermanent(g, owner, "Borrowed Creature")
	permanent.Controller = controller
	return permanent
}

// TestGarlandControlNotOwnStaticBuffsAndProtectsOnlyForeignCreatures proves the
// third ability applies its +2/+2 and can't-be-sacrificed shield only to
// creatures the controller controls but does not own; a creature they both own
// and control is untouched.
func TestGarlandControlNotOwnStaticBuffsAndProtectsOnlyForeignCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	garlandControlNotOwnStatic(g, game.Player1)

	foreign := controlNotOwnCreature(g, game.Player2, game.Player1)
	ownAndControl := makeCreaturePermanent(g, game.Player1, "Homegrown Creature")

	if got := effectivePower(g, foreign); got != 4 {
		t.Fatalf("control-but-don't-own creature power = %d, want base 2 + anthem 2 = 4", got)
	}
	if got, ok := effectiveToughness(g, foreign); !ok || got != 4 {
		t.Fatalf("control-but-don't-own creature toughness = %d (ok=%v), want base 2 + anthem 2 = 4", got, ok)
	}
	if !permanentCantBeSacrificed(g, foreign) {
		t.Fatal("control-but-don't-own creature should not be sacrificeable")
	}

	if got := effectivePower(g, ownAndControl); got != 2 {
		t.Fatalf("own-and-control creature power = %d, want unmodified base 2", got)
	}
	if got, ok := effectiveToughness(g, ownAndControl); !ok || got != 2 {
		t.Fatalf("own-and-control creature toughness = %d (ok=%v), want unmodified base 2", got, ok)
	}
	if permanentCantBeSacrificed(g, ownAndControl) {
		t.Fatal("a creature the controller both owns and controls must remain sacrificeable")
	}
}
