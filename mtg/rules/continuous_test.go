package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestChangelingHasEveryCreatureSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Changeling",
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Shapeshifter},
		StaticAbilities: []game.StaticAbility{game.ChangelingStaticBody},
	}})

	subtypes := effectivePermanentValues(g, permanent).subtypes
	for _, subtype := range []types.Sub{types.Shapeshifter, types.Elf, types.Zombie} {
		if !slices.Contains(subtypes, subtype) {
			t.Fatalf("changeling subtypes omit %s: %#v", subtype, subtypes)
		}
	}

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer:              game.LayerAbility,
		AffectedObjectID:   permanent.ObjectID,
		RemoveAllAbilities: true,
	})
	subtypes = effectivePermanentValues(g, permanent).subtypes
	if !slices.Contains(subtypes, types.Elf) {
		t.Fatalf("removing abilities erased Changeling's type-layer subtypes: %#v", subtypes)
	}
}

func TestChangelingMatchesArbitrarySubtypesSimultaneously(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Universal Automaton",
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Shapeshifter},
		StaticAbilities: []game.StaticAbility{game.ChangelingStaticBody},
	}})

	for _, subtype := range []types.Sub{types.Goblin, types.Elf} {
		if !permanentHasSubtype(g, permanent, subtype) {
			t.Fatalf("changeling not treated as %s", subtype)
		}
	}
}

func TestGroupAddSubtypeStaticKeepsOtherTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	ally := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin Ally",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerType,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
		AddSubtypes: []types.Sub{types.Sliver},
	})

	if !permanentHasSubtype(g, ally, types.Sliver) {
		t.Fatal("controlled creature did not gain the added Sliver subtype")
	}
	if !permanentHasSubtype(g, ally, types.Goblin) {
		t.Fatal("adding a subtype erased the creature's original Goblin subtype")
	}
	if permanentHasSubtype(g, opponentCreature, types.Sliver) {
		t.Fatal("opponent's creature incorrectly gained the added Sliver subtype")
	}
}

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

func TestPhasedOutStaticAbilitySourceDoesNotApply(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addAnthemPermanent(g, game.Player1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power with anthem = %d, want 3", got)
	}
	source.PhasedOut = true
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power with phased-out anthem = %d, want 2", got)
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

func TestDynamicStarPowerToughnessTracksLandCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	star := game.PT{IsStar: true}
	dynamic := game.DynamicValue{Kind: game.DynamicValueControllerLandCount}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Multani",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(star),
		Toughness:        opt.Val(star),
		DynamicPower:     opt.Val(dynamic),
		DynamicToughness: opt.Val(dynamic)},
	})
	for range 4 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
			Types: []types.Card{types.Land}}})
	}

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power = %d, want land count 4", got)
	}
	toughness, ok := effectiveToughness(g, creature)
	if !ok || toughness != 4 {
		t.Fatalf("effective toughness = %d (ok=%v), want land count 4", toughness, ok)
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types: []types.Card{types.Land}}})
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power after extra land = %d, want 5", got)
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
			Abilities: []game.Ability{&game.StaticAbility{Text: "Flying", KeywordAbilities: game.SimpleKeywords(game.Flying)}},
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
			&game.ActivatedAbility{
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
	if _, ok := values.abilities[0].(*game.ActivatedAbility); !ok {
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

func TestContinuousEffectDoublesGroupPowerAndToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Doubled Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:         1,
		Controller: game.Player1,
		Layer:      game.LayerPowerToughnessModify,
		Group: game.BattlefieldGroup(game.Selection{
			Controller:    game.ControllerYou,
			RequiredTypes: []types.Card{types.Creature},
		}),
		DoublePower:     true,
		DoubleToughness: true,
	})

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("doubled power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 6 {
		t.Fatalf("doubled toughness = %d ok=%v, want 6 true", got, ok)
	}
	if got := effectivePower(g, opponentCreature); got != 2 {
		t.Fatalf("opponent creature power = %d, want 2 (not doubled)", got)
	}
}

