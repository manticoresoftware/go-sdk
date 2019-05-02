package manticore

import "fmt"

// Keyword represents a keyword returned from BuildKeywords() call
type Keyword struct {
	Tokenized string	// token from the query
	Normalized string 	// normalized token after all stemming/lemming
	Querypos int		// position in the query
	Docs int   			// number of docs (from backend index)
	Hits int          	// number of hits (from backend index)
}

// Stringer interface for Keyword type
func (kw Keyword) String() string {
	return fmt.Sprintf("{Tok: '%v',\tNorm: '%v',\tQpos: %v; docs/hits %v/%v}\n", kw.Tokenized, kw.Normalized,
		kw.Querypos, kw.Docs, kw.Hits)
}

func buildKeywordsRequest(query, index string, hits bool) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(query)
		buf.putString(index)
		buf.putBoolDword(hits)

		buf.putBoolDword(false) // fixme! FoldLemmas
		buf.putBoolDword(false) // fixme! FoldBlended
		buf.putBoolDword(false) // fixme! FoldWildcards
		buf.putDword(0) // fixme! ExpansionLimit
	}
}

func parseKeywordsAnswer(hits bool) func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		nkeywords := answer.getInt()
		keywords := make([]Keyword, nkeywords)
		for j := 0; j < nkeywords; j++ {
			keywords[j].Tokenized = answer.getString()
			keywords[j].Normalized = answer.getString()
			keywords[j].Querypos = answer.getInt()
			if hits {
				keywords[j].Docs = answer.getInt()
				keywords[j].Hits = answer.getInt()
			}
		}
		return keywords
	}
}
