package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Comprehensive Rules effective 2026-09-25. CR 601.2b: "If a
// cost that will be paid as the spell is being cast includes Phyrexian mana
// symbols, the player announces whether they intend to pay 2 life or a
// corresponding colored mana cost for each of those symbols." CR 601.2f
// determines the total cost and CR 601.2h pays it "in any order". CR 107.4e:
// "a monocolored hybrid symbol such as {2/B} can be paid with either one black
// mana or two mana of any type"; CR 107.4f: Phyrexian mana symbols "are
// colored mana symbols" paid "either with one mana of its color or by paying 2
// life". CR 702.51a: convoke means "For each colored mana in this spell's
// total cost, you may tap an untapped creature of that color you control rather
// than pay that mana. For each generic mana in this spell's total cost, you may
// tap an untapped creature you control rather than pay that mana." CR 702.66a:
// delve means "For each generic mana in this spell's total cost, you may exile
// a card from your graveyard rather than pay that mana." CR 702.126a:
// improvise means "For each generic mana in this spell's total cost, you may
// tap an untapped artifact you control rather than pay that mana." CR
// 702.51b/702.66b: each "applies only after the total cost of the spell ... is
// determined", so it pays its portion of a mixed total cost and leaves every
// other symbol, and the announced payment of each Phyrexian symbol it doesn't
// pay, to the rest of the payment. CR 733.1 reverses a cast that can't be
// paid. Oracle (pinned corpus): Vault Skirge
// 64cf5d59-7bcd-4b0b-a160-c8468d4c0f60 {1}{B/P}; Porcelain Legionnaire
// 66442396-f804-47f3-8daa-0d43d8cfb482 {2}{W/P}; Dismember
// fd74f8eb-0253-42dd-8277-186d4934da38 {1}{B/P}{B/P}; Spectral Procession
// 376271a6-80f0-47a6-990a-5ec5517aae42 {2/W}{2/W}{2/W}; Merrow Skyswimmer
// d709f396-021f-49eb-adb5-374aa790eaae {3}{W/U}{W/U} "Convoke"; Archway of
// Innovation bfa20bc7-4626-4a52-87f4-6e2763cb8ed5 "{U}, {T}: The next spell
// you cast this turn has improvise."; Wand of the Worldsoul
// 3d175a59-c1b4-43b8-8a1f-f7cbcefe130e "{T}: The next spell you cast this turn
// has convoke."; Inspiring Statuary 78b2b68d-6e27-49f3-8373-6e19faf6f679
// "Nonartifact spells you cast have improvise."; Conduit of Worlds
// ed14be15-8f8d-4fe3-a147-f5da8ed873bf "{T}: Choose target nonland permanent
// card in your graveyard. If you haven't cast a spell this turn, you may cast
// that card."; Etherium Sculptor 96b87445-6362-4a17-91b2-cbf203fd03fd
// "Artifact spells you cast cost {1} less to cast."; Ichor Wellspring
// 5b5ef43b-13fd-4461-8d2d-18be65e9a790 (artifact); Vampire Nighthawk
// feb244f8-bcb1-44cf-9940-2719221a7309 (black creature); Marble Diamond
// 910488bf-66ab-415e-973b-1262b2ab7454 "{T}: Add {W}." (artifact); Birds of
// Paradise d3a0b660-358c-41bd-9cd2-41fbf3491b1a "{T}: Add one mana of any
// color." (green creature); Savannah Lions
// 60ba93eb-39e6-4af2-9c66-cd38f72daff2 (white); Fugitive Wizard
// 381fa2a9-69fc-4558-a0ac-fb99c6f8d77f (blue); Grizzly Bears
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0 (green). No registered card grants
// delve to another spell, so the delve scenario grants it with a rule effect.
// CR 106.12/702.126a: a permanent tapped for convoke or improvise isn't
// tapped for mana, so the mechanic must leave untapped the mana sources the
// rest of the payment needs. These scenarios cover Engine gap #3571 on the priority cast, the paid cast
// from a graveyard during resolution and clones.

