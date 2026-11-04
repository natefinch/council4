package rules

import (
	"maps"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus Oracle text, with Scryfall rulings retrieved
// for the scenario cards. Knight of Sorrows (RNA 14, oracle
// 50f644d6-2597-4ad6-8ea8-6b09056f90c0; one 2019-01-25 ruling on blocking with
// afterlife), Bilbo, Retired Burglar (LTR 196, oracle
// 2e7e3be5-04dd-4ff4-b121-3e65b3e00699; seven 2023-06-16 Ring-tempts
// rulings), Static Snare (TDM 26, oracle b5b742d2-f19b-47c9-84bd-e85d975900e6;
// no rulings) and Clone (M14 47, oracle 42226b87-0746-4ebf-9fd0-108d508462af;
// copied enters abilities trigger as Clone enters). CR 2026-09-25: each
// paragraph and each keyword of a keyword line is a separate ability (113.2c);
// afterlife (702.135a) and typecycling (702.29e) are single keyword abilities;
// "enters or leaves" names more than one trigger event of one paragraph
// (603.2, 603.6); the return of an "until" exile is part of the exiling
// ability's effect (610.3); a copy copies printed abilities and their text
// (707.2).
const (
	knightOfSorrowsAfterlifeText = "Afterlife 1 (When this creature dies, create a 1/1 white and black Spirit creature token with flying.)"
	bilboEntersOrLeavesText      = "When Bilbo enters or leaves the battlefield, the Ring tempts you."
	timelessDragonCyclingText    = "Plainscycling {2} ({2}, Discard this card: Search your library for a Plains card, reveal it, put it into your hand, then shuffle.)"
)

// Every untexted triggered, activated, loyalty and chapter ability of every
// registered face has its own identifiable Oracle text, so neither the printed
// ability nor a copy of it is shown with unrelated abilities' text.
func TestRegisteredUntextedAbilitiesHaveOwnParagraphs(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, name := range registry.All() {
		def := registry.Lookup(name)
		if def == nil {
			continue
		}
		faces := []*game.CardFace{&def.CardFace}
		if def.Back.Exists {
			faces = append(faces, &def.Back.Val)
		}
		if def.Alternate.Exists {
			faces = append(faces, &def.Alternate.Val)
		}
		for _, face := range faces {
			paragraphs := printedAbilityParagraphs(face, face.OracleText)
			for index := range face.AbilityCount() {
				body := face.BodyAt(index)
				if oracleRulesText(game.BodyText(body)) != "" {
					continue
				}
				switch body.(type) {
				case *game.TriggeredAbility, *game.ActivatedAbility, *game.LoyaltyAbility, *game.ChapterAbility:
					if paragraphs[index] == "" {
						t.Errorf("%s ability %d (%T) has no identifiable Oracle text", face.Name, index, body)
					}
				}
			}
		}
	}
}

func TestPrintedAbilityParagraphsOfKeywordAndSplitTriggerCards(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	tests := []struct {
		name string
		want map[int]string
	}{{
		name: "Hraesvelgr of the First Brood",
		want: map[int]string{
			0: "Shiva's Aid — When Hraesvelgr enters and whenever you cast a noncreature spell, target creature gets +1/+0 until end of turn and can't be blocked this turn.",
			1: "Shiva's Aid — When Hraesvelgr enters and whenever you cast a noncreature spell, target creature gets +1/+0 until end of turn and can't be blocked this turn.",
		},
	}, {
		name: "Nulldrifter",
		want: map[int]string{
			0: "When you cast this spell, draw two cards.",
			1: "Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)",
		},
	}, {
		name: "Power Depot",
		want: map[int]string{2: "Modular 1"},
	}, {
		name: "Timeless Dragon",
		want: map[int]string{0: timelessDragonCyclingText},
	}, {
		name: "Unshakable Tail",
		want: map[int]string{
			1: "When this creature enters and at the beginning of your upkeep, surveil 1.",
			2: "When this creature enters and at the beginning of your upkeep, surveil 1.",
			3: "Whenever one or more creature cards are put into your graveyard from your library, investigate. (Create a Clue token. It's an artifact with \"{2}, Sacrifice this token: Draw a card.\")",
		},
	}, {
		name: "Lassoed by the Law",
		want: map[int]string{
			0: "When this enchantment enters, exile target nonland permanent an opponent controls until this enchantment leaves the battlefield.",
			1: "When this enchantment enters, create a 1/1 red Mercenary creature token with \"{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery.\"",
		},
	}}
	for _, tt := range tests {
		def := registry.Lookup(tt.name)
		if def == nil {
			t.Fatalf("%s is not in the default registry", tt.name)
		}
		if got := printedAbilityParagraphs(&def.CardFace, def.OracleText); !maps.Equal(got, tt.want) {
			t.Errorf("%s paragraphs = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestPrintedAbilityParagraphsPairKeywordsAndTriggerEvents(t *testing.T) {
	when := game.TriggeredAbility{Trigger: game.TriggerCondition{Type: game.TriggerWhen}}
	drawWhen := when
	drawWhen.Content = game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1)}}}}.Ability()
	face := game.CardFace{
		StaticAbilities:    []game.StaticAbility{{Text: "Flying"}},
		TriggeredAbilities: []game.TriggeredAbility{when, drawWhen, drawWhen},
	}
	got := printedAbilityParagraphs(&face, `
		Flying, annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
		When this creature enters or dies, draw a card.
	`)
	want := map[int]string{
		face.TriggeredAbilityIndex(0): "Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)",
		face.TriggeredAbilityIndex(1): "When this creature enters or dies, draw a card.",
		face.TriggeredAbilityIndex(2): "When this creature enters or dies, draw a card.",
	}
	if !maps.Equal(got, want) {
		t.Fatalf("paragraphs = %v, want %v", got, want)
	}

	ambiguous := game.CardFace{TriggeredAbilities: []game.TriggeredAbility{drawWhen, drawWhen, drawWhen}}
	if got := printedAbilityParagraphs(&ambiguous, `
		When this creature enters or dies, draw a card.
		When this creature attacks or blocks, draw a card.
	`); len(got) != 0 {
		t.Fatalf("ambiguous trigger events paired as %v, want no pairing", got)
	}

	cycling := game.LandcyclingActivatedAbility(nil, game.SearchSpec{})
	tagged := game.CardFace{ActivatedAbilities: []game.ActivatedAbility{cycling, {}}}
	got = printedAbilityParagraphs(&tagged, `
		{1}: You gain 1 life.
		Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.)
	`)
	if len(got) != 0 {
		t.Fatalf("a cycling ability paired with another ability's paragraph: %v", got)
	}
}

