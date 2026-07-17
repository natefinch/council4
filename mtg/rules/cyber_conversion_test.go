package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func cyberConversionInstructions() []game.Instruction {
	return []game.Instruction{{Primitive: game.TurnFaceDown{
		Object: game.TargetPermanentReference(0),
		Characteristics: opt.Val(game.FaceDownCharacteristics{
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Cyberman},
			Power:     game.PT{Value: 2},
			Toughness: game.PT{Value: 2},
		}),
	}}}
}

func applyCyberConversion(g *game.Game, permanent *game.Permanent) {
	engine := NewEngine(nil)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(permanent.ObjectID)},
	}
	for _, instruction := range cyberConversionInstructions() {
		resolveInstruction(engine, g, obj, instruction.Primitive, &TurnLog{})
	}
}

func cyberConversionCreature(morphCost cost.Mana) *game.CardDef {
	pt := opt.Val(game.PT{Value: 4})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:       "Ruby Skybeast",
		ManaCost:   opt.Val(cost.Mana{cost.R, cost.R, cost.R}),
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dragon},
		Power:      pt,
		Toughness:  pt,
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}},
	}}
	if morphCost != nil {
		def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: morphCost}},
		})
	}
	return def
}

func TestCyberConversionFaceDownCharacteristicsAndStatePersistence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player2, cyberConversionCreature(nil))
	attachment := addCombatPermanent(g, game.Player1, vanillaCreature("Attachment", 1, 1))
	permanent.Attachments = []game.ObjectID{attachment.ObjectID}
	attachment.AttachedTo = opt.Val(permanent.ObjectID)
	permanent.Counters.Add(counter.PlusOnePlusOne, 1)
	permanent.MarkedDamage = 1
	permanent.Tapped = true
	permanent.Flipped = true
	originalID := permanent.CardInstanceID

	applyCyberConversion(g, permanent)

	if !permanent.FaceDown || permanent.FaceDownKind != game.FaceDownEffect ||
		permanent.CardInstanceID != originalID ||
		effectiveController(g, permanent) != game.Player2 ||
		!permanent.Tapped ||
		!permanent.Flipped ||
		permanent.MarkedDamage != 1 ||
		permanent.Counters.Get(counter.PlusOnePlusOne) != 1 ||
		!slices.Equal(permanent.Attachments, []game.ObjectID{attachment.ObjectID}) ||
		!attachment.AttachedTo.Exists || attachment.AttachedTo.Val != permanent.ObjectID {
		t.Fatalf("persistent state changed: %+v attachment=%+v", permanent, attachment)
	}
	toughness, toughnessOK := effectiveToughness(g, permanent)
	if permanentEffectiveName(g, permanent) != "" ||
		len(permanentEffectiveColors(g, permanent)) != 0 ||
		len(permanentEffectiveAbilities(g, permanent)) != 0 ||
		permanentHasSupertype(g, permanent, types.Legendary) ||
		!permanentHasType(g, permanent, types.Artifact) ||
		!permanentHasType(g, permanent, types.Creature) ||
		permanentHasType(g, permanent, types.Enchantment) ||
		!permanentHasSubtype(g, permanent, types.Cyberman) ||
		permanentHasSubtype(g, permanent, types.Dragon) ||
		effectivePower(g, permanent) != 3 ||
		!toughnessOK || toughness != 3 {
		t.Fatalf("effective characteristics name=%q colors=%v abilities=%d power/toughness=%d/%d",
			permanentEffectiveName(g, permanent), permanentEffectiveColors(g, permanent),
			len(permanentEffectiveAbilities(g, permanent)), effectivePower(g, permanent), toughness)
	}
	copyDef, ok := permanentCopyDef(g, permanent)
	if !ok ||
		copyDef.Name != "" ||
		copyDef.ManaValue() != 0 ||
		!slices.Equal(copyDef.Types, []types.Card{types.Artifact, types.Creature}) ||
		!slices.Equal(copyDef.Subtypes, []types.Sub{types.Cyberman}) ||
		len(copyDef.Colors) != 0 ||
		len(copyDef.StaticAbilities) != 0 ||
		!copyDef.Power.Exists || copyDef.Power.Val.Value != 2 {
		t.Fatalf("face-down copiable values = %+v, ok=%v", copyDef, ok)
	}
}

