package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestShieldCounterPreventsDamageBeforeMutationAndEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == target.ObjectID &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestShieldCounterReplacesDestroyBeforeZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	removed, ok := destroyPermanent(g, target.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want nil, false for replaced destroy", removed, ok)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("shield-replaced permanent left the battlefield")
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	if g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("shield-replaced permanent moved to graveyard")
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID &&
			event.FromZone == zone.Battlefield &&
			event.ToZone == zone.Graveyard
	})
	assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestPreventedCombatDamageDoesNotGrantLifelinkOrMarkDeathtouch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Deathtouch, game.Lifelink)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.Counters.Add(counter.Shield, 1)
	g.Players[game.Player1].Life = 40
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("lifelink controller life = %d, want 40", g.Players[game.Player1].Life)
	}
	if blocker.MarkedDamage != 0 || blocker.MarkedDeathtouchDamage {
		t.Fatalf("blocker damage = %d deathtouch = %v, want no marked damage", blocker.MarkedDamage, blocker.MarkedDeathtouchDamage)
	}
}

func TestProtectionFromColorPreventsDamageAndTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 2, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if protected.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", protected.MarkedDamage)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == protected.ObjectID &&
			event.Amount == 2
	})

	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Red Strike",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Red},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability())},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil) {
		t.Fatal("red spell could target a permanent with protection from red")
	}
}

func TestProtectionFromEverythingPreventsDamageAndTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Use a red source — protection from everything should block any color.
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	protected := addProtectionFromEverythingPermanent(g, game.Player2)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 3, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from everything)", dealt)
	}

	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil) {
		t.Fatal("spell could target permanent with protection from everything")
	}
}

func TestProtectionFromTypesPreventsDamageFromCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Source is a creature permanent.
	pt := game.PT{Value: 2}
	sourceDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
	sourcePerm := addCombatPermanent(g, game.Player1, sourceDef)
	protected := addProtectionFromTypesPermanent(g, game.Player2, types.Creature)

	// Use the permanent's ObjectID as source.
	dealt := dealPermanentDamage(g, 0, sourcePerm.ObjectID, game.Player1, protected, 2, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from creatures)", dealt)
	}
}

func TestProtectionFromTypesAllowsDamageFromNonCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Source is an instant spell (not a creature).
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	protected := addProtectionFromTypesPermanent(g, game.Player2, types.Creature)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 2, false)
	if dealt != 2 {
		t.Fatalf("dealt = %d, want 2 (instant is not a creature)", dealt)
	}
}

func TestProtectionFromSubtypesPreventsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	dragonDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Dragon",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
	dragonPerm := addCombatPermanent(g, game.Player1, dragonDef)
	protected := addProtectionFromSubtypesPermanent(g, game.Player2, types.Dragon)

	dealt := dealPermanentDamage(g, 0, dragonPerm.ObjectID, game.Player1, protected, 2, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from Dragons)", dealt)
	}
}

func TestProtectionFromMulticoloredPreventsDamageFromMulticoloredSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Two-color source card.
	multicolorID := g.IDGen.Next()
	g.CardInstances[multicolorID] = &game.CardInstance{
		ID: multicolorID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:   "Multicolor Source",
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.Red, color.Green},
		}},
		Owner: game.Player1,
	}
	pt := game.PT{Value: 2}
	protected := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Multicolored",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromMulticoloredStaticAbility()},
	}})

	dealt := dealPermanentDamage(g, multicolorID, 0, game.Player1, protected, 2, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from multicolored)", dealt)
	}

	// Single-color source should get through.
	monoID := addColoredSourceCard(g, game.Player1, color.Red)
	dealt2 := dealPermanentDamage(g, monoID, 0, game.Player1, protected, 2, false)
	if dealt2 != 2 {
		t.Fatalf("dealt = %d from mono source, want 2 (not multicolored)", dealt2)
	}
}

func TestProtectionFromMonocoloredPreventsDamageFromMonocoloredSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	monoID := addColoredSourceCard(g, game.Player1, color.Blue)
	pt := game.PT{Value: 2}
	protected := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Monocolored",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromMonocoloredStaticAbility()},
	}})

	dealt := dealPermanentDamage(g, monoID, 0, game.Player1, protected, 2, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from monocolored)", dealt)
	}

	// Two-color source should get through.
	multiID := g.IDGen.Next()
	g.CardInstances[multiID] = &game.CardInstance{
		ID: multiID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:   "Multicolor Source",
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.Red, color.Blue},
		}},
		Owner: game.Player1,
	}
	dealt2 := dealPermanentDamage(g, multiID, 0, game.Player1, protected, 2, false)
	if dealt2 != 2 {
		t.Fatalf("dealt = %d from 2-color source, want 2 (not monocolored)", dealt2)
	}
}

func TestProtectionFromEachColorPreventsDamageFromColoredSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, color.White)
	pt := game.PT{Value: 2}
	protected := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Each Color",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromEachColorStaticAbility()},
	}})

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 2, false)
	if dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (protection from each color)", dealt)
	}
}

