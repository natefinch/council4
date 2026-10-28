package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCastRouteReadFrameSeparatesContexts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	def := numericDamageSpell()
	def.Layout, def.Back = game.LayoutModalDFC, opt.Val(def.CardFace)
	below := addCardToLibrary(g, game.Player1, def)
	top := addCardToLibrary(g, game.Player1, def)

	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	for _, test := range []struct {
		name    string
		key     castRouteKey
		count   int
		payLife bool
	}{
		{"normal", castRouteKey{player: game.Player1, card: top, zone: zone.Library}, 1, true},
		{"overload", castRouteKey{player: game.Player1, card: top, zone: zone.Library, mode: castCostMode{overloaded: true}}, 0, false},
		{"bestow", castRouteKey{player: game.Player1, card: top, zone: zone.Library, mode: castCostMode{bestowed: true}}, 0, false},
		{"mutate", castRouteKey{player: game.Player1, card: top, zone: zone.Library, mode: castCostMode{mutate: true}}, 0, false},
		{"other player", castRouteKey{player: game.Player2, card: top, zone: zone.Library}, 0, false},
		{"not top", castRouteKey{player: game.Player1, card: below, zone: zone.Library}, 0, false},
		{"hand", castRouteKey{player: game.Player1, card: top, zone: zone.Hand}, 1, false},
		{"back face", castRouteKey{player: game.Player1, card: top, zone: zone.Library, face: game.FaceBack}, 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			key := test.key
			first := spellCastRoutes(g, key.player, key.card, key.zone, key.face, key.mode)
			if len(first) != test.count || len(first) > 0 && first[0].payLife != test.payLife {
				t.Fatalf("routes = %d, want %d with payLife=%v", len(first), test.count, test.payLife)
			}
			again := spellCastRoutes(g, key.player, key.card, key.zone, key.face, key.mode)
			if _, cached := frameCacheFor(g).castRoutes[key]; !cached {
				t.Fatal("read frame did not retain the route result, including empty results")
			}
			if len(first) > 0 && (&first[0] != &again[0] || cap(first) != len(first)) {
				t.Fatal("repeated query rebuilt routes or exposed writable spare capacity")
			}
		})
	}
}

func TestCastRouteReadFrameEndsBeforeMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	card := addLibraryTopOgre(g)
	var branch *game.Game
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		if routes := spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{}); len(routes) != 1 || !routes[0].payLife {
			t.Fatal("fixture did not expose the library permission")
		}
		branch = g.Clone()
	}()
	if branch.InStaticSourceFrame() {
		t.Fatal("clone inherited its parent's route cache")
	}
	clonedSource, ok := permanentByObjectID(branch, source.ObjectID)
	if !ok || !movePermanentToZone(branch, clonedSource, zone.Graveyard) {
		t.Fatal("could not remove the cloned permission source")
	}
	if len(spellCastRoutes(branch, game.Player1, card, zone.Library, game.FaceFront, castCostMode{})) != 0 ||
		len(spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{})) != 1 {
		t.Fatal("route data crossed clones or survived source departure")
	}
	addLibraryTopOgre(g)
	if routes := spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{}); len(routes) != 0 {
		t.Fatal("route cache retained a card that is no longer on top")
	}
	if g.InStaticSourceFrame() {
		t.Fatal("standalone route queries retained a read frame")
	}
}

func TestCastRoutesRefreshAfterPermissionUse(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
	sourceDef, ok := permanentCardDef(g, source)
	if !ok {
		t.Fatal("permission source has no card definition")
	}
	sourceDef.StaticAbilities[0].RuleEffects[0].OncePerTurn = true
	card := addLibraryTopOgre(g)
	var permission game.RuleEffect
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		routes := spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{})
		if len(routes) != 1 || !routes[0].permission.OncePerTurn {
			t.Fatal("fixture did not expose a once-per-turn route")
		}
		permission = routes[0].permission
	}()
	recordExilePlayPermissionUse(g, permission)
	if routes := spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{}); len(routes) != 0 {
		t.Fatal("used permission remained in the cast routes")
	}
	NewEngine(nil).advanceToNextTurn(g)
	if routes := spellCastRoutes(g, game.Player1, card, zone.Library, game.FaceFront, castCostMode{}); len(routes) != 1 {
		t.Fatal("fresh turn did not restore the limited permission")
	}
}

