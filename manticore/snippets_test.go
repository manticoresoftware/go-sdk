package manticore

import (
	"fmt"
	"testing"
)

func TestClient_BuildExcerpts_default(t *testing.T) {

	cl := NewClient()
	foo, err := cl.BuildExcerpts([]string{"10 word1 here", "20 word2 there"}, "lj", "word1 word2")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_BuildExcerpts_custom(t *testing.T) {

	cl := NewClient()
	opts := NewSnippetOptions()
	opts.BeforeMatch, opts.AfterMatch = "before", "after"
	opts.ChunkSeparator = "separator"
	opts.Limit = 10
	foo, err := cl.BuildExcerpts([]string{"10 word1 here", "20 word2 there"}, "lj", "word1 word2", *opts)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_BuildExcerpts_flags(t *testing.T) {

	cl := NewClient()
	opts := NewSnippetOptions()
	opts.Flags = ExcerptFlagExactphrase | ExcerptFlagUseboundaries | ExcerptFlagWeightorder
	foo, err := cl.BuildExcerpts([]string{"10 word1 here", "20 word2 there"}, "lj", "word1 word2", *opts)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}