func TestDamageAddendReplacementIncreasesMatchingSourceDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	replacementSource := addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	redSourceID := addColoredSourceCard(g, game.Player1, color.Red)
	blueSourceID := addColoredSourceCard(g, game.Player1, color.Blue)

	if dealt := dealPlayerDamage(g, redSourceID, 0, game.Player1, game.Player2, 2, false); dealt != 3 {
		t.Fatalf("red source damage = %d, want 3", dealt)
	}
	if dealt := dealPlayerDamage(g, blueSourceID, 0, game.Player1, game.Player2, 2, false); dealt != 2 {
		t.Fatalf("blue source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, redSourceID, 0, game.Player2, game.Player1, 2, false); dealt != 2 {
		t.Fatalf("opponent-controlled red source damage = %d, want 2", dealt)
	}
	if dealt := dealPlayerDamage(g, replacementSource.CardInstanceID, replacementSource.ObjectID, game.Player1, game.Player2, 2, false); dealt != 2 {
		t.Fatalf("replacement source's own red damage = %d, want 2", dealt)
	}
}

func TestDamageMultiplierReplacementDoublesPermanentDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Green)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 2, false)
	if dealt != 4 {
		t.Fatalf("damage dealt = %d, want 4", dealt)
	}
	if target.MarkedDamage != 4 {
		t.Fatalf("marked damage = %d, want 4", target.MarkedDamage)
	}
}

func TestDamageReplacementEffectsUseFallbackOrderingWhenNoChoiceIsAvailable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 2, false)
	if dealt != 6 {
		t.Fatalf("stacked damage replacements dealt = %d, want deterministic add-then-double result 6", dealt)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player2 {
		t.Fatalf("replacement decision player = %v, want damaged player", got)
	}
	decision := g.ReplacementDecisions[0]
	if !decision.UsedFallback || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want deterministic fallback order", decision)
	}
}

func TestSelectedReplacementEffectUsesRecordedOrder(t *testing.T) {
	first := game.ReplacementEffect{ID: 1, Description: "first"}
	second := game.ReplacementEffect{ID: 2, Description: "second"}

	selected := selectedReplacementEffect([]game.ReplacementEffect{first, second}, game.ReplacementDecision{Selected: []int{1, 0}})

	if selected.ID != second.ID {
		t.Fatalf("selected replacement = %+v, want second replacement", selected)
	}
}

func TestDamageReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	damageSourceID := addColoredSourceCard(g, game.Player1, color.Red)

	dealt := dealPlayerDamage(g, damageSourceID, 0, game.Player1, game.Player2, 2, false)
	if dealt != 2 {
		t.Fatalf("damage after replacement source leaves = %d, want 2", dealt)
	}
}

func TestDamageReplacementAppliesAfterPrevention(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	g.PreventionShields = append(g.PreventionShields, game.PreventionShield{
		ID:         g.IDGen.Next(),
		Controller: game.Player2,
		Player:     game.Player2,
		Amount:     1,
		Duration:   game.DurationUntilEndOfTurn,
	})

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 3, false)
	if dealt != 4 {
		t.Fatalf("damage after prevention then replacement = %d, want 4", dealt)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 1
	})
}

func TestHexproofPreventsOpponentTargetsButAllowsControllerTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hexproof := addHexproofPermanent(g, game.Player2)
	opponentSpell := addCardToHand(g, game.Player1, targetCreatureInstant())
	controllerSpell := addCardToHand(g, game.Player2, targetCreatureInstant())

	g.Turn.PriorityPlayer = game.Player1
	if engine.canCastSpell(g, game.Player1, opponentSpell, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target hexproof permanent")
	}

	g.Turn.PriorityPlayer = game.Player2
	if !engine.canCastSpell(g, game.Player2, controllerSpell, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil) {
		t.Fatal("controller spell could not target own hexproof permanent")
	}
}

func TestShroudPreventsAllTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shroud := addShroudPermanent(g, game.Player2)
	opponentSpell := addCardToHand(g, game.Player1, targetCreatureInstant())
	controllerSpell := addCardToHand(g, game.Player2, targetCreatureInstant())

	g.Turn.PriorityPlayer = game.Player1
	if engine.canCastSpell(g, game.Player1, opponentSpell, []game.Target{game.PermanentTarget(shroud.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target shroud permanent")
	}

	g.Turn.PriorityPlayer = game.Player2
	if engine.canCastSpell(g, game.Player2, controllerSpell, []game.Target{game.PermanentTarget(shroud.ObjectID)}, 0, nil) {
		t.Fatal("controller spell could target own shroud permanent")
	}
}

func TestShroudGrantedByContinuousEffectPreventsTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Shroud Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Shroud},
			}},
		}},
	}})
	spellID := addCardToHand(g, game.Player2, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player2

	if engine.canCastSpell(g, game.Player2, spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil) {
		t.Fatal("controller spell could target permanent granted shroud")
	}
}

func TestLegalActionsOmitOpponentHexproofTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hexproof := addHexproofPermanent(g, game.Player2)
	targetable := addCombatCreaturePermanent(g, game.Player3)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if actionsContain(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil)) {
		t.Fatalf("legal actions include hexproof target: %+v", legal)
	}
	if !actionsContain(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(targetable.ObjectID)}, 0, nil)) {
		t.Fatalf("legal actions omit non-hexproof target: %+v", legal)
	}
}

func TestHexproofCounterPreventsOpponentTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	target.Counters.Add(counter.Hexproof, 1)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target permanent with hexproof counter")
	}
}

// TestProtectionTargetingUsesEffectivePermanentSourceCharacteristics verifies
// that a permanent whose color is granted by a continuous effect (not the base
// CardDef) is treated as having that color when checking protection on targets
// of its abilities.
func TestProtectionTargetingUsesEffectivePermanentSourceCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Source: colorless artifact creature (no colors in base CardDef).
	pt := game.PT{Value: 2}
	sourceDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Colorless Artifacter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
	source := addCombatPermanent(g, game.Player1, sourceDef)

	// Target creature has protection from red.
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	// Before applying any color effect, the source is colorless — protection
	// from red should not apply.
	if permanentProtectedFromPermanentEffective(g, protected, source) {
		t.Fatal("colorless source should not trigger protection from red before color effect")
	}
	if targetProtectedFromSource(g, game.Player1, sourceDef, source.ObjectID, game.PermanentTarget(protected.ObjectID)) {
		t.Fatal("colorless source should not make protection block targeting before color effect")
	}

	// Apply a continuous effect that makes the source red.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerColor,
		AddColors:        []color.Color{color.Red},
		Duration:         game.DurationPermanent,
	})

	// Now protection from red must apply via effective characteristics.
	if !permanentProtectedFromPermanentEffective(g, protected, source) {
		t.Fatal("red source (via continuous effect) should trigger protection from red")
	}
	if !targetProtectedFromSource(g, game.Player1, sourceDef, source.ObjectID, game.PermanentTarget(protected.ObjectID)) {
		t.Fatal("effectively-red source should make protection block targeting")
	}
}

// TestProtectionSourceUsesSelectedStackFace verifies that when an adventure
// (alternate-face) spell is on the stack, protection checks use the selected
// face's characteristics rather than the root card def.
func TestProtectionSourceUsesSelectedStackFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Build an adventure card: front face is a red creature, alternate face is
	// a blue instant.
	redFace := game.CardFace{
		Name:   "Red Creature",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
		Power:  opt.Val(game.PT{Value: 2}),
	}
	blueFace := game.CardFace{
		Name:   "Blue Blast",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
	}
	adventureDef := &game.CardDef{
		CardFace:  redFace,
		Alternate: opt.Val(blueFace),
	}
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   adventureDef,
		Owner: game.Player1,
	}

	// Permanents with protection from each color.
	protFromRed := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	protFromBlue := addProtectionFromColorPermanent(g, game.Player2, color.Blue)

	// Push the card as its alternate (blue) face onto the stack.
	stackObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Face:       game.FaceAlternate,
		Controller: game.Player1,
	}
	g.Stack.Push(stackObj)

	// Protection from blue should apply when the blue face is on the stack.
	if !permanentProtectedFromSource(g, protFromBlue, 0, stackObj.ID) {
		t.Fatal("protection from blue should apply against blue adventure face on the stack")
	}
	// Protection from red should NOT apply (the spell face is blue, not red).
	if permanentProtectedFromSource(g, protFromRed, 0, stackObj.ID) {
		t.Fatal("protection from red should not apply against blue adventure face on the stack")
	}

	// Also verify the sourceID-only path: selectedFaceForCardInstance finds the
	// stack object and uses the alternate (blue) face.
	if !permanentProtectedFromSource(g, protFromBlue, cardID, 0) {
		t.Fatal("protection from blue should apply using sourceID with blue face on stack")
	}
	if permanentProtectedFromSource(g, protFromRed, cardID, 0) {
		t.Fatal("protection from red should not apply using sourceID with blue face on stack")
	}
}