func TestCastTargetReadFrameSeparatesModesAndNumbers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Cached numeric modes", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{MinModes: 1, MaxModes: 1, Modes: []game.Mode{
			{Targets: []game.TargetSpec{{Allow: game.TargetAllowPlayer, CountEqualsX: true, MinTargets: 1, MaxTargets: 20}}},
			{Targets: []game.TargetSpec{{Allow: game.TargetAllowPlayer, CountEqualsKickerPlusOne: true, MinTargets: 1, MaxTargets: 20}}},
		}}),
	}}
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		modes := []int{0}
		branch := game.CastBranch{XValue: opt.Val(2), KickerCount: opt.Val(3)}
		first := spellTargetSpecsInFrame(g, def, modes, branch)
		again := spellTargetSpecsInFrame(g, def, []int{0}, branch)
		if len(first) != 1 || first[0].MaxTargets != 2 || &first[0] != &again[0] {
			t.Fatal("same branch did not reuse its exact numeric target spec")
		}
		modes[0] = 1
		kicked := spellTargetSpecsInFrame(g, def, modes, branch)
		if kicked[0].MaxTargets != 4 || first[0].MaxTargets != 2 {
			t.Fatal("mutable mode input or one mode's targets overwrote another")
		}
		branch.XValue = opt.Val(5)
		if changed := spellTargetSpecsInFrame(g, def, []int{0}, branch); changed[0].MaxTargets != 5 {
			t.Fatal("different X reused the previous target cardinality")
		}
		branch.KickerCount = opt.Val(6)
		if changed := spellTargetSpecsInFrame(g, def, []int{1}, branch); changed[0].MaxTargets != 7 {
			t.Fatal("different kicker count reused the previous target cardinality")
		}
	}()
	def.SpellAbility.Val.Modes[0].Targets[0].CountEqualsX = false
	def.SpellAbility.Val.Modes[0].Targets[0].MaxTargets = 9
	if changed := spellTargetSpecsInFrame(g, def, []int{0}, game.CastBranch{XValue: opt.Val(2)}); changed[0].MaxTargets != 9 {
		t.Fatal("target cache survived the end of its frame")
	}
}

func TestSharedCastTargetChecksMatchIndividualChecks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player2, frameDiffCreature("Numeric target", 2, 2))
	for _, spec := range []game.TargetSpec{
		{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1},
		{Allow: game.TargetAllowPlayer, CountEqualsX: true, MinTargets: 0, MaxTargets: 20},
		{Allow: game.TargetAllowPlayer, CountEqualsKickerPlusOne: true, MinTargets: 1, MaxTargets: 21},
		{Allow: game.TargetAllowPlayer, Gate: game.TargetGateSpellKicked, MinTargets: 1, MaxTargets: 1},
		{Allow: game.TargetAllowPermanent, ManaValueEqualsX: true, MinTargets: 1, MaxTargets: 1},
	} {
		def := numericDamageSpell()
		def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{spec}
		func() {
			g.BeginStaticSourceFrame()
			defer g.EndStaticSourceFrame()
			for _, x := range []int{0, 1, 2} {
				for _, kicker := range []int{0, 1, 2} {
					branch := game.CastBranch{XValue: opt.Val(x), KickerCount: opt.Val(kicker), Kicked: kicker > 0}
					for _, targets := range [][]game.Target{
						nil,
						{game.PlayerTarget(game.Player2)},
						{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
						{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player2)},
						{game.PermanentTarget(creature.ObjectID)},
						{game.PermanentTarget(creature.ObjectID + 100)},
					} {
						want := targetsValidForSpell(g, game.Player1, def, nil, targets, branch) &&
							spellTargetCountsMatchX(g, game.Player1, def, nil, targets, x, branch) &&
							spellTargetCountsMatchKicker(g, game.Player1, def, nil, targets, kicker, branch) &&
							spellTargetsSatisfyManaValueX(g, game.Player1, def, nil, targets, x, branch)
						if got := spellTargetsValidForCast(g, game.Player1, def, nil, targets, x, kicker, branch); got != want {
							t.Fatalf("shared checks = %v, individual checks = %v for X=%d kicker=%d targets=%v spec=%+v",
								got, want, x, kicker, targets, spec)
						}
					}
				}
			}
		}()
	}
}

func BenchmarkSpellCastRoutesInReadFrame(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	_ = spellCastRoutes(g, game.Player1, 1, zone.Hand, game.FaceFront, castCostMode{})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if len(spellCastRoutes(g, game.Player1, 1, zone.Hand, game.FaceFront, castCostMode{})) != 1 {
			b.Fatal("ordinary hand cast lost its default route")
		}
	}
}

func BenchmarkSpellCastRoutesWithoutReadFrame(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if len(spellCastRoutes(g, game.Player1, 1, zone.Hand, game.FaceFront, castCostMode{})) != 1 {
			b.Fatal("ordinary hand cast lost its default route")
		}
	}
}
