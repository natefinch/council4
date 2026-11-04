package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: Oracle text from the run's pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 for
// Kiki-Jiki, Mirror Breaker (a34b7416-cfe3-4a1e-a8c1-a3056b747519): "Haste" /
// "{T}: Create a token that's a copy of target nonlegendary creature you
// control, except it has haste. Sacrifice it at the beginning of the next end
// step."; Goryo's Vengeance (3ea3c681-f0ad-45a2-bc25-2813efd8d560): "Return
// target legendary creature card from your graveyard to the battlefield. That
// creature gains haste. Exile it at the beginning of the next end step." /
// "Splice onto Arcane {2}{B} (...)"; Splinter Twin
// (2df3187b-4b91-4de9-9b26-d48a30700cec) and Blessed Wine
// (8079007c-a07a-4d71-bb08-e1dd215297de) are covered in
// observation_triggers_test.go. Scryfall rulings read for all four (Blessed
// Wine has none); none address the delayed trigger's text: Kiki-Jiki 2021-03-19
// "If Kiki-Jiki's ability creates multiple tokens due to a replacement effect
// ..., you'll sacrifice each of them."; Goryo's Vengeance 2018-12-07 "If the
// returned creature leaves the battlefield before the end step, it will remain
// in its current zone."; Splinter Twin 2010-06-15 "If the ability is activated
// during a turn's end step, the token will be exiled at the beginning of the
// following turn's end step." Comprehensive Rules effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 603.3 (a triggered ability on the stack "has the text of the ability that
// created it"), 603.7 (a delayed triggered ability "will contain 'when,'
// 'whenever,' or 'at,' although that word won't usually begin the ability"),
// 603.7a (created during resolution of spells or abilities), 603.7d/603.7e (its
// source is the creating spell or the creating ability's source) and 707.10 (a
// copy of a spell or ability copies its text). So a fixed-timing delayed
// trigger's own text is the sentence that states its timing, not the creating
// spell's or ability's other sentences, and the creating ability's own timing
// condition is not the delayed trigger's.

const (
	kikiJikiDelayedText = "Sacrifice it at the beginning of the next end step."
	goryoDelayedText    = "Exile it at the beginning of the next end step."
)

func requirePendingTexts(t *testing.T, g *game.Game, viewer game.PlayerID, source string, want ...string) {
	t.Helper()
	pending := observe(g, viewer).DelayedTriggers()
	var got []string
	for _, trigger := range pending {
		if trigger.SourceName != source {
			t.Fatalf("Player%d pending trigger source = %q, want %q", viewer, trigger.SourceName, source)
		}
		got = append(got, trigger.Text)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Player%d pending texts = %q, want %q", viewer, got, want)
	}
}

func requireStackTexts(t *testing.T, g *game.Game, viewer game.PlayerID, source string, want ...string) {
	t.Helper()
	var got []string
	for _, object := range observe(g, viewer).Stack() {
		if object.Name != source {
			t.Fatalf("Player%d stack object source = %q, want %q", viewer, object.Name, source)
		}
		got = append(got, object.Text)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Player%d stack texts = %q, want %q", viewer, got, want)
	}
}

// Kiki-Jiki's printed activated ability and a copy of it on the stack (CR
// 707.10) each schedule a delayed sacrifice that carries only "Sacrifice it at
// the beginning of the next end step.", pending and on the stack, with
// Kiki-Jiki as its source. A cloned game keeps that text when the triggers fire.
func TestFixedDelayedTriggerOwnTextFromActivatedAbilityAndCopy(t *testing.T) {
	g, engine := grantedHasteGame()
	kiki := addCombatPermanent(g, game.Player1, registryCard(t, "Kiki-Jiki, Mirror Breaker"))
	bear := addCombatPermanent(g, game.Player1, copyTestBear("Kiki Bear"))
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(kiki.ObjectID, 0, []game.Target{game.PermanentTarget(bear.ObjectID)}, 0)) {
		t.Fatal("could not activate Kiki-Jiki")
	}
	original, _ := g.Stack.Peek()
	if copied := copyStackObject(engine, g, original); copied == nil || copied.ID == original.ID {
		t.Fatal("copying Kiki-Jiki's ability did not put a copy on the stack")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if tokens := playerTokensNamed(g, game.Player1, "Kiki Bear"); len(tokens) != 2 {
		t.Fatalf("Kiki-Jiki and its copy created %d tokens, want 2", len(tokens))
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		requirePendingTexts(t, g, viewer, "Kiki-Jiki, Mirror Breaker", kikiJikiDelayedText, kikiJikiDelayedText)
	}

	clone := g.Clone()
	fireEndStep(t, clone, engine)
	requireStackTexts(t, clone, game.Player2, "Kiki-Jiki, Mirror Breaker", kikiJikiDelayedText, kikiJikiDelayedText)
	resolveStackWithTriggers(engine, clone, [game.NumPlayers]PlayerAgent{})
	if tokens := playerTokensNamed(clone, game.Player1, "Kiki Bear"); len(tokens) != 0 {
		t.Fatalf("cloned game kept %d Kiki-Jiki tokens after the end step", len(tokens))
	}
	requirePendingTexts(t, g, game.Player2, "Kiki-Jiki, Mirror Breaker", kikiJikiDelayedText, kikiJikiDelayedText)
}