// TestContinuousEffectDoublePowerStacksOnRunningValue verifies the doubling adds
// each creature's running power (after an earlier +1/+1 anthem in the same layer)
// back into itself, doubling the already-buffed value rather than the printed one.
func TestContinuousEffectDoublePowerStacksOnRunningValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:         1,
			Controller: game.Player1,
			Timestamp:  1,
			Layer:      game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				Controller:    game.ControllerYou,
				RequiredTypes: []types.Card{types.Creature},
			}),
			PowerDelta: 1,
		},
		game.ContinuousEffect{
			ID:         2,
			Controller: game.Player1,
			Timestamp:  2,
			Layer:      game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				Controller:    game.ControllerYou,
				RequiredTypes: []types.Card{types.Creature},
			}),
			DoublePower: true,
		},
	)

	if got := effectivePower(g, creature); got != 6 {
		t.Fatalf("power = %d, want 6 ((2+1) doubled)", got)
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

// TestContinuousEffectKeywordFilteredGroupMatchesOnlyKeywordHolders verifies a
// keyword-filtered group anthem ("Creatures with flying get +1/+1") buffs only
// the creatures that already have the keyword, and a "without" filter buffs only
// the creatures that lack it.
func TestContinuousEffectKeywordFilteredGroupMatchesOnlyKeywordHolders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flyer := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:         1,
			Controller: game.Player1,
			Layer:      game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Keyword:       game.Flying,
			}),
			PowerDelta: 1,
		},
		game.ContinuousEffect{
			ID:         2,
			Controller: game.Player1,
			Layer:      game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes:   []types.Card{types.Creature},
				ExcludedKeyword: game.Flying,
			}),
			PowerDelta: 10,
		},
	)

	if got := effectivePower(g, flyer); got != 3 {
		t.Fatalf("flyer power = %d, want 3 (only the with-flying anthem applies)", got)
	}
	if got := effectivePower(g, grounded); got != 12 {
		t.Fatalf("grounded power = %d, want 12 (only the without-flying anthem applies)", got)
	}
}

// TestContinuousEffectArtifactCreatureGroupMatchesOnlyArtifactCreatures verifies
// a multi-type group anthem ("Artifact creatures you control get +1/+1") buffs
// only creatures whose type line includes both Artifact and Creature.
func TestContinuousEffectArtifactCreatureGroupMatchesOnlyArtifactCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ornithopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	plainCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: artifactCreature.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Artifact, types.Creature},
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, artifactCreature); got != 3 {
		t.Fatalf("artifact creature power = %d, want 3", got)
	}
	if got := effectivePower(g, plainCreature); got != 2 {
		t.Fatalf("non-artifact creature power = %d, want 2 (unbuffed)", got)
	}
}

// TestContinuousEffectNontokenGroupExcludesTokens verifies a nontoken group
// anthem ("Nontoken creatures you control get +1/+1") buffs real cards but not
// token permanents.
func TestContinuousEffectNontokenGroupExcludesTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	nontoken := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	pt := game.PT{Value: 2}
	tokenPermanent := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player1,
		Controller: game.Player1,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:      "Soldier Token",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(pt),
			Toughness: opt.Val(pt),
		}},
	}
	g.Battlefield = append(g.Battlefield, tokenPermanent)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: nontoken.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			NonToken:      true,
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, nontoken); got != 3 {
		t.Fatalf("nontoken creature power = %d, want 3", got)
	}
	if got := effectivePower(g, tokenPermanent); got != 2 {
		t.Fatalf("token creature power = %d, want 2 (unbuffed)", got)
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

// TestContinuousEffectBattlefieldGroupCombatStateMatchesAttackers verifies that a
// group anthem filtered by attacking combat state buffs only attacking creatures.
func TestContinuousEffectBattlefieldGroupCombatStateMatchesAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:         1,
		Controller: game.Player1,
		Layer:      game.LayerPowerToughnessModify,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("attacking creature power = %d, want 3", got)
	}
	if got := effectivePower(g, bystander); got != 2 {
		t.Fatalf("non-attacking creature power = %d, want 2 (should not be buffed)", got)
	}
}

