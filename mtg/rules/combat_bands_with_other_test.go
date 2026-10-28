package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 702.22a-m. 702.22a: "bands with other [quality]" is a form of banding.
// 702.22b: an effect that makes a creature lose banding also removes all its
// "bands with other" abilities; one removing "bands with other" leaves banding.
// 702.22c: a band is either creatures with banding plus up to one without, or
// one or more attacking [quality] creatures with "bands with other [quality]"
// plus any number of other attacking [quality] creatures. 702.22e: a band
// survives its members losing these abilities. 702.22j-k: the defending
// (active) player divides the damage of a creature blocked by (blocking) a
// creature with banding, or both a [quality] creature with "bands with other
// [quality]" and another [quality] creature. 702.22m: multiple "bands with
// other" abilities of the same kind are redundant. Oracle text, pinned corpus:
// Mountain Stronghold af147bb6-aa46-4fc9-af96-de3740c19fd5 "Red legendary
// creatures you control have 'bands with other legendary creatures.'" (Cathedral
// of Serra, Adventurers' Guildhouse, Unholy Citadel and Seafarer's Quay are the
// same for white, green, black and blue); Master of the Hunt
// cdf89ad2-ce08-4708-88c2-8acdc861526d "{2}{G}{G}: Create a 1/1 green Wolf
// creature token named Wolves of the Hunt. It has 'bands with other creatures
// named Wolves of the Hunt.'"; Shelkin Brownie
// 507ce455-bd02-49f9-bf64-4b2d21ffe0ea "{T}: Target creature loses all 'bands
// with other' abilities until end of turn."; Tolaria
// 9879a4f3-3b9c-45cf-af03-7f2ae4c689b4 "{T}: Target creature loses banding and
// all 'bands with other' abilities until end of turn. Activate only during any
// upkeep step."; Old Fogey "bands with other Dinosaurs". Scryfall rulings
// retrieved for this unit: Shelkin Brownie 2004-10-04 "Can only remove 'bands
// with other' and not normal 'banding' ability."; Master of the Hunt 2004-10-04
// is about repeated activation only; Tolaria and Mountain Stronghold have none.
// These scenarios use test-local shapes of those effects; a later card unit
// admits the cards themselves.

var (
	bwoLegendary = game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}}
	bwoDinosaur  = game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Dinosaur}}
	bwoWolves    = game.Selection{RequiredTypes: []types.Card{types.Creature}, Name: "Wolves of the Hunt"}
)

type bwoFace struct {
	name      string
	colors    []color.Color
	legendary bool
	subtypes  []types.Sub
	power     int
	toughness int
	statics   []game.StaticAbility
}

func bwoCreature(g *game.Game, controller game.PlayerID, face bwoFace) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:            face.name,
		Colors:          face.colors,
		Types:           []types.Card{types.Creature},
		Subtypes:        face.subtypes,
		Power:           opt.Val(game.PT{Value: face.power}),
		Toughness:       opt.Val(game.PT{Value: face.toughness}),
		StaticAbilities: face.statics,
	}}
	if face.legendary {
		def.Supertypes = []types.Super{types.Legendary}
	}
	permanent := addCombatPermanent(g, controller, def)
	permanent.SummoningSick = false
	return permanent
}

// addStrongholdGrant adds a land shaped like Mountain Stronghold: red
// legendary creatures its controller controls have "bands with other legendary
// creatures."
func addStrongholdGrant(g *game.Game, controller game.PlayerID) *game.Permanent {
	body := game.BandsWithOtherStaticAbility(bwoLegendary, "legendary creatures")
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Stronghold Test Land",
		Types: []types.Card{types.Land},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Supertypes:    []types.Super{types.Legendary},
				ColorsAny:     []color.Color{color.Red},
			}),
			AddAbilities: []game.Ability{&body},
		}}}},
	}})
}

