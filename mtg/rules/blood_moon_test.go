package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const bloodMoonID = "94fac5fe-97d5-4c12-a80c-8efff9d853ae"

func bloodMoonDef(t *testing.T) *game.CardDef {
	t.Helper()
	return lifetimeDefinition(t, cards.NewDefaultRegistry(), bloodMoonID)
}

func addBloodMoon(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	return addCombatPermanent(g, game.Player1, bloodMoonDef(t))
}

func moonLandDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Ancient Observatory", Supertypes: []types.Super{types.Legendary, types.Snow},
		Types: []types.Card{types.Artifact, types.Land}, Subtypes: []types.Sub{types.Island, types.Locus},
		ManaAbilities:   []game.ManaAbility{game.TapManaAbility(mana.U)},
		StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text: "You gain 1 life.", Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func assertMoonLand(t *testing.T, g *game.Game, land *game.Permanent) {
	t.Helper()
	values := effectivePermanentValues(g, land)
	if !slices.Contains(values.subtypes, types.Mountain) || slices.Contains(values.subtypes, types.Island) ||
		slices.Contains(values.subtypes, types.Locus) || slices.Contains(values.supertypes, types.Basic) ||
		values.name != "Ancient Observatory" {
		t.Fatalf("land identity/types under Moon: %+v", values)
	}
	_, colors := abilitiesManaProduction(values.abilities, nil, nil)
	if !slices.Equal(colors, []color.Color{color.Red}) {
		t.Fatalf("mana colors = %v, want only red", colors)
	}
}

func TestBloodMoonProductionCastingSimulatorObservationAndPayment(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	land := addCombatPermanent(g, game.Player1, moonLandDef())
	basic := addBasicLandWithManaPermanent(g, game.Player2, "Snow-Covered Island", types.Island, mana.U)
	basicDef, _ := permanentCardDef(g, basic)
	basicDef.Supertypes = append(basicDef.Supertypes, types.Snow)
	card := addCardToHand(g, game.Player1, lifetimeDefinition(t, registry, bloodMoonID))
	hand := NewObservation(g, game.Player1).Hand()
	if len(hand) != 1 || hand[0].Name != "Blood Moon" || hand[0].ManaValue != 3 {
		t.Fatal("production card/cost missing from observation")
	}
	def := bloodMoonDef(t)
	effect := def.StaticAbilities[0].ContinuousEffects[0]
	if def.StaticAbilities[0].ZoneOfFunction != zone.Battlefield || effect.Layer != game.LayerType ||
		effect.Group.Selection().ExcludedSupertype != types.Basic ||
		!slices.Equal(effect.SetBasicLandTypes, []types.Sub{types.Mountain}) {
		t.Fatal("inspectable shared Selection/layer/SET declaration missing")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	cast := action.CastSpell(card, nil, 0, nil)
	if !containsAction(e.legalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("production Blood Moon could not be legally cast")
	}
	before := g.Clone()
	branch := e.Simulator().ResolvePriority(before, [game.NumPlayers]PlayerAgent{})
	branchLand, _ := permanentByObjectID(branch, land.ObjectID)
	assertMoonLand(t, branch, branchLand)
	if permanentByCardID(g, card) != nil || !permanentHasSubtype(g, land, types.Island) {
		t.Fatal("Simulator mutated original casting state")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	assertMoonLand(t, g, land)
	basicView := findPermanentView(t, NewObservation(g, game.Player1), basic.ObjectID)
	if !slices.Contains(basicView.Subtypes, types.Island) || slices.Contains(basicView.Subtypes, types.Mountain) ||
		!slices.Equal(basicView.ProducesColors, []color.Color{color.Blue}) {
		t.Fatal("basic land changed")
	}
	view := findPermanentView(t, NewObservation(g, game.Player1), land.ObjectID)
	if !slices.Contains(view.Subtypes, types.Mountain) || !slices.Equal(view.ProducesColors, []color.Color{color.Red}) ||
		hasKeyword(g, land, game.Indestructible) {
		t.Fatal("effective observations retained old abilities/types")
	}
	for _, body := range permanentEffectiveAbilitiesView(g, land) {
		if _, ok := body.(*game.ActivatedAbility); ok {
			t.Fatal("old non-mana activation survived SET")
		}
	}
	red := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Red payment witness", Types: []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{cost.R}),
	}})
	if !e.applyAction(g, game.Player1, action.CastSpell(red, nil, 0, nil)) || !land.Tapped {
		t.Fatal("Engine payment did not use the new Mountain mana ability")
	}
}