// Goryo's Vengeance and a copy of it with a new target (CR 707.10) each
// schedule a delayed exile carrying only "Exile it at the beginning of the next
// end step.", not the spell's return and haste sentences or its splice line,
// with the spell as its source (CR 603.7d), and each exiles its own creature.
func TestFixedDelayedTriggerOwnTextFromSpellAndCopy(t *testing.T) {
	g, engine := grantedHasteGame()
	legend := func(name string) *game.CardDef {
		def := copyTestBear(name)
		def.Supertypes = []types.Super{types.Legendary}
		return def
	}
	first := addCardToGraveyard(g, game.Player1, legend("First Legend"))
	second := addCardToGraveyard(g, game.Player1, legend("Second Legend"))
	goryo := addCardToHand(g, game.Player1, registryCard(t, "Goryo's Vengeance"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(goryo, []game.Target{{Kind: game.TargetCard, CardID: first}}, 0, nil)) {
		t.Fatal("could not cast Goryo's Vengeance")
	}
	original, _ := g.Stack.Peek()
	retarget := game.Target{Kind: game.TargetCard, CardID: second, CardZoneVersionSet: true}
	if copied := copyStackObject(engine, g, original, retarget); copied == nil || copied.ID == original.ID || !slices.Equal(copied.Targets, []game.Target{retarget}) {
		t.Fatal("copying Goryo's Vengeance did not put a copy on the stack")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	for _, viewer := range []game.PlayerID{game.Player1, game.Player3} {
		requirePendingTexts(t, g, viewer, "Goryo's Vengeance", goryoDelayedText, goryoDelayedText)
	}
	if g.Players[game.Player1].Graveyard.Contains(first) || g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("Goryo's Vengeance and its copy did not return both legends")
	}
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "Goryo's Vengeance", goryoDelayedText, goryoDelayedText)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	for _, cardID := range []id.ID{first, second} {
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("returned legend %d was not exiled at the end step", cardID)
		}
	}
}

// A printed triggered ability's own leading timing ("At the beginning of your
// upkeep, ...") is not the delayed trigger's, so the upkeep delayed trigger it
// creates carries only its later sentence, pending and once it fires.
func TestFixedDelayedTriggerOwnTextSkipsCreatingTriggerCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Step:       game.StepUpkeep,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{
		{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
			}}}.Ability(),
		}}},
	}, nil)
	card, _ := g.GetCardInstance(source.CardInstanceID)
	const own = "You gain 2 life at the beginning of the next turn's upkeep."
	card.Def.OracleText = "At the beginning of your upkeep, you gain 1 life. " + own
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseBeginning, game.StepUpkeep
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player1, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("upkeep trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	view := findPermanentView(t, observe(g, game.Player2), source.ObjectID)
	if len(view.PendingDelayedTriggers) != 1 || view.PendingDelayedTriggers[0].Text != own {
		t.Fatalf("pending delayed triggers = %+v, want text %q", view.PendingDelayedTriggers, own)
	}

	g.Turn.TurnNumber++
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player1, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the next upkeep did not fire the delayed trigger")
	}
	var texts []string
	for _, object := range observe(g, game.Player2).Stack() {
		texts = append(texts, object.Text)
	}
	if !slices.Contains(texts, own) {
		t.Fatalf("stack texts at the next upkeep = %q, want the delayed trigger's %q", texts, own)
	}
}

// Source review: Oracle text from the run's pinned corpus for Giant
// Caterpillar (bdc5eb8a-4531-4408-90fe-9b352d71a052): "{G}, Sacrifice this
// creature: Create a 1/1 green Insect creature token with flying named
// Butterfly at the beginning of the next end step."; Pact of Negation
// (1ed4c0bb-b710-44a1-b8bc-6bd11c27b8b8): "Counter target spell." / "At the
// beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the
// game."; Crazed Armodon (b83e4b36-57c1-493d-ab79-52075990b2d5): "{G}: This
// creature gets +3/+0 and gains trample until end of turn. Destroy this
// creature at the beginning of the next end step. Activate only once each
// turn."; Whip of Erebos (b738612f-4e78-4b19-9392-6c074956d6e4): "{2}{B}{B},
// {T}: Return target creature card from your graveyard to the battlefield. It
// gains haste. Exile it at the beginning of the next end step. If it would
// leave the battlefield, exile it instead of putting it anywhere else. Activate
// only as a sorcery." Scryfall rulings: Giant Caterpillar 2006-02-01 "Now it
// makes Insect tokens named Butterfly."; Pact of Negation 2021-03-19 "If Pact of
// Negation resolves with a legal target but fails to counter that spell ...,
// the delayed triggered ability will still trigger at the beginning of your
// next upkeep and you will have to pay."; Crazed Armodon has none; Whip of
// Erebos 2013-09-15 "At the beginning of the next end step, the creature
// returned ... is exiled. This is a delayed triggered ability. If the ability
// is countered, ... the replacement effect will still exile the creature when
// it eventually leaves the battlefield." CR 602.1a (an activation cost precedes
// the colon), 602.5b (activation restrictions are part of the activated
// ability), 603.7a and 608.2c (a later instruction may depend on an earlier
// one: "If you don't, you lose the game." is the delayed ability's). The
// creator's cost and restrictions are not the delayed ability's text; Pact's
// payment consequence is. A later sentence that cannot be attributed to either
// ability by its wording (Whip's replacement sentence) is recorded by the
// definition's DelayedTriggerDef.Text instead.