// TestContinuousEffectBattleCryBuffsOtherAttackersOnly verifies the Battle cry
// triggered ability shape: each OTHER attacking creature gets +1/+0, while the
// source creature (also attacking) is excluded and non-attackers are unaffected.
func TestContinuousEffectBattleCryBuffsOtherAttackersOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: otherAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.BattlefieldGroupExcluding(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}, game.SourcePermanentReference()),
		PowerDelta: 1,
	})

	if got := effectivePower(g, source); got != 2 {
		t.Fatalf("source attacker power = %d, want 2 (excluded from its own Battle cry)", got)
	}
	if got := effectivePower(g, otherAttacker); got != 3 {
		t.Fatalf("other attacking creature power = %d, want 3 (+1/+0 from Battle cry)", got)
	}
	if got := effectivePower(g, bystander); got != 2 {
		t.Fatalf("non-attacking creature power = %d, want 2 (should not be buffed)", got)
	}
}

// TestContinuousEffectBattlefieldGroupSubtypeMatchesSubtype verifies that a group
// anthem filtered by creature subtype buffs only creatures of that subtype.
func TestContinuousEffectChosenTypeGroupBuffsOnlyChosenSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Chosen-Type Anthem",
		Types: []types.Card{types.Artifact},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Elf},
	}
	elf := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Elf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	goblin := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Goblin",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypeChoice: game.SubtypeChoiceSourceEntry,
		}),
		PowerDelta:     1,
		ToughnessDelta: 1,
	})

	if got := effectivePower(g, elf); got != 3 {
		t.Fatalf("chosen-type creature power = %d, want 3", got)
	}
	if got := effectivePower(g, goblin); got != 2 {
		t.Fatalf("non-chosen-type creature power = %d, want 2 (should not be buffed)", got)
	}
}

func TestContinuousEffectChosenTypeGroupFailsClosedWithoutChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Chosen-Type Anthem",
		Types: []types.Card{types.Artifact},
	}})
	elf := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Elf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypeChoice: game.SubtypeChoiceSourceEntry,
		}),
		PowerDelta:     1,
		ToughnessDelta: 1,
	})

	if got := effectivePower(g, elf); got != 2 {
		t.Fatalf("creature power = %d, want 2 (no entry choice means no buff)", got)
	}
}

func TestContinuousEffectBattlefieldGroupSubtypeMatchesSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 2}
	sliver := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Sliver",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sliver},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	goblin := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Goblin",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:         1,
		Controller: game.Player1,
		Layer:      game.LayerPowerToughnessModify,
		Group: game.BattlefieldGroup(game.Selection{
			SubtypesAny: []types.Sub{types.Sliver},
		}),
		PowerDelta:     1,
		ToughnessDelta: 1,
	})

	if got := effectivePower(g, sliver); got != 3 {
		t.Fatalf("Sliver power = %d, want 3", got)
	}
	if got := effectivePower(g, goblin); got != 2 {
		t.Fatalf("non-Sliver power = %d, want 2 (should not be buffed)", got)
	}
}

func TestContinuousEffectAddsSubtypeChosenAsSourceEntered(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Choosing Creature",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Golem},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:                     game.LayerType,
				AffectedSource:            true,
				AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,
			}},
		}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {
			Kind:    game.ResolutionChoiceSubtype,
			Subtype: types.Elf,
		},
	}

	if !permanentHasSubtype(g, source, types.Golem) {
		t.Fatal("source lost its printed subtype")
	}
	if !permanentHasSubtype(g, source, types.Elf) {
		t.Fatal("source did not gain the creature type chosen as it entered")
	}
}

func TestContinuousEffectChosenSubtypeFailsClosedWithoutSubtypeChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Choosing Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:                     game.LayerType,
				AffectedSource:            true,
				AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,
			}},
		}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {
			Kind:    game.ResolutionChoiceMana,
			Subtype: types.Elf,
		},
	}

	if permanentHasSubtype(g, source, types.Elf) {
		t.Fatal("non-subtype entry choice added a creature type")
	}
}

