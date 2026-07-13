package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// commandersPlateEquipment mirrors the static abilities the executable backend
// generates for Commander's Plate: a LayerPowerToughnessModify +3/+3 effect and
// a LayerAbility grant carrying the commander-identity-complement protection
// marker, both scoped to the attached creature. The rules resolve the marker to
// the concrete complement of the controller's commander color identity.
func commandersPlateEquipment(g *game.Game, controller game.PlayerID) *game.Permanent {
	protection := game.ProtectionFromNonCommanderIdentityColorsStaticAbility()
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Commander's Plate",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerPowerToughnessModify,
					Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
					PowerDelta:     3,
					ToughnessDelta: 3,
				},
				{
					Layer:        game.LayerAbility,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddAbilities: []game.Ability{&protection},
				},
			},
		}},
	}})
}

// setPlayerCommander associates a modeled commander with the given color
// identity to the player, mirroring NewGame's command-zone setup so the
// identity-complement rewrite has a commander to read.
func setPlayerCommander(t *testing.T, g *game.Game, playerID game.PlayerID, identity color.Identity) {
	t.Helper()
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:          cardID,
		Def:         &game.CardDef{CardFace: game.CardFace{Name: "Commander"}, ColorIdentity: identity},
		Owner:       playerID,
		ZoneVersion: 0,
	}
	g.CommanderIDs[cardID] = true
	player, ok := playerByID(g, playerID)
	if !ok {
		t.Fatalf("player %v not found", playerID)
	}
	player.CommanderInstanceID = cardID
}

// addPartnerCommander adds an additional commander (Partner) to a player without
// replacing their primary CommanderInstanceID, exercising the identity union.
func addPartnerCommander(g *game.Game, playerID game.PlayerID, identity color.Identity) {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Partner Commander"}, ColorIdentity: identity},
		Owner: playerID,
	}
	g.CommanderIDs[cardID] = true
}

func plateBear(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

func attachPlate(creature, equipment *game.Permanent) {
	creature.Attachments = append(creature.Attachments, equipment.ObjectID)
	equipment.AttachedTo = opt.Val(creature.ObjectID)
}

func allFiveColors() []color.Color {
	return []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}
}

// TestCommanderIdentityProtectionColorlessProtectsAllFive verifies that a
// colorless commander (empty color identity) yields protection from all five
// colors on the equipped creature.
func TestCommanderIdentityProtectionColorlessProtectsAllFive(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity())
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	for _, c := range allFiveColors() {
		if !permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("colorless commander: creature lacks protection from %v", c)
		}
	}
}

// TestCommanderIdentityProtectionMonoProtectsOtherFour verifies that a
// mono-red commander yields protection from the other four colors but not red.
func TestCommanderIdentityProtectionMonoProtectsOtherFour(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(color.Red))
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	if permanentHasGrantedProtectionFromColor(g, creature, color.Red) {
		t.Fatal("mono-red commander: creature should not have protection from red")
	}
	for _, c := range []color.Color{color.White, color.Blue, color.Black, color.Green} {
		if !permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("mono-red commander: creature lacks protection from %v", c)
		}
	}
}

// TestCommanderIdentityProtectionPartnerUnion verifies that two commanders
// (Partner) union their color identities: a white primary and a blue partner
// leave the equipped creature protected from black, red, and green only.
func TestCommanderIdentityProtectionPartnerUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(color.White))
	addPartnerCommander(g, game.Player1, color.NewIdentity(color.Blue))
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	for _, c := range []color.Color{color.White, color.Blue} {
		if permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("partner union: creature should not have protection from identity color %v", c)
		}
	}
	for _, c := range []color.Color{color.Black, color.Red, color.Green} {
		if !permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("partner union: creature lacks protection from %v", c)
		}
	}
}

// TestCommanderIdentityProtectionFailsClosedWithoutCommander verifies that when
// the controller has no modeled commander the grant fails closed: the equipped
// creature gains protection from no color.
func TestCommanderIdentityProtectionFailsClosedWithoutCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	for _, c := range allFiveColors() {
		if permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("no commander: creature gained protection from %v, want fail closed", c)
		}
	}
}

// TestCommanderIdentityProtectionFiveColorProtectsNone verifies that a
// five-color commander leaves an empty complement, so no color is protected and
// the empty grant does not panic.
func TestCommanderIdentityProtectionFiveColorProtectsNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(allFiveColors()...))
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	for _, c := range allFiveColors() {
		if permanentHasGrantedProtectionFromColor(g, creature, c) {
			t.Fatalf("five-color commander: creature gained protection from %v, want none", c)
		}
	}
}