// Giant Caterpillar's delayed token creation is embedded after its activation
// cost, which belongs to the creating ability, so the delayed trigger carries
// only its effect, pending and on the stack, and creates the Butterfly.
func TestFixedDelayedTriggerOwnTextWithoutActivationCost(t *testing.T) {
	g, engine := grantedHasteGame()
	caterpillar := addCombatPermanent(g, game.Player1, registryCard(t, "Giant Caterpillar"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(caterpillar.ObjectID, 0, nil, 0)) {
		t.Fatal("could not activate Giant Caterpillar")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	const want = "Create a 1/1 green Insect creature token with flying named Butterfly at the beginning of the next end step."
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		requirePendingTexts(t, g, viewer, "Giant Caterpillar", want)
	}
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "Giant Caterpillar", want)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if tokens := playerTokensNamed(g, game.Player1, "Butterfly"); len(tokens) != 1 {
		t.Fatalf("the delayed trigger created %d Butterfly tokens, want 1", len(tokens))
	}
}

// Pact of Negation's delayed trigger keeps the sentence that depends on its
// payment ("If you don't, you lose the game."), pending and once it fires at
// the next upkeep, without the spell's counter line.
func TestFixedDelayedTriggerOwnTextKeepsDependentSentence(t *testing.T) {
	g, engine := grantedHasteGame()
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	pact := addCardToHand(g, game.Player1, registryCard(t, "Pact of Negation"))
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(pact, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil)) {
		t.Fatal("could not cast Pact of Negation")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("Pact of Negation did not counter its target")
	}
	const want = "At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game."
	for _, viewer := range []game.PlayerID{game.Player1, game.Player3} {
		requirePendingTexts(t, g, viewer, "Pact of Negation", want)
	}
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseBeginning, game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the next upkeep did not fire Pact of Negation's delayed trigger")
	}
	requireStackTexts(t, g, game.Player2, "Pact of Negation", want)
}

// Crazed Armodon's activation restriction belongs to its activated ability, not
// to the delayed destruction it creates, which carries only its own sentence
// and destroys the Armodon at the end step.
func TestFixedDelayedTriggerOwnTextOmitsActivationRestriction(t *testing.T) {
	g, engine := grantedHasteGame()
	armodon := addCombatPermanent(g, game.Player1, registryCard(t, "Crazed Armodon"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(armodon.ObjectID, 0, nil, 0)) {
		t.Fatal("could not activate Crazed Armodon")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	const want = "Destroy this creature at the beginning of the next end step."
	requirePendingTexts(t, g, game.Player2, "Crazed Armodon", want)
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "Crazed Armodon", want)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, armodon.ObjectID); ok {
		t.Fatal("Crazed Armodon survived its delayed destruction")
	}
}

// Source review (#3666): Oracle text from the run's pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 for Whip of
// Erebos (53987a39-c18c-4c13-b1ea-fd1b2a369f9e), quoted above; Moira and
// Teshar (445ad01c-c899-424d-8666-ee7bb2a6e535): "Whenever you cast a historic
// spell, return target nonland permanent card from your graveyard to the
// battlefield. It gains haste. Exile it at the beginning of the next end step.
// If it would leave the battlefield, exile it instead of putting it anywhere
// else. (...)"; Calamity, Galloping Inferno
// (eb37496e-374f-41b9-bd52-f52f37bdac31): "... create a tapped and attacking
// token that's a copy of it. Sacrifice that token at the beginning of the next
// end step. Repeat this process once."; Arcane Denial
// (ab1cc360-b9de-48d9-9983-4dfe4a7d2a37): "Counter target spell. Its
// controller may draw up to two cards at the beginning of the next turn's
// upkeep." / "You draw a card at the beginning of the next turn's upkeep.";
// Inkshield (da93264e-4e04-401a-9809-e4e1056cf604): "Prevent all combat damage
// that would be dealt to you this turn. For each 1 damage prevented this way,
// create a 2/1 white and black Inkling creature token with flying."; Simic
// Basilisk (5e0f0fd8-fdeb-44d9-b5e7-e3948f2dd10b): "{1}{G}: Until end of turn,
// target creature with a +1/+1 counter on it gains "Whenever this creature
// deals combat damage to a creature, destroy that creature at end of
// combat.""; Mangara's Blessing (5d9431de-7cf0-43d4-9b86-e0ef4a7756f9): "When a
// spell or ability an opponent controls causes you to discard this card, you
// gain 2 life, and you return this card from your graveyard to your hand at the
// beginning of the next end step."; Frenetic Sliver
// (b1af4207-e546-4bde-950b-3234dfffe278): All Slivers have "{0}: If this
// permanent is on the battlefield, flip a coin. If you win the flip, exile this
// permanent and return it to the battlefield under its owner's control at the
// beginning of the next end step. If you lose the flip, sacrifice it.";
// Arcum's Whistle (de691eed-d90b-48f9-bae4-f982767f2bfd): "... If they don't
// pay, the creature attacks this turn if able, and at the beginning of the next
// end step, destroy it if it didn't attack this turn. Activate only before
// attackers are declared."; Pygmy Hippo (0d39513a-1f05-4d9a-a772-745f4f1a3471):
// "... If you do, this creature assigns no combat damage this turn and at the
// beginning of your next main phase this turn, you add an amount of {C} equal
// to the amount of mana that player lost this way."; Kaya, Ghost Assassin
// (5e4c81b9-a31e-45e6-983f-30f11b8a05c0): "0: Exile Kaya or up to one target
// creature. Return that card to the battlefield under its owner's control at
// the beginning of your next upkeep. You lose 2 life."; Obzedat, Ghost Council
// (6c3df703-e96a-4ee4-b449-b8e1e3460bd3): "... If you do, return it to the
// battlefield under its owner's control at the beginning of your next upkeep.
// It gains haste."; Elemental Appeal (3bcfc82c-3523-40e3-b236-78ece8a8d69f):
// "... Exile it at the beginning of the next end step. If this spell was
// kicked, that creature gets +7/+0 until end of turn." Scryfall rulings: Whip
// of Erebos 2013-09-15 "neither of the 'exile' abilities is granted to that
// creature" and the replacement effect still applies if the delayed trigger is
// countered (so it is the creator's, not the delayed trigger's); Mangara's
// Blessing 2004-10-04 "The 2 life ... is gained when the triggered ability
// resolves, not at end of turn. Only the return to hand is done at end of
// turn."; Arcane Denial 2007-09-16 "The controller of the countered spell
// doesn't choose how many cards to draw until the relevant ability resolves.";
// Calamity's rulings, Inkshield's 2021-04-16 ruling and Arcum's Whistle's and
// Pygmy Hippo's rulings do not address the delayed text; Moira and Teshar and
// Simic Basilisk have none. CR 603.3, 603.7, 603.7a, 608.2c (later
// instructions in one sentence or later sentences may each be the creator's or
// the delayed ability's) and 113.2c/603.7a (two delayed abilities created by
// one spell). Wording that ends a coordinated creator instruction ("you gain 2
// life, and you return this card ...") cannot be separated from the creator's
// by syntax, since "and" may equally join two objects of one instruction;
// wording whose timing opens its own coordinated clause ("..., and at the
// beginning of the next end step, destroy it ...") can. Where wording cannot be
// attributed, the card definition records the delayed ability's text.