func TestCyberConversionTurnFaceUpPermissionsAndEffectEnding(t *testing.T) {
	t.Run("effect allows inherent morph and ends on turn up", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addCombatPermanent(g, game.Player1, cyberConversionCreature(cost.Mana{cost.G}))
		applyCyberConversion(g, permanent)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("morph permanent could not turn face up")
		}
		if permanent.FaceDown ||
			permanentEffectiveName(g, permanent) != "Ruby Skybeast" ||
			!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Red}) ||
			!hasKeyword(g, permanent, game.Flying) ||
			permanentHasType(g, permanent, types.Artifact) ||
			permanentHasSubtype(g, permanent, types.Cyberman) ||
			!permanentHasSubtype(g, permanent, types.Dragon) ||
			effectivePower(g, permanent) != 4 {
			t.Fatal("face-up converted characteristics are wrong")
		}
		copyDef, ok := permanentCopyDef(g, permanent)
		if !ok || copyDef.Name != "Ruby Skybeast" || copyDef.ManaValue() != 3 ||
			!slices.Equal(copyDef.Types, []types.Card{types.Creature}) ||
			!slices.Equal(copyDef.Subtypes, []types.Sub{types.Dragon}) {
			t.Fatalf("face-up copiable values = %+v, ok=%v", copyDef, ok)
		}
	})

	t.Run("effect grants no printed mana-cost permission", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addCombatPermanent(g, game.Player1, cyberConversionCreature(nil))
		applyCyberConversion(g, permanent)
		g.Players[game.Player1].ManaPool.Add(mana.R, 3)
		g.Turn.PriorityPlayer = game.Player1
		if engine.canTurnFaceUp(g, game.Player1, permanent.ObjectID) {
			t.Fatal("ordinary card gained a turn-face-up permission")
		}
	})

	t.Run("effect allows inherent disguise", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addCombatPermanent(g, game.Player1, disguiseCreature(cost.Mana{cost.W}))
		applyCyberConversion(g, permanent)
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		g.Turn.PriorityPlayer = game.Player1
		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("disguise permanent could not turn face up")
		}
	})

	t.Run("existing manifest and disguise provenance survives", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		manifested := addFaceDownPermanent(g, game.Player1, manifestCreature(cost.Mana{cost.G}), game.FaceDownManifest)
		disguised := addFaceDownPermanent(g, game.Player2, disguiseCreature(cost.Mana{cost.W}), game.FaceDownDisguise)
		cloaked := addFaceDownPermanent(g, game.Player3, manifestCreature(cost.Mana{cost.G}), game.FaceDownCloak)
		applyCyberConversion(g, manifested)
		applyCyberConversion(g, disguised)
		applyCyberConversion(g, cloaked)
		if manifested.FaceDownKind != game.FaceDownManifest ||
			disguised.FaceDownKind != game.FaceDownDisguise ||
			cloaked.FaceDownKind != game.FaceDownCloak ||
			!hasKeyword(g, disguised, game.Ward) ||
			!hasKeyword(g, cloaked, game.Ward) {
			t.Fatal("existing face-down provenance or ward was lost")
		}
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1
		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(manifested.ObjectID)) {
			t.Fatal("manifest mana-cost turn-up permission was lost")
		}
		if permanentHasType(g, manifested, types.Artifact) ||
			permanentHasSubtype(g, manifested, types.Cyberman) {
			t.Fatal("an already face-down permanent received Cyberman characteristics")
		}
	})
}

func TestCyberConversionFaceDownOverridesCopyAndCopiedMorphTurnsUp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, cyberConversionCreature(nil))
	applyCyberConversion(g, permanent)
	copiedPT := game.PT{Value: 5}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerCopy,
		CopyValues: opt.Val(game.CopyableValues{
			Name:      "Copied Morph",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(copiedPT),
			Toughness: opt.Val(copiedPT),
			Abilities: []game.Ability{&game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.G}}},
			}},
		}),
	})

	if permanentEffectiveName(g, permanent) != "" ||
		!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{}) ||
		!permanentHasType(g, permanent, types.Artifact) ||
		!permanentHasSubtype(g, permanent, types.Cyberman) ||
		permanentHasSubtype(g, permanent, types.Beast) ||
		len(permanentEffectiveAbilities(g, permanent)) != 0 ||
		effectivePower(g, permanent) != 2 {
		t.Fatal("copy effect exposed values through face-down characteristics")
	}
	copyDef, ok := permanentCopyDef(g, permanent)
	if !ok ||
		copyDef.Name != "" ||
		!slices.Equal(copyDef.Types, []types.Card{types.Artifact, types.Creature}) ||
		!slices.Equal(copyDef.Subtypes, []types.Sub{types.Cyberman}) {
		t.Fatalf("face-down copy values = %+v, ok=%v", copyDef, ok)
	}

	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
		t.Fatal("copied morph cost could not turn the permanent face up")
	}
	if permanentEffectiveName(g, permanent) != "Copied Morph" ||
		!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Green}) ||
		!permanentHasSubtype(g, permanent, types.Beast) ||
		permanentHasSubtype(g, permanent, types.Cyberman) ||
		effectivePower(g, permanent) != 5 {
		t.Fatal("face-up copied characteristics are wrong")
	}
}

