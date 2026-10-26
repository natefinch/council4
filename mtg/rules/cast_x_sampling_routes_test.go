package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Legal-action X sampling on every cast route (#3378): the X values offered for
// a spell are the ones cast validation accepts once the total cost is lowered
// by cost reductions and paid as a cast pays it, with spell-restricted mana and
// with convoke, improvise or delve.
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     107.3a and 601.2b (the caster announces X), 601.2f (total cost is the
//     mana or alternative cost plus increases minus reductions, never below
//     {0}), 118.7a (generic reductions affect only the generic component),
//     702.51a-b (convoke: tap a creature rather than pay a colored mana of its
//     color or one generic mana, applied after the total cost is determined),
//     702.126a-b (improvise: tap an artifact rather than pay one generic mana,
//     likewise), 702.66a-b (delve: exile a card from your graveyard rather than
//     pay one generic mana, likewise).
//   - Mistcutter Hydra, printing 3c844ed7-8ed2-4a39-9134-e14e476ab0c4
//     (oracle_id 30b5b764-478a-46fa-8f95-a1f515925a2d), {X}{G} Creature —
//     Hydra: "This spell can't be countered. / Haste, protection from blue /
//     This creature enters with X +1/+1 counters on it."
//   - Heartless Summoning, printing c50fdb09-e06c-463c-87dd-3c46f4db889a
//     (oracle_id fa1be67e-a07e-42ce-88be-446acd643dc6), {1}{B}: "Creature
//     spells you cast cost {2} less to cast. / Creatures you control get
//     -1/-1." Ruling 2025-01-24: it reduces only the generic mana portion.
//   - Humble Naturalist, printing 8f56705d-eb64-4cef-b716-edbeac60bf79
//     (oracle_id ff3a69a7-a244-4958-9e8b-d153942c1ba6): "{T}: Add one mana of
//     any color. Spend this mana only to cast a creature spell."
//   - Vodalian Arcanist, printing a52e90c0-b012-4ce5-8462-1e33c7143de5
//     (oracle_id 1d07856b-d9ef-418c-bd06-3888dbcaddf4): "{T}: Add {C}. Spend
//     this mana only to cast an instant or sorcery spell."
//   - Infirmary Healer // Stream of Life, printing
//     911442e3-3003-4683-a766-e791e9553667 (oracle_id
//     4cbbdebc-b807-4d7b-92dd-312637d2a5de): "This creature enters prepared."
//     // Stream of Life {X}{G} Sorcery: "Target player gains X life." Ruling
//     2026-03-20: the prepared creature's controller casts the copy of its
//     prepare spell from exile; doing so isn't an alternative cost.
//   - Chord of Calling, printing b18fe7e0-8344-40cc-b242-83f01c6be7a6
//     (oracle_id 6789a170-f2c5-4fc0-8a45-2b2361e67410), {X}{G}{G}{G} Instant:
//     "Convoke ... Search your library for a creature card with mana value X or
//     less, put it onto the battlefield, then shuffle." Rulings 2024-01-12:
//     choose X first; X and cost changes determine the total cost, then
//     creatures can be tapped to help pay it, a green creature paying {G}.
//   - Grolnok, the Omnivore, printing 7aaba76b-9cec-4c2b-b0eb-8f44201f6422
//     (oracle_id 53029460-50a0-4934-8fa9-0271517cd5d0): "You may play lands and
//     cast spells from among cards you own in exile with croak counters on
//     them."
//   - Whir of Invention, printing 0279fd3c-9252-4958-9d7a-5f33aa25907e
//     (oracle_id 152b91c9-cc07-4ca8-944f-9bc2242a2283), {X}{U}{U}{U} Instant:
//     "Improvise ... Search your library for an artifact card with mana value X
//     or less, put it onto the battlefield, then shuffle." Rulings 2017-02-09:
//     choose X first, then tap artifacts to pay generic mana of the total cost;
//     improvise can't pay {U}.
//   - Empty the Pits, printing e94b1ac7-dd7b-46a1-a74e-aa0563bab3cd
//     (oracle_id 089d91cf-ed7a-4859-8967-cad975a5127e), {X}{X}{B}{B}{B}{B}
//     Instant: "Delve ... Cast this spell only during your turn. Create X
//     tapped 2/2 black Zombie creature tokens." Ruling 2021-03-19: cards can
//     be exiled to pay only generic mana.
//   - Grizzly Bears, printing 409f9b88-f03e-40b6-9883-68c14c37c0de (green
//     vanilla creature); Ornithopter, printing
//     305078a5-ac18-4721-bba2-3434eba5b1cf ({0} artifact creature, Flying).
//   - Comprehensive Rules 107.3b (a spell cast paying neither its mana cost nor
//     an alternative cost that includes X must choose X=0), 118.9a (only one
//     alternative cost applies), 702.34a (flashback: cast from the graveyard
//     by paying [cost] rather than its mana cost).
//   - Firecat Blitz, printing d4e1d485-02d5-4a07-bcc6-d2a8d95763e8
//     (oracle_id 8f262a8b-2553-45c9-b37f-218a5547fcff), {X}{R}{R} Sorcery:
//     "Create X 1/1 red Elemental Cat creature tokens with haste. Exile them at
//     the beginning of the next end step. / Flashback—{R}{R}, Sacrifice X
//     Mountains." Ruling 2021-03-19: the total cost starts with the mana cost
//     or alternative cost (such as a flashback cost) being paid, adds
//     increases, then applies reductions.
//   - Goblin Electromancer, printing bd9eca71-2430-4112-a23a-fba76ca2efb4
//     (oracle_id 81f06f84-1580-43c0-89d5-08d34541a519): "Instant and sorcery
//     spells you cast cost {1} less to cast." Ruling 2018-10-05: it reduces
//     only generic mana, starting from the mana or alternative cost paid.
//   - Thalia, Guardian of Thraben, printing
//     c9f8b8fb-1cd8-450e-a1fe-892e7a323479 (oracle_id
//     9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee): "First strike / Noncreature
//     spells cost {1} more to cast." Ruling 2021-11-19: it affects your own
//     spells, increases applying to the mana or alternative cost paid.
//   - Court of Locthwain, printing a8e1f698-1c04-4f31-aeaa-b796c5eb5070
//     (oracle_id 6106ba65-f07b-41ec-8b59-e052bbde5529): "... exile the top
//     card of target opponent's library. You may play that card for as long as
//     it remains exiled, and mana of any type can be spent to cast it. If
//     you're the monarch, until end of turn, you may cast a spell from among
//     cards exiled with this enchantment without paying its mana cost."
//   - Exsanguinate, printing f11d7311-4066-4a5d-ba28-9857fa707a0b (oracle_id
//     8164b1e8-3350-465e-8a17-75f57d326344), {X}{B}{B} Sorcery: "Each
//     opponent loses X life. You gain life equal to the life lost this way."
//   - Bolas's Citadel, printing d2124603-d20e-40eb-97f0-a66323397ac2
//     (oracle_id 2bd111bb-ce02-414c-b5b7-e0e037d8d96b): "... If you cast a
//     spell this way, pay life equal to its mana value rather than pay its mana
//     cost." Ruling 2019-05-03: a spell with {X} in its mana cost cast this way
//     must choose X=0. The scenario uses the test permission
//     castSpellsFromLibraryTopPayLifePermanent with the same rule effect.
//   - Comprehensive Rules 601.2b (kicker intentions are announced with X),
//     702.33c-d (multikicker may be paid any number of times; each payment is
//     an additional cost).
//   - Comet Storm, printing bb5f586d-6bf0-4590-ad73-2d46b2a45b1a (oracle_id
//     9c2b262b-53c0-41a1-8f1a-0c336453d4bf), {X}{R}{R} Instant: "Multikicker
//     {1} ... Choose any target, then choose another target for each time this
//     spell was kicked. Comet Storm deals X damage to each of them." Rulings
//     2024-11-08: the total cost starts with the mana cost, adds increases
//     such as kicker, then applies reductions; 2010-03-01: X and the number of
//     kicks are declared together, then targets, then costs are paid.

// assertCastXRange gives player sorcery-speed priority and checks that casting
// cardID's face from source with targets is listed by legal-action
// enumeration and accepted by direct application exactly for X below limit.
func assertCastXRange(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, source zone.Type, face game.FaceIndex, targets []game.Target, limit int, label string) {
	t.Helper()
	giveMainPhasePriority(g, player)
	actions := engine.legalActions(g, player)
	for x := 0; x <= limit; x++ {
		act := action.CastSpellFaceFromZone(cardID, source, face, targets, x, nil)
		listed := actionsContain(actions, act)
		applied := engine.applyAction(g.Clone(), player, act)
		want := x < limit
		if listed != want || applied != want {
			t.Fatalf("%s X=%d: listed=%v applied=%v, want %v", label, x, listed, applied, want)
		}
	}
}

// castXAndCheckStack casts cardID for xValue and checks it is on the stack
// with that X.
func castXAndCheckStack(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, source zone.Type, face game.FaceIndex, targets []game.Target, xValue int) {
	t.Helper()
	giveMainPhasePriority(g, player)
	if !engine.applyAction(g, player, action.CastSpellFaceFromZone(cardID, source, face, targets, xValue, nil)) {
		t.Fatalf("casting the enumerated spell for X=%d failed", xValue)
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardID || top.Controller != player || top.XValue != xValue {
		t.Fatalf("stack top = %+v, want player %v's spell from card %v with X=%d", top, player, cardID, xValue)
	}
}

// TestCastXSamplingCostReductionRoutes covers CR 601.2f and 118.7a on every
// non-exile cast route: Heartless Summoning's {2} reduction lets one Forest
// cast Mistcutter Hydra for X up to 2 from hand, graveyard, the top of the
// library and the command zone, and each boundary X is listed exactly when
// validation accepts it.
func TestCastXSamplingCostReductionRoutes(t *testing.T) {
	for _, source := range []zone.Type{zone.Hand, zone.Graveyard, zone.Library, zone.Command} {
		t.Run(source.String(), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			groupExileRegistryPermanent(t, g, game.Player1, "Heartless Summoning")
			addBasicLandPermanent(g, game.Player1, types.Forest)
			hydra := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
			if source != zone.Hand {
				g.Players[game.Player1].Hand.Remove(hydra)
				destination, _ := destinationZone(g, game.Player1, source)
				destination.Add(hydra)
			}
			switch source {
			case zone.Graveyard:
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
					CastFromZone: zone.Graveyard, AffectedCardID: hydra, Duration: game.DurationPermanent,
				})
			case zone.Library:
				castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
			case zone.Command:
				trackCommanderID(g, game.Player1, hydra)
			default:
			}

			assertCastXRange(t, g, engine, game.Player1, hydra, source, game.FaceFront, nil, 3, "reduced Hydra")
			castXAndCheckStack(t, g, engine, game.Player1, hydra, source, game.FaceFront, nil, 2)
		})
	}
}