// addResolvedAbilityEffect adds an until-end-of-turn layer-6 effect from a
// resolved ability, timestamped after everything already on the battlefield.
func addResolvedAbilityEffect(g *game.Game, permanent *game.Permanent, effect game.ContinuousEffect) {
	effect.ID = g.IDGen.Next()
	effect.Layer = game.LayerAbility
	effect.Duration = game.DurationUntilEndOfTurn
	effect.AffectedObjectID = permanent.ObjectID
	effect.Timestamp = game.Timestamp(g.IDGen.Next())
	g.ContinuousEffects = append(g.ContinuousEffects, effect)
}

// bandLegal reports whether the creatures may attack Player2 as one band.
func bandLegal(g *game.Game, members ...*game.Permanent) bool {
	attack := bandAttack(members...)
	payload, _ := action.DeclareAttackersWithBands(attack.Attackers, attack.Bands).DeclareAttackersPayload()
	return combatEngine{NewEngine(nil)}.attackersLegal(g, game.Player1, payload)
}

// bandChoicePolicy attacks Player2 with every creature in attack through the
// staged chooser, records which creatures each band request offers, and adds
// the creatures in addTo[anchor] in order.
type bandChoicePolicy struct {
	simPassPolicy
	attack  map[id.ID]bool
	addTo   map[id.ID][]id.ID
	offered map[id.ID][][]id.ID
}

func (p *bandChoicePolicy) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.ChooseAttackers()
}

func (p *bandChoicePolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceCombatDeclaration:
		attacking := p.attack[request.CombatAttacker]
		for _, option := range request.Options {
			if option.Attack.Exists == attacking && (!attacking || option.Attack.Val.Target == game.AttackTarget{Player: game.Player2}) {
				return []int{option.Index}
			}
		}
	case game.ChoiceAttackingBand:
		var offered []id.ID
		for _, option := range request.Options[1:] {
			offered = append(offered, option.Attack.Val.Attacker)
		}
		slices.Sort(offered)
		p.offered[request.CombatAttacker] = append(p.offered[request.CombatAttacker], offered)
		for _, next := range p.addTo[request.CombatAttacker] {
			for _, option := range request.Options[1:] {
				if option.Attack.Val.Attacker == next {
					return []int{option.Index}
				}
			}
		}
		return []int{0}
	default:
	}
	return request.DefaultSelection
}

func stageBands(t *testing.T, g *game.Game, attackers []*game.Permanent, addTo map[*game.Permanent][]*game.Permanent) *bandChoicePolicy {
	t.Helper()
	policy := &bandChoicePolicy{attack: make(map[id.ID]bool), addTo: make(map[id.ID][]id.ID), offered: make(map[id.ID][][]id.ID)}
	for _, attacker := range attackers {
		policy.attack[attacker.ObjectID] = true
	}
	for anchor, members := range addTo {
		for _, member := range members {
			policy.addTo[anchor.ObjectID] = append(policy.addTo[anchor.ObjectID], member.ObjectID)
		}
	}
	agents := simPassPolicies()
	agents[game.Player1] = policy
	combatEngine{NewEngine(nil)}.declareAttackers(g, agents, &TurnLog{})
	if len(g.Combat.Attackers) != len(attackers) {
		t.Fatalf("declared %d attackers, want %d", len(g.Combat.Attackers), len(attackers))
	}
	return policy
}

func objectIDs(permanents ...*game.Permanent) []id.ID {
	ids := make([]id.ID, 0, len(permanents))
	for _, permanent := range permanents {
		ids = append(ids, permanent.ObjectID)
	}
	slices.Sort(ids)
	return ids
}