// TestContinuousEffectObjectControlledGroupTokenOnlyMatchesTokens verifies that a
// controlled-group anthem filtered by TokenOnly buffs only token creatures.
func TestContinuousEffectObjectControlledGroupTokenOnlyMatchesTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	pt := game.PT{Value: 2}
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
	token := addCombatPermanent(g, game.Player1, tokenDef)
	token.Token = true
	token.TokenDef = tokenDef
	nonToken := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			TokenOnly:     true,
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, token); got != 3 {
		t.Fatalf("token creature power = %d, want 3", got)
	}
	if got := effectivePower(g, nonToken); got != 2 {
		t.Fatalf("non-token creature power = %d, want 2 (should not be buffed)", got)
	}
}

// TestContinuousEffectObjectControlledGroupSupertypeMatchesLegendary verifies
// that a controlled-group anthem filtered by the Legendary supertype buffs only
// legendary creatures.
func TestContinuousEffectObjectControlledGroupSupertypeMatchesLegendary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	pt := game.PT{Value: 2}
	legendary := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Test Legend",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
	}})
	ordinary := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Supertypes:    []types.Super{types.Legendary},
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, legendary); got != 3 {
		t.Fatalf("legendary creature power = %d, want 3", got)
	}
	if got := effectivePower(g, ordinary); got != 2 {
		t.Fatalf("nonlegendary creature power = %d, want 2 (should not be buffed)", got)
	}
}

// TestContinuousEffectObjectControlledGroupUntappedMatchesUntapped verifies that
// a controlled-group anthem filtered by untapped state buffs only untapped
// creatures.
func TestContinuousEffectObjectControlledGroupUntappedMatchesUntapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anchor := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	untapped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	tapped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	tapped.Tapped = true
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Tapped:        game.TriFalse,
		}),
		ToughnessDelta: 2,
	})

	if got, ok := effectiveToughness(g, untapped); !ok || got != 4 {
		t.Fatalf("untapped creature toughness = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, tapped); !ok || got != 2 {
		t.Fatalf("tapped creature toughness = %d, want 2 (should not be buffed)", got)
	}
}

func TestStaticSourceTiedControlGrantOnAttachedObject(t *testing.T) {
	t.Parallel()
	// Mirror a generated "You control enchanted creature." Aura: a static
	// ability whose LayerControl continuous effect targets the attached object
	// and grants control to the Aura's controller.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Mind Grip",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:         game.LayerControl,
				NewController: opt.Val(game.Player1), // sentinel: replaced with source controller
				Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
			}},
		}},
	}}
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	aura := addCombatPermanent(g, game.Player1, auraDef)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	if got := effectiveController(g, creature); got != game.Player1 {
		t.Fatalf("controller while Aura attached = %v, want Player1", got)
	}

	// The new controller is the Aura's controller, not the sentinel: a Player2
	// Aura grants control to Player2.
	aura.Controller = game.Player2
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("controller for Player2-controlled Aura = %v, want Player2", got)
	}
	aura.Controller = game.Player1

	// When the source Aura leaves the battlefield the grant stops applying
	// immediately and control reverts to the creature's owner.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("controller after Aura leaves = %v, want Player2 (original)", got)
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

// TestStaticCharacteristicGrantsOnAttachedObject mirrors a generated Aura whose
// static ability composes layer-preserving characteristic changes on the
// enchanted creature: it sets base power/toughness, adds a color, and adds a
// subtype. The runtime applies each layer to the attached object.
func TestStaticCharacteristicGrantsOnAttachedObject(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Beastly Curse",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:        game.LayerPowerToughnessSet,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetPower:     opt.Val(game.PT{Value: 0}),
					SetToughness: opt.Val(game.PT{Value: 2}),
				},
				{
					Layer:     game.LayerColor,
					Group:     game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddColors: []color.Color{color.Black},
				},
				{
					Layer:       game.LayerType,
					Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddSubtypes: []types.Sub{types.Zombie},
				},
			},
		}},
	}}
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	aura := addCombatPermanent(g, game.Player1, auraDef)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	if got := effectivePower(g, creature); got != 0 {
		t.Fatalf("effective power = %d, want 0 (base set overrides printed)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 2 {
		t.Fatalf("effective toughness = %d (ok=%v), want 2", got, ok)
	}
	if !slices.Contains(permanentEffectiveColors(g, creature), color.Black) {
		t.Fatalf("colors = %#v, want to contain black", permanentEffectiveColors(g, creature))
	}
	if !permanentHasSubtype(g, creature, types.Zombie) {
		t.Fatal("enchanted creature should gain the Zombie subtype")
	}

	// When the Aura leaves the battlefield the grants stop applying.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power after Aura leaves = %d, want 4 (printed)", got)
	}
	if slices.Contains(permanentEffectiveColors(g, creature), color.Black) {
		t.Fatal("colors should revert after Aura leaves")
	}
}