// TestProtectionAppliesDuringAlternateFaceSpellResolution verifies that after a
// StackSpell is popped and effects resolve, protection checks against the
// resolving spell still use the selected (alternate) face's characteristics
// rather than falling back to the front face.
func TestProtectionAppliesDuringAlternateFaceSpellResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Build an adventure card: front face is a red creature, alternate face is
	// a blue instant that deals 3 damage.
	blueDamageSpell := game.CardFace{
		Name:   "Blue Blast",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability()),
	}
	redFront := game.CardFace{
		Name:   "Red Creature",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
		Power:  opt.Val(game.PT{Value: 2}),
	}
	adventureDef := &game.CardDef{
		CardFace:  redFront,
		Alternate: opt.Val(blueDamageSpell),
	}
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   adventureDef,
		Owner: game.Player1,
	}

	// Target: a creature with protection from blue.
	protFromBlue := addProtectionFromColorPermanent(g, game.Player2, color.Blue)
	// Also a creature with protection from red (should NOT prevent damage from the blue face).
	protFromRed := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	// Cast the alternate (blue) face as a spell targeting the blue-protected creature.
	stackID := g.IDGen.Next()
	g.Stack.Push(&game.StackObject{
		ID:         stackID,
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Face:       game.FaceAlternate,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(protFromBlue.ObjectID)},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	// Protection from blue MUST prevent damage even though the spell was popped.
	if protFromBlue.MarkedDamage != 0 {
		t.Fatalf("blue-protected creature took %d damage, want 0 (blue face protection not applied after pop)", protFromBlue.MarkedDamage)
	}
	// Protection from red must NOT affect the blue-face spell.
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Face:       game.FaceAlternate,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(protFromRed.ObjectID)},
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if protFromRed.MarkedDamage == 0 {
		t.Fatal("red-protected creature incorrectly had damage prevented against blue-face spell")
	}
}

// TestProtectionChecksLKIForDepartedSourcePermanent verifies that if the damage
// source was a permanent that has since left the battlefield, protection checks
// use the departed permanent's last-known characteristics.
func TestProtectionChecksLKIForDepartedSourcePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	// A red permanent that will be moved off the battlefield before damage
	// processing. We capture its LKI manually (as zones.go would in normal play).
	sourcePermanent := addMulticoloredSourcePermanent(g, game.Player1, color.Red)
	snapshot := snapshotPermanent(g, sourcePermanent, 0)
	rememberLastKnown(g, &snapshot)
	sourceObjID := sourcePermanent.ObjectID

	// Remove it from the battlefield to simulate departure.
	g.Battlefield = slices.DeleteFunc(g.Battlefield, func(p *game.Permanent) bool {
		return p.ObjectID == sourcePermanent.ObjectID
	})

	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	// Damage event with departed source (objectID present but not on battlefield).
	result := permanentProtectedFromSource(g, protected, sourcePermanent.CardInstanceID, sourceObjID)
	if !result {
		t.Fatal("protection from red should apply against departed red source via LKI")
	}

	// Also verify a non-red protection does not match.
	notProtected := addProtectionFromColorPermanent(g, game.Player2, color.Blue)
	result = permanentProtectedFromSource(g, notProtected, sourcePermanent.CardInstanceID, sourceObjID)
	if result {
		t.Fatal("protection from blue should not apply against departed red source via LKI")
	}
}

// TestProtectionConsultedForDepartedTriggeredAbilitySource exercises the full
// resolveInstruction path: a triggered ability's damage must be blocked by
// protection even when the source permanent has left the battlefield, using LKI
// to identify its effective characteristics.
func TestProtectionConsultedForDepartedTriggeredAbilitySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Red creature whose triggered ability will deal damage.
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Vandal",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})

	protFromRed := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	notProtected := addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(protFromRed.ObjectID)},
	}

	// Move the source off the battlefield — LKI is stored by movePermanentToZone.
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("source should move to graveyard")
	}

	resolveInstruction(engine, g, obj,
		game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}, &TurnLog{})

	if protFromRed.MarkedDamage != 0 {
		t.Fatalf("protection from red should block damage from departed red source via LKI, got %d damage", protFromRed.MarkedDamage)
	}

	// Confirm a creature without protection takes damage (sanity check).
	obj2 := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(notProtected.ObjectID)},
	}
	resolveInstruction(engine, g, obj2,
		game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}, &TurnLog{})
	if notProtected.MarkedDamage != 3 {
		t.Fatalf("unprotected creature should take 3 damage, got %d", notProtected.MarkedDamage)
	}
}

// TestProtectionConsultedForDepartedTokenAbilitySource verifies that protection
// checks consult LKI for a token source that has left the battlefield. Token
// permanents have no SourceCardID; the source object ID is the sole identity.
func TestProtectionConsultedForDepartedTokenAbilitySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Manually create a red token permanent.
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Dragon Token",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}}
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Token:      true,
		TokenDef:   tokenDef,
		Owner:      game.Player1,
		Controller: game.Player1,
	}
	g.Battlefield = append(g.Battlefield, token)

	protFromRed := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	obj := &game.StackObject{
		Kind:     game.StackTriggeredAbility,
		SourceID: token.ObjectID,
		// SourceCardID intentionally 0: tokens have no card instance.
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(protFromRed.ObjectID)},
	}

	// Token departs — LKI records its effective characteristics.
	if !movePermanentToZone(g, token, zone.Graveyard) {
		t.Fatal("token should move to graveyard")
	}

	resolveInstruction(engine, g, obj,
		game.Damage{Amount: game.Fixed(5), Recipient: game.AnyTargetDamageRecipient(0)}, &TurnLog{})

	if protFromRed.MarkedDamage != 0 {
		t.Fatalf("protection from red should block token damage via LKI, got %d damage", protFromRed.MarkedDamage)
	}
}

