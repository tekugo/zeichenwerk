package zeichenwerk

import "testing"

// ── fuzzyMatch ───────────────────────────────────────────────────────────────

func TestFuzzyMatch_NoMatch_MissingChar(t *testing.T) {
	ok, _ := fuzzyMatch("oz", "Open file")
	if ok {
		t.Error("fuzzyMatch should return false when 'z' is not in the name")
	}
}

func TestFuzzyMatch_NoMatch_WrongOrder(t *testing.T) {
	ok, _ := fuzzyMatch("fo", "Of") // 'f' comes after 'o' in name
	if ok {
		t.Error("fuzzyMatch should return false when query characters appear out of order")
	}
}

func TestFuzzyMatch_Match_AllCharsPresent(t *testing.T) {
	ok, _ := fuzzyMatch("opfi", "Open file")
	if !ok {
		t.Error("fuzzyMatch should return true when all query chars appear in order")
	}
}

func TestFuzzyMatch_Match_WithSpaceInQuery(t *testing.T) {
	ok, _ := fuzzyMatch("op fi", "Open file")
	if !ok {
		t.Error("fuzzyMatch should handle spaces in query as regular characters to match")
	}
}

func TestFuzzyMatch_EmptyQuery_AlwaysMatches(t *testing.T) {
	ok, score := fuzzyMatch("", "Anything")
	if !ok {
		t.Error("empty query should always match")
	}
	if score != 0 {
		t.Errorf("empty query score = %d; want 0", score)
	}
}

func TestFuzzyMatch_WordBoundary_Position0(t *testing.T) {
	// 'O' is at position 0 → word-boundary bonus +5; case match +1 → total ≥ 6
	_, score := fuzzyMatch("O", "Open")
	if score < 5 {
		t.Errorf("score = %d; want ≥ 5 for match at word boundary (position 0)", score)
	}
}

func TestFuzzyMatch_WordBoundary_AfterSpace(t *testing.T) {
	// Match 'f' at position 5 in "Open file" (preceded by space) → +5
	_, scoreSpace := fuzzyMatch("f", "Open file")
	// Match 'i' at position 2 in "file" (not a word boundary) → 0 boundary bonus
	_, scoreNoWB := fuzzyMatch("i", "file")
	if scoreSpace <= scoreNoWB {
		t.Errorf("score after space (%d) should be > score not at boundary (%d)", scoreSpace, scoreNoWB)
	}
}

func TestFuzzyMatch_WordBoundary_AfterDash(t *testing.T) {
	_, score := fuzzyMatch("b", "foo-bar")
	if score < 5 {
		t.Errorf("score = %d; want ≥ 5 for match after '-'", score)
	}
}

func TestFuzzyMatch_WordBoundary_AfterUnderscore(t *testing.T) {
	_, score := fuzzyMatch("b", "foo_bar")
	if score < 5 {
		t.Errorf("score = %d; want ≥ 5 for match after '_'", score)
	}
}

func TestFuzzyMatch_AdjacencyBonus(t *testing.T) {
	// "op" matches positions 0,1 in "Open" → adjacency bonus for position 1 (+3)
	// plus boundary bonus for position 0 (+5) plus case match on 'o' at 0 (0, lowercase)
	_, scoreAdjacent := fuzzyMatch("op", "Open")
	// "oe" matches positions 0 and 3 in "Open" — not adjacent
	_, scoreGap := fuzzyMatch("oe", "Open")
	if scoreAdjacent <= scoreGap {
		t.Errorf("adjacent score (%d) should be > gap score (%d)", scoreAdjacent, scoreGap)
	}
}

func TestFuzzyMatch_CaseMatchBonus(t *testing.T) {
	// Exact case match adds +1 per char
	_, scoreLower := fuzzyMatch("open", "Open")
	_, scoreExact := fuzzyMatch("Open", "Open")
	if scoreExact <= scoreLower {
		t.Errorf("exact case score (%d) should be > lowercase query score (%d)", scoreExact, scoreLower)
	}
}