// TestStrongholdLegendariesBandAndDivideBlockerDamage is the Mountain
// Stronghold shape on the attacking side (CR 702.22c, 702.22k). The red
// legendary gains "bands with other legendary creatures"; the green legendary
// and a red non-legendary don't. The red and green legendaries may band, but no
// band may include the non-legendary, nor form without the red holder. The
// staged chooser offers the red anchor only the green legendary and doesn't let
// a creature without the ability anchor a band. In live combat a 3/3 blocking
// only the green legendary blocks the whole band, and the active player is
// asked to divide its damage, all 3 onto the red legendary.
func TestStrongholdLegendariesBandAndDivideBlockerDamage(t *testing.T) {
	g := bandCombatGame()
	addStrongholdGrant(g, game.Player1)
	red := bwoCreature(g, game.Player1, bwoFace{name: "Red Legend", colors: []color.Color{color.Red}, legendary: true, power: 2, toughness: 4})
	green := bwoCreature(g, game.Player1, bwoFace{name: "Green Legend", colors: []color.Color{color.Green}, legendary: true, power: 2, toughness: 2})
	bear := bwoCreature(g, game.Player1, bwoFace{name: "Red Bear", colors: []color.Color{color.Red}, power: 2, toughness: 2})
	otherGreen := bwoCreature(g, game.Player1, bwoFace{name: "Other Green Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})

	if !hasKeyword(g, red, game.BandsWithOther) || hasKeyword(g, green, game.BandsWithOther) || hasKeyword(g, bear, game.BandsWithOther) {
		t.Fatal("only the red legendary should have bands with other")
	}
	for _, test := range []struct {
		name    string
		members []*game.Permanent
		legal   bool
	}{
		{"red and green legendaries", []*game.Permanent{red, green}, true},
		{"three legendaries", []*game.Permanent{red, green, otherGreen}, true},
		{"with a non-legendary", []*game.Permanent{red, green, bear}, false},
		{"holder and non-legendary", []*game.Permanent{red, bear}, false},
		{"legendaries without the holder", []*game.Permanent{green, otherGreen}, false},
	} {
		if got := bandLegal(g, test.members...); got != test.legal {
			t.Errorf("%s: legal = %v, want %v", test.name, got, test.legal)
		}
	}

	staged := g.Clone()
	policy := stageBands(t, staged, []*game.Permanent{red, green, bear}, map[*game.Permanent][]*game.Permanent{red: {green}})
	// Once the green legendary joins no candidate is left, so there's no second ask.
	if got := policy.offered[red.ObjectID]; len(policy.offered) != 1 || len(got) != 1 || !slices.Equal(got[0], objectIDs(green)) {
		t.Fatalf("band offers = %v, want one request anchored at the red legendary offering only the green legendary", policy.offered)
	}
	if !sameBands(staged.Combat.Bands, []game.AttackingBand{bandOf(red, green)}) {
		t.Fatalf("bands = %+v, want red and green", staged.Combat.Bands)
	}

	blocker := bandingTestCreature(g, game.Player2, 3, 3)
	attackerAgent := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{red.CardInstanceID: 3})
	}}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: green.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(red, green), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if _, ok := blockEvents(after.Events, game.EventAttackerBecameBlocked)[[2]id.ID{blocker.ObjectID, red.ObjectID}]; !ok {
		t.Fatal("the red legendary did not become blocked with its band")
	}
	if len(defender.requests) != 0 || len(attackerAgent.requests) != 1 {
		t.Fatalf("requests defender/active = %d/%d, want 0/1", len(defender.requests), len(attackerAgent.requests))
	}
	request := attackerAgent.requests[0]
	if request.Kind != game.ChoiceCombatDamageDivision || request.Player != game.Player1 ||
		request.CombatBlocker != blocker.ObjectID || len(request.Options) != 2 {
		t.Fatalf("division request = %+v, want Player1 dividing the 3/3's damage between both legendaries", request)
	}
	if got := markedDamage(t, after, red); got != 3 {
		t.Fatalf("red legendary damage = %d, want the active player's 3", got)
	}
	if got := markedDamage(t, after, green); got != 0 {
		t.Fatalf("green legendary damage = %d, want 0", got)
	}
}

