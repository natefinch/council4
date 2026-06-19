package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

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

func TestTemporaryControlledPermanentKeywordGrantRulesBehavior(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	affected := addCombatCreaturePermanent(g, game.Player1)
	unrelated := addCombatCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.BattlefieldGroup(game.Selection{
				Controller: game.ControllerYou,
			}),
			AddKeywords: []game.Keyword{game.Hexproof, game.Indestructible},
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !hasKeyword(g, affected, game.Hexproof) || !hasKeyword(g, affected, game.Indestructible) {
		t.Fatal("controlled permanent did not gain both temporary keywords")
	}
	if hasKeyword(g, unrelated, game.Hexproof) || hasKeyword(g, unrelated, game.Indestructible) {
		t.Fatal("opponent permanent gained controlled-permanent keywords")
	}
	later := addCombatCreaturePermanent(g, game.Player1)
	if hasKeyword(g, later, game.Hexproof) || hasKeyword(g, later, game.Indestructible) {
		t.Fatal("later entrant incorrectly joined the resolution snapshot")
	}

	opponentSpell := addCardToHand(g, game.Player2, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player2
	if engine.canCastSpell(g, game.Player2, opponentSpell, []game.Target{game.PermanentTarget(affected.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target permanent with temporary hexproof")
	}

	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Object: game.TargetPermanentReference(0),
	}, []game.Target{game.PermanentTarget(affected.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, affected.ObjectID); !ok {
		t.Fatal("temporary indestructible did not prevent destruction")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if hasKeyword(g, affected, game.Hexproof) || hasKeyword(g, affected, game.Indestructible) {
		t.Fatal("temporary keywords did not expire during cleanup")
	}
	g.Turn.PriorityPlayer = game.Player2
	if !engine.canCastSpell(g, game.Player2, opponentSpell, []game.Target{game.PermanentTarget(affected.ObjectID)}, 0, nil) {
		t.Fatal("expired hexproof still prevented opponent targeting")
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
