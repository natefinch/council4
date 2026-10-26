package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3380. Bolas's Citadel (oracle
// 2bd111bb-ce02-414c-b5b7-e0e037d8d96b): "You may play lands and cast spells
// from the top of your library. If you cast a spell this way, pay life equal to
// its mana value rather than pay its mana cost." Gwenom, Remorseless (oracle
// c48cf135-b280-4e9f-a9d9-225104870741) grants the same pay-life permission;
// Future Sight (oracle 93a0e938-b916-47ec-9803-369e885e253c) is a plain
// cast-from-top permission. Citadel rulings 2019-05-03 (repeated for Gwenom
// 2025-09-19): a spell cast for another cost "rather than pay its mana cost"
// can't be cast for any alternative cost but may pay additional costs such as
// kicker; X must be 0 when casting without paying the mana cost. CR (effective
// 2026-09-25) 601.2b: the caster announces alternative/additional costs and
// can't apply two alternative methods of casting or two alternative costs;
// 118.9a: only one alternative cost applies to a spell; 118.14: a permission's
// rider applies only to spells cast that way; 107.3b: X is 0 when paying
// neither the mana cost nor an alternative cost that includes X; 601.2f: the
// total cost is the chosen mana or alternative cost plus additional costs.
// Each matching cast-from-library permission is a separate route; only the
// Citadel route replaces the mana cost with life.

func libraryRouteGame(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	return g, NewEngine(nil)
}

func addLibraryTopOgre(g *game.Game) id.ID {
	return addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Top Ogre",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
	}})
}

const payLifeOptionLabel = "Pay life equal to mana value"

// libraryRouteAgent picks the spell-cost option whose label is (or, with
// payLife false, is not) the Citadel pay-life option, recording the offered
// labels.
func libraryRouteAgent(t *testing.T, payLife bool, offered *[]string) [game.NumPlayers]PlayerAgent {
	t.Helper()
	return [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(request game.ChoiceRequest) []int {
			if request.Kind != game.ChoicePayment || request.Prompt != "Choose spell cost" {
				return request.DefaultSelection
			}
			for _, option := range request.Options {
				*offered = append(*offered, option.Label)
			}
			for _, option := range request.Options {
				if (option.Label == payLifeOptionLabel) == payLife {
					return []int{option.Index}
				}
			}
			t.Fatalf("no option with payLife=%v offered: %+v", payLife, request.Options)
			return nil
		},
	}}
}

func TestLibraryCastPlainPermissionPaysManaWhateverPermissionOrder(t *testing.T) {
	for _, citadelFirst := range []bool{true, false} {
		name := map[bool]string{true: "citadel-first", false: "plain-first"}[citadelFirst]
		t.Run(name, func(t *testing.T) {
			g, engine := libraryRouteGame(t)
			if citadelFirst {
				castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
				castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
			} else {
				castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
				castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
			}
			spellID := addLibraryTopOgre(g)
			g.Players[game.Player1].Life = 1
			g.Players[game.Player1].ManaPool.Add(mana.R, 3)

			act := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)
			if !containsAction(engine.legalActions(g, game.Player1), act) {
				t.Fatal("casting the top card through the plain permission is not listed")
			}
			clone := g.Clone()
			for _, current := range []*game.Game{g, clone} {
				if !engine.applyAction(current, game.Player1, act) {
					t.Fatal("casting the top card through the plain permission failed")
				}
				obj, ok := current.Stack.Peek()
				if !ok || obj.SourceID != spellID || obj.ManaSpentToCast != 3 {
					t.Fatalf("stack top = %#v, want the spell cast for its mana cost", obj)
				}
				if life, pool := current.Players[game.Player1].Life, current.Players[game.Player1].ManaPool.Total(); life != 1 || pool != 0 {
					t.Fatalf("life=%d pool=%d, want life 1 and the mana cost paid", life, pool)
				}
			}
		})
	}
}