// TestConditionalSourceControlsMatchingTokenStatic proves a self static gated on
// a "you control a token" condition applies its source power/toughness and
// keyword modification only while the controller controls a token, exercising
// the ControlsMatching TokenOnly selection produced by the cardgen pipeline.
func TestConditionalSourceControlsMatchingTokenStatic(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Token Watcher",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{TokenOnly: true},
				}),
			}),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerPowerToughnessModify,
					AffectedSource: true,
					PowerDelta:     2,
				},
				{
					Layer:          game.LayerAbility,
					AffectedSource: true,
					AddKeywords:    []game.Keyword{game.Trample},
				},
			},
		}},
	}})

	if got := effectivePower(g, source); got != 1 {
		t.Fatalf("power without a token = %d, want 1", got)
	}
	if hasKeyword(g, source, game.Trample) {
		t.Fatal("source gained trample without controlling a token")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Creature",
		Types: []types.Card{types.Creature},
	}})
	if got := effectivePower(g, source); got != 1 {
		t.Fatalf("power with only a non-token permanent = %d, want 1", got)
	}

	token := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Spirit",
		Types: []types.Card{types.Creature},
	}})
	token.Token = true
	if got := effectivePower(g, source); got != 3 {
		t.Fatalf("power while controlling a token = %d, want 3", got)
	}
	if !hasKeyword(g, source, game.Trample) {
		t.Fatal("source did not gain trample while controlling a token")
	}

	movePermanentToZone(g, token, zone.Graveyard)
	if got := effectivePower(g, source); got != 1 {
		t.Fatalf("power after the token left = %d, want 1", got)
	}
	if hasKeyword(g, source, game.Trample) {
		t.Fatal("source retained trample after the token left")
	}
}

// TestConditionalSourceLifeThresholdStatic proves a self static produced from a
// leading "As long as you have N or more life, ..." clause applies its source
// power/toughness and keyword only while the controller meets the life
// threshold.
func TestConditionalSourceLifeThresholdStatic(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Devout Defender",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{ControllerLifeAtLeast: 30}),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerPowerToughnessModify,
					AffectedSource: true,
					PowerDelta:     5,
					ToughnessDelta: 5,
				},
				{
					Layer:          game.LayerAbility,
					AffectedSource: true,
					AddKeywords:    []game.Keyword{game.Flying},
				},
			},
		}},
	}})

	g.Players[game.Player1].Life = 29
	if got := effectivePower(g, source); got != 1 {
		t.Fatalf("power below the life threshold = %d, want 1", got)
	}
	if hasKeyword(g, source, game.Flying) {
		t.Fatal("source gained flying below the life threshold")
	}

	g.Players[game.Player1].Life = 30
	if got := effectivePower(g, source); got != 6 {
		t.Fatalf("power at the life threshold = %d, want 6", got)
	}
	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("source did not gain flying at the life threshold")
	}
}

// TestSourceIsAllColorsSetsFiveColors proves the "<source> is all colors" self
// static shape (a color-layer SetColors of all five colors affecting the source)
// replaces the permanent's printed colors with exactly white, blue, black, red,
// and green.
func TestSourceIsAllColorsSetsFiveColors(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Transguild Courier",
		Colors:    nil,
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerColor,
					AffectedSource: true,
					SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
				},
			},
		}},
	}})

	got := permanentEffectiveColors(g, source)
	slices.Sort(got)
	want := []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("effective colors = %#v, want all five colors", got)
	}
}