// TestCastXSamplingSpellRestrictedManaRoutes covers CR 106.6 and 601.2g-h on
// non-exile routes: Humble Naturalist's creature-only mana plus one Forest
// casts Mistcutter Hydra from hand for X up to 1, and Vodalian Arcanist's
// instant-or-sorcery mana plus one Forest casts a prepared Infirmary Healer's
// Stream of Life copy for X up to 1.
func TestCastXSamplingSpellRestrictedManaRoutes(t *testing.T) {
	t.Run("Humble Naturalist mana for a Hydra in hand", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		naturalist := groupExileRegistryPermanent(t, g, game.Player1, "Humble Naturalist")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		hydra := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		activateRestrictedManaAbility(t, g, engine, game.Player1, naturalist)

		assertCastXRange(t, g, engine, game.Player1, hydra, zone.Hand, game.FaceFront, nil, 2, "Hydra with Naturalist mana")
		castXAndCheckStack(t, g, engine, game.Player1, hydra, zone.Hand, game.FaceFront, nil, 1)
		if !g.Players[game.Player1].ManaPool.IsEmpty() {
			t.Fatalf("pool = %+v, want the Naturalist mana spent on the Hydra", g.Players[game.Player1].ManaPool)
		}
	})

	t.Run("Vodalian Arcanist mana for a prepared copy", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		arcanist := groupExileRegistryPermanent(t, g, game.Player1, "Vodalian Arcanist")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		healer := groupExileRegistryPermanent(t, g, game.Player1, "Infirmary Healer")
		healer.Prepared = true
		activateRestrictedManaAbility(t, g, engine, game.Player1, arcanist)
		atSelf := []game.Target{game.PlayerTarget(game.Player1)}

		assertCastXRange(t, g, engine, game.Player1, healer.CardInstanceID, zone.Battlefield, game.FaceAlternate, atSelf, 2, "Stream of Life with Arcanist mana")
		castXAndCheckStack(t, g, engine, game.Player1, healer.CardInstanceID, zone.Battlefield, game.FaceAlternate, atSelf, 1)
	})
}