func TestBloodMoonExternalGrantsAnimatedLandAndManaLoss(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		t.Run(map[bool]string{false: "later animation", true: "earlier animation"}[earlier], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			land := addCombatPermanent(g, game.Player1, moonLandDef())
			green := game.TapManaAbility(mana.G)
			timestamp := game.Timestamp(100000)
			if earlier {
				timestamp = 1
			}
			g.ContinuousEffects = append(g.ContinuousEffects,
				game.ContinuousEffect{ID: g.IDGen.Next(), Timestamp: timestamp,
					AffectedObjectID: land.ObjectID, Layer: game.LayerType,
					AddTypes: []types.Card{types.Creature}, AddSubtypes: []types.Sub{types.Elf}},
				game.ContinuousEffect{ID: g.IDGen.Next(), Timestamp: timestamp,
					AffectedObjectID: land.ObjectID, Layer: game.LayerPowerToughnessSet,
					SetPower: opt.Val(game.PT{Value: 4}), SetToughness: opt.Val(game.PT{Value: 5})},
				game.ContinuousEffect{ID: g.IDGen.Next(), Timestamp: 1,
					AffectedObjectID: land.ObjectID, Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Flying},
					AddAbilities: []game.Ability{&green}},
			)
			addBloodMoon(t, g)
			values := effectivePermanentValues(g, land)
			if !slices.Contains(values.types, types.Creature) || !slices.Contains(values.types, types.Artifact) ||
				!slices.Contains(values.types, types.Land) || !slices.Contains(values.subtypes, types.Elf) ||
				!slices.Contains(values.supertypes, types.Legendary) || !slices.Contains(values.supertypes, types.Snow) ||
				values.power != 4 || values.toughness != 5 || !values.keywords.has(game.Flying) ||
				values.keywords.has(game.Indestructible) {
				t.Fatal("SET erased animation, unrelated types/supertypes or external grants")
			}
			_, colors := abilitiesManaProduction(values.abilities, nil, nil)
			if !slices.Contains(colors, color.Red) || !slices.Contains(colors, color.Green) || slices.Contains(colors, color.Blue) {
				t.Fatalf("mana colors = %v; SET should preserve external green but not intrinsic blue", colors)
			}
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), Timestamp: 200000, AffectedObjectID: land.ObjectID,
				Layer: game.LayerAbility, RemoveAllAbilities: true,
			})
			if len(permanentEffectiveAbilitiesView(g, land)) != 0 || hasKeyword(g, land, game.Flying) {
				t.Fatal("type-derived mana/earlier external grants reappeared after layer-6 ability loss")
			}
		})
	}
}

func compileUrborg(t *testing.T) *game.CardDef {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Urborg, Tomb of Yawgmoth", Layout: "normal", TypeLine: "Legendary Land",
		OracleText: "Each land is a Swamp in addition to its other land types.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("retained supported Urborg: %v, %v", err, diagnostics)
	}
	return defs[0]
}

