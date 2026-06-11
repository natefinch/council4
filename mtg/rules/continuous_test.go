package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestStaticPTEffectAffectsCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addAnthemPermanent(g, game.Player1)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: opponentCreature.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		},
	}
	log := TurnLog{}

	NewEngine(nil).resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending Player2 life = %d, want 37", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Life != 38 {
		t.Fatalf("defending Player1 life = %d, want 38", g.Players[game.Player1].Life)
	}
}

func TestConditionalSourceKeywordEffectTracksCondition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      "Conditional Flier",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []types.Sub{types.Mountain},
				},
			}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerAbility,
				AffectedSource: true,
				AddKeywords:    []game.Keyword{game.Flying},
			}},
		}},
	}}
	source := addCombatPermanent(g, game.Player1, def)
	otherControllerSource := addCombatPermanent(g, game.Player2, def)
	mountain := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain},
	}})

	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("source did not gain flying while its controller controlled a Mountain")
	}
	if hasKeyword(g, otherControllerSource, game.Flying) {
		t.Fatal("same card definition gained flying for a controller without a Mountain")
	}

	movePermanentToZone(g, mountain, zone.Graveyard)
	if hasKeyword(g, source, game.Flying) {
		t.Fatal("source retained flying after its controller lost the Mountain")
	}
}

func TestConditionalSourceKeywordEffectUsesEffectiveCharacteristics(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Conditional Flier",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					Types:     []types.Card{types.Creature},
					ColorsAny: []color.Color{color.Red},
				},
			}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerAbility,
				AffectedSource: true,
				AddKeywords:    []game.Keyword{game.Flying},
			}},
		}},
	}})
	qualifier := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Changed Creature",
		Types: []types.Card{types.Creature},
	}})

	if hasKeyword(g, source, game.Flying) {
		t.Fatal("source gained flying from a creature with the wrong controller and color")
	}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			AffectedObjectID: qualifier.ObjectID,
			Layer:            game.LayerControl,
			NewController:    opt.Val(game.Player1),
		},
		game.ContinuousEffect{
			AffectedObjectID: qualifier.ObjectID,
			Layer:            game.LayerColor,
			AddColors:        []color.Color{color.Red},
		},
	)
	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("source did not gain flying from an effectively controlled red creature")
	}
}

func TestSourceContinuousEffectWithGroupDoesNotApply(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Invalid Granter",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerAbility,
				AffectedSource: true,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Flying},
			}},
		}},
	}})

	if hasKeyword(g, source, game.Flying) {
		t.Fatal("invalid source-and-group continuous effect applied")
	}
}

func TestStaticPTEffectRaisesLethalDamageThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addAnthemPermanent(g, game.Player2)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
		t.Fatal("anthem-pumped blocker died to nonlethal marked damage")
	}
	for _, death := range deaths {
		if death.Permanent == blocker.ObjectID {
			t.Fatalf("blocker death = %+v, want blocker to survive anthem-raised toughness", death)
		}
	}
}

func TestStaticDomainDynamicPTUsesLayerBoundedValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	domain := opt.Val(game.DynamicAmount{
		Kind:       game.DynamicAmountControllerBasicLandTypeCount,
		Multiplier: 1,
	})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Domain Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:                 game.LayerPowerToughnessModify,
				AffectedSource:        true,
				PowerDeltaDynamic:     domain,
				ToughnessDeltaDynamic: domain,
			}},
		}},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Dual Land",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Plains, types.Island},
	}})

	if got := effectivePower(g, source); got != 4 {
		t.Fatalf("effective power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, source); !ok || got != 4 {
		t.Fatalf("effective toughness = %d ok=%v, want 4 true", got, ok)
	}
}

func TestStaticCovenConditionUsesLayerBoundedValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Coven Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{ControllerCreaturePowerDiversityAtLeast: 3}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerPowerToughnessModify,
				AffectedSource: true,
				PowerDelta:     1,
				ToughnessDelta: 1,
			}},
		}},
	}})
	counterCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	counterCreature.Counters.Add(counter.PlusOnePlusOne, 1)
	temporaryCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	temporaryCreature.TemporaryPowerModifier = 2

	if got := effectivePower(g, source); got != 2 {
		t.Fatalf("effective power = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, source); !ok || got != 2 {
		t.Fatalf("effective toughness = %d ok=%v, want 2 true", got, ok)
	}
}