// activateNextSpellGrant activates the registered permanent's "the next spell
// you cast this turn has ..." ability and resolves it.
func activateNextSpellGrant(t *testing.T, e *Engine, g *game.Game, source *game.Permanent) {
	t.Helper()
	size := g.Stack.Size()
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("activating the next-spell grant failed")
	}
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Stack.Size() != size {
		t.Fatal("the next-spell grant did not resolve")
	}
}

// castFromHandWithAnnouncements casts cardID from Player1's hand at targets
// with agent's Phyrexian announcements and reports whether it was cast.
func castFromHandWithAnnouncements(e *Engine, g *game.Game, cardID id.ID, targets []game.Target, agent *phyrexianWindowAgent) bool {
	return e.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, targets, 0, nil), reboundKickerAgents(agent), &TurnLog{})
}

func requireTapped(t *testing.T, want map[*game.Permanent]bool) {
	t.Helper()
	for permanent, tapped := range want {
		if permanent.Tapped != tapped {
			t.Fatalf("permanent %v tapped=%t, want %t", permanent.ObjectID, permanent.Tapped, tapped)
		}
	}
}

func requireSpellOnStack(t *testing.T, g *game.Game, cardID id.ID, manaSpent int) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID || obj.ManaSpentToCast != manaSpent {
		t.Fatalf("stack top = %+v, want the spell %v cast with %d mana", obj, cardID, manaSpent)
	}
}

func requireCastReversed(t *testing.T, g *game.Game, cardID id.ID, life int) {
	t.Helper()
	if !g.Players[game.Player1].Hand.Contains(cardID) || !g.Stack.IsEmpty() || g.Players[game.Player1].Life != life {
		t.Fatalf("hand has spell=%t stack=%d life=%d, want the cast reversed to hand at life %d",
			g.Players[game.Player1].Hand.Contains(cardID), g.Stack.Size(), g.Players[game.Player1].Life, life)
	}
}

// TestConduitCastImprovisePaysGenericBesidePhyrexian is the #3571 repro:
// Conduit of Worlds casts the registered Vault Skirge ({1}{B/P}) from the
// graveyard after Archway of Innovation gave the next spell improvise, with
// Ichor Wellspring the only artifact. Announcing 2 life for {B/P}, Wellspring
// pays the {1} (CR 702.126a) and life pays {B/P}; Waterlogged Grove, whose
// mana costs life, stays untapped. Announcing {B} with {B} floating, the pool
// pays {B/P} and Wellspring the {1}. Each cast runs in a clone without
// changing the original game.
func TestConduitCastImprovisePaysGenericBesidePhyrexian(t *testing.T) {
	for _, tc := range []struct {
		name          string
		agent         *phyrexianWindowAgent
		poolB         int
		wantLifeLost  int
		wantManaSpent int
	}{
		{name: "announcing 2 life", agent: &phyrexianWindowAgent{payLife: true}, wantLifeLost: 2},
		{name: "announcing mana", agent: &phyrexianWindowAgent{payMana: true}, poolB: 1, wantManaSpent: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g, cardID := conduitPaidCastGame(t, "Vault Skirge", "Ichor Wellspring", "Archway of Innovation")
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			activateNextSpellGrant(t, e, g, reboundPermanentNamed(t, g, "Archway of Innovation"))
			g.Players[game.Player1].ManaPool.Add(mana.B, tc.poolB)
			life := g.Players[game.Player1].Life
			clone := g.Clone()

			e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(tc.agent), &TurnLog{})

			requirePaidCast(t, clone, cardID, zone.Graveyard, tc.wantManaSpent)
			requireTapped(t, map[*game.Permanent]bool{
				reboundPermanentNamed(t, clone, "Ichor Wellspring"):  true,
				reboundPermanentNamed(t, clone, "Waterlogged Grove"): false,
			})
			if clone.Players[game.Player1].Life != life-tc.wantLifeLost {
				t.Fatalf("life=%d, want %d", clone.Players[game.Player1].Life, life-tc.wantLifeLost)
			}
			if reboundPermanentNamed(t, g, "Ichor Wellspring").Tapped || !g.Players[game.Player1].Graveyard.Contains(cardID) ||
				g.Players[game.Player1].Life != life {
				t.Fatal("casting in the clone changed the original game")
			}
		})
	}
}

