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
	m.Set("go", "0")
	m.Set("java", "0")
	m.Set("javaScript", "0")
	m.Set("php", "0")
	// for i := 0; i < 10; i++ {
	// 	m.Set(fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
	// }
}