// TestStrongholdBlockersControlAttackerDamage is the defending side of CR
// 702.22j: a red legendary with "bands with other legendary creatures" and
// another legendary block a 4/4 trampler, so the defending player divides its
// damage and the protective default spreads it 2/2 with nothing trampling
// over. With a non-legendary second blocker the ability doesn't apply and the
// active player assigns lethal damage in order, 3/1. In live combat the
// division request goes to the defending player.
func TestStrongholdBlockersControlAttackerDamage(t *testing.T) {
	setup := func(secondLegendary bool) (*game.Game, *game.Permanent, *game.Permanent, *game.Permanent) {
		g := bandCombatGame()
		addStrongholdGrant(g, game.Player2)
		attacker := bandingTestCreature(g, game.Player1, 4, 4, game.Trample)
		attacker.SummoningSick = false
		first := bwoCreature(g, game.Player2, bwoFace{name: "Red Legend", colors: []color.Color{color.Red}, legendary: true, power: 1, toughness: 3})
		second := bwoCreature(g, game.Player2, bwoFace{name: "Second Blocker", colors: []color.Color{color.Green}, legendary: secondLegendary, power: 1, toughness: 3})
		return g, attacker, first, second
	}
	for _, test := range []struct {
		name            string
		secondLegendary bool
		first, second   int
		life            int
	}{
		{"two legendaries", true, 2, 2, 40},
		{"one legendary", false, 3, 1, 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, attacker, first, second := setup(test.secondLegendary)
			g.Combat = doubleBlockCombat(attacker, first, second)
			NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
			if first.MarkedDamage != test.first || second.MarkedDamage != test.second || g.Players[game.Player2].Life != test.life {
				t.Fatalf("damage %d/%d, life %d; want %d/%d, life %d", first.MarkedDamage, second.MarkedDamage,
					g.Players[game.Player2].Life, test.first, test.second, test.life)
			}
		})
	}

	g, attacker, first, second := setup(true)
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID}, {Blocker: second.ObjectID, Blocking: attacker.ObjectID},
	}}
	defender.answer = func(request game.ChoiceRequest) []int {
		if request.Kind != game.ChoiceCombatDamageDivision {
			return request.DefaultSelection
		}
		return divisionSelection(request, map[id.ID]int{first.CardInstanceID: 4})
	}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(attacker, game.AttackTarget{Player: game.Player2})}}
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if len(attackerAgent.requests) != 0 {
		t.Fatalf("active player was asked %+v; the defending player divides", attackerAgent.requests)
	}
	if !slices.ContainsFunc(defender.requests, func(request game.ChoiceRequest) bool {
		return request.Kind == game.ChoiceCombatDamageDivision && request.Player == game.Player2 && request.CombatAttacker == attacker.ObjectID
	}) {
		t.Fatalf("defender requests = %+v, want a division of the 4/4's damage", defender.requests)
	}
	if permanentWithCardOnBattlefield(after, first.CardInstanceID) {
		t.Fatal("the red legendary should die to the 4 damage its controller assigned to it")
	}
	if got := markedDamage(t, after, second); got != 0 {
		t.Fatalf("second blocker damage = %d, want 0", got)
	}
}