// TestRemoveKeywordsSuppressesProtectionSemantics verifies that a continuous
// effect removing Protection (e.g., "loses all abilities") prevents the
// protection ability body from blocking damage/targeting.
func TestRemoveKeywordsSuppressesProtectionSemantics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	// Without the removal effect, protection prevents damage.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 3, false); dealt != 0 {
		t.Fatalf("before removal: dealt = %d, want 0", dealt)
	}
	protected.MarkedDamage = 0
	g.Events = nil

	// Apply a RemoveKeywords effect that strips Protection.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: protected.ObjectID,
		Layer:            game.LayerAbility,
		RemoveKeywords:   []game.Keyword{game.Protection},
	})

	// Now damage should go through.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 3, false); dealt != 3 {
		t.Fatalf("after removal: dealt = %d, want 3 (Protection removed but body still blocked damage)", dealt)
	}
}

func TestPreventionShieldPreventsTrackedAmountAndExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(2), Object: game.TargetPermanentReference(0)}, nil)
	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 5, false)

	if dealt != 3 {
		t.Fatalf("dealt damage = %d, want 3 after prevention shield", dealt)
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID && event.Amount == 2
	})

	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields after cleanup = %+v, want expired", g.PreventionShields)
	}
}

func TestMultiplePreventionShieldsRecordDeterministicReplacementOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)
	resolveInstruction(engine, g, obj, game.PreventDamage{Amount: game.Fixed(1), Object: game.TargetPermanentReference(0)}, nil)

	dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one deterministic prevention order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player2 || !decision.UsedFallback || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want Player2 fallback order [0 1]", decision)
	}
}

func TestRegenerationReplacesDestroyAndRemovesFromCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.MarkedDamage = 2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: []id.ID{blocker.ObjectID},
		},
	}

	resolveInstruction(engine, g, &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(blocker.ObjectID)},
	}, game.Regenerate{Object: game.TargetPermanentReference(0)}, nil)
	removed, ok := destroyPermanent(g, blocker.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want regenerated replacement", removed, ok)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
		t.Fatal("regenerated blocker left battlefield")
	}
	if !blocker.Tapped || blocker.MarkedDamage != 0 || blocker.RegenerationShields != 0 {
		t.Fatalf("regenerated blocker tapped=%v damage=%d shields=%d, want tapped, no damage, no shields", blocker.Tapped, blocker.MarkedDamage, blocker.RegenerationShields)
	}
	if len(g.Combat.Blockers) != 0 || len(g.Combat.BlockerOrder[attacker.ObjectID]) != 0 {
		t.Fatalf("combat after regeneration blockers=%+v order=%+v, want blocker removed", g.Combat.Blockers, g.Combat.BlockerOrder)
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.Event) bool {
		return event.PermanentID == blocker.ObjectID
	})
}

func TestRegenerationShieldExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.RegenerationShields = 1

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.RegenerationShields != 0 {
		t.Fatalf("regeneration shields = %d, want cleanup expiry", creature.RegenerationShields)
	}
}

func TestShieldAndRegenerationReplacementOrderIsRecorded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.Shield, 1)
	creature.RegenerationShields = 1

	destroyPermanent(g, creature.ObjectID)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one shield/regeneration order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || !decision.UsedFallback || len(decision.Selected) != 2 {
		t.Fatalf("replacement decision = %+v, want Player1 fallback order", decision)
	}
	if creature.Counters.Get(counter.Shield) != 0 || creature.RegenerationShields != 1 {
		t.Fatalf("shield counters=%d regeneration=%d, want shield used before regeneration", creature.Counters.Get(counter.Shield), creature.RegenerationShields)
	}
}

func TestRegenerationReplacesLethalDamageSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2
	creature.RegenerationShields = 1

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("deaths = %+v, want regeneration to replace lethal-damage destruction", deaths)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok || !creature.Tapped || creature.MarkedDamage != 0 {
		t.Fatalf("creature after regeneration = %+v, want tapped on battlefield with no damage", creature)
	}
}

func TestPermanentEntersTappedAndWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{Name: "Tapped Walker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("Tapped Walker enters tapped."),
			game.EntersWithCountersReplacement("Tapped Walker enters with two +1/+1 counters.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
		}},
	}

	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)

	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want enters tapped", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
}

func TestEntersTappedUnlessPaidPaysLifeByDefault(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if permanent.Tapped {
		t.Fatalf("permanent = %+v, want untapped after paying life", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("life = %d, want 37", got)
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one ETB payment choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.Kind != game.ChoiceMay || choice.Request.Prompt != "Pay 3 life?" || len(choice.Selected) != 1 || choice.Selected[0] != 1 || !choice.UsedFallback {
		t.Fatalf("choice = %+v, want fallback yes for ETB payment", choice)
	}
}

func TestEntersTappedUnlessPaidDeclinedEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, agents, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped after declining payment", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want 40", got)
	}
	if len(log.Choices) != 1 || len(log.Choices[0].Selected) != 1 || log.Choices[0].Selected[0] != 0 || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want explicit no", log.Choices)
	}
}

