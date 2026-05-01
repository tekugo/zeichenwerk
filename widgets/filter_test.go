package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ---- helpers ---------------------------------------------------------------

func newFilter() *Filter {
	return NewFilter("f", "")
}

func newFilterWithList(items ...string) (*Filter, *List) {
	f := newFilter()
	l := NewList("l", "", items)
	f.Bind(l)
	return f, l
}

// typeIntoFilter simulates the user typing text into a Filter by setting the
// input text and dispatching EvtChange.
func typeIntoFilter(f *Filter, text string) {
	f.Input.Set(text)
	f.Input.Dispatch(&f.Input, EvtChange, text)
}

// ---- Filter / List integration ---------------------------------------------

func TestFilter_TypingFiltersListItems(t *testing.T) {
	f, l := newFilterWithList("apple", "banana", "apricot", "cherry")

	typeIntoFilter(f, "ap")

	items := l.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items after filter \"ap\", got %d: %v", len(items), items)
	}
	if items[0] != "apple" || items[1] != "apricot" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestFilter_ClearRestoresOriginalItems(t *testing.T) {
	original := []string{"apple", "banana", "apricot"}
	f, l := newFilterWithList(original...)

	typeIntoFilter(f, "ap")
	f.Clear()

	items := l.Items()
	if len(items) != len(original) {
		t.Fatalf("after Clear expected %d items, got %d", len(original), len(items))
	}
	for i, item := range original {
		if items[i] != item {
			t.Errorf("item[%d]: got %q want %q", i, items[i], item)
		}
	}
}

func TestFilter_FilterEmptyStringRestoresItems(t *testing.T) {
	original := []string{"foo", "bar", "baz"}
	f, l := newFilterWithList(original...)

	typeIntoFilter(f, "ba")
	typeIntoFilter(f, "")

	if len(l.Items()) != len(original) {
		t.Fatalf("Filter(\"\") should restore all %d items, got %d", len(original), len(l.Items()))
	}
}

func TestFilter_CaseInsensitiveMatch(t *testing.T) {
	_, l := newFilterWithList("Apple", "BANANA", "apricot")

	l.Filter("ap")

	items := l.Items()
	if len(items) != 2 {
		t.Fatalf("case-insensitive filter: expected 2, got %d: %v", len(items), items)
	}
}

func TestFilter_SubstringMatch(t *testing.T) {
	_, l := newFilterWithList("foobar", "barbaz", "hello")

	l.Filter("bar")

	items := l.Items()
	if len(items) != 2 {
		t.Fatalf("substring filter: expected 2 items, got %d: %v", len(items), items)
	}
}

// ---- Unbind ----------------------------------------------------------------

func TestFilter_UnbindRestoresListAndDetaches(t *testing.T) {
	original := []string{"one", "two", "three"}
	f, l := newFilterWithList(original...)

	typeIntoFilter(f, "on")
	if len(l.Items()) != 1 {
		t.Fatal("expected list to be filtered before Unbind")
	}

	f.Unbind()

	if len(l.Items()) != len(original) {
		t.Fatalf("Unbind should restore all items, got %d", len(l.Items()))
	}
	if f.Bound() != nil {
		t.Fatal("Bound() should be nil after Unbind")
	}

	// Typing after Unbind should not affect the list
	typeIntoFilter(f, "on")
	if len(l.Items()) != len(original) {
		t.Fatal("typing after Unbind should not filter the list")
	}
}

// ---- List.Suggest ----------------------------------------------------------

func TestList_Suggest_PrefixMatch(t *testing.T) {
	l := NewList("l", "", []string{"apple", "apricot", "banana"})

	got := l.Suggest("ap")
	if len(got) != 2 {
		t.Fatalf("Suggest(\"ap\"): expected 2, got %d: %v", len(got), got)
	}
}

func TestList_Suggest_EmptyReturnsNil(t *testing.T) {
	l := NewList("l", "", []string{"apple"})
	if l.Suggest("") != nil {
		t.Fatal("Suggest(\"\") should return nil")
	}
}

func TestList_Suggest_NoMatchReturnsNil(t *testing.T) {
	l := NewList("l", "", []string{"apple", "banana"})
	if l.Suggest("xyz") != nil {
		t.Fatal("Suggest with no match should return nil")
	}
}

func TestList_Suggest_SearchesOriginalWhileFiltered(t *testing.T) {
	l := NewList("l", "", []string{"apple", "apricot", "banana"})
	l.Filter("app") // only "apple" visible

	got := l.Suggest("ap")
	if len(got) != 2 {
		t.Fatalf("Suggest should search unfiltered items, got %d: %v", len(got), got)
	}
}

// ---- Ghost text wiring -----------------------------------------------------

func TestFilter_GhostTextAppearsWithSuggester(t *testing.T) {
	f, _ := newFilterWithList("apple", "apricot", "banana")

	typeIntoFilter(f, "ap")

	// Typeahead stores the ghost-text candidate in its suggestion field.
	if f.Typeahead.suggestion == "" {
		t.Fatal("ghost text should appear when bound widget implements Suggester")
	}
}

