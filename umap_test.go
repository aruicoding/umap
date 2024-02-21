package umap_test

import (
	"fmt"
	"math/rand"
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

func TestGrow(t *testing.T) {
	m := umap.New(0)
	arr := make([]string, 0)
	testNum := 3500000
	for i := 0; i < testNum; i++ {
		k := fmt.Sprintf("key%08d-%09d", i, i*rand.Intn(10))
		m.Set(k, i)
		arr = append(arr, k)
	}
	var ks int
	for i, k := range arr {
		if v := m.Get(k); v != nil {
			if i != v {
				t.Log(k, v)
			}
			ks++
		}
	}
	t.Log(ks == testNum, ks)
	ks = 0
	m.Iteration(func(k any, v any) {
		if m.Get(k) == v {
			ks++
		}
	})
	t.Log(ks == testNum, ks)
}