func TestEntersTappedUnlessPaidCannotPayEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 2
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped when payment is not payable", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 2 {
		t.Fatalf("life = %d, want 2", got)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt for unpayable ETB payment", log.Choices)
	}
}

func TestEntersTappedUnlessRevealMatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player2)
	forestID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	cardID := addCardToHand(g, game.Player2, revealETBLand())
	engine := NewEngine(nil)

	if !engine.applyPlayLand(g, game.Player2, cardID) {
		t.Fatal("applyPlayLand() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if permanent.Tapped {
		t.Fatalf("permanent = %+v, want untapped after revealing Forest", permanent)
	}
	if !g.Players[game.Player2].Hand.Contains(forestID) {
		t.Fatal("revealed Forest left its owner's hand")
	}
	if !eventRevealedCardFromZone(g, game.Player2, cardID, forestID, zone.Hand) {
		t.Fatal("revealing Forest did not emit a reveal event")
	}
}

func eventRevealedCardFromZone(g *game.Game, player game.PlayerID, sourceID, cardID id.ID, from zone.Type) bool {
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed &&
			event.Controller == player &&
			event.Player == player &&
			event.SourceID == sourceID &&
			event.CardID == cardID &&
			event.FromZone == from {
			return true
		}
	}
	return false
}

func TestEntersTappedUnlessRevealRejectsNonmatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	cardID := addCardToHand(g, game.Player1, revealETBLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceFront, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped without a matching card", permanent)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt when reveal cost is unpayable", log.Choices)
	}
}

func revealETBLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Reveal Land",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedUnlessPaidReplacement(
				"As this land enters, you may reveal a Forest or Mountain card from your hand. If you don't, this land enters tapped.",
				game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{{
						Kind:        cost.AdditionalReveal,
						SubtypesAny: cost.SubtypeSet{types.Forest, types.Mountain},
						Source:      zone.Hand,
					}},
				},
			),
		},
	}}
}

func TestGenericReplacementChangesZoneDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not redirect away from graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not move card to exile")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID && event.ToZone == zone.Exile
	})
}

func TestStaticSelfZoneReplacementMovesPermanentToLibrary(t *testing.T) {
	g := game.NewGameWithRand([game.NumPlayers]game.PlayerConfig{}, rand.New(rand.NewPCG(1, 2)))
	bottomID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	target := addCombatPermanent(g, game.Player1, selfLibraryReplacementCardDef())

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("self replacement did not redirect away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(target.CardInstanceID) {
		t.Fatal("self replacement did not move card to library")
	}
	if top, ok := g.Players[game.Player1].Library.Top(); !ok || top != bottomID {
		t.Fatalf("library top = %v, %v; want existing card on top after deterministic shuffle", top, ok)
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID &&
			event.FromZone == zone.Battlefield &&
			event.ToZone == zone.Library
	})
}

func TestStaticSelfZoneReplacementDoesNotApplyFaceDownPermanentAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addFaceDownPermanent(g, game.Player1, selfLibraryReplacementCardDef(), game.FaceDownMorph)

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player1].Library.Contains(target.CardInstanceID) {
		t.Fatal("face-down permanent used its hidden self zone replacement")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("face-down permanent did not move to graveyard")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == target.ObjectID &&
			event.ToZone == zone.Graveyard
	})
}

func TestStaticSelfZoneReplacementAppliesWhenDiscardedFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, selfLibraryReplacementCardDef())

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("self replacement did not redirect discarded card away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("self replacement did not move discarded card to library")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == cardID && event.Player == game.Player1
	})
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Library
	})
}

func TestStaticSelfZoneReplacementAppliesToGenericZoneMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, selfLibraryReplacementCardDef())

	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("self replacement did not redirect generic zone move away from graveyard")
	}
	if !g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("self replacement did not move generic zone move to library")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == cardID && event.Player == game.Player1
	})
}

func TestTokenCreationReplacementDoublesTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 2, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 4 {
		t.Fatalf("created tokens = %d, want 4", got)
	}
}

func TestTokenCreationReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("created tokens after source leaves = %d, want 1", got)
	}
}

func TestTokenCreationReplacementStacksAndRecordsOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 4 {
		t.Fatalf("created tokens = %d, want 4", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player1 {
		t.Fatalf("replacement decision player = %v, want Player1", got)
	}
}

func TestTokenCreationReplacementDoesNotAffectOpponentTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, token, 1, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("opponent-created tokens = %d, want 1", got)
	}
}

func TestTokenCreationReplacementUsesCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
		Duration:         game.DurationPermanent,
	})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player1) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("old controller tokens = %d, want 1", got)
	}
	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, token, 1, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player2) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 3 {
		t.Fatalf("tokens after new controller creates one = %d, want 3", got)
	}
}