func TestBloodMoonUrborgDependencyAndSourceRestoration(t *testing.T) {
	for _, moonFirst := range []bool{false, true} {
		for _, renamed := range []bool{false, true} {
			t.Run(map[bool]string{false: "Urborg first", true: "Moon first"}[moonFirst]+
				map[bool]string{false: "/named", true: "/renamed"}[renamed], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				basic := addBasicLandWithManaPermanent(g, game.Player2, "Island", types.Island, mana.U)
				land := addCombatPermanent(g, game.Player1, moonLandDef())
				urborgDef := compileUrborg(t)
				if renamed {
					urborgDef.Name = "Anonymous swamp source"
				}
				var moon *game.Permanent
				if moonFirst {
					moon = addBloodMoon(t, g)
				}
				urborg := addCombatPermanent(g, game.Player2, urborgDef)
				if !moonFirst {
					if !permanentHasSubtype(g, basic, types.Swamp) || !permanentHasSubtype(g, land, types.Island) {
						t.Fatal("ADD control did not retain old types or affect basics")
					}
					moon = addBloodMoon(t, g)
				}
				assertMoonLand(t, g, land)
				if permanentHasSubtype(g, basic, types.Swamp) || permanentHasSubtype(g, urborg, types.Swamp) ||
					!permanentHasSubtype(g, urborg, types.Mountain) || len(permanentEffectiveAbilitiesView(g, urborg)) != 1 {
					t.Fatal("dependency did not eliminate Urborg before it could apply")
				}
				clone := g.Clone()
				movePermanentToZone(g, moon, zone.Exile)
				if !permanentHasSubtype(g, basic, types.Swamp) || !permanentHasSubtype(g, land, types.Island) ||
					!permanentHasSubtype(g, land, types.Swamp) || !hasKeyword(g, land, game.Indestructible) {
					t.Fatal("Moon departure did not restore Urborg ADD or original abilities")
				}
				clonedLand, _ := permanentByObjectID(clone, land.ObjectID)
				assertMoonLand(t, clone, clonedLand)
			})
		}
	}
}

func TestBloodMoonCopiesAndAbilityLossRespectLayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, moonLandDef())
	moon := addBloodMoon(t, g)
	copiedSource := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Physical copy", Types: []types.Card{types.Artifact}}})
	becomeStaticCopy(t, g, copiedSource, moon, false)
	movePermanentToZone(g, moon, zone.Exile)
	assertMoonLand(t, g, land)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 100000, AffectedObjectID: copiedSource.ObjectID,
		Layer: game.LayerAbility, RemoveAllAbilities: true,
	})
	if len(permanentEffectiveAbilitiesView(g, copiedSource)) != 0 {
		t.Fatal("copied source did not actually lose its ability")
	}
	// CR 613.6: layer 6 cannot stop a type-changing effect already begun in 4.
	assertMoonLand(t, g, land)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 200000, AffectedObjectID: copiedSource.ObjectID,
		Layer: game.LayerCopy, CopyValues: opt.Val(copyableValuesFromDef(&game.CardDef{CardFace: game.CardFace{
			Name: "Blank copied face", Types: []types.Card{types.Artifact},
		}})),
	})
	if !permanentHasSubtype(g, land, types.Island) || permanentHasSubtype(g, land, types.Mountain) {
		t.Fatal("overwriting source abilities in layer 1 did not stop the declaration")
	}
}

func TestBloodMoonLandCopyStripsCopiedAbilitiesButKeepsCopiedTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Copy host", Types: []types.Card{types.Land},
	}})
	blueprint := addCombatPermanent(g, game.Player2, moonLandDef())
	becomeStaticCopy(t, g, land, blueprint, false)
	addBloodMoon(t, g)
	assertMoonLand(t, g, land)
	if hasKeyword(g, land, game.Indestructible) || !permanentHasSupertype(g, land, types.Legendary) ||
		!permanentHasType(g, land, types.Artifact) {
		t.Fatal("SET did not strip copied abilities while retaining copied types")
	}
}