// TestPriorityCastImproviseKeepsAnnouncedPhyrexianPayment casts the
// registered Porcelain Legionnaire ({2}{W/P}) from hand after Archway of
// Innovation's grant. Mana alone can't pay it, so the payment retries with
// improvise, and that retry must keep the announced {W/P} payment (CR
// 601.2b): announcing 2 life, Ichor Wellspring and a Swamp pay {2}, and with
// two Wellsprings a Plains stays untapped. Announcing {W}, the Plains pays it
// while two Wellsprings pay {2}; with only Wellsprings the announced {W} can't
// be paid, so the cast is reversed without paying life (CR 733.1). A single
// Plains beside one Wellspring is the separate mana-before-life defect #3397.
// Etherium Sculptor's reduction (CR 601.2f) leaves {1}{W/P}, and the Sculptor
// itself, an artifact, can be tapped for the {1}.
func TestPriorityCastImproviseKeepsAnnouncedPhyrexianPayment(t *testing.T) {
	for _, tc := range []struct {
		name        string
		agent       *phyrexianWindowAgent
		wellsprings int
		plains      bool
		swamp       bool
		sculptor    bool
		wantCast    bool
		wantLife    int
		wantPlains  bool
		wantMana    int
		wantTapped  int
	}{
		{name: "life with a Swamp", agent: &phyrexianWindowAgent{payLife: true}, wellsprings: 1, swamp: true,
			wantCast: true, wantLife: 2, wantMana: 1, wantTapped: 1},
		{name: "life with two artifacts", agent: &phyrexianWindowAgent{payLife: true}, wellsprings: 2, plains: true,
			wantCast: true, wantLife: 2, wantTapped: 2},
		{name: "mana with a Plains", agent: &phyrexianWindowAgent{payMana: true}, wellsprings: 2, plains: true,
			wantCast: true, wantPlains: true, wantMana: 1, wantTapped: 2},
		{name: "mana without a white source", agent: &phyrexianWindowAgent{payMana: true}, wellsprings: 2},
		{name: "life with a cost reduction", agent: &phyrexianWindowAgent{payLife: true}, sculptor: true,
			wantCast: true, wantLife: 2, wantTapped: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			archway := addCombatPermanent(g, game.Player1, reg.Lookup("Archway of Innovation"))
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			activateNextSpellGrant(t, e, g, archway)
			var artifacts []*game.Permanent
			for range tc.wellsprings {
				artifacts = append(artifacts, addCombatPermanent(g, game.Player1, reg.Lookup("Ichor Wellspring")))
			}
			if tc.sculptor {
				artifacts = append(artifacts, addCombatPermanent(g, game.Player1, reg.Lookup("Etherium Sculptor")))
			}
			var plains *game.Permanent
			if tc.plains {
				plains = addBasicLandPermanent(g, game.Player1, types.Plains)
			}
			var swamp *game.Permanent
			if tc.swamp {
				swamp = addBasicLandPermanent(g, game.Player1, types.Swamp)
			}
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Porcelain Legionnaire"))
			life := g.Players[game.Player1].Life

			cast := castFromHandWithAnnouncements(e, g, cardID, nil, tc.agent)

			if cast != tc.wantCast {
				t.Fatalf("cast=%t, want %t", cast, tc.wantCast)
			}
			if !tc.wantCast {
				requireCastReversed(t, g, cardID, life)
				for _, artifact := range artifacts {
					requireTapped(t, map[*game.Permanent]bool{artifact: false})
				}
				return
			}
			requireSpellOnStack(t, g, cardID, tc.wantMana)
			if g.Players[game.Player1].Life != life-tc.wantLife {
				t.Fatalf("life=%d, want %d", g.Players[game.Player1].Life, life-tc.wantLife)
			}
			if plains != nil {
				requireTapped(t, map[*game.Permanent]bool{plains: tc.wantPlains})
			}
			if swamp != nil {
				requireTapped(t, map[*game.Permanent]bool{swamp: true})
			}
			tapped := 0
			for _, artifact := range artifacts {
				if artifact.Tapped {
					tapped++
				}
			}
			if tapped != tc.wantTapped {
				t.Fatalf("tapped %d artifacts, want %d", tapped, tc.wantTapped)
			}
		})
	}
}

