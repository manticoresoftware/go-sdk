package manticore

/*
ExcerptFlags is bitmask for SnippetOptions.Flags
Different values have to be combined with '+' or '|' operation from following constants:

ExcerptFlagExactphrase

Whether to highlight exact query phrase matches only instead of individual keywords.

ExcerptFlagUseboundaries

Whether to additionally break passages by phrase boundary characters, as configured in
index settings with phrase_boundary directive.

ExcerptFlagWeightorder

Whether to sort the extracted passages in order of relevance (decreasing weight), or in order
of appearance in the document (increasing position).

ExcerptFlagQuery

Whether to handle 'words' as a query in extended syntax, or as a bag of words (default behavior).
For instance, in query mode "(one two | three four)" will only highlight and include those occurrences one two or
three four when the two words from each pair are adjacent to each other. In default mode, any single occurrence of
one, two, three, or four would be highlighted.

ExcerptFlagForceAllWords

Ignores the snippet length limit until it includes all the keywords.

ExcerptFlagLoadFiles

Whether to handle 'docs' as data to extract snippets from (default behavior), or to treat it as
file names, and load data from specified files on the server side. Up to dist_threads worker threads per request
will be created to parallelize the work when this flag is enabled. To parallelize snippets build between remote agents,
configure ``dist_threads'' param of searchd to value
greater than 1, and then invoke the snippets generation over the distributed index, which contain only one(!) local
agent and several remotes. The ``snippets_file_prefix'' param of remote daemons is also in the game and the final filename is calculated
by concatenation of the prefix with given name.

ExcerptFlagAllowEmpty

Allows empty string to be returned as highlighting result when a snippet could not be
generated (no keywords match, or no passages fit the limit). By default, the beginning of original text would be
returned instead of an empty string.

ExcerptFlagEmitZones

Emits an HTML tag with an enclosing zone name before each passage.

ExcerptFlagFilesScattered

It works only with distributed snippets generation with remote agents. The source files
for snippets could be distributed among different agents, and the main daemon will merge together all non-erroneous
results. So, if one agent of the distributed index has ‘file1.txt’, another has ‘file2.txt’ and you call for the
snippets with both these files, the sphinx will merge results from the agents together, so you will get the snippets
from both ‘file1.txt’ and ‘file2.txt’.

If the load_files is also set, the request will return the error in case if any of the files is not available
anywhere. Otherwise (if 'load_files' is not set) it will just return the empty strings for all absent files. The
master instance reset this flag when distributes the snippets among agents. So, for agents the absence of a file is
not critical error, but for the master it is so. If you want to be sure that all snippets are actually created,
set both `load_files_scattered` and `load_files`. If the absence of some snippets caused by some agents is not critical
for you - set just `load_files_scattered`, leaving `load_files` not set.

ExcerptFlagForcepassages

Whether to generate passages for snippet even if limits allow to highlight whole text.

Confusion and deprecation

*/
type ExcerptFlags uint32

const (
	_ ExcerptFlags = (1 << iota) // was: ExcerptFlagRemovespaces. Actually lost in implementation may years ago
	ExcerptFlagExactphrase
	_ // was: ExcerptFlagSinglepassage. Use LimitPassages=1 instead
	ExcerptFlagUseboundaries
	ExcerptFlagWeightorder
	ExcerptFlagQuery
	ExcerptFlagForceAllWords
	ExcerptFlagLoadFiles
	ExcerptFlagAllowEmpty
	ExcerptFlagEmitZones
	ExcerptFlagFilesScattered
	ExcerptFlagForcepassages
)

// SnippetOptions used to tune snippet's generation. All fields are exported and have meaning described below.
//
// BeforeMatch
//
// A string to insert before a keyword match. A '%PASSAGE_ID%' macro can be used in this string.
// The first match of the macro is replaced with an incrementing passage number within a current snippet.
// Numbering starts at 1 by default but can be overridden with start_passage_id option. In a multi-document call,
// '%PASSAGE_ID%' would restart at every given document.
//
// AfterMatch
//
// A string to insert after a keyword match. %PASSAGE_ID% macro can be used in this string.
//
// ChunkSeparator
//
// A string to insert between snippet chunks (passages).
//
// HtmlStripMode
//
// HTML stripping mode setting.
// Possible values are `index`, which means that index settings will be used, `none` and `strip`,
// that forcibly skip or apply stripping irregardless of index settings;
// and `retain`, that retains HTML markup and protects it from highlighting.
// The retain mode can only be used when highlighting full documents and thus requires that
// no snippet size limits are set. String, allowed values are none, strip, index, and retain.
//
// PassageBoundary
//
// Ensures that passages do not cross a sentence, paragraph, or zone boundary
// (when used with an index that has the respective indexing settings enabled).
// Allowed values are `sentence`, `paragraph`, and `zone`.
//
// Limit
//
// Maximum snippet size, in runes (codepoints).
//
// LimitPassages
//
// Limits the maximum number of passages that can be included into the snippet.
//
// LimitWords
//
// Limits the maximum number of words that can be included into the snippet.
// Note the limit applies to any words, and not just the matched keywords to highlight.
// For example, if we are highlighting Mary and a passage Mary had a little lamb is selected,
// then it contributes 5 words to this limit, not just 1
//
// Around
//
// How much words to pick around each matching keywords block.
//
// StartPassageId
//
// Specifies the starting value of `%PASSAGE_ID%` macro
// (that gets detected and expanded in before_match, after_match strings).
//
// Flags
//
// Bitmask. Individual bits described in `type ExcerptFlags` constants.
type SnippetOptions struct {
	BeforeMatch,
	AfterMatch,
	ChunkSeparator,
	HtmlStripMode,
	PassageBoundary string
	Limit,
	LimitPassages,
	LimitWords,
	Around,
	StartPassageId int32
	Flags ExcerptFlags
}

// Create default SnippetOptions with following defaults:
//
//  BeforeMatch: "<b>"
//  AfterMatch: "</b>"
//  ChunkSeparator: " ... "
//  HtmlStripMode: "index"
//  PassageBoundary: "none"
//  Limit: 256
//  Around: 5
//  StartPassageId: 1
//  //  Rest of the fields: 0, or "" (depends from type)
func NewSnippetOptions() *SnippetOptions {
	res := SnippetOptions{
		"<b>",
		"</b>",
		" ... ",
		"index",
		"none",
		256,
		0,
		0,
		5,
		1,
		0,
	}
	return &res
}

type snippetQuery struct {
	opts         *SnippetOptions
	docs         []string
	index, words string
}

func buildSnippetRequest(popts *SnippetOptions, docs []string, index, words string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putDword(0) // mode = 0
		buf.putDword(uint32(popts.Flags))
		buf.putString(index)
		buf.putString(words)

		// options
		buf.putString(popts.BeforeMatch)
		buf.putString(popts.AfterMatch)
		buf.putString(popts.ChunkSeparator)
		buf.putInt(popts.Limit)
		buf.putInt(popts.Around)
		buf.putInt(popts.LimitPassages)
		buf.putInt(popts.LimitWords)
		buf.putInt(popts.StartPassageId)
		buf.putString(popts.HtmlStripMode)
		buf.putString(popts.PassageBoundary)

		// documents
		ndocs := len(docs)
		buf.putLen(ndocs)
		for j := 0; j < ndocs; j++ {
			buf.putString(docs[j])
		}
	}
}

func parseSnippetAnswer(nreqs int) func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		snippets := make([]string, nreqs)
		for j := 0; j < nreqs; j++ {
			snippets[j] = answer.getString()
		}
		return snippets
	}
}