func TestLibraryCastCasterChoosesPayLifeOrManaRoute(t *testing.T) {
	for _, payLife := range []bool{true, false} {
		name := map[bool]string{true: "pay-life", false: "pay-mana"}[payLife]
		t.Run(name, func(t *testing.T) {
			g, engine := libraryRouteGame(t)
			castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
			castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
			spellID := addLibraryTopOgre(g)
			g.Players[game.Player1].Life = 20
			g.Players[game.Player1].ManaPool.Add(mana.R, 3)

			var offered []string
			act := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)
			if !engine.applyActionWithChoices(g, game.Player1, act, libraryRouteAgent(t, payLife, &offered), &TurnLog{}) {
				t.Fatal("casting the top card failed")
			}
			if len(offered) != 2 {
				t.Fatalf("offered options = %q, want the pay-life route and the mana route", offered)
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.SourceID != spellID {
				t.Fatalf("stack top = %#v, want the cast spell", obj)
			}
			wantLife, wantPool, wantSpent := 20, 0, 3
			if payLife {
				wantLife, wantPool, wantSpent = 17, 3, 0
			}
			if life, pool := g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total(); life != wantLife || pool != wantPool || obj.ManaSpentToCast != wantSpent {
				t.Fatalf("life=%d pool=%d spent=%d, want %d/%d/%d", life, pool, obj.ManaSpentToCast, wantLife, wantPool, wantSpent)
			}
		})
	}
}

func TestLibraryCastPayLifeRouteExcludesOverload(t *testing.T) {
	setup := func(t *testing.T, plain bool) (*game.Game, *Engine, id.ID, *game.Permanent) {
		g, engine := libraryRouteGame(t)
		castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
		if plain {
			castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
		}
		spellID := addCardToLibrary(g, game.Player1, vandalblastDefinition())
		artifact := addOverloadArtifact(g, game.Player2, false)
		g.Players[game.Player1].Life = 20
		g.Players[game.Player1].ManaPool.Add(mana.R, 5)
		return g, engine, spellID, artifact
	}

	t.Run("citadel-only", func(t *testing.T) {
		g, engine, spellID, artifact := setup(t, false)
		overload := action.CastOverloadedSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil)
		if containsAction(engine.legalActions(g, game.Player1), overload) || engine.applyAction(g, game.Player1, overload) {
			t.Fatal("overload must not combine with Citadel's pay-life alternative cost")
		}
		normal := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
		if !engine.applyAction(g, game.Player1, normal) {
			t.Fatal("the normal cast through Citadel failed")
		}
		if life, pool := g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total(); life != 19 || pool != 5 {
			t.Fatalf("life=%d pool=%d, want 1 life paid and no mana", life, pool)
		}
	})

	t.Run("with-plain-permission", func(t *testing.T) {
		g, engine, spellID, _ := setup(t, true)
		overload := action.CastOverloadedSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil)
		if !containsAction(engine.legalActions(g, game.Player1), overload) {
			t.Fatal("overload through the plain permission is not listed")
		}
		if !engine.applyAction(g, game.Player1, overload) {
			t.Fatal("overload through the plain permission failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || !obj.Overloaded || obj.ManaSpentToCast != 5 {
			t.Fatalf("stack top = %#v, want an overloaded spell paid with five mana", obj)
		}
		if life := g.Players[game.Player1].Life; life != 20 {
			t.Fatalf("life = %d, want 20 (overload pays no life)", life)
		}
	})
}