// TestConvokePaysHybridAndGenericSymbols casts the registered Merrow
// Skyswimmer ({3}{W/U}{W/U}, convoke). Savannah Lions and Fugitive Wizard pay
// the hybrid symbols with their colors and the Grizzly Bears pay generic mana
// (CR 702.51a), a Plains paying the {1} the creatures don't.
func TestConvokePaysHybridAndGenericSymbols(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bears  int
		plains bool
	}{
		{name: "creatures pay the whole cost", bears: 3},
		{name: "a Plains pays the rest", bears: 2, plains: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			creatures := []*game.Permanent{
				addCombatPermanent(g, game.Player1, reg.Lookup("Savannah Lions")),
				addCombatPermanent(g, game.Player1, reg.Lookup("Fugitive Wizard")),
			}
			for range tc.bears {
				creatures = append(creatures, addCombatPermanent(g, game.Player1, reg.Lookup("Grizzly Bears")))
			}
			wantMana := 0
			if tc.plains {
				addBasicLandPermanent(g, game.Player1, types.Plains)
				wantMana = 1
			}
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Merrow Skyswimmer"))

			if !castFromHandWithAnnouncements(e, g, cardID, nil, &phyrexianWindowAgent{}) {
				t.Fatal("Merrow Skyswimmer cast failed")
			}
			requireSpellOnStack(t, g, cardID, wantMana)
			for _, creature := range creatures {
				requireTapped(t, map[*game.Permanent]bool{creature: true})
			}
		})
	}
}

// TestConvokePaysPhyrexianAnnouncedAsMana casts the registered Vault Skirge
// after Wand of the Worldsoul gave the next spell convoke. Announcing {B}, the
// black Vampire Nighthawk pays {B/P} and Grizzly Bears the {1} (CR 702.51a,
// 107.4f); with only two Bears the announced {B} can't be paid and the cast is
// reversed without paying life (CR 733.1). Announcing 2 life, life pays {B/P}
// and one creature the {1}.
func TestConvokePaysPhyrexianAnnouncedAsMana(t *testing.T) {
	for _, tc := range []struct {
		name       string
		agent      *phyrexianWindowAgent
		nighthawk  bool
		wantCast   bool
		wantLife   int
		wantTapped int
	}{
		{name: "mana with a black creature", agent: &phyrexianWindowAgent{payMana: true}, nighthawk: true, wantCast: true, wantTapped: 2},
		{name: "mana without a black creature", agent: &phyrexianWindowAgent{payMana: true}},
		{name: "life", agent: &phyrexianWindowAgent{payLife: true}, nighthawk: true, wantCast: true, wantLife: 2, wantTapped: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			wand := addCombatPermanent(g, game.Player1, reg.Lookup("Wand of the Worldsoul"))
			wand.Tapped = false
			activateNextSpellGrant(t, e, g, wand)
			var creatures []*game.Permanent
			if tc.nighthawk {
				creatures = append(creatures, addCombatPermanent(g, game.Player1, reg.Lookup("Vampire Nighthawk")))
			} else {
				creatures = append(creatures, addCombatPermanent(g, game.Player1, reg.Lookup("Grizzly Bears")))
			}
			creatures = append(creatures, addCombatPermanent(g, game.Player1, reg.Lookup("Grizzly Bears")))
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Vault Skirge"))
			life := g.Players[game.Player1].Life

			cast := castFromHandWithAnnouncements(e, g, cardID, nil, tc.agent)

			if cast != tc.wantCast {
				t.Fatalf("cast=%t, want %t", cast, tc.wantCast)
			}
			if !tc.wantCast {
				requireCastReversed(t, g, cardID, life)
				return
			}
			requireSpellOnStack(t, g, cardID, 0)
			tapped := 0
			for _, creature := range creatures {
				if creature.Tapped {
					tapped++
				}
			}
			if tapped != tc.wantTapped || g.Players[game.Player1].Life != life-tc.wantLife {
				t.Fatalf("tapped=%d life=%d, want %d tapped and life %d", tapped, g.Players[game.Player1].Life, tc.wantTapped, life-tc.wantLife)
			}
			if tc.agent.payMana && !creatures[0].Tapped {
				t.Fatal("the black creature did not pay the {B/P} announced as mana")
			}
		})
	}
}