const whipDelayedText = "Exile it at the beginning of the next end step."

// Whip of Erebos's replacement sentence after its delayed exile is the
// creating ability's, not the delayed trigger's, which the definition records:
// the delayed exile carries only its own sentence, pending and on the stack,
// for every viewer and in a cloned game, and exiles the returned creature.
func TestFixedDelayedTriggerRecordedTextForUnattributableSentence(t *testing.T) {
	g, engine := grantedHasteGame()
	whip := addCombatPermanent(g, game.Player1, registryCard(t, "Whip of Erebos"))
	bear := addCardToGraveyard(g, game.Player1, copyTestBear("Whipped Bear"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(whip.ObjectID, 0, []game.Target{{Kind: game.TargetCard, CardID: bear}}, 0)) {
		t.Fatal("could not activate Whip of Erebos")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		requirePendingTexts(t, g, viewer, "Whip of Erebos", whipDelayedText)
	}
	clone := g.Clone()
	fireEndStep(t, clone, engine)
	requireStackTexts(t, clone, game.Player2, "Whip of Erebos", whipDelayedText)
	resolveStackWithTriggers(engine, clone, [game.NumPlayers]PlayerAgent{})
	if !clone.Players[game.Player1].Exile.Contains(bear) {
		t.Fatal("Whip of Erebos's returned creature was not exiled at the end step")
	}
	requirePendingTexts(t, g, game.Player2, "Whip of Erebos", whipDelayedText)
}

// Moira and Teshar's triggered ability shares Whip's shape: its delayed exile
// carries only its recorded sentence, pending and on the stack, and exiles the
// returned permanent.
func TestFixedDelayedTriggerRecordedTextFromTriggeredAbility(t *testing.T) {
	g, engine := grantedHasteGame()
	addCombatPermanent(g, game.Player1, registryCard(t, "Moira and Teshar"))
	bear := addCardToGraveyard(g, game.Player1, copyTestBear("Historic Return"))
	trinket := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Free Trinket", Types: []types.Card{types.Artifact}}})
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(trinket, nil, 0, nil)) {
		t.Fatal("could not cast the historic spell")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Fatal("Moira and Teshar did not return the permanent card")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player3} {
		requirePendingTexts(t, g, viewer, "Moira and Teshar", whipDelayedText)
	}
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "Moira and Teshar", whipDelayedText)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Exile.Contains(bear) {
		t.Fatal("Moira and Teshar's returned permanent was not exiled at the end step")
	}
}