// TestCastXSamplingConvokeImproviseDelve covers CR 702.51, 702.126 and 702.66:
// each pays part of the total cost after X is chosen, so the X values those
// payments afford are offered. Chord of Calling convokes from hand with
// generic-paying and color-paying creatures, and from exile under Grolnok with
// Grolnok itself convoking; Whir of Invention improvises with Ornithopters;
// Empty the Pits delves its {X}{X} from the graveyard.
func TestCastXSamplingConvokeImproviseDelve(t *testing.T) {
	t.Run("Chord of Calling convoked from hand for generic mana", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		for range 3 {
			addBasicLandPermanent(g, game.Player1, types.Forest)
		}
		groupExileRegistryPermanent(t, g, game.Player1, "Grizzly Bears")
		groupExileRegistryPermanent(t, g, game.Player1, "Grizzly Bears")
		chord := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Chord of Calling"))

		assertCastXRange(t, g, engine, game.Player1, chord, zone.Hand, game.FaceFront, nil, 3, "convoked Chord")
		castXAndCheckStack(t, g, engine, game.Player1, chord, zone.Hand, game.FaceFront, nil, 2)
		for _, permanent := range g.Battlefield {
			if !permanent.Tapped {
				t.Fatalf("permanent %v untapped after convoking Chord for X=2; want every land and creature tapped", permanent.ObjectID)
			}
		}
	})

	t.Run("Chord of Calling convoked from hand for green mana", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		for range 3 {
			groupExileRegistryPermanent(t, g, game.Player1, "Grizzly Bears")
		}
		chord := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Chord of Calling"))

		assertCastXRange(t, g, engine, game.Player1, chord, zone.Hand, game.FaceFront, nil, 2, "green-convoked Chord")
	})

	t.Run("Chord of Calling convoked from exile under Grolnok", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		groupExileRegistryPermanent(t, g, game.Player1, "Grizzly Bears")
		for range 3 {
			addBasicLandPermanent(g, game.Player1, types.Forest)
		}
		chord := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Chord of Calling"))
		g.AddExileCounter(chord, counter.Croak, 1)

		assertCastXRange(t, g, engine, game.Player1, chord, zone.Exile, game.FaceFront, nil, 3, "convoked croaked Chord")
		castXAndCheckStack(t, g, engine, game.Player1, chord, zone.Exile, game.FaceFront, nil, 2)
	})

	t.Run("Whir of Invention improvised from hand", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		for range 3 {
			addBasicLandPermanent(g, game.Player1, types.Island)
		}
		groupExileRegistryPermanent(t, g, game.Player1, "Ornithopter")
		groupExileRegistryPermanent(t, g, game.Player1, "Ornithopter")
		whir := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Whir of Invention"))

		assertCastXRange(t, g, engine, game.Player1, whir, zone.Hand, game.FaceFront, nil, 3, "improvised Whir")
		castXAndCheckStack(t, g, engine, game.Player1, whir, zone.Hand, game.FaceFront, nil, 2)
	})

	t.Run("Empty the Pits delved from hand", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		for range 4 {
			addBasicLandPermanent(g, game.Player1, types.Swamp)
			addCardToGraveyard(g, game.Player1, groupExileRegistryDef(t, "Grizzly Bears"))
		}
		pits := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Empty the Pits"))

		assertCastXRange(t, g, engine, game.Player1, pits, zone.Hand, game.FaceFront, nil, 3, "delved Empty the Pits")
		castXAndCheckStack(t, g, engine, game.Player1, pits, zone.Hand, game.FaceFront, nil, 2)
		if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
			t.Fatalf("graveyard size = %d after delving Empty the Pits for X=2, want 0", got)
		}
	})
}