// ── filterCommands ────────────────────────────────────────────────────────────

func makeTestCommands() *Commands {
	c := &Commands{maxItems: 10}
	c.entries = []*Command{
		{Name: "New File", Shortcut: "Ctrl+N"},
		{Name: "Open File", Shortcut: "Ctrl+O"},
		{Name: "Save File", Shortcut: "Ctrl+S"},
		{Name: "Toggle Theme"},
		{Name: "Quit", Shortcut: "Ctrl+Q"},
	}
	return c
}

func TestFilterCommands_EmptyQuery_ReturnsAll(t *testing.T) {
	c := makeTestCommands()
	result := c.filterCommands("")
	if len(result) != len(c.entries) {
		t.Errorf("filterCommands(\"\") returned %d; want %d", len(result), len(c.entries))
	}
}

func TestFilterCommands_EmptyQuery_RegistrationOrder(t *testing.T) {
	c := makeTestCommands()
	result := c.filterCommands("")
	for i, r := range result {
		if r.cmd.Name != c.entries[i].Name {
			t.Errorf("result[%d].Name = %q; want %q", i, r.cmd.Name, c.entries[i].Name)
		}
	}
}

func TestFilterCommands_Query_FiltersNonMatching(t *testing.T) {
	c := makeTestCommands()
	result := c.filterCommands("quit")
	if len(result) != 1 || result[0].cmd.Name != "Quit" {
		t.Errorf("filterCommands(\"quit\") = %v; want [Quit]", result)
	}
}

func TestFilterCommands_SortsByScoreDesc(t *testing.T) {
	c := &Commands{maxItems: 10}
	// "Open File" should score higher for query "op" than "Copy" because 'o','p' are
	// at a word boundary in "Open File"
	c.entries = []*Command{
		{Name: "Copy"},
		{Name: "Open File"},
	}
	result := c.filterCommands("op")
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(result))
	}
	if result[0].cmd.Name != "Open File" {
		t.Errorf("result[0] = %q; want %q (higher score first)", result[0].cmd.Name, "Open File")
	}
}

func TestFilterCommands_Groups_PreservesGroupOrder(t *testing.T) {
	c := &Commands{maxItems: 10}
	c.entries = []*Command{
		{Name: "New File", Group: "File"},
		{Name: "Open File", Group: "File"},
		{Name: "Toggle Theme", Group: "View"},
		{Name: "Split Pane", Group: "View"},
	}
	result := c.filterCommands("")
	// Expected: File header, New File, Open File, View header, Toggle Theme, Split Pane
	if len(result) != 6 {
		t.Fatalf("expected 6 items (2 headers + 4 commands), got %d", len(result))
	}
	if !result[0].isHeader || result[0].cmd.Name != "File" {
		t.Errorf("result[0] should be File group header, got %+v", result[0])
	}
	if result[1].isHeader || result[1].cmd.Name != "New File" {
		t.Errorf("result[1] should be 'New File', got %+v", result[1])
	}
	if !result[3].isHeader || result[3].cmd.Name != "View" {
		t.Errorf("result[3] should be View group header, got %+v", result[3])
	}
}

func TestFilterCommands_Groups_HeadersSkippedWhenGroupEmpty(t *testing.T) {
	c := &Commands{maxItems: 10}
	c.entries = []*Command{
		{Name: "New File", Group: "File"},
		{Name: "Toggle Theme", Group: "View"},
	}
	// Query "new" matches only "New File" → only File group should appear
	result := c.filterCommands("new")
	for _, r := range result {
		if r.isHeader && r.cmd.Name == "View" {
			t.Error("View group header should not appear when no View commands match")
		}
	}
	if len(result) != 2 { // File header + New File
		t.Errorf("expected 2 items (1 header + 1 command), got %d", len(result))
	}
}