// Arcane Denial creates two delayed triggers with the same upkeep timing, so
// neither sentence can be singled out by its timing; each carries its recorded
// sentence, pending and on the stack, and each draws for its own player.
func TestFixedDelayedTriggersWithSameTimingCarryRecordedTexts(t *testing.T) {
	const (
		theirDraw = "Its controller may draw up to two cards at the beginning of the next turn's upkeep."
		yourDraw  = "You draw a card at the beginning of the next turn's upkeep."
	)
	g, engine := grantedHasteGame()
	for range 3 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Denial Draw"}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Denial Draw"}})
	}
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	denial := addCardToHand(g, game.Player1, registryCard(t, "Arcane Denial"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(denial, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil)) {
		t.Fatal("could not cast Arcane Denial")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("Arcane Denial did not counter its target")
	}
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		requirePendingTexts(t, g, viewer, "Arcane Denial", theirDraw, yourDraw)
	}
	yourHand, theirHand := g.Players[game.Player1].Hand.Size(), g.Players[game.Player2].Hand.Size()
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase, g.Turn.Step = game.PhaseBeginning, game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the next upkeep did not fire Arcane Denial's delayed triggers")
	}
	var stacked []string
	for _, object := range observe(g, game.Player3).Stack() {
		stacked = append(stacked, object.Text)
	}
	slices.Sort(stacked)
	if want := []string{theirDraw, yourDraw}; !slices.Equal(stacked, want) {
		t.Fatalf("stack texts = %q, want %q", stacked, want)
	}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player2] = numberChoiceAgent{number: 2}
	resolveStackWithTriggers(engine, g, agents)
	if got := g.Players[game.Player1].Hand.Size() - yourHand; got != 1 {
		t.Fatalf("Arcane Denial's controller drew %d cards, want 1", got)
	}
	if got := g.Players[game.Player2].Hand.Size() - theirHand; got != 2 {
		t.Fatalf("the countered spell's controller drew %d cards, want 2", got)
	}
}

// A granted triggered ability records its own text, so the end-of-combat
// destruction it creates carries only its own wording ("destroy that creature
// at end of combat"), on the stack after combat damage and pending, and
// destroys the damaged creature at end of combat.
func TestFixedDelayedTriggerOwnTextFromGrantedTriggeredAbility(t *testing.T) {
	const (
		granted = "Whenever this creature deals combat damage to a creature, destroy that creature at end of combat."
		want    = "Destroy that creature at end of combat."
	)
	g, engine := grantedHasteGame()
	basilisk := addCombatPermanent(g, game.Player1, registryCard(t, "Simic Basilisk"))
	basilisk.Counters.Add(counter.PlusOnePlusOne, 1)
	victim := addSturdyCombatCreature(g, game.Player2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(basilisk.ObjectID, 0, []game.Target{game.PermanentTarget(basilisk.ObjectID)}, 0)) {
		t.Fatal("could not activate Simic Basilisk")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepCombatDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: basilisk.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: victim.ObjectID, Blocking: basilisk.ObjectID}},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the granted combat damage trigger did not trigger")
	}
	requireStackTexts(t, g, game.Player2, "Simic Basilisk", granted)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		requirePendingTexts(t, g, viewer, "Simic Basilisk", want)
	}
	fireEndOfCombat(t, g, engine)
	requireStackTexts(t, g, game.Player3, "Simic Basilisk", want)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("the damaged creature survived end of combat")
	}
}

// An ability on the stack has its own text, independent of its source (CR
// 113.7a, 405.4), and that text was public when it was put on the stack, so
// turning the source face down before the ability resolves hides only the
// source's identity (CR 708.2, 708.5): Moira and Teshar's delayed exile keeps
// its recorded sentence, pending and on the stack, with no source name, and
// still exiles the returned permanent.
func TestFixedDelayedTriggerRecordedTextPublicAfterSourceTurnedFaceDown(t *testing.T) {
	g, engine := grantedHasteGame()
	moira := addCombatPermanent(g, game.Player1, registryCard(t, "Moira and Teshar"))
	bear := addCardToGraveyard(g, game.Player1, copyTestBear("Historic Return"))
	trinket := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Free Trinket", Types: []types.Card{types.Artifact}}})
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(trinket, nil, 0, nil)) {
		t.Fatal("could not cast the historic spell")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("Moira and Teshar's historic trigger did not trigger")
	}
	if stacked := observe(g, game.Player2).Stack(); len(stacked) != 2 || stacked[1].Text == "" {
		t.Fatalf("Moira and Teshar's creating ability = %+v, want public text", stacked)
	}
	applyCyberConversion(g, moira)
	if !moira.FaceDown {
		t.Fatal("Moira and Teshar was not turned face down")
	}
	if stacked := observe(g, game.Player2).Stack(); len(stacked) != 2 || stacked[1].Name != "" {
		t.Fatalf("face-down source's creating ability = %+v, want no source name", stacked)
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Fatal("Moira and Teshar did not return the permanent card")
	}
	for _, viewer := range []game.PlayerID{game.Player2, game.Player3} {
		requirePendingTexts(t, g, viewer, "", whipDelayedText)
	}
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "", whipDelayedText)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Exile.Contains(bear) {
		t.Fatal("the returned permanent was not exiled at the end step")
	}
}