// TestCastXSamplingAlternativeCostRoutes covers CR 601.2b, 601.2f, 107.3b and
// 118.9a: X candidates come from every cost option the cast may choose, not
// only the printed mana cost. Firecat Blitz's flashback cost, lowered by Goblin
// Electromancer against Thalia's tax, affords X up to the Mountains it can
// sacrifice though the reduced printed cost affords less; a free cast from
// exile under Court of Locthwain and a pay-life cast from the top of the
// library under a Bolas's Citadel permission each afford exactly X=0 with no
// mana; and a cast from exile pays under its own permission's spending rider.
func TestCastXSamplingAlternativeCostRoutes(t *testing.T) {
	t.Run("Firecat Blitz flashback reduced by Goblin Electromancer", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		electromancer := groupExileRegistryPermanent(t, g, game.Player1, "Goblin Electromancer")
		groupExileRegistryPermanent(t, g, game.Player2, "Thalia, Guardian of Thraben")
		for range 2 {
			addBasicLandPermanent(g, game.Player1, types.Mountain).Tapped = true
		}
		g.Players[game.Player1].ManaPool.Add(mana.R, 2)
		blitz := addCardToGraveyard(g, game.Player1, groupExileRegistryDef(t, "Firecat Blitz"))

		unreduced := g.Clone()
		for _, permanent := range unreduced.Battlefield {
			if permanent.ObjectID == electromancer.ObjectID {
				permanent.PhasedOut = true
			}
		}
		giveMainPhasePriority(unreduced, game.Player1)
		unreducedActions := engine.legalActions(unreduced, game.Player1)
		for x := 0; x <= 3; x++ {
			act := action.CastSpellFaceFromZone(blitz, zone.Graveyard, game.FaceFront, nil, x, nil)
			if actionsContain(unreducedActions, act) || engine.applyAction(unreduced.Clone(), game.Player1, act) {
				t.Fatalf("flashback X=%d payable without Goblin Electromancer's reduction against Thalia's tax", x)
			}
		}

		assertCastXRange(t, g, engine, game.Player1, blitz, zone.Graveyard, game.FaceFront, nil, 3, "reduced flashback Firecat Blitz")
		castXAndCheckStack(t, g, engine, game.Player1, blitz, zone.Graveyard, game.FaceFront, nil, 2)
		top, _ := g.Stack.Peek()
		if !top.Flashback {
			t.Fatal("Firecat Blitz was not cast for its flashback cost")
		}
		if got := g.Players[game.Player1].Graveyard.Size(); got != 2 {
			t.Fatalf("graveyard size = %d after flashback for X=2, want the two sacrificed Mountains", got)
		}
		for _, permanent := range g.Battlefield {
			if permanent.Controller == game.Player1 && permanent.ObjectID != electromancer.ObjectID {
				t.Fatalf("permanent %v still on the battlefield after flashback for X=2; want both Mountains sacrificed", permanent.ObjectID)
			}
		}
		if !g.Players[game.Player1].ManaPool.IsEmpty() {
			t.Fatalf("pool = %+v, want {R}{R} spent on the flashback cost", g.Players[game.Player1].ManaPool)
		}
	})

	t.Run("Court of Locthwain free cast of Exsanguinate from exile", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		court := groupExileRegistryPermanent(t, g, game.Player1, "Court of Locthwain")
		exsanguinate := addCardToLibrary(g, game.Player2, groupExileRegistryDef(t, "Exsanguinate"))
		g.Players[game.Player1].IsMonarch = true
		obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
			SourceID: court.ObjectID, SourceCardID: court.CardInstanceID,
			Targets: []game.Target{game.PlayerTarget(game.Player2)}}
		def := groupExileRegistryDef(t, "Court of Locthwain")
		for _, instruction := range def.TriggeredAbilities[1].Content.Modes[0].Sequence {
			engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}

		assertCastXRange(t, g, engine, game.Player1, exsanguinate, zone.Exile, game.FaceFront, nil, 1, "free Exsanguinate")
		castXAndCheckStack(t, g, engine, game.Player1, exsanguinate, zone.Exile, game.FaceFront, nil, 0)
	})

	t.Run("Exsanguinate from exile under its permission's spending rider", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		exsanguinate := riderExileCard(g, groupExileRegistryDef(t, "Exsanguinate"), game.ManaSpendAsAnyColor)
		g.Players[game.Player1].ManaPool.Add(mana.U, 3)

		assertCastXRange(t, g, engine, game.Player1, exsanguinate, zone.Exile, game.FaceFront, nil, 2, "Exsanguinate paid with blue mana")
		castXAndCheckStack(t, g, engine, game.Player1, exsanguinate, zone.Exile, game.FaceFront, nil, 1)
	})

	t.Run("Mistcutter Hydra cast from the library top for life", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
		hydra := addCardToLibrary(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.Players[game.Player1].Life = 20

		assertCastXRange(t, g, engine, game.Player1, hydra, zone.Library, game.FaceFront, nil, 1, "Hydra for life")
		castXAndCheckStack(t, g, engine, game.Player1, hydra, zone.Library, game.FaceFront, nil, 0)
		if got := g.Players[game.Player1].Life; got != 19 {
			t.Fatalf("life = %d after casting the X=0 Hydra for life, want 19", got)
		}
	})
}