func TestStaticPTEffectDisappearingChangesLethalDamageThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anthem := addAnthemPermanent(g, game.Player1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2
	engine := NewEngine(nil)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("anthem-pumped creature died before anthem left")
	}
	if len(deaths) != 0 {
		t.Fatalf("deaths before anthem leaves = %+v, want none", deaths)
	}

	movePermanentToZone(g, anthem, zone.Graveyard)
	_, deaths = engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature survived after anthem left and marked damage became lethal")
	}
	if len(deaths) != 1 || deaths[0].Permanent != creature.ObjectID || deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("deaths after anthem leaves = %+v, want creature lethal damage death", deaths)
	}
}

func TestContinuousEffectsApplyInLayerOrderBeforeTimestamp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	animatedLand := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Animated Forest",
		Types: []types.Card{types.Land}},
	})
	two := game.PT{Value: 2}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: animatedLand.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(two),
			SetToughness:     opt.Val(two),
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: animatedLand.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessModify,
			PowerDelta:       3,
			ToughnessDelta:   3,
		},
	)

	if got := effectivePower(g, animatedLand); got != 5 {
		t.Fatalf("effective power = %d, want layer-ordered 5", got)
	}
	if got, ok := effectiveToughness(g, animatedLand); !ok || got != 5 {
		t.Fatalf("effective toughness = %d ok=%v, want layer-ordered 5 true", got, ok)
	}
}

func TestContinuousEffectDependenciesOverrideTimestampWithinLayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	four := game.PT{Value: 4}
	one := game.PT{Value: 1}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               10,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(four),
			SetToughness:     opt.Val(four),
		},
		game.ContinuousEffect{
			ID:               11,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        10,
			DependsOn:        []id.ID{10},
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(one),
			SetToughness:     opt.Val(one),
		},
	)

	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want dependency-ordered 1", got)
	}
}

func TestTypeAndPTContinuousEffectsAffectCombatAndSBAs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Living Land",
		Types: []types.Card{types.Land}},
	})
	two := game.PT{Value: 2}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: land.ObjectID,
			Layer:            game.LayerType,
			AddTypes:         []types.Card{types.Creature},
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: land.ObjectID,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(two),
			SetToughness:     opt.Val(two),
		},
	)
	land.MarkedDamage = 2

	if !canAttackWith(g, land, game.Player1) {
		t.Fatal("animated land could not attack as an effective creature")
	}
	_, deaths := NewEngine(nil).applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, land.ObjectID); ok {
		t.Fatal("animated land survived lethal marked damage")
	}
	if len(deaths) != 1 || deaths[0].Permanent != land.ObjectID {
		t.Fatalf("deaths = %+v, want animated land lethal damage death", deaths)
	}
}

func TestDynamicStarPowerAffectsCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	star := game.PT{IsStar: true}
	dynamic := game.DynamicValue{Kind: game.DynamicValueControllerHandSize}
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Hand Avatar",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(star),
		Toughness:        opt.Val(star),
		DynamicPower:     opt.Val(dynamic),
		DynamicToughness: opt.Val(dynamic)},
	})
	for range 3 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card in Hand"}})
	}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending life = %d, want dynamic-star combat damage to set it to 37", g.Players[game.Player2].Life)
	}
}

func TestCopyEffectChangesEffectiveCombatKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	copier := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	copyPower := game.PT{Value: 4}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: copier.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues: opt.Val(game.CopyableValues{
			Name:      "Copied Dragon",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(copyPower),
			Toughness: opt.Val(copyPower),
			Abilities: []game.Ability{game.StaticAbility{Text: "Flying", KeywordAbilities: game.SimpleKeywords(game.Flying)}},
		}),
	})

	if got := permanentEffectiveName(g, copier); got != "Copied Dragon" {
		t.Fatalf("effective name = %q, want copied name", got)
	}
	if got := effectivePower(g, copier); got != 4 {
		t.Fatalf("effective power = %d, want copied power 4", got)
	}
	if !hasKeyword(g, copier, game.Flying) {
		t.Fatal("copy effect did not grant copied Flying keyword")
	}
}