// TestDelvePaysGenericBesidePhyrexian casts the registered Dismember
// ({1}{B/P}{B/P}) at an opponent's Grizzly Bears with delve granted by a rule
// effect. Announcing 2 life for each {B/P}, one graveyard card is exiled for
// the {1} (CR 702.66a) and the other stays. Announcing {B} for each, two
// Swamps pay them and one card still pays the {1}.
func TestDelvePaysGenericBesidePhyrexian(t *testing.T) {
	for _, tc := range []struct {
		name     string
		agent    *phyrexianWindowAgent
		swamps   int
		wantLife int
	}{
		{name: "life for both", agent: &phyrexianWindowAgent{payLife: true}, wantLife: 4},
		{name: "mana for both", agent: &phyrexianWindowAgent{payMana: true}, swamps: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			grant := nextSpellImproviseGrant(g, game.Selection{})
			grant.GrantedKeyword = game.Delve
			g.RuleEffects = append(g.RuleEffects, grant)
			bears := addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
			first := addCardToGraveyard(g, game.Player1, reg.Lookup("Grizzly Bears"))
			second := addCardToGraveyard(g, game.Player1, reg.Lookup("Savannah Lions"))
			var swamps []*game.Permanent
			for range tc.swamps {
				swamps = append(swamps, addBasicLandPermanent(g, game.Player1, types.Swamp))
			}
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Dismember"))
			life := g.Players[game.Player1].Life
			clone := g.Clone()

			if !castFromHandWithAnnouncements(e, clone, cardID, []game.Target{game.PermanentTarget(bears.ObjectID)}, tc.agent) {
				t.Fatal("Dismember cast failed")
			}
			player := clone.Players[game.Player1]
			exiled := 0
			for _, graveID := range []id.ID{first, second} {
				if player.Exile.Contains(graveID) {
					exiled++
				}
			}
			if exiled != 1 || player.Graveyard.Size() != 1 || player.Life != life-tc.wantLife {
				t.Fatalf("exiled=%d graveyard=%d life=%d, want one card exiled and life %d", exiled, player.Graveyard.Size(), player.Life, life-tc.wantLife)
			}
			requireSpellOnStack(t, clone, cardID, tc.swamps)
			for _, swamp := range swamps {
				if cloned, ok := clone.PermanentByID(swamp.ObjectID); !ok || !cloned.Tapped {
					t.Fatal("a Swamp did not pay a {B/P} announced as mana")
				}
			}
			if g.Players[game.Player1].Graveyard.Size() != 2 || g.Players[game.Player1].Life != life || !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("casting in the clone changed the original game")
			}
		})
	}
}

// TestImprovisePaysMonocoloredHybridGenericHalves casts the registered
// Spectral Procession ({2/W}{2/W}{2/W}) with Inspiring Statuary's improvise
// grant. With one Plains and four artifacts, two {2/W} symbols are paid as
// {2} by the artifacts (CR 107.4e, 702.126a) and the Plains pays the third
// with {W}. Three artifacts can't pay enough, so that cast is reversed.
func TestImprovisePaysMonocoloredHybridGenericHalves(t *testing.T) {
	for _, tc := range []struct {
		name        string
		wellsprings int
		wantCast    bool
	}{
		{name: "four artifacts and a Plains", wellsprings: 3, wantCast: true},
		{name: "three artifacts and a Plains", wellsprings: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			artifacts := []*game.Permanent{addCombatPermanent(g, game.Player1, reg.Lookup("Inspiring Statuary"))}
			for range tc.wellsprings {
				artifacts = append(artifacts, addCombatPermanent(g, game.Player1, reg.Lookup("Ichor Wellspring")))
			}
			plains := addBasicLandPermanent(g, game.Player1, types.Plains)
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Spectral Procession"))
			life := g.Players[game.Player1].Life

			cast := castFromHandWithAnnouncements(e, g, cardID, nil, &phyrexianWindowAgent{})

			if cast != tc.wantCast {
				t.Fatalf("cast=%t, want %t", cast, tc.wantCast)
			}
			if !tc.wantCast {
				requireCastReversed(t, g, cardID, life)
				return
			}
			requireSpellOnStack(t, g, cardID, 1)
			requireTapped(t, map[*game.Permanent]bool{plains: true})
			for _, artifact := range artifacts {
				requireTapped(t, map[*game.Permanent]bool{artifact: true})
			}
			if !slices.ContainsFunc(g.Events, func(event game.Event) bool {
				return event.Kind == game.EventSpellCast && event.CardID == cardID
			}) {
				t.Fatal("no spell-cast event")
			}
		})
	}
}