// The same holds for wording singled out of the creating ability's printed
// text rather than recorded: Kiki-Jiki is turned face down after its ability is
// activated, and the delayed sacrifice still carries "Sacrifice it at the
// beginning of the next end step." for every player, pending and on the stack,
// without Kiki-Jiki's name, and sacrifices the token.
func TestFixedDelayedTriggerOwnTextPublicAfterSourceTurnedFaceDown(t *testing.T) {
	g, engine := grantedHasteGame()
	kiki := addCombatPermanent(g, game.Player1, registryCard(t, "Kiki-Jiki, Mirror Breaker"))
	bear := addCombatPermanent(g, game.Player1, copyTestBear("Kiki Bear"))
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(kiki.ObjectID, 0, []game.Target{game.PermanentTarget(bear.ObjectID)}, 0)) {
		t.Fatal("could not activate Kiki-Jiki")
	}
	applyCyberConversion(g, kiki)
	if !kiki.FaceDown {
		t.Fatal("Kiki-Jiki was not turned face down")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if tokens := playerTokensNamed(g, game.Player1, "Kiki Bear"); len(tokens) != 1 {
		t.Fatalf("Kiki-Jiki's ability created %d tokens, want 1", len(tokens))
	}
	for _, viewer := range []game.PlayerID{game.Player2, game.Player3} {
		requirePendingTexts(t, g, viewer, "", kikiJikiDelayedText)
	}
	fireEndStep(t, g, engine)
	requireStackTexts(t, g, game.Player2, "", kikiJikiDelayedText)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if tokens := playerTokensNamed(g, game.Player1, "Kiki Bear"); len(tokens) != 0 {
		t.Fatalf("Kiki-Jiki's token survived the end step: %d left", len(tokens))
	}
}

// An ability granted to a face-down creature is not part of the hidden card,
// so its public text, and the delayed wording singled out of it, stay visible:
// Simic Basilisk's granted ability on a face-down creature schedules "Destroy
// that creature at end of combat." for every player.
func TestFixedDelayedTriggerTextFromAbilityGrantedToFaceDownCreature(t *testing.T) {
	const want = "Destroy that creature at end of combat."
	g, engine := grantedHasteGame()
	basilisk := addCombatPermanent(g, game.Player1, registryCard(t, "Simic Basilisk"))
	hidden := addCombatPermanent(g, game.Player1, copyTestBear("Hidden Attacker"))
	hidden.Counters.Add(counter.PlusOnePlusOne, 1)
	applyCyberConversion(g, hidden)
	if !hidden.FaceDown {
		t.Fatal("the attacker was not turned face down")
	}
	victim := addSturdyCombatCreature(g, game.Player2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(basilisk.ObjectID, 0, []game.Target{game.PermanentTarget(hidden.ObjectID)}, 0)) {
		t.Fatal("could not activate Simic Basilisk")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepCombatDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: hidden.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: victim.ObjectID, Blocking: hidden.ObjectID}},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the granted combat damage trigger did not trigger")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	for _, viewer := range []game.PlayerID{game.Player2, game.Player3} {
		requirePendingTexts(t, g, viewer, "", want)
	}
	fireEndOfCombat(t, g, engine)
	requireStackTexts(t, g, game.Player2, "", want)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("the damaged creature survived end of combat")
	}
}

// delayedTriggerOwnWording prefers the definition's recorded text, and
// otherwise reports whether the creating text's wording can be singled out.
func TestDelayedTriggerOwnWordingPrefersRecordedText(t *testing.T) {
	whip := registryCard(t, "Whip of Erebos").ActivatedAbilities[0].Text
	def := &game.DelayedTriggerDef{Timing: game.DelayedAtBeginningOfNextEndStep}
	if got, ok := delayedTriggerOwnWording(whip, def, false); ok {
		t.Fatalf("unrecorded Whip wording = %q, true; want not attributable", got)
	}
	def.Text = whipDelayedText
	if got, ok := delayedTriggerOwnWording(whip, def, false); !ok || got != whipDelayedText {
		t.Fatalf("recorded Whip wording = %q, %v; want %q, true", got, ok, whipDelayedText)
	}
}

// Every registered card's fixed-timing delayed trigger has its own text: the
// definition records it or the creating spell's or ability's wording singles it
// out, so no pending or triggered delayed ability falls back to the whole
// creating text. A granted ability's delayed trigger is created from that
// granted ability's own text.
func TestRegisteredFixedDelayedTriggersHaveOwnText(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, name := range reg.All() {
		def := reg.Lookup(name)
		faces := []*game.CardFace{&def.CardFace}
		if def.Back.Exists {
			faces = append(faces, &def.Back.Val)
		}
		if def.Alternate.Exists {
			faces = append(faces, &def.Alternate.Val)
		}
		for _, face := range faces {
			for i := range face.AbilityCount() {
				body := face.BodyAt(i)
				_, spell := body.(*game.AbilityContent)
				text := game.BodyText(body)
				switch {
				case spell:
					text = oracleRulesText(face.OracleText)
				case text == "":
					text = printedAbilityParagraph(face, face.OracleText, i)
				}
				visitDelayedTriggers(reflect.ValueOf(body), text, spell, true, map[uintptr]bool{}, func(trigger game.DelayedTriggerDef, text string, spell bool) {
					if trigger.EventPattern.Exists {
						return
					}
					if _, ok := delayedTriggerOwnWording(text, &trigger, spell); !ok {
						t.Errorf("%s: fixed-timing delayed trigger has no own text in %q; record DelayedTriggerDef.Text", name, text)
					}
				})
			}
		}
	}
}