func TestCyberConversionTokensDoubleFacedAndMergedPermanents(t *testing.T) {
	t.Run("tokens use their normal turn-up permissions", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		ordinary, ok := createTokenPermanent(g, game.Player1, cyberConversionCreature(nil))
		if !ok {
			t.Fatal("create ordinary token")
		}
		morph, ok := createTokenPermanent(g, game.Player1, cyberConversionCreature(cost.Mana{cost.G}))
		if !ok {
			t.Fatal("create morph token")
		}
		applyCyberConversion(g, ordinary)
		applyCyberConversion(g, morph)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1
		if !ordinary.FaceDown || engine.canTurnFaceUp(g, game.Player1, ordinary.ObjectID) {
			t.Fatal("ordinary face-down token gained a turn-up permission")
		}
		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(morph.ObjectID)) ||
			morph.FaceDown ||
			permanentHasSubtype(g, morph, types.Cyberman) ||
			effectivePower(g, morph) != 4 {
			t.Fatal("morph token could not use its inherent turn-up permission")
		}
	})

	t.Run("double-faced permanent remains face up but is converted", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		def := cyberConversionCreature(nil)
		def.Layout = game.LayoutTransform
		def.Back = opt.Val(game.CardFace{Name: "Back", Types: []types.Card{types.Creature}})
		permanent := addCombatPermanent(g, game.Player2, def)
		applyCyberConversion(g, permanent)
		if permanent.FaceDown ||
			permanentEffectiveName(g, permanent) != "Ruby Skybeast" ||
			permanentHasType(g, permanent, types.Artifact) ||
			permanentHasSubtype(g, permanent, types.Cyberman) ||
			effectivePower(g, permanent) != 4 {
			t.Fatal("double-faced permanent handling is wrong")
		}
	})

	t.Run("merged component morph permission and reveal", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addCombatPermanent(g, game.Player1, vanillaCreature("Top", 3, 3))
		lowerID := addCardInstance(g, game.Player1, morphCreature(cost.Mana{cost.G}))
		permanent.MergedCards = []game.MergedCard{{
			CardInstanceID: lowerID,
			Face:           game.FaceFront,
			Owner:          game.Player1,
		}}
		applyCyberConversion(g, permanent)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1
		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("merged permanent could not use a visible component's morph cost")
		}
		assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
			return event.CardID == lowerID
		})
	})
}

func TestCyberConversionObservationCommanderAndZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player3, cyberConversionCreature(nil))
	permanent.Controller = game.Player2
	lowerCommanderID := addCardInstance(g, game.Player2, vanillaCreature("Merged Commander", 5, 5))
	g.Players[game.Player2].CommanderInstanceID = lowerCommanderID
	permanent.MergedCards = []game.MergedCard{{
		CardInstanceID: lowerCommanderID,
		Face:           game.FaceFront,
		Owner:          game.Player2,
	}}
	applyCyberConversion(g, permanent)

	for _, observer := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
		view := findPermanentView(t, observe(g, observer), permanent.ObjectID)
		if !view.FaceDown ||
			view.Name != "" ||
			view.HasKeyword(game.Flying) ||
			!slices.Equal(view.Types, []types.Card{types.Artifact, types.Creature}) ||
			!slices.Equal(view.Subtypes, []types.Sub{types.Cyberman}) ||
			!slices.Contains(view.CommanderInstanceIDs, lowerCommanderID) {
			t.Fatalf("observer %v view = %+v", observer, view)
		}
		if observer == game.Player2 {
			if len(view.FaceDownCards) != 2 ||
				view.FaceDownCards[0].Name != "Ruby Skybeast" ||
				view.FaceDownCards[1].Name != "Merged Commander" {
				t.Fatalf("controller private face-down cards = %+v", view.FaceDownCards)
			}
		} else if len(view.FaceDownCards) != 0 {
			t.Fatalf("opponent saw private face-down cards = %+v", view.FaceDownCards)
		}
	}

	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("move converted permanent to graveyard")
	}
	revealed := map[game.ObjectID]bool{}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			revealed[event.CardID] = true
		}
	}
	if !revealed[permanent.CardInstanceID] || !revealed[lowerCommanderID] {
		t.Fatalf("revealed cards = %v, want top and merged components", revealed)
	}
	if len(g.Battlefield) != 0 ||
		!g.Players[game.Player3].Graveyard.Contains(permanent.CardInstanceID) ||
		!g.Players[game.Player2].CommandZone.Contains(lowerCommanderID) {
		t.Fatal("zone change did not create independent destination objects")
	}
}

func TestCyberConversionTargetLegalityAndFizzle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Target", 2, 2))
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Cyber Conversion Test",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
					}),
				}},
				Sequence: cyberConversionInstructions(),
			}.Ability()),
		}},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	})
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("remove target before resolution")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("fizzled spell installed %d continuous effects", len(g.ContinuousEffects))
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("fizzled instant did not go to its owner's graveyard")
	}
}