func TestKeywordAddRemoveEffectsStickAfterLayerSix(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerAbility,
			RemoveKeywords:   []game.Keyword{game.Flying},
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerAbility,
			AddKeywords:      []game.Keyword{game.Trample},
		},
	)

	if hasKeyword(g, creature, game.Flying) {
		t.Fatal("remove-keyword continuous effect did not remove Flying")
	}
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("add-keyword continuous effect did not add Trample")
	}
}

func TestAbilityLayerAddsTypedAbilityBody(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ability Recipient",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerAbility,
		AddAbilities: []game.Ability{
			game.ActivatedAbility{
				Text: "{2}: This creature gets +1/+0 until end of turn.",
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.ModifyPT{
								Object:     game.SourcePermanentReference(),
								PowerDelta: game.Fixed(1),
								Duration:   game.DurationUntilEndOfTurn,
							},
						},
					},
				}.Ability(),
			},
		},
	})

	values := effectivePermanentValues(g, creature)
	if len(values.abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(values.abilities))
	}
	if _, ok := values.abilities[0].(game.ActivatedAbility); !ok {
		t.Fatalf("ability body = %T, want game.ActivatedAbilityBody", values.abilities[0])
	}
}

func TestControlChangeEffectsAffectLegalityAndSelectors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	newController := game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(newController),
	})

	if canAttackWith(g, creature, game.Player1) {
		t.Fatal("old controller can attack with control-changed creature")
	}
	if !canAttackWith(g, creature, game.Player2) {
		t.Fatal("new effective controller cannot attack with control-changed creature")
	}
	values := effectivePermanentValues(g, creature)
	effect := &game.ContinuousEffect{
		Controller: game.Player2,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}
	if !continuousEffectApplies(g, creature, &values, effect) {
		t.Fatal("creatures-you-control group did not use effective controller")
	}
}

func TestContinuousEffectBattlefieldGroupMatchesCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:         1,
		Controller: game.Player1,
		Layer:      game.LayerAbility,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
		AddKeywords: []game.Keyword{game.Haste},
	})

	if !hasKeyword(g, creature, game.Haste) {
		t.Fatal("creature did not gain haste from battlefield creature group")
	}
	if hasKeyword(g, artifact, game.Haste) {
		t.Fatal("noncreature artifact incorrectly matched battlefield creature group")
	}
}

func TestCopyEffectPreservesDynamicStarValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	copier := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	star := game.PT{IsStar: true}
	dynamic := game.DynamicValue{Kind: game.DynamicValueControllerHandSize}
	for range 4 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card in Hand"}})
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: copier.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues:       opt.Val(game.CopyableValues{Name: "Copied Star", Types: []types.Card{types.Creature}, Power: opt.Val(star), Toughness: opt.Val(star), DynamicPower: opt.Val(dynamic), DynamicToughness: opt.Val(dynamic)}),
	})

	if got := effectivePower(g, copier); got != 4 {
		t.Fatalf("copied dynamic-star power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, copier); !ok || got != 4 {
		t.Fatalf("copied dynamic-star toughness = %d ok=%v, want 4 true", got, ok)
	}
}

// TestContinuousEffectObjectControlledGroupMatchesOwnedCreatures verifies that
// a GroupDomainObjectControlled effect only applies to permanents controlled by
// the same player who controls the anchor permanent.
func TestContinuousEffectObjectControlledGroupMatchesOwnedCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	allyCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
		AddKeywords: []game.Keyword{game.Haste},
	})

	if !hasKeyword(g, allyCreature, game.Haste) {
		t.Fatal("creature controlled by same player did not gain haste from ObjectControlled group")
	}
	if !hasKeyword(g, anchor, game.Haste) {
		t.Fatal("anchor creature itself did not gain haste from ObjectControlled group")
	}
	if hasKeyword(g, opponentCreature, game.Haste) {
		t.Fatal("opponent's creature incorrectly gained haste from ObjectControlled group")
	}
}