// TestPolymorphRemovesAbilitiesAndSetsCharacteristics proves the polymorph
// static shape an Aura lowers ("loses all abilities and is a blue Frog creature
// with base power and toughness 1/1"): the enchanted creature loses its printed
// keyword, ability, color, and creature type, and gains exactly the set color,
// creature type, subtype, and base power/toughness. The grants revert when the
// Aura leaves the battlefield.
func TestPolymorphRemovesAbilitiesAndSetsCharacteristics(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearPT := game.PT{Value: 4}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bear",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(bearPT),
		Toughness: opt.Val(bearPT),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Trample),
		}},
	}})

	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Frogify",
		Colors:   []color.Color{color.Blue},
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:              game.LayerAbility,
					Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
					RemoveAllAbilities: true,
				},
				{
					Layer:     game.LayerColor,
					Group:     game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetColors: []color.Color{color.Blue},
				},
				{
					Layer:       game.LayerType,
					Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetTypes:    []types.Card{types.Creature},
					SetSubtypes: []types.Sub{types.Frog},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetPower:     opt.Val(game.PT{Value: 1}),
					SetToughness: opt.Val(game.PT{Value: 1}),
				},
			},
		}},
	}}
	aura := addCombatPermanent(g, game.Player1, auraDef)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	if hasKeyword(g, creature, game.Trample) {
		t.Fatal("polymorph should remove the printed Trample keyword")
	}
	if abilities := permanentEffectiveAbilities(g, creature); len(abilities) != 0 {
		t.Fatalf("effective abilities = %#v, want none after losing all abilities", abilities)
	}
	if colors := permanentEffectiveColors(g, creature); !slices.Equal(colors, []color.Color{color.Blue}) {
		t.Fatalf("effective colors = %#v, want exactly blue", colors)
	}
	if !permanentHasSubtype(g, creature, types.Frog) {
		t.Fatal("polymorph should set the Frog creature type")
	}
	if permanentHasSubtype(g, creature, types.Bear) {
		t.Fatal("polymorph should remove the printed Bear creature type")
	}
	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1 (base set)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 1 {
		t.Fatalf("effective toughness = %d (ok=%v), want 1", got, ok)
	}

	// When the Aura leaves the battlefield the creature reverts entirely.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("Trample should return after the Aura leaves")
	}
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power after Aura leaves = %d, want 4 (printed)", got)
	}
	if !permanentHasSubtype(g, creature, types.Bear) {
		t.Fatal("Bear creature type should return after the Aura leaves")
	}
	if permanentHasSubtype(g, creature, types.Frog) {
		t.Fatal("Frog creature type should stop applying after the Aura leaves")
	}
	if colors := permanentEffectiveColors(g, creature); !slices.Contains(colors, color.Green) {
		t.Fatalf("colors after Aura leaves = %#v, want to contain printed green", colors)
	}
}

