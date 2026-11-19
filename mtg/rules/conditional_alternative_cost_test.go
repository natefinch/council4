package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Mogg Salvage, Oracle ID 0923ad85-2ea4-4463-8f71-b7d0a162b486;
// MH2 #282, revision 2b93e65e-8a10-444f-8b0a-9327dd30cce2, pinned corpus
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle (verified 2026-11-19):
// https://api.scryfall.com/cards/2b93e65e-8a10-444f-8b0a-9327dd30cce2
// {2}{R} Instant. "If an opponent controls an Island and you control a Mountain,
// you may cast this spell without paying its mana cost. Destroy target artifact."
// https://api.scryfall.com/cards/2b93e65e-8a10-444f-8b0a-9327dd30cce2/rulings
// (verified 2026-11-19): any opponent may control the Island, irrespective of
// the target; land control matters only during casting; mana value remains 3.
// The unrelated Nevinyrral's Disk ruling returned by the endpoint does not apply.
// Wizards CR, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// 109.5, 110.2, 806.1 and 903.2: use actual caster and current controllers in
// free-for-all. 205.3i, 305.6 and 702.26b: check current land subtypes and
// ignore phased-out lands. 118.9a-d, 202.3 and 601.2a-c,f-h: announce one
// alternative, retain printed mana value, choose targets, then lock and pay
// the total including additional costs. 608.2b-c and 701.8a: recheck the
// target, not the cast condition, and destroy the legal target. 707.10 and
// 707.12: a stack copy is not cast, but a copy of a card can be cast.
// This definition is an Engine fixture, not a card registration.
func conditionalAlternativeFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Conditional artifact destruction",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
		Types:    []types.Card{types.Instant},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Cast without paying its mana cost",
			ManaCost: opt.Val(cost.Mana{}),
			Conditions: []cost.AlternativePredicate{
				{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Mountain},
				{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Island},
			},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
}

func TestMoggSalvageConditionalAlternativeCost(t *testing.T) {
	const noLand = game.PlayerID(-1)
	for _, tc := range []struct {
		name        string
		mountain    game.PlayerID
		island      game.PlayerID
		islandTaken bool
		phasedOut   bool
		want        bool
	}{
		{name: "Mountain only", mountain: game.Player1, island: noLand},
		{name: "Island only", mountain: noLand, island: game.Player3},
		{name: "caster's own Island", mountain: game.Player1, island: game.Player1},
		{name: "opponent's Mountain", mountain: game.Player2, island: game.Player3},
		{name: "phased-out Island", mountain: game.Player1, island: game.Player3, phasedOut: true},
		{name: "Island belongs to an opponent other than the target controller", mountain: game.Player1, island: game.Player3, want: true},
		{name: "opponent controls caster-owned Island", mountain: game.Player1, island: game.Player1, islandTaken: true, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, conditionalAlternativeFixture())
			artifact := addArtifactPermanent(g, game.Player2)
			if tc.mountain != noLand {
				addBasicLandPermanent(g, tc.mountain, types.Mountain).Tapped = true
			}
			var island *game.Permanent
			if tc.island != noLand {
				island = addBasicLandPermanent(g, tc.island, types.Island)
				island.PhasedOut = tc.phasedOut
				if tc.islandTaken {
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: island.ObjectID,
						NewController: opt.Val(game.Player3),
					})
				}
			}
			act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
			if got := containsAction(engine.legalActions(g, game.Player1), act); got != tc.want {
				t.Fatalf("free cast offered = %t, want %t", got, tc.want)
			}
			if got := engine.applyAction(g, game.Player1, act); got != tc.want {
				t.Fatalf("free cast applied = %t, want %t", got, tc.want)
			}
			if !tc.want {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Hand.Contains(spellID) {
					t.Fatal("rejected cast committed state")
				}
				return
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.ManaSpentToCast != 0 {
				t.Fatalf("free cast did not put a spell on the stack for zero mana: %+v", obj)
			}
			if value, ok := stackObjectManaValue(g, obj); !ok || value != 3 {
				t.Fatalf("mana value on the stack = %d (%t), want 3", value, ok)
			}
			island.PhasedOut = true // The already-cast spell must not recheck this gate.
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(spellID) {
				t.Fatal("the target artifact and resolving spell did not reach their owners' graveyards")
			}
		})
	}
}