func TestFilterCommands_Groups_StableSortWithinGroup(t *testing.T) {
	c := &Commands{maxItems: 10}
	// Both "Save File" and "Save As" should match "sa"; "Save File" registered first
	c.entries = []*Command{
		{Name: "Save File", Group: "File"},
		{Name: "Save As", Group: "File"},
	}
	result := c.filterCommands("sa")
	// Both match "sa" at same positions. Result should have File header, then both.
	// Scores: "Save File" — 's' at pos 0 (+5+1), 'a' at pos 1 (+3+1) = 10
	//         "Save As"   — 's' at pos 0 (+5+1), 'a' at pos 1 (+3+1) = 10
	// Equal scores → registration order preserved (stable sort)
	if len(result) < 3 {
		t.Fatalf("expected header + 2 commands, got %d", len(result))
	}
	if result[1].cmd.Name != "Save File" {
		t.Errorf("result[1] = %q; want %q (registration order for equal scores)", result[1].cmd.Name, "Save File")
	}
}

// ── Register / Unregister ─────────────────────────────────────────────────────

func TestCommands_Register_Appends(t *testing.T) {
	c := &Commands{}
	cmd1 := c.Register("", "First", "", nil)
	cmd2 := c.Register("", "Second", "Ctrl+S", nil)
	if len(c.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(c.entries))
	}
	if c.entries[0] != cmd1 || c.entries[1] != cmd2 {
		t.Error("Register should append commands in order")
	}
}

func TestCommands_RegisterGroup_SetsGroup(t *testing.T) {
	c := &Commands{}
	cmd := c.Register("MyGroup", "Do Thing", "Ctrl+D", nil)
	if cmd.Group != "MyGroup" {
		t.Errorf("Group = %q; want %q", cmd.Group, "MyGroup")
	}
}

func TestCommands_Unregister_RemovesFirst(t *testing.T) {
	c := &Commands{}
	c.Register("", "Alpha", "", nil)
	c.Register("", "Beta", "", nil)
	c.Register("", "Alpha", "", nil) // duplicate
	ok := c.Unregister("Alpha")
	if !ok {
		t.Fatal("Unregister returned false; want true")
	}
	if len(c.entries) != 2 {
		t.Fatalf("expected 2 entries after unregister, got %d", len(c.entries))
	}
	// Second "Alpha" should still be present
	if c.entries[1].Name != "Alpha" {
		t.Errorf("second Alpha was incorrectly removed; entries = %v", c.entries)
	}
}

func TestCommands_Unregister_ReturnsFalseNotFound(t *testing.T) {
	c := &Commands{}
	ok := c.Unregister("NonExistent")
	if ok {
		t.Error("Unregister should return false when name not found")
	}
}

func TestCommands_Unregister_NoOpWhenOpen(t *testing.T) {
	c := &Commands{open: true}
	c.Register("", "Foo", "", nil)
	ok := c.Unregister("Foo")
	if ok {
		t.Error("Unregister should be a no-op while the palette is open")
	}
	if len(c.entries) != 1 {
		t.Error("entry should not have been removed while palette is open")
	}
}

func TestCommands_All_ReturnsSnapshot(t *testing.T) {
	c := &Commands{}
	c.Register("", "A", "", nil)
	c.Register("", "B", "", nil)
	snap := c.All()
	if len(snap) != 2 {
		t.Fatalf("All() len = %d; want 2", len(snap))
	}
	// Snapshot is independent of future mutations
	c.Register("", "C", "", nil)
	if len(snap) != 2 {
		t.Error("All() snapshot should not be affected by subsequent registrations")
	}
}

func TestCommands_SetMaxItems_Clamps(t *testing.T) {
	c := &Commands{maxItems: 10}
	c.SetMaxItems(1) // below minimum of 3
	if c.maxItems != 3 {
		t.Errorf("maxItems = %d; want 3 (clamped to minimum)", c.maxItems)
	}
}

