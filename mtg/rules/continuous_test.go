package rules

import (
	"testing"

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
								TargetIndex: game.TargetIndexSourcePermanent,
								PowerDelta:  game.Fixed(1),
								Duration:    game.DurationUntilEndOfTurn,
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
	if !permanentMatchesSelectorForSource(g, nil, game.Player2, creature, game.EffectSelectorCreaturesYouControl) {
		t.Fatal("creatures-you-control selector did not use effective controller")
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
						Layer:          game.LayerPowerToughnessModify,
						Selector:       game.EffectSelectorOtherCreaturesYouControl,
						PowerDelta:     1,
						ToughnessDelta: 1,
					},
				},
			},
		}},
	})
}