// TestCommanderIdentityProtectionRecomputesOnControlChange verifies that the
// dynamic complement is recomputed from the new controller's commander identity
// after the Equipment changes control: player 1 has a red commander, player 2 a
// blue one, and the resolved protection follows the controller.
func TestCommanderIdentityProtectionRecomputesOnControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(color.Red))
	setPlayerCommander(t, g, game.Player2, color.NewIdentity(color.Blue))
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(creature, equipment)

	if permanentHasGrantedProtectionFromColor(g, creature, color.Red) {
		t.Fatal("player 1 control: should not protect from red (in identity)")
	}
	if !permanentHasGrantedProtectionFromColor(g, creature, color.Blue) {
		t.Fatal("player 1 control: should protect from blue (not in identity)")
	}

	// Control of the Equipment (and creature) moves to player 2.
	equipment.Controller = game.Player2
	creature.Controller = game.Player2

	if permanentHasGrantedProtectionFromColor(g, creature, color.Blue) {
		t.Fatal("player 2 control: should not protect from blue (in identity)")
	}
	if !permanentHasGrantedProtectionFromColor(g, creature, color.Red) {
		t.Fatal("player 2 control: should protect from red (not in identity)")
	}
}

// TestCommanderIdentityPlateBuffsAndProtectsOnlyWhileAttached verifies the
// +3/+3 and protection apply only while the Plate is attached and only to the
// equipped creature, and stop once the source leaves the battlefield.
func TestCommanderIdentityPlateBuffsAndProtectsOnlyWhileAttached(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(color.Red))
	creature := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("power before attach = %d, want 2", got)
	}
	if permanentHasGrantedProtectionFromColor(g, creature, color.White) {
		t.Fatal("creature protected before the Plate is attached")
	}

	attachPlate(creature, equipment)

	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("power while attached = %d, want 5", got)
	}
	if tough, ok := effectiveToughness(g, creature); !ok || tough != 5 {
		t.Fatalf("toughness while attached = %d (ok=%v), want 5", tough, ok)
	}
	if !permanentHasGrantedProtectionFromColor(g, creature, color.White) {
		t.Fatal("creature lacks protection while the Plate is attached")
	}

	// Source leaves the battlefield: buff and protection stop applying.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("power after Plate leaves = %d, want 2", got)
	}
	if permanentHasGrantedProtectionFromColor(g, creature, color.White) {
		t.Fatal("creature retains protection after the Plate leaves the battlefield")
	}
}

// TestCommanderIdentityProtectionBlockingEnforced verifies the resolved dynamic
// protection color set flows into the runtime blocking check: an equipped
// attacker with a mono-red commander controller has protection from green (not
// in identity) so a green creature cannot block it, while a red creature can.
func TestCommanderIdentityProtectionBlockingEnforced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	setPlayerCommander(t, g, game.Player1, color.NewIdentity(color.Red))

	attacker := plateBear(g, game.Player1)
	equipment := commandersPlateEquipment(g, game.Player1)
	attachPlate(attacker, equipment)

	pt := game.PT{Value: 2}
	greenBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Green Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	redBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	if canBlockAttacker(g, greenBlocker, attacker) {
		t.Fatal("green creature can block an attacker with protection from green, want false")
	}
	if !canBlockAttacker(g, redBlocker, attacker) {
		t.Fatal("red creature cannot block an attacker without protection from red, want true")
	}
}

// TestEquipCommanderTargetsOnlyCommander verifies the two coexisting equip
// abilities apply different target legality: the Equip commander ability matches
// only a commander the controller controls, while the ordinary Equip ability
// matches any creature the controller controls (including a commander).
func TestEquipCommanderTargetsOnlyCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := plateBear(g, game.Player1)
	g.CommanderIDs[commander.CardInstanceID] = true
	ordinary := plateBear(g, game.Player1)

	equipCommander := game.EquipCommanderActivatedAbility(cost.Mana{cost.O(3)})
	equipOrdinary := game.EquipActivatedAbility(cost.Mana{cost.O(5)})
	commanderSel := game.BodyTargets(&equipCommander)[0].Selection.Val
	ordinarySel := game.BodyTargets(&equipOrdinary)[0].Selection.Val

	if !matchSelectionForPermanent(g, game.Player1, commanderSel, commander) {
		t.Fatal("Equip commander should be legal targeting a commander you control")
	}
	if matchSelectionForPermanent(g, game.Player1, commanderSel, ordinary) {
		t.Fatal("Equip commander should not be legal targeting a non-commander creature")
	}
	if !matchSelectionForPermanent(g, game.Player1, ordinarySel, ordinary) {
		t.Fatal("ordinary Equip should be legal targeting any creature you control")
	}
	if !matchSelectionForPermanent(g, game.Player1, ordinarySel, commander) {
		t.Fatal("ordinary Equip should also be legal targeting a commander creature")
	}
}