func TestBloodMoonCopiedLandSourceDoesNotTruncateItsBegunEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "First physical object", Types: []types.Card{types.Artifact},
	}})
	land := addCombatPermanent(g, game.Player1, moonLandDef())
	values := copyableValuesFromDef(bloodMoonDef(t))
	values.Types = []types.Card{types.Land, types.Enchantment}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: source.Timestamp(), Layer: game.LayerCopy,
		AffectedObjectID: source.ObjectID, CopyValues: opt.Val(values),
	})
	if len(permanentEffectiveAbilitiesView(g, source)) != 1 || !permanentHasSubtype(g, source, types.Mountain) {
		t.Fatal("land copy source did not lose its copied Moon declaration")
	}
	assertMoonLand(t, g, land)
}

func TestBloodMoonExternalEntryReplacementAndResolvedReplacementPersist(t *testing.T) {
	g, e, _ := lifetimeGame(t)
	replacement := game.EntersTappedGroupReplacement("Other lands enter tapped.", game.TriggerControllerAny, types.Land)
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Land replacement source", Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{replacement},
	}})
	registerPermanentReplacementEffects(g, source)
	enchantment := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "External replacement", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{replacement},
	}})
	registerPermanentReplacementEffects(g, enchantment)
	moon := addBloodMoon(t, g)
	enter := func() *game.Permanent {
		cardID := addCardToHand(g, game.Player1, moonLandDef())
		card, _ := g.GetCardInstance(cardID)
		g.Players[game.Player1].Hand.Remove(cardID)
		land, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
		if !ok {
			t.Fatal("land entry failed")
		}
		return land
	}
	if !enter().Tapped {
		t.Fatal("Moon erased an external enters-tapped replacement")
	}
	movePermanentToZone(g, enchantment, zone.Exile)
	if enter().Tapped {
		t.Fatal("erased intrinsic land replacement continued to affect entrants")
	}
	runtime := replacement.Replacement
	runtime.Duration = game.DurationPermanent
	resolveInstruction(e, g, &game.StackObject{
		Kind: game.StackActivatedAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
	}, game.CreateReplacement{Replacement: &runtime}, nil)
	if !enter().Tapped {
		t.Fatal("Moon incorrectly ended a replacement created by prior resolution")
	}
	g.ReplacementEffects = slices.Delete(g.ReplacementEffects, len(g.ReplacementEffects)-1, len(g.ReplacementEffects))
	movePermanentToZone(g, moon, zone.Exile)
	if !enter().Tapped {
		t.Fatal("intrinsic registered replacement did not return after Moon departure")
	}
}

func TestBloodMoonMultipleSetBasicTypesAndSameSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, moonLandDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: land.ObjectID,
		SetBasicLandTypes: []types.Sub{types.Island, types.Mountain},
	})
	values := effectivePermanentValues(g, land)
	_, colors := abilitiesManaProduction(values.abilities, nil, nil)
	if len(values.abilities) != 2 || !slices.Contains(colors, color.Red) || !slices.Contains(colors, color.Blue) ||
		slices.Contains(values.subtypes, types.Locus) || values.keywords.has(game.Indestructible) {
		t.Fatal("setting a retained basic subtype omitted mana, duplicated it or kept old text")
	}
}

func TestBloodMoonEntryReplacementsTriggersAndDeparture(t *testing.T) {
	g, e, _ := lifetimeGame(t)
	moon := addBloodMoon(t, g)
	def := moonLandDef()
	def.ReplacementAbilities = []game.ReplacementAbility{game.EntersTappedReplacement("This land enters tapped."),
		{Text: "As this land enters, choose a color.", Replacement: game.ReplacementEffect{
			MatchEvent: game.EventPermanentEnteredBattlefield, EntryColorChoice: true,
		}}}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "When this land enters, you gain 1 life.",
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}}
	cardID := addCardToHand(g, game.Player1, def)
	if !e.applyAction(g, game.Player1, action.PlayLand(cardID)) {
		t.Fatal("land could not enter while Moon was active")
	}
	land := permanentByCardID(g, cardID)
	assertMoonLand(t, g, land)
	if land.Tapped || len(land.EntryChoices) != 0 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("intrinsic entry replacement/choice/trigger survived Moon")
	}
	movePermanentToZone(g, moon, zone.Graveyard)
	if !permanentHasSubtype(g, land, types.Island) || !hasKeyword(g, land, game.Indestructible) || land.Tapped {
		t.Fatal("departure did not restore abilities without retroactive entry replacement")
	}
	movePermanentToZone(g, land, zone.Hand)
	g.Turn.LandsPlayedThisTurn = 0
	if !e.applyAction(g, game.Player1, action.PlayLand(cardID)) {
		t.Fatal("land failed to reenter without Moon")
	}
	land = permanentByCardID(g, cardID)
	if !land.Tapped || len(land.EntryChoices) == 0 || !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("intrinsic entry behavior control failed after Moon departure")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 {
		t.Fatal("restored ETB trigger did not resolve")
	}
}

