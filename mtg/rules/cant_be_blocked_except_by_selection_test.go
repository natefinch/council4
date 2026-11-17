package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-11-06 (verified copy retained in
// the unit's .cardwork/cr-20261106.txt), read 509.1b (a block declaration that
// disobeys a restriction is illegal; evasion abilities restrict what can block
// the attacker and are cumulative), 205.3m (creature types) and 702.73a
// (changeling: the object is every creature type). Effect shapes (pinned corpus
// and current Scryfall Oracle text, retained in .cardwork/card-*.json):
// Deathcult Rogue ("This creature can't be blocked except by Rogues."), Amrou
// Seekers ("... except by artifact creatures and/or white creatures."), Skirk
// Shaman (".. and/or red creatures."), Elven Riders ("... except by Walls
// and/or creatures with flying."), Invisibility ("Enchanted creature can't be
// blocked except by Walls."), Prowler's Helm ("Equipped creature can't be
// blocked except by Walls."), Shifting Sliver ("Slivers can't be blocked except
// by Slivers."), Serpent of Yawning Depths ("Krakens, Leviathans, Octopuses,
// and Serpents you control can't be blocked except by Krakens, Leviathans,
// Octopuses, and Serpents."), Departed Deckhand ("{3}{U}: Another target
// creature you control can't be blocked this turn except by Spirits.").
// Rulings (retained in .cardwork/rulings-*.json): Deathcult Rogue 2013-01-24 —
// a blocker that stops being a Rogue after blocks are declared keeps blocking;
// Elven Riders 2007-05-01 — reach isn't flying, so a reach creature can't
// block it; Shifting Sliver 2013-07-01 — a Sliver that stops being a Sliver
// isn't affected; Departed Deckhand 2018-07-13 — activating after a non-Spirit
// has blocked doesn't remove the blocker.

func exceptBySelection(sel game.Selection) game.BlockerRestriction {
	return game.BlockerRestriction{Kind: game.BlockerRestrictionSelection, Selection: sel}
}

func exceptBySubtypes(subtypes ...types.Sub) game.BlockerRestriction {
	return exceptBySelection(game.Selection{SubtypesAny: subtypes})
}

// addSubtypedCombatCreature returns a power-2 creature with the given creature
// subtypes, colors and keywords.
func addSubtypedCombatCreature(g *game.Game, controller game.PlayerID, subtypes []types.Sub, colors []color.Color, keywords ...game.Keyword) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Typed Creature",
		Types:           []types.Card{types.Creature},
		Subtypes:        subtypes,
		Colors:          colors,
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keywords...)}},
	}})
}

// addSelfExceptByCreature declares a creature whose own static ability says
// "This creature can't be blocked except by ...".
func addSelfExceptByCreature(g *game.Game, controller game.PlayerID, subtypes []types.Sub, restriction game.BlockerRestriction) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Except-By Creature",
		Types:     []types.Card{types.Creature},
		Subtypes:  subtypes,
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBeBlockedExceptBy,
			AffectedSource:     true,
			Duration:           game.DurationPermanent,
			BlockerRestriction: restriction,
		}}}},
	}})
}

// offeredBlockers returns the blockers that the legal declare-blockers actions
// offer against attacker.
func offeredBlockers(t *testing.T, g *game.Game, attacker *game.Permanent) []id.ID {
	t.Helper()
	var offered []id.ID
	for _, act := range (combatEngine{}).legalBlockers(g, game.Player2) {
		payload, ok := act.DeclareBlockersPayload()
		if !ok {
			t.Fatalf("legal blocker action %+v has no declare-blockers payload", act)
		}
		for _, block := range payload.Blockers {
			if block.Blocking == attacker.ObjectID && !slices.Contains(offered, block.Blocker) {
				offered = append(offered, block.Blocker)
			}
		}
	}
	return offered
}

// checkExceptByBlockers asserts that exactly the allowed blockers can block
// attacker, both in the offered legal actions and in submitted declarations.
func checkExceptByBlockers(t *testing.T, g *game.Game, attacker *game.Permanent, allowed, prohibited map[string]*game.Permanent) {
	t.Helper()
	offered := offeredBlockers(t, g, attacker)
	for name, blocker := range allowed {
		if !slices.Contains(offered, blocker.ObjectID) {
			t.Errorf("%s was not offered as a blocker", name)
		}
		if !blocksLegal(g, blockPairs([2]*game.Permanent{blocker, attacker})) {
			t.Errorf("%s's block was rejected", name)
		}
	}
	for name, blocker := range prohibited {
		if slices.Contains(offered, blocker.ObjectID) {
			t.Errorf("%s was offered as a blocker", name)
		}
		if blocksLegal(g, blockPairs([2]*game.Permanent{blocker, attacker})) {
			t.Errorf("%s's block was accepted", name)
		}
	}
}