func TestFilter_NoGhostTextWithoutSuggester(t *testing.T) {
	f := newFilter()
	// Bind a Filterable that does not implement Suggester.
	f.Bind(&noSuggestImpl{})

	typeIntoFilter(f, "ap")

	if f.Typeahead.suggestion != "" {
		t.Fatal("no ghost text expected when bound widget does not implement Suggester")
	}
}

// noSuggestImpl is a minimal Filterable with no Suggest method.
type noSuggestImpl struct{}

func (n *noSuggestImpl) Filter(_ string) {}

// ---- Tab accepts ghost text ------------------------------------------------

func TestFilter_TabAcceptsGhostText(t *testing.T) {
	f, _ := newFilterWithList("apple", "apricot")
	typeIntoFilter(f, "ap")

	// There should be a ghost text suggestion now
	if f.Typeahead.suggestion == "" {
		t.Skip("no ghost text to accept")
	}

	var accepted string
	f.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			accepted, _ = data[0].(string)
		}
		return false
	})

	// Simulate Tab key
	tab := tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)
	f.Typeahead.handleKey(tab)

	if accepted == "" {
		t.Fatal("Tab should dispatch EvtAccept with the accepted suggestion")
	}
	if f.Get() != accepted {
		t.Errorf("input text %q should match accepted suggestion %q", f.Get(), accepted)
	}
}

// ---- Tree.Filter -----------------------------------------------------------

func TestTree_FilterShowsMatchingNodes(t *testing.T) {
	apple := NewTreeNode("apple")
	banana := NewTreeNode("banana")
	tr := newTree(apple, banana)

	tr.Filter("ap")

	if len(tr.flat) != 1 {
		t.Fatalf("expected 1 matching node, got %d", len(tr.flat))
	}
	if tr.flat[0].node.text != "apple" {
		t.Errorf("expected \"apple\", got %q", tr.flat[0].node.text)
	}
}

func TestTree_FilterExposesParentOfMatchingDescendant(t *testing.T) {
	parent := NewTreeNode("parent")
	child := NewTreeNode("needle")
	parent.Add(child)
	tr := newTree(parent)

	tr.Filter("needle")

	// parent and child should both appear
	if len(tr.flat) != 2 {
		t.Fatalf("expected 2 flat items (parent + child), got %d", len(tr.flat))
	}
	if tr.flat[0].node != parent {
		t.Error("first item should be parent")
	}
	if tr.flat[1].node != child {
		t.Error("second item should be matching child")
	}
}

func TestTree_FilterDoesNotMutateExpandedState(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("needle"))
	// parent starts collapsed
	tr := newTree(parent)

	tr.Filter("needle")

	if parent.Expanded() {
		t.Fatal("Filter should not mutate node.expanded")
	}
}

func TestTree_FilterEmptyStringRestoresAll(t *testing.T) {
	a := NewTreeNode("apple")
	b := NewTreeNode("banana")
	tr := newTree(a, b)

	tr.Filter("ap")
	tr.Filter("")

	if len(tr.flat) != 2 {
		t.Fatalf("Filter(\"\") should restore all nodes, got %d", len(tr.flat))
	}
}

func TestTree_FilterCaseInsensitive(t *testing.T) {
	tr := newTree(NewTreeNode("Apple"), NewTreeNode("banana"))

	tr.Filter("ap")

	if len(tr.flat) != 1 || tr.flat[0].node.text != "Apple" {
		t.Fatalf("filter should be case-insensitive, got %v", tr.flat)
	}
}

// ---- Tree.Suggest ----------------------------------------------------------

func TestTree_Suggest_PrefixMatch(t *testing.T) {
	parent := NewTreeNode("parent")
	parent.Add(NewTreeNode("apple"))
	parent.Add(NewTreeNode("apricot"))
	parent.Add(NewTreeNode("banana"))
	tr := newTree(parent)

	got := tr.Suggest("ap")
	if len(got) != 2 {
		t.Fatalf("Suggest(\"ap\"): expected 2, got %d: %v", len(got), got)
	}
}

func TestTree_Suggest_SearchesAllNodesRegardlessOfExpanded(t *testing.T) {
	parent := NewTreeNode("parent") // collapsed
	parent.Add(NewTreeNode("apple"))
	tr := newTree(parent)

	got := tr.Suggest("ap")
	if len(got) != 1 || got[0] != "apple" {
		t.Fatalf("Suggest should search collapsed subtrees, got %v", got)
	}
}

func TestTree_Suggest_EmptyReturnsNil(t *testing.T) {
	tr := newTree(NewTreeNode("apple"))
	if tr.Suggest("") != nil {
		t.Fatal("Suggest(\"\") should return nil")
	}
}

func TestTree_Suggest_NoMatchReturnsNil(t *testing.T) {
	tr := newTree(NewTreeNode("apple"))
	if tr.Suggest("xyz") != nil {
		t.Fatal("Suggest with no match should return nil")
	}
}
