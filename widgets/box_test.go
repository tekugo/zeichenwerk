package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

func TestBox_Insert_ReplacesAtZero(t *testing.T) {
	b := NewBox("box", "", "")
	c1 := NewComponent("c1", "")
	c2 := NewComponent("c2", "")
	if err := b.Insert(0, c1); err != nil {
		t.Fatalf("Insert(0, c1) returned error: %v", err)
	}
	if b.child != c1 {
		t.Error("Insert(0) should set box.child")
	}
	if err := b.Insert(0, c2); err != nil {
		t.Fatalf("Insert(0, c2) returned error: %v", err)
	}
	if b.child != c2 {
		t.Error("Insert(0) on populated box should replace child")
	}
	if c1.Parent() != nil {
		t.Error("replaced child's parent should be cleared")
	}
}

func TestBox_Insert_NonZeroIsFull(t *testing.T) {
	b := NewBox("box", "", "")
	c := NewComponent("c", "")
	if err := b.Insert(1, c); err != ErrFull {
		t.Errorf("Insert(1) on Box = %v; want ErrFull", err)
	}
}

func TestBox_Insert_Nil(t *testing.T) {
	b := NewBox("box", "", "")
	if err := b.Insert(0, nil); err != ErrChildIsNil {
		t.Errorf("Insert(0, nil) = %v; want ErrChildIsNil", err)
	}
}

func TestBox_Remove(t *testing.T) {
	b := NewBox("box", "", "")
	c := NewComponent("c", "")
	b.Add(c)
	if err := b.Remove(c); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if b.child != nil {
		t.Error("Remove should clear box.child")
	}
	if c.Parent() != nil {
		t.Error("Remove should clear the removed child's parent")
	}
}

func TestBox_Remove_NotFound(t *testing.T) {
	b := NewBox("box", "", "")
	c := NewComponent("c", "")
	stranger := NewComponent("stranger", "")
	b.Add(c)
	if err := b.Remove(stranger); err != ErrNotFound {
		t.Errorf("Remove(stranger) = %v; want ErrNotFound", err)
	}
}