// TestDeathcultRogueBlockableOnlyByRogues plays Deathcult Rogue's text: a Rogue
// and a changeling (every creature type, CR 702.73a) can block it; a non-Rogue
// can't.
func TestDeathcultRogueBlockableOnlyByRogues(t *testing.T) {
	g := mustBlockGame()
	rogue := addSelfExceptByCreature(g, game.Player1, []types.Sub{types.Human, types.Rogue}, exceptBySubtypes(types.Rogue))
	attackingDefender(g, game.Player2, rogue)
	defendingRogue := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Rogue}, nil)
	changeling := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Shapeshifter}, nil, game.Changeling)
	human := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Human}, nil)

	checkExceptByBlockers(t, g, rogue,
		map[string]*game.Permanent{"Rogue": defendingRogue, "changeling": changeling},
		map[string]*game.Permanent{"non-Rogue": human})
}

// TestAmrouSeekersArtifactOrWhiteUnion plays Amrou Seekers' and Skirk Shaman's
// type-or-color union, reading colors live.
func TestAmrouSeekersArtifactOrWhiteUnion(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color color.Color
	}{{"Amrou Seekers", color.White}, {"Skirk Shaman", color.Red}} {
		t.Run(tc.name, func(t *testing.T) {
			g := mustBlockGame()
			seekers := addSelfExceptByCreature(g, game.Player1, nil, exceptBySelection(game.Selection{AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
				{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{tc.color}},
			}}))
			attackingDefender(g, game.Player2, seekers)
			pt := game.PT{Value: 1}
			artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name:      "Blue Artifact Creature",
				Types:     []types.Card{types.Artifact, types.Creature},
				Colors:    []color.Color{color.Blue},
				Power:     opt.Val(pt),
				Toughness: opt.Val(pt),
			}})
			matching := addColoredCombatCreature(g, game.Player2, "Matching Creature", tc.color)
			green := addColoredCombatCreature(g, game.Player2, "Green Creature", color.Green)
			painted := addColoredCombatCreature(g, game.Player2, "Painted Creature", color.Green)

			checkExceptByBlockers(t, g, seekers,
				map[string]*game.Permanent{"artifact creature": artifact, "matching color": matching},
				map[string]*game.Permanent{"green": green, "green before painting": painted})

			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), Layer: game.LayerColor, AffectedObjectID: painted.ObjectID, SetColors: []color.Color{tc.color},
			})
			checkExceptByBlockers(t, g, seekers,
				map[string]*game.Permanent{"recolored creature": painted},
				map[string]*game.Permanent{"green": green})
		})
	}
}

// TestElvenRidersWallOrFlyingUnion plays Elven Riders: a Wall or a flier can
// block, a reach creature can't (2007-05-01 ruling), and granted flying counts.
func TestElvenRidersWallOrFlyingUnion(t *testing.T) {
	g := mustBlockGame()
	riders := addSelfExceptByCreature(g, game.Player1, []types.Sub{types.Elf}, exceptBySelection(game.Selection{AnyOf: []game.Selection{
		{SubtypesAny: []types.Sub{types.Wall}},
		{Keyword: game.Flying},
	}}))
	attackingDefender(g, game.Player2, riders)
	wall := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Wall}, nil, game.Defender)
	flier := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bird}, nil, game.Flying)
	spider := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Spider}, nil, game.Reach)
	bear := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bear}, nil)

	checkExceptByBlockers(t, g, riders,
		map[string]*game.Permanent{"Wall": wall, "flier": flier},
		map[string]*game.Permanent{"reach creature": spider, "ground non-Wall": bear})

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerAbility, AffectedObjectID: bear.ObjectID, AddKeywords: []game.Keyword{game.Flying},
	})
	checkExceptByBlockers(t, g, riders,
		map[string]*game.Permanent{"creature granted flying": bear},
		map[string]*game.Permanent{"reach creature": spider})
}