// A copied keyword trigger shows its keyword line (CR 113.2c, 707.2): Clone
// copying Knight of Sorrows has afterlife, and when the copy dies its trigger
// shows exactly "Afterlife 1 (...)", not the blocking ability's text.
func TestObservationCopiedKeywordTriggerShowsItsKeywordLine(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	knight := registry.Lookup("Knight of Sorrows")
	clone := registry.Lookup("Clone")
	if knight == nil || clone == nil {
		t.Fatal("Knight of Sorrows or Clone is not in the default registry")
	}
	if knight.TriggeredAbilities[0].Text != "" {
		t.Fatal("Knight of Sorrows's afterlife has its own text; this scenario needs an untexted keyword trigger")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, knight)
	addImplementationSpellToStack(g, game.Player1, clone, nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	if copied.Controller != game.Player1 || permanentEffectiveName(g, copied) != knight.Name {
		t.Fatalf("Clone entered as %q, want a copy of %q", permanentEffectiveName(g, copied), knight.Name)
	}
	if knight.TriggeredAbilities[0].Text != "" {
		t.Fatal("copying wrote text into the shared card definition")
	}
	g.TriggerEventCursor = len(g.Events)
	if !movePermanentToZone(g, copied, zone.Graveyard) {
		t.Fatal("could not move the copy to the graveyard")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the copied afterlife did not trigger")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		stack := observe(g, viewer).Stack()
		if len(stack) != 1 || stack[0].Text != knightOfSorrowsAfterlifeText {
			t.Fatalf("player %d stack = %+v, want the copied afterlife with text %q", viewer, stack, knightOfSorrowsAfterlifeText)
		}
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	spirits := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == game.Player1 && permanent.Token {
			spirits++
		}
	}
	if spirits != 1 {
		t.Fatalf("the copied afterlife created %d tokens, want one Spirit", spirits)
	}
}

// One paragraph naming two trigger events is the text of both triggered
// abilities the engine models it as: Clone copying Bilbo shows Bilbo's "enters
// or leaves" paragraph when it enters and again when it leaves, never the
// combat damage paragraph.
func TestObservationCopiedSplitTriggerShowsItsParagraph(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	bilbo := registry.Lookup("Bilbo, Retired Burglar")
	clone := registry.Lookup("Clone")
	if bilbo == nil || clone == nil {
		t.Fatal("Bilbo or Clone is not in the default registry")
	}
	for i := range bilbo.TriggeredAbilities {
		if bilbo.TriggeredAbilities[i].Text != "" {
			t.Fatalf("Bilbo trigger %d has its own text; this scenario needs untexted triggers", i)
		}
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, bilbo)
	g.TriggerEventCursor = len(g.Events)
	addImplementationSpellToStack(g, game.Player1, clone, nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	if copied.Controller != game.Player1 || permanentEffectiveName(g, copied) != bilbo.Name {
		t.Fatalf("Clone entered as %q, want a copy of %q", permanentEffectiveName(g, copied), bilbo.Name)
	}
	wantStack := func(when string) {
		t.Helper()
		if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatalf("the copied Bilbo did not trigger when it %s", when)
		}
		for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
			stack := observe(g, viewer).Stack()
			if len(stack) != 1 || stack[0].Text != bilboEntersOrLeavesText {
				t.Fatalf("player %d stack after the copy %s = %+v, want text %q", viewer, when, stack, bilboEntersOrLeavesText)
			}
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Stack.IsEmpty() {
			t.Fatalf("the copied Bilbo trigger did not resolve when it %s", when)
		}
	}
	wantStack("entered")
	g.TriggerEventCursor = len(g.Events)
	if !movePermanentToZone(g, copied, zone.Graveyard) {
		t.Fatal("could not move the copy to the graveyard")
	}
	wantStack("left")
	if bilbo.TriggeredAbilities[0].Text != "" || bilbo.TriggeredAbilities[1].Text != "" {
		t.Fatal("copying wrote text into the shared card definition")
	}
}

// A typecycling ability activated from hand is labeled with its own keyword
// line (CR 702.29e), not the card's other paragraphs.
func TestTypecyclingFromHandShowsItsKeywordLine(t *testing.T) {
	dragon := cards.NewDefaultRegistry().Lookup("Timeless Dragon")
	if dragon == nil {
		t.Fatal("Timeless Dragon is not in the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, dragon)
	activate := action.ActivateAbilityAction{SourceID: cardID, AbilityIndex: dragon.ActivatedAbilityIndex(0)}
	if text := activatedAbilityText(g, game.Player1, activate); text != timelessDragonCyclingText {
		t.Fatalf("plainscycling label = %q, want %q", text, timelessDragonCyclingText)
	}
}

// A printed ability activated from hand or graveyard keeps its own identified
// Oracle text on the stack for every player, on the cycling, other hand and
// graveyard activation routes alike: its stack object is that printed ability,
// not an unrecognizable copy of it.
func TestPrintedHandAndGraveyardActivationsShowOwnTextOnStack(t *testing.T) {
	t.Run("typecycling from hand", func(t *testing.T) {
		dragon := cards.NewDefaultRegistry().Lookup("Timeless Dragon")
		if dragon == nil {
			t.Fatal("Timeless Dragon is not in the default registry")
		}
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToHand(g, game.Player1, dragon)
		addBasicLandPermanent(g, game.Player1, types.Plains)
		addBasicLandPermanent(g, game.Player1, types.Plains)
		g.Turn.PriorityPlayer = game.Player1
		activateAndRequireStackText(t, g, action.ActivateAbility(cardID, dragon.ActivatedAbilityIndex(0), nil, 0), timelessDragonCyclingText)
	})
	t.Run("ability word activation from hand", func(t *testing.T) {
		const channelText = "Channel — {1}, Discard this card: You gain 2 life."
		gain := func(amount int) game.AbilityContent {
			return game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(amount), Player: game.ControllerReference()},
			}}}.Ability()
		}
		def := &game.CardDef{CardFace: game.CardFace{
			Name:  "Channel Test Creature",
			Types: []types.Card{types.Creature},
			ActivatedAbilities: []game.ActivatedAbility{{
				ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
				ZoneOfFunction: zone.Hand,
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalDiscard, Text: "Discard this card", Amount: 1, Source: zone.Hand,
				}},
				Content: gain(2),
			}},
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}},
				Content: gain(1),
			}},
			OracleText: "When this creature enters, you gain 1 life.\n" + channelText,
		}}
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToHand(g, game.Player1, def)
		addBasicLandPermanent(g, game.Player1, types.Plains)
		g.Turn.PriorityPlayer = game.Player1
		activateAndRequireStackText(t, g, action.ActivateAbility(cardID, def.ActivatedAbilityIndex(0), nil, 0), channelText)
	})
	t.Run("keyword activation from graveyard", func(t *testing.T) {
		const eternalizeText = "Eternalize {2}{W}{W} ({2}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Dragon with no mana cost. Eternalize only as a sorcery.)"
		dragon := cards.NewDefaultRegistry().Lookup("Timeless Dragon")
		if dragon == nil {
			t.Fatal("Timeless Dragon is not in the default registry")
		}
		untexted := *dragon
		untexted.ActivatedAbilities = slices.Clone(dragon.ActivatedAbilities)
		untexted.ActivatedAbilities[1].Text = ""
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToGraveyard(g, game.Player1, &untexted)
		for range 4 {
			addBasicLandPermanent(g, game.Player1, types.Plains)
		}
		setMainPhasePriority(g, game.Player1)
		activateAndRequireStackText(t, g, action.ActivateAbility(cardID, untexted.ActivatedAbilityIndex(1), nil, 0), eternalizeText)
	})
}

