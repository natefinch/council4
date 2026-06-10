package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSagaEntersWithLoreCounterAndTriggersFirstChapter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player1, []game.ChapterAbility{
		sagaDrawChapter(1),
		sagaDrawChapter(2),
	})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	permanent, ok := createCardPermanentFace(g, card, game.Player1, zone.Stack, game.FaceFront)
	if !ok {
		t.Fatal("create Saga permanent failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 1 {
		t.Fatalf("lore counters = %d, want 1", got)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first chapter was not put on the stack")
	}
	object, ok := g.Stack.Peek()
	if !ok || !object.SagaChapter || object.SourceID != permanent.ObjectID {
		t.Fatalf("stack object = %+v, want first Saga chapter", object)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want first chapter to draw", got)
	}
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed before its final chapter")
	}
}

func TestSagaCrossesMultipleChaptersAndSacrificesAfterTheyLeaveStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{
		sagaDrawChapter(1),
		sagaDrawChapter(2),
		sagaDrawChapter(3),
	}))
	permanent.Counters.Add(counter.Lore, 1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})

	addCountersToPermanent(g, permanent, counter.Lore, 2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("crossed chapters were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want two crossed chapter abilities", got)
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed while chapter abilities remained on the stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed while one chapter ability remained on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("Saga remained after its final chapter ability left the stack")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want both crossed chapters to resolve", got)
	}
}

func TestSagaSharedChapterAbilityTriggersOnceWhenSeveralNumbersAreCrossed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{{
		Text:     "II, III — Draw a card.",
		Chapters: []int{2, 3},
		Content:  sagaDrawChapter(2).Content,
	}}))
	permanent.Counters.Add(counter.Lore, 1)

	addCountersToPermanent(g, permanent, counter.Lore, 2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("shared chapter ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want shared chapter ability once", got)
	}
}

func TestSagaChapterResolvesAfterSourceLeavesBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{
		sagaDrawChapter(1),
	}))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	addCountersToPermanent(g, permanent, counter.Lore, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chapter ability was not put on the stack")
	}
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("moving Saga to graveyard failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want chapter to resolve from inline source data", got)
	}
}

func TestAdvanceSagasAddsLoreOnlyToActivePlayersSagas(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(2)}))
	second := addCombatPermanent(g, game.Player2, sagaCard([]game.ChapterAbility{sagaDrawChapter(2)}))
	first.Counters.Add(counter.Lore, 1)
	second.Counters.Add(counter.Lore, 1)

	advanceSagas(g, game.Player1)

	if got := first.Counters.Get(counter.Lore); got != 2 {
		t.Fatalf("active player's lore counters = %d, want 2", got)
	}
	if got := second.Counters.Get(counter.Lore); got != 1 {
		t.Fatalf("nonactive player's lore counters = %d, want 1", got)
	}
}

func addSagaCardInstance(g *game.Game, owner game.PlayerID, chapters []game.ChapterAbility) *game.CardInstance {
	card := &game.CardInstance{
		ID:    g.IDGen.Next(),
		Owner: owner,
		Def:   sagaCard(chapters),
	}
	g.CardInstances[card.ID] = card
	return card
}

func sagaCard(chapters []game.ChapterAbility) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:             "Test Saga",
		Types:            []types.Card{types.Enchantment},
		Subtypes:         []types.Sub{types.Saga},
		ChapterAbilities: chapters,
	}}
}

func sagaDrawChapter(number int) game.ChapterAbility {
	return game.ChapterAbility{
		Text:     "Draw a card.",
		Chapters: []int{number},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
}