// TestNonManaTapsLeaveManaSourcesForRemainingCost casts the registered
// Porcelain Legionnaire ({2}{W/P}) announcing {W} for {W/P}, with {C} floating.
// With improvise from Archway of Innovation, Ichor Wellspring pays one generic
// mana and Marble Diamond's {W} pays {W/P}; with convoke from Wand of the
// Worldsoul, Grizzly Bears pays one generic mana and Birds of Paradise taps for
// {W}. Tapping the mana source for the mechanic too would leave {W/P}
// unpayable, so the mechanic pays less generic mana and {C} pays the rest (CR
// 601.2h, 702.51a, 702.126a). Without the {C} the cost can't be paid and the
// cast is reversed (CR 733.1).
func TestNonManaTapsLeaveManaSourcesForRemainingCost(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		grant, tapper, manaSource string
		grantMana                 mana.Color
		floating                  bool
		improvise                 bool
	}{
		{name: "improvise", grant: "Archway of Innovation", grantMana: mana.U, tapper: "Ichor Wellspring", manaSource: "Marble Diamond", floating: true, improvise: true},
		{name: "improvise without the floating mana", grant: "Archway of Innovation", grantMana: mana.U, tapper: "Ichor Wellspring", manaSource: "Marble Diamond", improvise: true},
		{name: "convoke", grant: "Wand of the Worldsoul", tapper: "Grizzly Bears", manaSource: "Birds of Paradise", floating: true},
		{name: "convoke without the floating mana", grant: "Wand of the Worldsoul", tapper: "Grizzly Bears", manaSource: "Birds of Paradise"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := cards.NewDefaultRegistry()
			e, g := newEffectiveCascadeGame()
			grant := addCombatPermanent(g, game.Player1, reg.Lookup(tc.grant))
			if tc.grantMana != "" {
				g.Players[game.Player1].ManaPool.Add(tc.grantMana, 1)
			}
			activateNextSpellGrant(t, e, g, grant)
			tapper := addCombatPermanent(g, game.Player1, reg.Lookup(tc.tapper))
			source := addCombatPermanent(g, game.Player1, reg.Lookup(tc.manaSource))
			if tc.floating {
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			}
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Porcelain Legionnaire"))
			life := g.Players[game.Player1].Life
			clone := g.Clone()

			cast := castFromHandWithAnnouncements(e, clone, cardID, nil, &phyrexianWindowAgent{payMana: true})

			if cast != tc.floating {
				t.Fatalf("cast=%t, want %t", cast, tc.floating)
			}
			if !tc.floating {
				requireCastReversed(t, clone, cardID, life)
				return
			}
			requireSpellOnStack(t, clone, cardID, 2)
			for _, permanent := range []*game.Permanent{tapper, source} {
				if cloned, ok := clone.PermanentByID(permanent.ObjectID); !ok || !cloned.Tapped {
					t.Fatalf("permanent %v was not tapped", permanent.ObjectID)
				}
			}
			if clone.Players[game.Player1].Life != life || clone.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatalf("life=%d pool=%d, want life %d and the {C} spent", clone.Players[game.Player1].Life, clone.Players[game.Player1].ManaPool.Total(), life)
			}
			if tapper.Tapped || source.Tapped || g.Players[game.Player1].ManaPool.Total() != 1 || !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("casting in the clone changed the original game")
			}
		})
	}
}