// visitDelayedTriggers calls visit for each delayed trigger definition reachable
// from v with the text of the spell or ability that creates it: an ability
// nested in v (a granted or token ability) creates its delayed triggers from
// its own text.
func visitDelayedTriggers(v reflect.Value, text string, spell, root bool, seen map[uintptr]bool, visit func(game.DelayedTriggerDef, string, bool)) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		visitDelayedTriggers(v.Elem(), text, spell, root, seen, visit)
		return
	case reflect.Interface:
		if !v.IsNil() {
			visitDelayedTriggers(v.Elem(), text, spell, root, seen, visit)
		}
		return
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			visitDelayedTriggers(v.Index(i), text, spell, false, seen, visit)
		}
		return
	case reflect.Map:
		for _, key := range v.MapKeys() {
			visitDelayedTriggers(v.MapIndex(key), text, spell, false, seen, visit)
		}
		return
	case reflect.Struct:
	default:
		return
	}
	switch x := v.Interface().(type) {
	case game.CreateDelayedTrigger:
		visit(x.Trigger, text, spell)
	case game.ActivatedAbility:
		if !root {
			text, spell = x.Text, false
		}
	case game.TriggeredAbility:
		if !root {
			text, spell = x.Text, false
		}
	}
	for i := range v.NumField() {
		if v.Type().Field(i).IsExported() {
			visitDelayedTriggers(v.Field(i), text, spell, false, seen, visit)
		}
	}
}