func TestBloodMoonTypeBegunAnimationContinuesAfterIntrinsicAbilityLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Self-animating land", Types: []types.Card{types.Land},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerType, AffectedSource: true, AddTypes: []types.Card{types.Creature}},
			{Layer: game.LayerPowerToughnessSet, AffectedSource: true,
				SetPower: opt.Val(game.PT{Value: 3}), SetToughness: opt.Val(game.PT{Value: 3})},
		}}},
	}})
	addBloodMoon(t, g)
	// Dependency on the SET removes an unstarted intrinsic land static.
	if permanentHasType(g, land, types.Creature) {
		t.Fatal("unstarted type static on a nonbasic land survived source erasure")
	}
	// Basic supertype keeps the source out of Moon's group. Its begun effect
	// then continues into 7b even if a layer-6 effect removes its ability.
	landDef, _ := permanentCardDef(g, land)
	landDef.Supertypes = []types.Super{types.Basic}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 100000, AffectedObjectID: land.ObjectID,
		Layer: game.LayerAbility, RemoveAllAbilities: true,
	})
	if !permanentHasType(g, land, types.Creature) || effectivePower(g, land) != 3 ||
		len(permanentEffectiveAbilitiesView(g, land)) != 0 {
		t.Fatal("CR 613.6 did not continue the begun animation into 7b")
	}
}

func TestBloodMoonFrameSnapshotSurvivesLaterColorMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Two-color witness", Types: []types.Card{types.Artifact},
		Colors: []color.Color{color.Red, color.Blue},
	}})
	addBloodMoon(t, g)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: permanent.ObjectID, Layer: game.LayerColor,
		RemoveColors: []color.Color{color.Red},
	})
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	if !slices.Equal(effectivePermanentValues(g, permanent).colors, []color.Color{color.Blue}) {
		t.Fatal("later color removal failed")
	}
	if !slices.Equal(permanentValuesBeforeLayer(g, permanent, game.LayerColor).colors, []color.Color{color.Red, color.Blue}) {
		t.Fatal("later color removal mutated the cached earlier-layer snapshot")
	}
}

func TestBloodMoonFrameDirectRecomputationAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	lands := []*game.Permanent{
		addBasicLandWithManaPermanent(g, game.Player1, "Island", types.Island, mana.U),
		addCombatPermanent(g, game.Player2, compileUrborg(t)),
		addCombatPermanent(g, game.Player3, moonLandDef()),
	}
	moon := addBloodMoon(t, g)
	expected := make(map[id.ID]permanentEffectiveValues)
	for _, land := range lands {
		expected[land.ObjectID] = effectivePermanentValues(g, land)
	}
	g.BeginStaticSourceFrame()
	for i := len(lands) - 1; i >= 0; i-- {
		land := lands[i]
		if !reflect.DeepEqual(effectivePermanentValues(g, land), expected[land.ObjectID]) {
			t.Fatal("framed result diverged from direct uncached recomputation")
		}
	}
	g.EndStaticSourceFrame()
	movePermanentToZone(g, moon, zone.Exile)
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	if !permanentHasSubtype(g, lands[0], types.Swamp) || !permanentHasSubtype(g, lands[2], types.Island) {
		t.Fatal("new frame reused stale land type results after source departure")
	}
}