func TestLibraryCastNonzeroXUsesOnlyManaRoute(t *testing.T) {
	addXSpell := func(g *game.Game) id.ID {
		return addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Top Hydra",
			Types:    []types.Card{types.Creature},
			ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
		}})
	}

	t.Run("citadel-only", func(t *testing.T) {
		g, engine := libraryRouteGame(t)
		castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
		spellID := addXSpell(g)
		g.Players[game.Player1].Life = 20
		g.Players[game.Player1].ManaPool.Add(mana.R, 3)
		legal := engine.legalActions(g, game.Player1)
		if containsAction(legal, action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 2, nil)) {
			t.Fatal("Citadel's pay-life route must announce X=0")
		}
		if !containsAction(legal, action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)) {
			t.Fatal("X=0 through Citadel is not listed")
		}
	})

	t.Run("with-plain-permission", func(t *testing.T) {
		g, engine := libraryRouteGame(t)
		castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
		castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
		spellID := addXSpell(g)
		g.Players[game.Player1].Life = 20
		g.Players[game.Player1].ManaPool.Add(mana.R, 3)
		act := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 2, nil)
		if !containsAction(engine.legalActions(g, game.Player1), act) {
			t.Fatal("X=2 through the plain permission is not listed")
		}
		var offered []string
		if !engine.applyActionWithChoices(g, game.Player1, act, libraryRouteAgent(t, false, &offered), &TurnLog{}) {
			t.Fatal("X=2 through the plain permission failed")
		}
		if len(offered) != 0 {
			t.Fatalf("offered %q, want no route choice: only the mana route can announce X=2", offered)
		}
		obj, _ := g.Stack.Peek()
		if obj.XValue != 2 || obj.ManaSpentToCast != 3 || g.Players[game.Player1].Life != 20 {
			t.Fatalf("X=%d spent=%d life=%d, want X=2 paid with three mana", obj.XValue, obj.ManaSpentToCast, g.Players[game.Player1].Life)
		}
	})
}

func TestLibraryCastPayLifeRouteCombinesWithKicker(t *testing.T) {
	g, engine := libraryRouteGame(t)
	castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Top Kicker",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost: cost.Mana{cost.G},
				BonusContent: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
				}.Ability(),
			}},
		}},
	}})
	g.Players[game.Player1].Life = 20
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)

	act := action.CastKickedSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("the kicked cast through Citadel is not listed")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("the kicked cast through Citadel failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.KickerPaid || obj.ManaSpentToCast != 1 {
		t.Fatalf("stack top = %#v, want a kicked spell paying only the kicker mana", obj)
	}
	if life, pool := g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total(); life != 18 || pool != 0 {
		t.Fatalf("life=%d pool=%d, want 2 life for mana value and the kicker paid with mana", life, pool)
	}
}

func TestLibraryCastOncePerTurnPermissionSpendsOnlyItsOwnUse(t *testing.T) {
	addOncePerTurn := func(g *game.Game) *game.Permanent {
		return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Test Once Seer",
			StaticAbilities: []game.StaticAbility{{
				Text: "Once each turn, you may cast a spell from the top of your library.",
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCastSpellsFromZone,
					AffectedPlayer: game.PlayerYou,
					CastFromZone:   zone.Library,
					TopCardOnly:    true,
					OncePerTurn:    true,
				}},
			}},
		}})
	}
	castTop := func(g *game.Game, engine *Engine) bool {
		top, _ := g.Players[game.Player1].Library.Top()
		return engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(top, zone.Library, game.FaceFront, nil, 0, nil))
	}

	t.Run("once-per-turn-only", func(t *testing.T) {
		g, engine := libraryRouteGame(t)
		addOncePerTurn(g)
		addLibraryTopOgre(g)
		addLibraryTopOgre(g)
		g.Players[game.Player1].ManaPool.Add(mana.R, 6)
		if !castTop(g, engine) {
			t.Fatal("the first cast through the once-per-turn permission failed")
		}
		if castTop(g, engine) {
			t.Fatal("the once-per-turn permission authorized a second cast this turn")
		}
	})

	t.Run("with-plain-permission", func(t *testing.T) {
		g, engine := libraryRouteGame(t)
		source := addOncePerTurn(g)
		castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
		addLibraryTopOgre(g)
		g.Players[game.Player1].ManaPool.Add(mana.R, 3)
		if !castTop(g, engine) {
			t.Fatal("casting the top card failed")
		}
		if g.ExilePlayPermissionUsedThisTurn[source.ObjectID] {
			t.Fatal("the default cast spent the once-per-turn use while an unlimited permission offered the same cost")
		}
	})
}