func TestCounterPlacementReplacementDoublesSpecificCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4", got)
	}
	if !addCountersToPermanent(g, creature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanent(stun) = false, want true")
	}
	if got := creature.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("stun counters = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementDoublesETBCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Entering Creature",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement(
				"Entering Creature enters with a +1/+1 counter on it.",
				game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
			),
		},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("entering card instance missing")
	}

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent() = false, want true")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("ETB +1/+1 counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesAllCounterKinds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanent(stun) = false, want true")
	}
	if got := creature.Counters.Get(counter.Stun); got != 2 {
		t.Fatalf("stun counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementUsesPlacingController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})
	controllerCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controller Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanentControlledBy(g, game.Player1, opponentCreature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanentControlledBy(Player1) = false, want true")
	}
	if got := opponentCreature.Counters.Get(counter.Stun); got != 2 {
		t.Fatalf("opponent creature stun counters = %d, want 2", got)
	}
	if !addCountersToPermanentControlledBy(g, game.Player2, controllerCreature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanentControlledBy(Player2) = false, want true")
	}
	if got := controllerCreature.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("controller creature stun counters from opponent = %d, want 1", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesProliferatedCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})
	opponentCreature.Counters.Add(counter.Stun, 1)
	g.Players[game.Player2].PoisonCounters = 1

	if !addProliferatedCounter(g, game.Player1, proliferateTarget{
		permanentID: opponentCreature.ObjectID,
		counters:    []counter.Kind{counter.Stun},
	}, counter.Stun) {
		t.Fatal("addProliferatedCounter(permanent) = false, want true")
	}
	if got := opponentCreature.Counters.Get(counter.Stun); got != 3 {
		t.Fatalf("proliferated stun counters = %d, want 3", got)
	}
	if !addProliferatedCounter(g, game.Player1, proliferateTarget{
		player:   game.Player2,
		counters: []counter.Kind{counter.Poison},
	}, counter.Poison) {
		t.Fatal("addProliferatedCounter(player) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 3 {
		t.Fatalf("proliferated poison counters = %d, want 3", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesToxicCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Toxic Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			{KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: 1}}},
		},
	}})

	markPlayerCombatDamage(g, source, game.Player2, 1, &TurnLog{})
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("toxic poison counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesWitherDamageCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Wither)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 7)

	dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, target, 3, false)
	if dealt != 3 {
		t.Fatalf("damage dealt = %d, want 3", dealt)
	}
	if got := target.Counters.Get(counter.MinusOneMinusOne); got != 6 {
		t.Fatalf("-1/-1 counters = %d, want 6", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesPlayerCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())

	if !addCountersToPlayerControlledBy(g, game.Player1, g.Players[game.Player2], counter.Poison, 1) {
		t.Fatal("addCountersToPlayerControlledBy(Player1) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("poison counters = %d, want 2", got)
	}
}

func TestCounterPlacementReplacementOnlyMatchesCreatureRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Noncreature Artifact",
		Types: []types.Card{types.Artifact},
	}})

	if !addCountersToPermanent(g, artifact, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(noncreature) = false, want true")
	}
	if got := artifact.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("noncreature +1/+1 counters = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementUsesRecipientController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controller Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanentControlledBy(g, game.Player2, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanentControlledBy(opponent) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters from opponent effect = %d, want 2", got)
	}
}

func TestCounterPlacementReplacementStacksAndRecordsOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player1 {
		t.Fatalf("replacement decision player = %v, want Player1", got)
	}
}

func TestCounterPlacementReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after source leaves = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementUsesCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
		Duration:         game.DurationPermanent,
	})
	oldControllerCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Old Creature",
		Types: []types.Card{types.Creature},
	}})
	newControllerCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "New Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, oldControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(old controller) = false, want true")
	}
	if got := oldControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("old controller +1/+1 counters = %d, want 1", got)
	}
	if !addCountersToPermanent(g, newControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(new controller) = false, want true")
	}
	if got := newControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("new controller +1/+1 counters = %d, want 2", got)
	}
}

func TestReplacementRegistrationSkipsETBReplacementEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Tapped Bear",
		Types:                []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedReplacement("This creature enters tapped.")},
	}}

	permanent := addReplacementPermanent(t, g, game.Player1, def)
	if !permanent.Tapped {
		t.Fatal("ETB replacement did not tap entering permanent")
	}
	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("registered replacement effects = %d, want 0", len(g.ReplacementEffects))
	}
}

func TestReplacementRegistrationSkipsSelfZoneReplacementEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, selfLibraryReplacementCardDef())
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other Creature",
		Types: []types.Card{types.Creature},
	}})

	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("registered replacement effects = %d, want 0", len(g.ReplacementEffects))
	}
	if !movePermanentToZone(g, other, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(other.CardInstanceID) {
		t.Fatal("other permanent was not put into graveyard")
	}
	if g.Players[game.Player1].Library.Contains(other.CardInstanceID) {
		t.Fatal("self-zone replacement affected a different permanent")
	}
}

func selfLibraryReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Darksteel Colossus",
		Types: []types.Card{types.Artifact, types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{{
			Text: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
			Replacement: game.ReplacementEffect{
				MatchEvent:         game.EventZoneChanged,
				MatchToZone:        true,
				ToZone:             zone.Graveyard,
				ReplaceToZone:      zone.Library,
				ShuffleIntoLibrary: true,
				RevealSource:       true,
				Duration:           game.DurationPermanent,
			},
		}},
	}}
}

func tokenDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Anointed Procession",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacement(
				"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
				2,
				game.TriggerControllerYou,
			),
		},
	}}
}

func damageAddendReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:   "Embermaw Hellion",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamageReplacementExcludingSource(
				"If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
				0,
				1,
				[]color.Color{color.Red},
				game.TriggerControllerYou,
			),
		},
	}}
}

func damageMultiplierReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Angrath's Marauders",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamageReplacement(
				"If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
				2,
				0,
				nil,
				game.TriggerControllerYou,
			),
		},
	}}
}

func counterDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Branching Evolution",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.CounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
				2,
				counter.PlusOnePlusOne,
				game.TriggerControllerYou,
			),
		},
	}}
}

func anyCounterDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Vorinclex",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.AnyCounterPlacementReplacement(
				"If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
				2,
				game.TriggerControllerYou,
			),
		},
	}}
}

func addReplacementPermanent(t *testing.T, g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], controller, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent() = false, want true")
	}
	return permanent
}

func countTokenPermanentsNamed(g *game.Game, name string) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == name {
			count++
		}
	}
	return count
}

func payLifeETBModalLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Front Spell // Pay Life Land",

		Types: []types.Card{types.Sorcery}}, Layout: game.LayoutModalDFC,

		Back: opt.Val(game.CardFace{
			Name:  "Pay Life Land",
			Types: []types.Card{types.Land},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 3 life. If you don't, it enters tapped.", game.ResolutionPayment{
					Prompt: "Pay 3 life?",
					AdditionalCosts: []cost.Additional{
						{Kind: cost.AdditionalPayLife, Amount: 3, Text: "Pay 3 life"},
					},
				}),
			},
		}),
	}
}

func TestGenericETBReplacementAppliesTappedAndCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:        "enter modified",
			MatchEvent:         game.EventPermanentEnteredBattlefield,
			MatchToZone:        true,
			ToZone:             zone.Battlefield,
			EntersTapped:       true,
			EntersWithCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}},
		},
	}, nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Entering Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped by replacement", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
}

func TestMultipleGenericReplacementsRecordOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, replacement := range []game.ReplacementEffect{
		{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
		{
			Description:   "hand instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Hand,
		},
	} {
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{
			Replacement: &replacement,
		}, nil)
	}

	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one order decision", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want deterministic Player1 order", decision)
	}
	if !g.Players[game.Player1].Hand.Contains(target.CardInstanceID) {
		t.Fatal("second replacement in fallback order should move card to hand")
	}
}

func TestPermanentSourceReplacementStopsAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Replacement Source",
		Types: []types.Card{types.Enchantment}},
	})
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	resolveInstruction(engine, g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}, game.CreateReplacement{
		Replacement: &game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("source should leave battlefield")
	}
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("target should move to graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement from departed source should not apply")
	}
}

func TestSkipStepEffectSkipsNextDrawStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Would Draw"}})
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.SkipStep{
		Player: game.ControllerReference(),
		Step:   game.StepDraw,
	}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want skipped draw step", got)
	}
	if g.Players[game.Player1].Library.Size() != 1 {
		t.Fatalf("library size = %d, want card not drawn", g.Players[game.Player1].Library.Size())
	}
}

func addColoredSourceCard(g *game.Game, owner game.PlayerID, sourceColor color.Color) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Colored Source",
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{sourceColor}},
		},
		Owner: owner,
	}
	return cardID
}

func addProtectionFromColorPermanent(g *game.Game, controller game.PlayerID, protectedColor color.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Protected Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.ProtectionKeyword{FromColors: []color.Color{protectedColor}}},
		}}},
	})
}

func addHexproofPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Hexproof Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.HexproofStaticBody}},
	})
}

func addShroudPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Shroud Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ShroudStaticBody}},
	})
}

func addProtectionFromEverythingPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Everything",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromEverythingStaticAbility()},
	}})
}

func addProtectionFromTypesPermanent(g *game.Game, controller game.PlayerID, cardTypes ...types.Card) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Types",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromTypesStaticAbility(cardTypes...)},
	}})
}

func addProtectionFromSubtypesPermanent(g *game.Game, controller game.PlayerID, subtypes ...types.Sub) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected from Subtypes",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromSubtypesStaticAbility(subtypes...)},
	}})
}

func addMulticoloredSourcePermanent(g *game.Game, controller game.PlayerID, colors ...color.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Multicolored Creature",
		Types:     []types.Card{types.Creature},
		Colors:    colors,
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
}

func targetCreatureInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Target Creature Instant",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
		}.Ability())},
	}
}