// TestWolvesOfTheHuntTokensBandOnlyWithEachOther creates two Wolves of the
// Hunt tokens through CreateToken, as Master of the Hunt does. Each has "bands
// with other creatures named Wolves of the Hunt", so they may band together but
// not with a differently named Wolf (CR 702.22c). A creature blocking one of
// the banded wolves blocks both, and the active player divides its damage.
func TestWolvesOfTheHuntTokensBandOnlyWithEachOther(t *testing.T) {
	g := bandCombatGame()
	wolfDef := &game.CardDef{CardFace: game.CardFace{
		Name:            "Wolves of the Hunt",
		Colors:          []color.Color{color.Green},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Wolf},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.BandsWithOtherStaticAbility(bwoWolves, "creatures named Wolves of the Hunt")},
	}}
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.CreateToken{Amount: game.Fixed(2), Source: game.TokenDef(wolfDef)}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	var wolves []*game.Permanent
	for _, permanent := range g.Battlefield {
		if def, ok := permanentCardDef(g, permanent); ok && permanent.Token && def.Name == "Wolves of the Hunt" {
			wolves = append(wolves, permanent)
		}
	}
	if len(wolves) != 2 {
		t.Fatalf("created %d Wolves of the Hunt, want 2", len(wolves))
	}
	otherWolf := bwoCreature(g, game.Player1, bwoFace{name: "Grizzled Wolf", colors: []color.Color{color.Green}, subtypes: []types.Sub{types.Wolf}, power: 1, toughness: 1})

	if !bandLegal(g, wolves[0], wolves[1]) {
		t.Fatal("two Wolves of the Hunt can't band")
	}
	if bandLegal(g, wolves[0], otherWolf) || bandLegal(g, wolves[0], wolves[1], otherWolf) {
		t.Fatal("a Wolves of the Hunt banded with a differently named creature")
	}

	blocker := bandingTestCreature(g, game.Player2, 2, 2)
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: wolves[0].ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	after, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(wolves[0], wolves[1]), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if _, ok := blockEvents(after.Events, game.EventAttackerBecameBlocked)[[2]id.ID{blocker.ObjectID, wolves[1].ObjectID}]; !ok {
		t.Fatal("the second wolf did not become blocked with its band")
	}
	if len(defender.requests) != 0 || len(attackerAgent.requests) != 1 || attackerAgent.requests[0].CombatBlocker != blocker.ObjectID {
		t.Fatalf("requests defender/active = %+v/%+v, want only the active player dividing the blocker's damage",
			defender.requests, attackerAgent.requests)
	}
}