// An ability granted to a card in hand is not in that card's Oracle text, so an
// untexted granted cycling ability on the stack has no text, while the card's
// own printed cycling ability keeps its keyword line.
func TestGrantedHandActivationIsNotShownWithPrintedText(t *testing.T) {
	const printedText = "Cycling {1} ({1}, Discard this card: Draw a card.)"
	setup := func(t *testing.T) (*game.Game, *game.CardDef, game.ObjectID) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		granter := addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Creature}}, cost.Mana{cost.R})
		granterCard, ok := g.GetCardInstance(granter.CardInstanceID)
		if !ok {
			t.Fatal("missing cycling granter card")
		}
		granterCard.Def.StaticAbilities[0].RuleEffects[0].GrantedAbility.Text = ""
		def := typedCyclingCard(types.Creature, cost.Mana{cost.O(1)})
		def.ActivatedAbilities[0].Text = ""
		def.OracleText = printedText
		cardID := addCardToHand(g, game.Player1, def)
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		g.Turn.PriorityPlayer = game.Player1
		return g, def, cardID
	}
	t.Run("printed", func(t *testing.T) {
		g, def, cardID := setup(t)
		activateAndRequireStackText(t, g, action.ActivateAbility(cardID, def.ActivatedAbilityIndex(0), nil, 0), printedText)
	})
	t.Run("granted", func(t *testing.T) {
		g, def, cardID := setup(t)
		activateAndRequireStackText(t, g, action.ActivateAbility(cardID, def.AbilityCount(), nil, 0), "")
	})
}

func activateAndRequireStackText(t *testing.T, g *game.Game, activate action.Action, want string) {
	t.Helper()
	if !NewEngine(nil).applyAction(g, game.Player1, activate) {
		t.Fatalf("could not activate %+v", activate)
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		stack := observe(g, viewer).Stack()
		if len(stack) != 1 || stack[0].Text != want {
			t.Errorf("player %d stack = %+v, want one ability with text %q", viewer, stack, want)
		}
	}
}
