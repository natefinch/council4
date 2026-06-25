package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRemovePermanentFromBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)

	removed, ok := removePermanentFromBattlefield(g, first.ObjectID)

	if !ok || removed != first {
		t.Fatalf("removed permanent = %+v, want %+v", removed, first)
	}
	if len(g.Battlefield) != 1 || g.Battlefield[0] != second {
		t.Fatalf("battlefield = %+v, want only second permanent", g.Battlefield)
	}
}

func TestMovePermanentToZoneMovesCardBackedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1)

	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if !g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("card-backed permanent did not move to owner's graveyard")
	}
}

func TestMovePermanentsToZoneSimultaneouslyPreservesPreMoveLastKnownInformation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	granter := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dying Mentor",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				AddKeywords: []game.Keyword{game.Deathtouch},
			}},
		}},
	}})
	creature := addCombatCreaturePermanent(g, game.Player1)

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{granter, creature}, zone.Graveyard) {
		t.Fatal("movePermanentsToZoneSimultaneously() = false, want true")
	}

	snapshot, ok := lastKnownObject(g, creature.ObjectID)
	if !ok || !slices.Contains(snapshot.Keywords, game.Deathtouch) {
		t.Fatalf("last known keywords = %v, want deathtouch from pre-move battlefield", snapshot.Keywords)
	}
}

func TestMovePermanentToZoneMovesTokenObjectIDToDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player1,
		Controller: game.Player1,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{Name: "Token",
			Types: []types.Card{types.Creature}},
		},
	}
	g.Battlefield = append(g.Battlefield, token)

	if !movePermanentToZone(g, token, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if !g.Players[game.Player1].Graveyard.Contains(token.ObjectID) {
		t.Fatal("token object ID did not move to graveyard")
	}
}

func TestDestroyPermanentMovesToOwnersGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1)

	removed, ok := destroyPermanent(g, permanent.ObjectID)

	if !ok {
		t.Fatal("destroyPermanent() ok = false, want true")
	}
	if removed != permanent {
		t.Fatalf("destroyed permanent = %+v, want %+v", removed, permanent)
	}
	if !g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("destroyed permanent did not move to graveyard")
	}
}

func TestDestroyPermanentDoesNotMoveIndestructiblePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)

	removed, ok := destroyPermanent(g, permanent.ObjectID)

	if ok {
		t.Fatal("destroyPermanent() ok = true, want false")
	}
	if removed != nil {
		t.Fatalf("destroyed permanent = %+v, want nil", removed)
	}
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("indestructible permanent left the battlefield")
	}
	if g.Players[game.Player1].Graveyard.Contains(permanent.CardInstanceID) {
		t.Fatal("indestructible permanent moved to graveyard")
	}
}

func TestAttachPermanentAttachesAuraToLegalCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	aura := addAuraPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player2)

	if !attachPermanent(g, aura, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("aura attached to = %v, want %v", aura.AttachedTo, creature.ObjectID)
	}
	if len(creature.Attachments) != 1 || creature.Attachments[0] != aura.ObjectID {
		t.Fatalf("creature attachments = %+v, want aura %v", creature.Attachments, aura.ObjectID)
	}
}

func TestAuraWithoutEnchantTargetCannotAttach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Targetless Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura}},
	})
	creature := addCombatCreaturePermanent(g, game.Player2)

	if canAttachPermanent(g, aura, creature) {
		t.Fatal("Aura without explicit EnchantTarget attached via implicit default")
	}
}

func TestEnchantTargetRestrictsAuraAttachment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	aura := addCombatPermanent(g, game.Player1, landAuraCard())
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	creature := addCombatCreaturePermanent(g, game.Player2)

	if !canAttachPermanent(g, aura, land) {
		t.Fatal("canAttachPermanent() = false for matching enchant land target, want true")
	}
	if canAttachPermanent(g, aura, creature) {
		t.Fatal("canAttachPermanent() = true for creature with enchant land, want false")
	}
}

func TestEnchantUnionTargetMatchesAnyAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	aura := addCombatPermanent(g, game.Player1, creatureOrVehicleAuraCard())
	creature := addCombatCreaturePermanent(g, game.Player2)
	vehicle := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Vehicle",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Vehicle},
	}})
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Artifact",
		Types: []types.Card{types.Artifact},
	}})

	if !canAttachPermanent(g, aura, creature) {
		t.Fatal("canAttachPermanent() = false for creature with enchant creature-or-Vehicle, want true")
	}
	if !canAttachPermanent(g, aura, vehicle) {
		t.Fatal("canAttachPermanent() = false for Vehicle with enchant creature-or-Vehicle, want true")
	}
	if canAttachPermanent(g, aura, artifact) {
		t.Fatal("canAttachPermanent() = true for non-Vehicle artifact with enchant creature-or-Vehicle, want false")
	}
}

func creatureOrVehicleAuraCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Creature Or Vehicle Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					AnyOf: []game.Selection{
						{RequiredTypesAny: []types.Card{types.Creature}},
						{SubtypesAny: []types.Sub{types.Vehicle}},
					},
				}),
			}}},
		}}},
	}
}

func TestAuraSpellUsesEnchantTargetForCastLegality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	auraID := addCardToHand(g, game.Player1, landAuraCard())
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	creature := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, auraID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil) {
		t.Fatal("canCastSpell() = true for creature target with enchant land, want false")
	}
	if !engine.canCastSpell(g, game.Player1, auraID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil) {
		t.Fatal("canCastSpell() = false for land target with enchant land, want true")
	}
}

func TestAuraSpellAttachesToEnchantTargetOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	auraID := addCardToHand(g, game.Player1, landAuraCard())
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)) {
		t.Fatal("casting land Aura failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatal("Aura did not enter the battlefield")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != land.ObjectID {
		t.Fatalf("aura attached to = %v, want land %v", aura.AttachedTo, land.ObjectID)
	}
}

func TestIllegalEnchantTargetStateBasedActionMovesAuraToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	aura := addCombatPermanent(g, game.Player1, landAuraCard())
	creature := addCombatCreaturePermanent(g, game.Player2)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("illegally attached land Aura remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("illegally attached land Aura did not move to graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != aura.ObjectID || deaths[0].Reason != PermanentDeathReasonIllegalAura {
		t.Fatalf("death logs = %+v, want illegal aura death", deaths)
	}
}

func TestIllegalAuraStateBasedActionMovesAuraToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	aura := addAuraPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player2)
	if !attachPermanent(g, aura, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}

	movePermanentToZone(g, creature, zone.Graveyard)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("unattached aura remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("unattached aura did not move to graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != aura.ObjectID || deaths[0].Reason != PermanentDeathReasonIllegalAura {
		t.Fatalf("death logs = %+v, want illegal aura death", deaths)
	}
}

func TestEquipmentRemainsWhenEquippedCreatureLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addEquipmentPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent() = false, want true")
	}

	movePermanentToZone(g, creature, zone.Graveyard)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("death logs = %+v, want no equipment death", deaths)
	}
	if _, ok := permanentByObjectID(g, equipment.ObjectID); !ok {
		t.Fatal("equipment left battlefield when equipped creature left")
	}
	if equipment.AttachedTo.Exists {
		t.Fatalf("equipment attached to = %v, want absent", equipment.AttachedTo.Val)
	}
	if len(creature.Attachments) != 0 {
		t.Fatalf("removed creature attachments = %+v, want none", creature.Attachments)
	}
}

func TestRemovePermanentFromBattlefieldMissingPermanentReturnsNil(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	if removed, ok := removePermanentFromBattlefield(g, 999); ok || removed != nil {
		t.Fatalf("removed permanent = %+v, %v, want nil, false", removed, ok)
	}
}

func addAuraPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Test Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
				}),
			}}},
		}}},
	})
}

func landAuraCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Land Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Land},
				}),
			}}},
		}}},
	}
}

func findPermanentByCardID(g *game.Game, cardID id.ID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent, true
		}
	}
	return nil, false
}

func addEquipmentPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Test Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment}},
	})
}