func attachedExceptByWalls(name string, subtype types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{subtype},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBeBlockedExceptBy,
			AffectedAttached:   true,
			Duration:           game.DurationPermanent,
			BlockerRestriction: exceptBySubtypes(types.Wall),
		}}}},
	}}
}

// TestInvisibilityAndProwlersHelmRestrictOnlyTheAttachedCreature plays
// Invisibility (Aura) and Prowler's Helm (Equipment): the attached creature can
// be blocked only by Walls until the Aura leaves or the Equipment moves.
func TestInvisibilityAndProwlersHelmRestrictOnlyTheAttachedCreature(t *testing.T) {
	t.Run("Invisibility", func(t *testing.T) {
		g := mustBlockGame()
		enchanted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, enchanted)
		def := attachedExceptByWalls("Invisibility", types.Aura)
		def.Types = []types.Card{types.Enchantment}
		def.StaticAbilities = append(def.StaticAbilities, game.EnchantStaticAbility(&game.TargetSpec{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "creature",
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
		}))
		aura := addCombatPermanent(g, game.Player1, def)
		if !attachPermanent(g, aura, enchanted) {
			t.Fatal("failed to attach Invisibility")
		}
		wall := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Wall}, nil, game.Defender)
		bear := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bear}, nil)

		checkExceptByBlockers(t, g, enchanted,
			map[string]*game.Permanent{"Wall": wall},
			map[string]*game.Permanent{"non-Wall": bear})

		if !movePermanentToZone(g, aura, zone.Graveyard) {
			t.Fatal("failed to move Invisibility to the graveyard")
		}
		checkExceptByBlockers(t, g, enchanted,
			map[string]*game.Permanent{"Wall": wall, "non-Wall after Aura left": bear}, nil)
	})
	t.Run("Prowler's Helm", func(t *testing.T) {
		g := mustBlockGame()
		equipped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, equipped, other)
		helm := addCombatPermanent(g, game.Player1, attachedExceptByWalls("Prowler's Helm", types.Equipment))
		if !attachPermanent(g, helm, equipped) {
			t.Fatal("failed to equip Prowler's Helm")
		}
		wall := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Wall}, nil, game.Defender)
		bear := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bear}, nil)

		checkExceptByBlockers(t, g, equipped,
			map[string]*game.Permanent{"Wall": wall},
			map[string]*game.Permanent{"non-Wall": bear})
		checkExceptByBlockers(t, g, other,
			map[string]*game.Permanent{"Wall": wall, "non-Wall": bear}, nil)

		if !attachPermanent(g, helm, other) {
			t.Fatal("failed to move Prowler's Helm")
		}
		checkExceptByBlockers(t, g, equipped,
			map[string]*game.Permanent{"non-Wall after Helm moved": bear}, nil)
		checkExceptByBlockers(t, g, other,
			map[string]*game.Permanent{"Wall": wall},
			map[string]*game.Permanent{"non-Wall": bear})
	})
}

// TestShiftingSliverRestrictsEverySliver plays Shifting Sliver: every Sliver,
// whoever controls it, can be blocked only by Slivers while it's on the
// battlefield.
func TestShiftingSliverRestrictsEverySliver(t *testing.T) {
	g := mustBlockGame()
	pt := game.PT{Value: 2}
	shifting := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Shifting Sliver",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sliver},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBeBlockedExceptBy,
			AffectedController: game.ControllerAny,
			PermanentTypes:     []types.Card{types.Creature},
			AffectedSelection:  game.Selection{SubtypesAny: []types.Sub{types.Sliver}},
			Duration:           game.DurationPermanent,
			BlockerRestriction: exceptBySubtypes(types.Sliver),
		}}}},
	}})
	sliver := addSubtypedCombatCreature(g, game.Player1, []types.Sub{types.Sliver}, nil)
	bear := addSubtypedCombatCreature(g, game.Player1, []types.Sub{types.Bear}, nil)
	attackingDefender(g, game.Player2, sliver, bear)
	defendingBear := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bear}, nil)

	checkExceptByBlockers(t, g, sliver,
		map[string]*game.Permanent{"Shifting Sliver": shifting},
		map[string]*game.Permanent{"non-Sliver": defendingBear})
	checkExceptByBlockers(t, g, bear,
		map[string]*game.Permanent{"non-Sliver blocking a non-Sliver": defendingBear}, nil)

	if !movePermanentToZone(g, shifting, zone.Graveyard) {
		t.Fatal("failed to move Shifting Sliver to the graveyard")
	}
	checkExceptByBlockers(t, g, sliver,
		map[string]*game.Permanent{"non-Sliver after Shifting Sliver left": defendingBear}, nil)
}