// TestContinuousEffectObjectControlledGroupExclusion verifies that the exclusion
// ObjectReference removes a specific permanent from the ObjectControlled group.
func TestContinuousEffectObjectControlledGroupExclusion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	// The effect excludes the source (anchor) from the group it creates.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.ObjectControlledGroupExcluding(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		),
		AddKeywords: []game.Keyword{game.Haste},
	})

	if !hasKeyword(g, other, game.Haste) {
		t.Fatal("non-excluded creature did not gain haste")
	}
	if hasKeyword(g, anchor, game.Haste) {
		t.Fatal("excluded source creature incorrectly gained haste from ObjectControlled group")
	}
}

// TestContinuousEffectObjectControlledGroupUsesEffectiveController verifies that
// control-change effects are respected when matching ObjectControlled groups.
func TestContinuousEffectObjectControlledGroupUsesEffectiveController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	stolen := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	// Control-change effect gives Player1 control of stolen, then an anthem-style
	// effect from anchor grants haste to all creatures Player1 controls.
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: stolen.ObjectID,
			Layer:            game.LayerControl,
			NewController:    opt.Val(game.Player1),
		},
		game.ContinuousEffect{
			ID:             2,
			Controller:     game.Player1,
			SourceObjectID: anchor.ObjectID,
			Layer:          game.LayerAbility,
			Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			}),
			AddKeywords: []game.Keyword{game.Haste},
		},
	)

	if !hasKeyword(g, stolen, game.Haste) {
		t.Fatal("control-changed creature should gain haste from ObjectControlled group")
	}
}

func TestEquippedCreatureStaticPTBuff(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	equipment := addEquipmentWithPTBuff(g, game.Player1, 2, 0)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	creature.Attachments = append(creature.Attachments, equipment.ObjectID)
	equipment.AttachedTo = opt.Val(creature.ObjectID)

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want 3", got)
	}
	if got := effectivePower(g, equipment); got != 0 {
		t.Fatalf("equipment effective power = %d, want 0", got)
	}
}

func TestStaticPTBuffSelectionSupportsTokensAndOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Selective Anthem",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer: game.LayerPowerToughnessModify,
					Group: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{TokenOnly: true},
					),
					PowerDelta:     1,
					ToughnessDelta: 1,
				},
				{
					Layer: game.LayerPowerToughnessModify,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerOpponent,
					}),
					PowerDelta: -1,
				},
			},
		}},
	}})
	token, _ := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Creature Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	nontoken := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponent := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if got := effectivePower(g, token); got != 3 {
		t.Fatalf("token effective power = %d, want 3", got)
	}
	if got := effectivePower(g, nontoken); got != 2 {
		t.Fatalf("nontoken effective power = %d, want 2", got)
	}
	if got := effectivePower(g, opponent); got != 1 {
		t.Fatalf("opponent creature effective power = %d, want 1", got)
	}
	if got := effectivePower(g, source); got != 0 {
		t.Fatalf("source effective power = %d, want 0", got)
	}
}

func addAnthemPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Anthem Captain",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{
			{
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer: game.LayerPowerToughnessModify,
						Group: game.BattlefieldGroup(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Controller:    game.ControllerYou,
							ExcludeSource: true,
						}),
						PowerDelta:     1,
						ToughnessDelta: 1,
					},
				},
			},
		}},
	})
}

func addEquipmentWithPTBuff(g *game.Game, controller game.PlayerID, powerDelta, toughnessDelta int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Buffing Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerPowerToughnessModify,
				Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
			}},
		}},
	}})
}

func TestGainControlApplyContinuousSubstitutesController(t *testing.T) {
	t.Parallel()
	// Verify that applyTypedContinuousEffects substitutes the sentinel Player1
	// NewController value with the actual resolving controller.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Player2 controls a creature.
	creature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Target Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	// Player1 casts a gain-control spell targeting the creature.
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:         game.LayerControl,
				NewController: opt.Val(game.Player1), // sentinel: replaced with actual controller
			}},
			Duration: game.DurationUntilEndOfTurn,
		},
	}}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectiveController(g, creature); got != game.Player1 {
		t.Fatalf("controller after gain-control = %v, want Player1", got)
	}
	// After cleanup, the control effect should expire and the original
	// controller is restored.
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("controller after cleanup = %v, want Player2 (original)", got)
	}
}