// TestRemovalAuraSetsTypeColorlessAndGrantsMana proves the removal-aura static
// shape ("Enchanted permanent is a colorless land with '{T}: Add {C}' and loses
// all other card types and abilities"): the enchanted creature becomes a
// colorless Land (no longer a creature), loses its printed abilities, and gains
// exactly the granted colorless mana ability. Everything reverts when the Aura
// leaves the battlefield.
func TestRemovalAuraSetsTypeColorlessAndGrantsMana(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearPT := game.PT{Value: 4}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bear",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(bearPT),
		Toughness: opt.Val(bearPT),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Trample),
		}},
	}})

	grantedMana := game.TapManaAbility(mana.C)
	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Imprisoned in the Moon",
		Colors:   []color.Color{color.Blue},
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:              game.LayerAbility,
					Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
					RemoveAllAbilities: true,
				},
				{
					Layer:        game.LayerColor,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetColorless: true,
				},
				{
					Layer:    game.LayerType,
					Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetTypes: []types.Card{types.Land},
				},
				{
					Layer:        game.LayerAbility,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddAbilities: []game.Ability{&grantedMana},
				},
			},
		}},
	}}
	aura := addCombatPermanent(g, game.Player1, auraDef)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	if !permanentHasType(g, creature, types.Land) {
		t.Fatal("removal aura should set the Land card type")
	}
	if permanentHasType(g, creature, types.Creature) {
		t.Fatal("removal aura should remove the printed Creature card type")
	}
	if colors := permanentEffectiveColors(g, creature); len(colors) != 0 {
		t.Fatalf("effective colors = %#v, want colorless (none)", colors)
	}
	if hasKeyword(g, creature, game.Trample) {
		t.Fatal("removal aura should remove the printed Trample keyword")
	}
	abilities := permanentEffectiveAbilities(g, creature)
	if len(abilities) != 1 {
		t.Fatalf("effective abilities = %#v, want exactly the granted mana ability", abilities)
	}
	manaBody, ok := abilities[0].(*game.ManaAbility)
	if !ok || !game.IsTapColorlessManaAbility(manaBody) {
		t.Fatalf("effective ability = %#v, want the granted {T}: Add {C} mana ability", abilities[0])
	}

	// When the Aura leaves the battlefield the creature reverts entirely.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if !permanentHasType(g, creature, types.Creature) {
		t.Fatal("Creature card type should return after the Aura leaves")
	}
	if permanentHasType(g, creature, types.Land) {
		t.Fatal("Land card type should stop applying after the Aura leaves")
	}
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("Trample should return after the Aura leaves")
	}
	if colors := permanentEffectiveColors(g, creature); !slices.Contains(colors, color.Green) {
		t.Fatalf("colors after Aura leaves = %#v, want to contain printed green", colors)
	}
}

// TestRemovalAuraSetsColorlessBasePowerToughness proves the base-P/T removal-aura
// shape ("Enchanted creature loses all abilities and is a colorless <subtype>
// with base power and toughness N/N"): the enchanted creature loses its printed
// abilities, becomes colorless, gains the named creature subtype, and has its
// base power and toughness set. Everything reverts when the Aura leaves.
func TestRemovalAuraSetsColorlessBasePowerToughness(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearPT := game.PT{Value: 4}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bear",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(bearPT),
		Toughness: opt.Val(bearPT),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Trample),
		}},
	}})

	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Noggle the Mind",
		Colors:   []color.Color{color.Blue},
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:              game.LayerAbility,
					Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
					RemoveAllAbilities: true,
				},
				{
					Layer:        game.LayerColor,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetColorless: true,
				},
				{
					Layer:       game.LayerType,
					Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetSubtypes: []types.Sub{types.Noggle},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetPower:     opt.Val(game.PT{Value: 1}),
					SetToughness: opt.Val(game.PT{Value: 1}),
				},
			},
		}},
	}}
	aura := addCombatPermanent(g, game.Player1, auraDef)
	aura.AttachedTo = opt.Val(creature.ObjectID)
	creature.Attachments = append(creature.Attachments, aura.ObjectID)

	if hasKeyword(g, creature, game.Trample) {
		t.Fatal("removal aura should remove the printed Trample keyword")
	}
	if abilities := permanentEffectiveAbilities(g, creature); len(abilities) != 0 {
		t.Fatalf("effective abilities = %#v, want none after losing all abilities", abilities)
	}
	if colors := permanentEffectiveColors(g, creature); len(colors) != 0 {
		t.Fatalf("effective colors = %#v, want colorless (none)", colors)
	}
	if !permanentHasSubtype(g, creature, types.Noggle) {
		t.Fatal("removal aura should set the Noggle creature type")
	}
	if permanentHasSubtype(g, creature, types.Bear) {
		t.Fatal("removal aura should remove the printed Bear creature type")
	}
	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1 (base set)", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 1 {
		t.Fatalf("effective toughness = %d (ok=%v), want 1", got, ok)
	}

	// When the Aura leaves the battlefield the creature reverts entirely.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("Trample should return after the Aura leaves")
	}
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power after Aura leaves = %d, want 4 (printed)", got)
	}
	if !permanentHasSubtype(g, creature, types.Bear) {
		t.Fatal("Bear creature type should return after the Aura leaves")
	}
	if colors := permanentEffectiveColors(g, creature); !slices.Contains(colors, color.Green) {
		t.Fatalf("colors after Aura leaves = %#v, want to contain printed green", colors)
	}
}