// TestSerpentOfYawningDepthsRestrictsOnlyYourSeaCreatures plays Serpent of
// Yawning Depths: your Krakens, Leviathans, Octopuses and Serpents can be
// blocked only by those types; an opponent's Serpent is unaffected.
func TestSerpentOfYawningDepthsRestrictsOnlyYourSeaCreatures(t *testing.T) {
	g := mustBlockGame()
	seaTypes := []types.Sub{types.Kraken, types.Leviathan, types.Octopus, types.Serpent}
	pt := game.PT{Value: 6}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Serpent of Yawning Depths",
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Serpent},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBeBlockedExceptBy,
			AffectedController: game.ControllerYou,
			PermanentTypes:     []types.Card{types.Creature},
			AffectedSelection:  game.Selection{SubtypesAny: seaTypes},
			Duration:           game.DurationPermanent,
			BlockerRestriction: exceptBySubtypes(seaTypes...),
		}}}},
	}})
	kraken := addSubtypedCombatCreature(g, game.Player1, []types.Sub{types.Kraken}, nil)
	attackingDefender(g, game.Player2, kraken)
	octopus := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Octopus}, nil)
	human := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Human}, nil)

	checkExceptByBlockers(t, g, kraken,
		map[string]*game.Permanent{"Octopus": octopus},
		map[string]*game.Permanent{"Human": human})

	opposingSerpent := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Serpent}, nil)
	defendingHuman := addSubtypedCombatCreature(g, game.Player1, []types.Sub{types.Human}, nil)
	if !canBlockAttacker(g, defendingHuman, opposingSerpent) {
		t.Fatal("an opponent's Serpent was restricted by Serpent of Yawning Depths")
	}
}

// TestDepartedDeckhandAbilityRestrictsTargetThisTurn resolves Departed
// Deckhand's activated ability: the targeted creature can be blocked only by
// Spirits this turn and is unrestricted after the turn ends.
func TestDepartedDeckhandAbilityRestrictsTargetThisTurn(t *testing.T) {
	g := mustBlockGame()
	deckhand := addSelfExceptByCreature(g, game.Player1, []types.Sub{types.Spirit, types.Pirate}, exceptBySubtypes(types.Spirit))
	pirate := addSubtypedCombatCreature(g, game.Player1, []types.Sub{types.Pirate}, nil)
	attackingDefender(g, game.Player2, pirate)
	spirit := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Spirit}, nil)
	bear := addSubtypedCombatCreature(g, game.Player2, []types.Sub{types.Bear}, nil)

	checkExceptByBlockers(t, g, pirate,
		map[string]*game.Permanent{"Spirit": spirit, "non-Spirit before activation": bear}, nil)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   deckhand.ObjectID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(pirate.ObjectID)},
	}
	resolveInstruction(NewEngine(nil), g, obj, game.ApplyRule{
		Object: opt.Val(game.TargetPermanentReference(0)),
		RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBeBlockedExceptBy,
			BlockerRestriction: exceptBySubtypes(types.Spirit),
		}},
		Duration: game.DurationThisTurn,
	}, nil)

	checkExceptByBlockers(t, g, pirate,
		map[string]*game.Permanent{"Spirit": spirit},
		map[string]*game.Permanent{"non-Spirit": bear})

	cloned := g.Clone()
	clonedPirate, _ := permanentByObjectID(cloned, pirate.ObjectID)
	clonedBear, _ := permanentByObjectID(cloned, bear.ObjectID)
	clonedSpirit, _ := permanentByObjectID(cloned, spirit.ObjectID)
	checkExceptByBlockers(t, cloned, clonedPirate,
		map[string]*game.Permanent{"cloned Spirit": clonedSpirit},
		map[string]*game.Permanent{"cloned non-Spirit": clonedBear})

	expireRuleEffects(g)
	checkExceptByBlockers(t, g, pirate,
		map[string]*game.Permanent{"Spirit": spirit, "non-Spirit next turn": bear}, nil)
}