// delayedTriggerTimingText singles out the one sentence stating a fixed
// delayed timing and keeps wording that cannot be singled out whole.
func TestDelayedTriggerTimingTextSentences(t *testing.T) {
	for _, tc := range []struct {
		name   string
		text   string
		timing game.DelayedTriggerTiming
		spell  bool
		want   string
		ok     bool
	}{
		{
			name:   "later sentence of an activated ability",
			text:   "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Exile that token at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "spell sentence beginning with at",
			text:   "Exile X target creature cards from graveyards. For each creature card exiled this way, create a token that's a copy of it. At the beginning of your next end step, sacrifice those tokens.",
			timing: game.DelayedAtBeginningOfYourNextEndStep,
			spell:  true,
			want:   "At the beginning of your next end step, sacrifice those tokens.",
			ok:     true,
		},
		{
			name:   "spell paragraph's first sentence",
			text:   "You gain 1 life.\nDraw a card at the beginning of the next turn's upkeep.",
			timing: game.DelayedAtBeginningOfNextUpkeep,
			spell:  true,
			want:   "Draw a card at the beginning of the next turn's upkeep.",
			ok:     true,
		},
		{
			name:   "wording embedded after the creating trigger condition",
			text:   "At the beginning of your upkeep, you gain 1 life at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "You gain 1 life at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "wording embedded after the creating trigger condition and intervening if",
			text:   "Split — When this creature dies, if it had two or more +1/+1 counters on it, create a token that's a copy of it at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Create a token that's a copy of it at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "wording embedded after the activation cost",
			text:   "{G}, Sacrifice this creature: Create a 1/1 green Insect creature token with flying named Butterfly at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Create a 1/1 green Insect creature token with flying named Butterfly at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "wording embedded after the loyalty cost",
			text:   "−7: For each artifact you control, create a token that's a copy of it. Those tokens gain haste. Exile those tokens at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Exile those tokens at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "delayed wording that opens an activated ability's effect",
			text:   "{5}: At the beginning of your next upkeep, choose a basic land type.",
			timing: game.DelayedAtBeginningOfYourNextUpkeep,
			want:   "At the beginning of your next upkeep, choose a basic land type.",
			ok:     true,
		},
		{
			name:   "condition checked by the creator is not the delayed trigger's",
			text:   "At the beginning of your end step, you may exile this creature. If you do, return it to the battlefield under its owner's control at the beginning of your next upkeep.",
			timing: game.DelayedAtBeginningOfYourNextUpkeep,
			want:   "Return it to the battlefield under its owner's control at the beginning of your next upkeep.",
			ok:     true,
		},
		{
			name:   "creating trigger condition without a comma is skipped",
			text:   "At the beginning of your end step you may exile this creature. Return it at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Return it at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "delayed trigger's own intervening if is kept",
			text:   "Exile target card. At the beginning of the next end step, if you haven't cast it, return it to its owner's hand.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
			want:   "At the beginning of the next end step, if you haven't cast it, return it to its owner's hand.",
			ok:     true,
		},
		{
			name:   "sentences depending on the delayed instruction belong to it",
			text:   "Counter target spell.\nAt the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.",
			timing: game.DelayedAtBeginningOfNextUpkeep,
			spell:  true,
			want:   "At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.",
			ok:     true,
		},
		{
			name:   "chained and reflexive continuations belong to it",
			text:   "Exile target creature. At the beginning of the next end step, sacrifice that Vehicle. When you sacrifice a creature this way, it deals damage equal to its power to any target. Then you may draw a card. If you do, discard a card.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
			want:   "At the beginning of the next end step, sacrifice that Vehicle. When you sacrifice a creature this way, it deals damage equal to its power to any target. Then you may draw a card. If you do, discard a card.",
			ok:     true,
		},
		{
			name:   "activation restriction is the creator's",
			text:   "{G}: This creature gets +3/+0 and gains trample until end of turn. Destroy this creature at the beginning of the next end step. Activate only once each turn.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Destroy this creature at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "unattributable later sentence keeps the whole text",
			text:   "0: Exile up to one target creature. Return that card to the battlefield under its owner's control at the beginning of your next upkeep. You lose 2 life.",
			timing: game.DelayedAtBeginningOfYourNextUpkeep,
		},
		{
			name:   "unattributable later sentence after embedded wording keeps the whole text",
			text:   "At the beginning of your upkeep, you gain 1 life at the beginning of the next end step. You lose 1 life.",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			name:   "creator's replacement sentence is not attributable",
			text:   "{2}{B}{B}, {T}: Return target creature card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step. If it would leave the battlefield, exile it instead of putting it anywhere else. Activate only as a sorcery.",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			name:   "unattributable later sentence after a condition checked by the creator",
			text:   "At the beginning of your end step, you may exile Obzedat. If you do, return it to the battlefield under its owner's control at the beginning of your next upkeep. It gains haste.",
			timing: game.DelayedAtBeginningOfYourNextUpkeep,
		},
		{
			name:   "unattributable kicker sentence",
			text:   "Kicker {5} (You may pay an additional {5} as you cast this spell.)\nCreate a 7/1 red Elemental creature token with trample and haste. Exile it at the beginning of the next end step. If this spell was kicked, that creature gets +7/+0 until end of turn.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
		},
		{
			name:   "timing ending a coordinated creator instruction is not attributable",
			text:   "When a spell or ability an opponent controls causes you to discard this card, you gain 2 life, and you return this card from your graveyard to your hand at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			name:   "coordinated instructions after a condition are not attributable",
			text:   "{0}: If this permanent is on the battlefield, flip a coin. If you win the flip, exile this permanent and return it to the battlefield under its owner's control at the beginning of the next end step. If you lose the flip, sacrifice it.",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			name:   "then-coordinated creator instruction is not attributable",
			text:   "Exile target creature, then return it to the battlefield at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
		},
		{
			name:   "timing opening a coordinated clause after a comma",
			text:   "{3}, {T}: Choose target non-Wall creature the active player has controlled continuously since the beginning of the turn. That player may pay {X}, where X is that creature's mana value. If they don't pay, the creature attacks this turn if able, and at the beginning of the next end step, destroy it if it didn't attack this turn. Activate only before attackers are declared.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "At the beginning of the next end step, destroy it if it didn't attack this turn.",
			ok:     true,
		},
		{
			name:   "timing opening a coordinated clause",
			text:   "Whenever this creature attacks and isn't blocked, you may have defending player activate a mana ability of each land they control and lose all unspent mana. If you do, this creature assigns no combat damage this turn and at the beginning of your next main phase this turn, you add an amount of {C} equal to the amount of mana that player lost this way.",
			timing: game.DelayedAtBeginningOfNextMainPhase,
			want:   "At the beginning of your next main phase this turn, you add an amount of {C} equal to the amount of mana that player lost this way.",
			ok:     true,
		},
		{
			name:   "timing opening a then clause",
			text:   "Exile target creature, then at the beginning of the next end step, return it to the battlefield.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
			want:   "At the beginning of the next end step, return it to the battlefield.",
			ok:     true,
		},
		{
			name:   "coordination inside quoted wording does not block attribution",
			text:   `Exile target creature. Return it to the battlefield with "Whenever this creature attacks, draw a card and discard a card." at the beginning of the next end step.`,
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
			want:   `Return it to the battlefield with "Whenever this creature attacks, draw a card and discard a card." at the beginning of the next end step.`,
			ok:     true,
		},
		{
			name:   "modal line",
			text:   "Choose one —\n• Exile target creature. Return it at the beginning of the next end step.\n• Draw a card.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
			want:   "Return it at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "creating trigger's own end step condition is skipped",
			text:   "At the beginning of your end step, exile target creature. Return it to the battlefield at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			want:   "Return it to the battlefield at the beginning of the next end step.",
			ok:     true,
		},
		{
			name:   "timing selects among delayed sentences",
			text:   "Whenever this creature attacks, create a token. Exile it at end of combat. Draw a card at the beginning of the next end step.",
			timing: game.DelayedAtEndOfCombat,
			want:   "Exile it at end of combat.",
			ok:     true,
		},
		{
			name:   "next main phase",
			text:   "Untap all creatures you control. Add {R} at the beginning of your next main phase.",
			timing: game.DelayedAtBeginningOfNextMainPhase,
			spell:  true,
			want:   "Add {R} at the beginning of your next main phase.",
			ok:     true,
		},
		{
			name:   "granted quoted wording is not this ability's",
			text:   `Target creature gains "{T}: Sacrifice this creature at the beginning of the next end step." until end of turn.`,
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
		},
		{
			name:   "reminder text is not singled out",
			text:   "Encore {6}{W}{W} (For each opponent, create a token copy. Sacrifice them at the beginning of the next end step.)",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			name:   "two sentences with the same timing are ambiguous",
			text:   "Create a token. Sacrifice it at the beginning of the next end step. Return a card at the beginning of the next end step.",
			timing: game.DelayedAtBeginningOfNextEndStep,
			spell:  true,
		},
		{
			name:   "no timing wording",
			text:   "{T}: Create a token that's a copy of this creature.",
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := delayedTriggerTimingText(tc.text, tc.timing, tc.spell)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("delayedTriggerTimingText() = %q, %v; want %q, %v", got, ok, tc.want, tc.ok)
			}
		})
	}
}
