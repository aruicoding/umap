package umap_test

import (
	"testing"

	"github.com/aruicoding/umap"
)

func TestNew(t *testing.T) {
	t.Log(umap.New(0))
	t.Log(umap.New(5))
	t.Log(umap.New(10))
}

func TestSet(t *testing.T) {
	cap := uint(10)
	m := umap.New(cap)
	{
		m.Set("go", "0")
		m.Set("go", "1")
		t.Log(m.Get("go"))
	}
	{
		m.Set(struct{ name string }{"sandwich"}, "food")
		t.Log(m.Get(struct{ name string }{"sandwich"}))
		m.Set("food", struct{ name string }{"sandwich"})
		t.Log(m.Get("food"))
		m.Set([]string{"apple", "banana", "peach"}, "fruits")
		t.Log(m.Get([]string{"apple", "banana", "peach"}))
		m.Set("fruits", []string{"apple", "banana", "peach"})
		t.Log(m.Get("fruits"))
		m.Set(map[string]int{"bird": 0, "cat": 1}, true)
		t.Log(m.Get(map[string]int{"bird": 0, "cat": 1}))
		m.Set(true, map[string]int{"bird": 0, "cat": 1})
		t.Log(m.Get(true))
	}
}

func TestDel(t *testing.T) {
	um := umap.New(1)
	{
		um.Set("id", 1)
		t.Log(um.Get("id"))
		um.Del("id")
		t.Log(um.Get("id"))
		um.Set("id", 1)
		t.Log(um.Get("id"))
	}

}