// TestCastXSamplingMultikickedReducedSpell covers CR 601.2b, 601.2f and
// 702.33c-d for the combined-cost route: Comet Storm's multikicker payments add
// to the total cost that Goblin Electromancer's {1} reduction lowers, so three
// Mountains cast it for X up to 2 unkicked, 1 kicked once and 0 kicked twice,
// and each kick count's boundary X is listed exactly when validation accepts it.
func TestCastXSamplingMultikickedReducedSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	groupExileRegistryPermanent(t, g, game.Player1, "Goblin Electromancer")
	var mountains []*game.Permanent
	for range 3 {
		mountains = append(mountains, addBasicLandPermanent(g, game.Player1, types.Mountain))
	}
	storm := addCardToHand(g, game.Player1, groupExileRegistryDef(t, "Comet Storm"))
	giveMainPhasePriority(g, game.Player1)
	actions := engine.legalActions(g, game.Player1)
	opponents := []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3), game.PlayerTarget(game.Player4)}
	cast := func(kicks, x int) action.Action {
		return action.CastMultikickedSpellFaceFromZone(storm, zone.Hand, game.FaceFront, opponents[:kicks+1], x, nil, kicks)
	}
	for kicks, limit := range []int{3, 2, 1} {
		for x := 0; x <= limit; x++ {
			act := cast(kicks, x)
			listed := actionsContain(actions, act)
			applied := engine.applyAction(g.Clone(), game.Player1, act)
			want := x < limit
			if listed != want || applied != want {
				t.Fatalf("Comet Storm kicked %d times X=%d: listed=%v applied=%v, want %v", kicks, x, listed, applied, want)
			}
		}
	}

	if !engine.applyAction(g, game.Player1, cast(1, 1)) {
		t.Fatal("casting the enumerated once-kicked Comet Storm for X=1 failed")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != storm || top.XValue != 1 || top.KickerCount != 1 || len(top.Targets) != 2 {
		t.Fatalf("stack top = %+v, want Comet Storm kicked once for X=1 with two targets", top)
	}
	for _, mountain := range mountains {
		if !mountain.Tapped {
			t.Fatalf("Mountain %v untapped after paying {1}{R}{R} plus {1} kicker minus {1}; want all three tapped", mountain.ObjectID)
		}
	}
}