// ── commandsPanel navigation ──────────────────────────────────────────────────

func makePanel() *commandsPanel {
	p := &commandsPanel{maxItems: 10, index: -1, lastClickIndex: -1}
	return p
}

func TestCommandsPanel_SetItems_FirstNonHeader(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "File"}, isHeader: true},
		{cmd: &Command{Name: "New File"}},
		{cmd: &Command{Name: "Open File"}},
	})
	if p.index != 1 {
		t.Errorf("index = %d after SetItems; want 1 (first non-header)", p.index)
	}
}

func TestCommandsPanel_SetItems_AllHeaders_IndexMinusOne(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "Group"}, isHeader: true},
	})
	if p.index != -1 {
		t.Errorf("index = %d; want -1 when all items are headers", p.index)
	}
}

func TestCommandsPanel_Move_SkipsHeaders(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "New File"}},             // 0
		{cmd: &Command{Name: "View"}, isHeader: true}, // 1
		{cmd: &Command{Name: "Toggle"}},               // 2
	})
	// Start at 0, move down — should skip header at 1, land on 2
	p.move(+1)
	if p.index != 2 {
		t.Errorf("index = %d after move(+1); want 2 (skipped header)", p.index)
	}
}

func TestCommandsPanel_Move_ClampsAtEnd(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "A"}},
		{cmd: &Command{Name: "B"}},
	})
	p.move(+1) // 0 → 1
	p.move(+1) // 1 → stays at 1 (no more items)
	if p.index != 1 {
		t.Errorf("index = %d; want 1 (clamped at end)", p.index)
	}
}

func TestCommandsPanel_Move_ClampsAtStart(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "A"}},
		{cmd: &Command{Name: "B"}},
	})
	p.move(+1) // 0 → 1
	p.move(-1) // 1 → 0
	p.move(-1) // 0 → stays at 0
	if p.index != 0 {
		t.Errorf("index = %d; want 0 (clamped at start)", p.index)
	}
}

func TestCommandsPanel_Home_JumpsToFirst(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "A"}},
		{cmd: &Command{Name: "B"}},
		{cmd: &Command{Name: "C"}},
	})
	p.move(+2) // go to C
	p.home()
	if p.index != 0 {
		t.Errorf("index = %d after home(); want 0", p.index)
	}
}

func TestCommandsPanel_End_JumpsToLast(t *testing.T) {
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "A"}},
		{cmd: &Command{Name: "B"}},
		{cmd: &Command{Name: "C"}, isHeader: false},
	})
	p.end()
	if p.index != 2 {
		t.Errorf("index = %d after end(); want 2", p.index)
	}
}

func TestCommandsPanel_End_SkipsTrailingHeader(t *testing.T) {
	// Trailing headers should be skipped by end()
	p := makePanel()
	p.SetItems([]rankedCommand{
		{cmd: &Command{Name: "A"}},
		{cmd: &Command{Name: "B"}},
		{cmd: &Command{Name: "Trailing"}, isHeader: true},
	})
	p.end()
	if p.index != 1 {
		t.Errorf("index = %d after end() with trailing header; want 1", p.index)
	}
}

func TestCommandsPanel_Focused_ReturnsNilWhenEmpty(t *testing.T) {
	p := makePanel()
	if p.focused() != nil {
		t.Error("focused() should return nil when no items set")
	}
}

func TestCommandsPanel_GroupHeaders_InRankedSlice(t *testing.T) {
	c := &Commands{maxItems: 10}
	c.entries = []*Command{
		{Name: "New", Group: "File"},
		{Name: "Zoom", Group: "View"},
	}
	result := c.filterCommands("")
	hasHeader := false
	for _, r := range result {
		if r.isHeader {
			hasHeader = true
			break
		}
	}
	if !hasHeader {
		t.Error("filterCommands should include isHeader entries when groups are used")
	}
}