// TestLosingBandingLosesBandsWithOther is the Tolaria shape (CR 702.22b): an
// effect removing only banding also removes "bands with other", so the red
// legendary can't band with other legendaries and two blocking legendaries no
// longer control the attacker's damage. Only abilities gained after the
// removal in timestamp order remain: a later grant of "bands with other
// Dinosaurs" works, while the earlier legendary grant stays lost.
func TestLosingBandingLosesBandsWithOther(t *testing.T) {
	g := bandCombatGame()
	addStrongholdGrant(g, game.Player1)
	red := bwoCreature(g, game.Player1, bwoFace{name: "Red Legendary Dinosaur", colors: []color.Color{color.Red}, legendary: true, subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 3})
	legend := bwoCreature(g, game.Player1, bwoFace{name: "Green Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 3})
	dinosaur := bwoCreature(g, game.Player1, bwoFace{name: "Plain Dinosaur", colors: []color.Color{color.Green}, subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 3})
	if !bandLegal(g, red, legend) || bandLegal(g, red, dinosaur) {
		t.Fatal("fixture: the red legendary should band only with the legendary")
	}
	// The blocker check reads only the blocking group, so the attacking-side
	// creatures stand in for blockers.
	if !defendingPlayerAssignsAttackerDamage(g, []*game.Permanent{red, legend}) {
		t.Fatal("fixture: blocking legendaries with bands with other should control the attacker's damage")
	}

	addResolvedAbilityEffect(g, red, game.ContinuousEffect{RemoveKeywords: []game.Keyword{game.Banding}})
	if hasKeyword(g, red, game.BandsWithOther) || hasKeyword(g, red, game.Banding) {
		t.Fatal("losing banding left bands with other")
	}
	if bandLegal(g, red, legend) {
		t.Fatal("a creature that lost banding still banded with other legendaries")
	}
	staged := g.Clone()
	if policy := stageBands(t, staged, []*game.Permanent{red, legend}, nil); len(policy.offered) != 0 {
		t.Fatalf("band offers = %v, want none", policy.offered)
	}
	if defendingPlayerAssignsAttackerDamage(g, []*game.Permanent{red, legend}) {
		t.Fatal("blockers without banding or bands with other control the attacker's damage")
	}

	later := game.BandsWithOtherStaticAbility(bwoDinosaur, "Dinosaurs")
	addResolvedAbilityEffect(g, red, game.ContinuousEffect{AddAbilities: []game.Ability{&later}})
	if !hasKeyword(g, red, game.BandsWithOther) || !bandLegal(g, red, dinosaur) {
		t.Fatal("bands with other gained after the removal was lost")
	}
	if bandLegal(g, red, legend) {
		t.Fatal("bands with other legendary creatures gained before the removal came back")
	}
}

func mustPermanent(t *testing.T, g *game.Game, permanent *game.Permanent) *game.Permanent {
	t.Helper()
	current, ok := permanentByObjectID(g, permanent.ObjectID)
	if !ok {
		t.Fatalf("%v is not on the battlefield", permanent.ObjectID)
	}
	return current
}

// TestLosingBandsWithOtherKeepsBanding is the Shelkin Brownie shape and its
// ruling: removing "bands with other" leaves printed banding (CR 702.22b). Its
// two "bands with other legendary creatures" abilities are redundant (702.22m)
// and both are lost. Three legendaries, only one with banding, may band through
// "bands with other" before the removal but not after, while that creature can
// still band with one creature without banding.
func TestLosingBandsWithOtherKeepsBanding(t *testing.T) {
	g := bandCombatGame()
	addStrongholdGrant(g, game.Player1)
	holder := bwoCreature(g, game.Player1, bwoFace{
		name: "Banding Red Legend", colors: []color.Color{color.Red}, legendary: true, power: 1, toughness: 1,
		statics: []game.StaticAbility{
			{KeywordAbilities: game.SimpleKeywords(game.Banding)},
			game.BandsWithOtherStaticAbility(bwoLegendary, "legendary creatures"),
		},
	})
	first := bwoCreature(g, game.Player1, bwoFace{name: "First Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})
	second := bwoCreature(g, game.Player1, bwoFace{name: "Second Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})
	if len(bandsWithOtherQualities(g, holder)) != 2 || !bandLegal(g, holder, first, second) {
		t.Fatal("fixture: the holder should have both abilities and band with two legendaries")
	}

	addResolvedAbilityEffect(g, holder, game.ContinuousEffect{RemoveKeywords: []game.Keyword{game.BandsWithOther}})
	if !hasKeyword(g, holder, game.Banding) {
		t.Fatal("losing bands with other removed banding")
	}
	if hasKeyword(g, holder, game.BandsWithOther) || len(bandsWithOtherQualities(g, holder)) != 0 {
		t.Fatal("a bands with other ability survived its removal")
	}
	if bandLegal(g, holder, first, second) {
		t.Fatal("two creatures without banding joined a band after bands with other was lost")
	}
	if !bandLegal(g, holder, first) {
		t.Fatal("the creature's banding no longer forms a band")
	}
}

// TestBandsWithOtherBandPersistsAfterLoss is CR 702.22e: a band of legendaries
// formed through "bands with other" stays a band after the holder loses banding
// and "bands with other", so a creature blocking the other member blocks both.
// The active player no longer controls the blocker's damage once neither
// member has the ability.
func TestBandsWithOtherBandPersistsAfterLoss(t *testing.T) {
	g := bandCombatGame()
	addStrongholdGrant(g, game.Player1)
	red := bwoCreature(g, game.Player1, bwoFace{name: "Red Legend", colors: []color.Color{color.Red}, legendary: true, power: 1, toughness: 1})
	green := bwoCreature(g, game.Player1, bwoFace{name: "Green Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})
	blocker := bandingTestCreature(g, game.Player2, 3, 3)
	attack := bandAttack(red, green)
	engine := NewEngine(nil)
	declared, ok := engine.Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	members := []*game.Permanent{mustPermanent(t, declared, red), mustPermanent(t, declared, green)}
	if !activePlayerAssignsBlockerDamage(declared, members) {
		t.Fatal("fixture: the band's bands with other should give the active player the blocker's division")
	}

	addResolvedAbilityEffect(declared, red, game.ContinuousEffect{RemoveKeywords: []game.Keyword{game.Banding}})
	declared.Turn.Step = game.StepDeclareBlockers
	if !(combatEngine{engine}).applyBlockers(declared, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: green.ObjectID}}}) {
		t.Fatal("block rejected")
	}
	want := []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: green.ObjectID}, {Blocker: blocker.ObjectID, Blocking: red.ObjectID}}
	if len(declared.Combat.Blockers) != 2 || !slices.Contains(declared.Combat.Blockers, want[0]) || !slices.Contains(declared.Combat.Blockers, want[1]) {
		t.Fatalf("blockers = %+v, want the block to reach the whole band %+v", declared.Combat.Blockers, want)
	}
	members = []*game.Permanent{mustPermanent(t, declared, red), mustPermanent(t, declared, green)}
	if activePlayerAssignsBlockerDamage(declared, members) {
		t.Fatal("the active player still divides after the band lost bands with other")
	}
}

// TestDifferentBandsWithOtherQualitiesAreNotRedundant: a legendary Dinosaur
// with both "bands with other legendary creatures" and "bands with other
// Dinosaurs" (CR 702.22m) may band with a legendary or with a Dinosaur, but not
// with both, since no single quality covers all three.
func TestDifferentBandsWithOtherQualitiesAreNotRedundant(t *testing.T) {
	g := bandCombatGame()
	holder := bwoCreature(g, game.Player1, bwoFace{
		name: "Legendary Dinosaur", legendary: true, subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 1,
		statics: []game.StaticAbility{
			game.BandsWithOtherStaticAbility(bwoLegendary, "legendary creatures"),
			game.BandsWithOtherStaticAbility(bwoDinosaur, "Dinosaurs"),
		},
	})
	legend := bwoCreature(g, game.Player1, bwoFace{name: "Legend", legendary: true, power: 1, toughness: 1})
	dinosaur := bwoCreature(g, game.Player1, bwoFace{name: "Dinosaur", subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 1})
	if !bandLegal(g, holder, legend) || !bandLegal(g, holder, dinosaur) {
		t.Fatal("each quality should allow its own band")
	}
	if bandLegal(g, holder, legend, dinosaur) {
		t.Fatal("a band mixed two qualities")
	}
	staged := g.Clone()
	policy := stageBands(t, staged, []*game.Permanent{holder, legend, dinosaur}, map[*game.Permanent][]*game.Permanent{holder: {legend}})
	// After the legendary joins, the Dinosaur can't, so there's no second ask.
	if got := policy.offered[holder.ObjectID]; len(policy.offered) != 1 || len(got) != 1 || !slices.Equal(got[0], objectIDs(legend, dinosaur)) {
		t.Fatalf("band offers = %v, want one request offering both", policy.offered)
	}
	if !sameBands(staged.Combat.Bands, []game.AttackingBand{bandOf(holder, legend)}) {
		t.Fatalf("bands = %+v, want the holder and the legendary", staged.Combat.Bands)
	}
}

// TestBandsWithOtherHolderMustJoinBand: a band formed through "bands with other
// [quality]" needs the holder among its members (CR 702.22c). A legendary
// holder attacking separately doesn't let two other legendaries band, nor
// let a white legendary with banding take two legendaries without banding. The staged
// chooser offers the banding anchor either legendary first, but after one
// joins only the holder remains addable; once the holder joins, the other
// legendary may too, and the declaration is legal.
func TestBandsWithOtherHolderMustJoinBand(t *testing.T) {
	g := bandCombatGame()
	addStrongholdGrant(g, game.Player1)
	anchor := bwoCreature(g, game.Player1, bwoFace{name: "Banding Anchor", colors: []color.Color{color.White}, legendary: true, power: 1, toughness: 1, statics: []game.StaticAbility{game.BandingStaticBody}})
	first := bwoCreature(g, game.Player1, bwoFace{name: "First Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})
	second := bwoCreature(g, game.Player1, bwoFace{name: "Second Legend", colors: []color.Color{color.Green}, legendary: true, power: 1, toughness: 1})
	holder := bwoCreature(g, game.Player1, bwoFace{name: "Red Legend", colors: []color.Color{color.Red}, legendary: true, power: 1, toughness: 1})

	if bandLegal(g, first, second) {
		t.Fatal("two legendaries banded through a holder outside the band")
	}
	if bandLegal(g, anchor, first, second) {
		t.Fatal("banding took two creatures without banding through a holder outside the band")
	}
	if !bandLegal(g, first, second, holder) || !bandLegal(g, anchor, first, second, holder) {
		t.Fatal("the holder joining should make the band legal")
	}

	attackers := []*game.Permanent{anchor, first, second, holder}
	staged := g.Clone()
	policy := stageBands(t, staged, attackers, map[*game.Permanent][]*game.Permanent{anchor: {first, second}})
	offers := policy.offered[anchor.ObjectID]
	if len(offers) != 2 || !slices.Equal(offers[0], objectIDs(first, second, holder)) || !slices.Equal(offers[1], objectIDs(holder)) {
		t.Fatalf("anchor offers = %v, want all three, then only the holder", offers)
	}
	if !sameBands(staged.Combat.Bands, []game.AttackingBand{bandOf(anchor, first)}) {
		t.Fatalf("bands = %+v, want the anchor and first legendary", staged.Combat.Bands)
	}

	staged = g.Clone()
	stageBands(t, staged, attackers, map[*game.Permanent][]*game.Permanent{anchor: {first, holder, second}})
	if !sameBands(staged.Combat.Bands, []game.AttackingBand{bandOf(anchor, first, second, holder)}) {
		t.Fatalf("bands = %+v, want all four", staged.Combat.Bands)
	}
}

// TestOneBodyWithDistinctBandsWithOtherQualities: a single static ability body
// carrying both "bands with other legendary creatures" and "bands with other
// Dinosaurs" grants both abilities (CR 702.22m), whether printed or gained. Its
// holder bands with a legendary or with a Dinosaur but not both, and blocking
// alongside a creature of either quality gives the defending player the
// attacker's damage division (CR 702.22j).
func TestOneBodyWithDistinctBandsWithOtherQualities(t *testing.T) {
	body := game.StaticAbility{
		Text: "Bands with other legendary creatures and bands with other Dinosaurs",
		KeywordAbilities: []game.KeywordAbility{
			game.BandsWithOtherKeyword{Quality: bwoLegendary},
			game.BandsWithOtherKeyword{Quality: bwoDinosaur},
		},
	}
	for _, granted := range []bool{false, true} {
		g := bandCombatGame()
		face := bwoFace{name: "Legendary Dinosaur", legendary: true, subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 1}
		if !granted {
			face.statics = []game.StaticAbility{body}
		}
		holder := bwoCreature(g, game.Player1, face)
		if granted {
			gained := body
			addResolvedAbilityEffect(g, holder, game.ContinuousEffect{AddAbilities: []game.Ability{&gained}})
		}
		legend := bwoCreature(g, game.Player1, bwoFace{name: "Legend", legendary: true, power: 1, toughness: 1})
		dinosaur := bwoCreature(g, game.Player1, bwoFace{name: "Dinosaur", subtypes: []types.Sub{types.Dinosaur}, power: 1, toughness: 1})
		if !bandLegal(g, holder, legend) || !bandLegal(g, holder, dinosaur) {
			t.Fatalf("granted=%v: each quality on the body should allow its own band", granted)
		}
		if bandLegal(g, holder, legend, dinosaur) {
			t.Fatalf("granted=%v: a band mixed two qualities", granted)
		}
		if !combatGroupControlsDamageDivision(g, []*game.Permanent{holder, dinosaur}) ||
			!combatGroupControlsDamageDivision(g, []*game.Permanent{holder, legend}) {
			t.Fatalf("granted=%v: blocking with a creature of either quality should control damage division", granted)
		}
	}
}