func TestConditionalAlternativeCostCombinesWithAdditionalCosts(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		available            int
		kicked, normal, want bool
		spent                int
	}{
		{name: "free with kicker", available: 1, kicked: true, want: true, spent: 1},
		{name: "free without kicker", want: true},
		{name: "unaffordable kicker", kicked: true},
		{name: "choose printed cost", available: 3, normal: true, want: true, spent: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := conditionalAlternativeFixture()
			def.AdditionalCosts = []cost.Additional{{
				Kind: cost.AdditionalSacrifice, Amount: 1, MatchPermanentType: true, PermanentType: types.Land,
			}}
			def.StaticAbilities = []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}}},
			}}
			spellID := addCardToHand(g, game.Player1, def)
			mountain := addBasicLandPermanent(g, game.Player1, types.Mountain)
			mountain.Tapped = true
			addBasicLandPermanent(g, game.Player3, types.Island)
			artifact := addArtifactPermanent(g, game.Player2)
			g.Players[game.Player1].ManaPool.Add(mana.R, tc.available)
			act := action.CastSpellAction{CardID: spellID, SourceZone: zone.Hand, KickerPaid: tc.kicked,
				Targets: []game.Target{game.PermanentTarget(artifact.ObjectID)}}
			choseNormal := false
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
				choose: func(request game.ChoiceRequest) []int {
					if tc.normal && request.Prompt == "Choose spell cost" {
						for _, option := range request.Options {
							if strings.HasPrefix(option.Label, "Normal cost") {
								choseNormal = true
								return []int{option.Index}
							}
						}
						t.Fatalf("printed cost is missing from cast choices: %+v", request.Options)
					}
					return request.DefaultSelection
				},
			}}
			if got := engine.applyCastSpellWithChoices(g, game.Player1, act, agents, &TurnLog{}); got != tc.want {
				t.Fatalf("combined cast applied = %t, want %t", got, tc.want)
			}
			if got := g.Players[game.Player1].Graveyard.Contains(mountain.CardInstanceID); got != tc.want {
				t.Fatalf("sacrificed the Mountain = %t, want %t", got, tc.want)
			}
			if tc.want {
				obj, ok := g.Stack.Peek()
				if !ok || obj.ManaSpentToCast != tc.spent || obj.KickerPaid != tc.kicked || choseNormal != tc.normal {
					t.Fatalf("combined payment or choices did not survive casting: %+v", obj)
				}
			}
		})
	}
}

func TestConditionalAlternativeCostExileRoutes(t *testing.T) {
	for _, route := range []string{"normal exile permission", "effect-granted free cast", "effect-granted free cast of card copy"} {
		t.Run(route, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToExile(g, game.Player1, conditionalAlternativeFixture())
			artifact := addArtifactPermanent(g, game.Player2)
			targets := []game.Target{game.PermanentTarget(artifact.ObjectID)}
			log := &TurnLog{}
			switch route {
			case "normal exile permission":
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectPlayFromZone, Controller: game.Player1,
					AffectedPlayer: game.PlayerYou, Duration: game.DurationPermanent, CastFromZone: zone.Exile,
					AffectedCardID: spellID, ExpiresFor: game.Player1,
				})
				act := action.CastSpellFromZone(spellID, zone.Exile, targets, 0, nil)
				if containsAction(engine.legalActions(g, game.Player1), act) {
					t.Fatal("ordinary exile permission incorrectly waived conditions")
				}
				addBasicLandPermanent(g, game.Player1, types.Mountain).Tapped = true
				addBasicLandPermanent(g, game.Player3, types.Island)
				if !containsAction(engine.legalActions(g, game.Player1), act) || !engine.applyAction(g, game.Player1, act) {
					t.Fatal("ordinary exile permission did not allow the now-legal alternative cost")
				}
			case "effect-granted free cast":
				cast := action.CastSpellAction{CardID: spellID, SourceZone: zone.Exile, Targets: targets}
				if !engine.castFreeOption(g, g.Players[game.Player1], game.Player1, g.CardInstances[spellID],
					cast, false, [game.NumPlayers]PlayerAgent{}, log) {
					t.Fatal("external free alternative incorrectly required the printed alternative's conditions")
				}
			case "effect-granted free cast of card copy":
				if !castExiledCardCopy(engine, g, game.Player1, spellID, targets) {
					t.Fatal("external free cast of a card copy incorrectly required the printed alternative's conditions")
				}
			}
			obj, ok := g.Stack.Peek()
			copyCast := route == "effect-granted free cast of card copy"
			if !ok || obj.CastCopy != copyCast || obj.ManaSpentToCast != 0 ||
				countEventsOfKind(g.Events, game.EventSpellCast) != 1 {
				t.Fatalf("unexpected cast state or spell-cast event count: %+v", obj)
			}
			engine.resolveTopOfStack(g, log)
			if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
				g.Players[game.Player1].Exile.Contains(spellID) != copyCast ||
				g.Players[game.Player1].Graveyard.Contains(spellID) == copyCast {
				t.Fatal("wrong resolution or source-card destination")
			}
		})
	}
}
