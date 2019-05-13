//
// Copyright (c) 2019, Manticore Software LTD (http://manticoresearch.com)
// All rights reserved
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License. You should have
// received a copy of the GPL license along with this program; if you
// did not, you can find it at http://www.gnu.org/
//

package manticore

import "fmt"

// Known commands
type eSearchdcommand uint16

const (
	commandSearch eSearchdcommand = iota
	commandExcerpt
	commandUpdate
	commandKeywords
	commandPersist
	commandStatus
	_
	commandFlushattrs
	commandSphinxql
	commandPing
	_ // commandDelete not exposed
	commandUvar
	_ // commandInsert not exposed
	_ // commandReplace not exposed
	_ // commandCommit not exposed
	_ // commandSuggest not exposed
	commandJson
	commandCallpq
	commandClusterpq

	commandTotal
	commandWrong = commandTotal
)

func (vl eSearchdcommand) String() string {
	switch vl {
	case commandSearch:
		return "search"
	case commandExcerpt:
		return "excerpt"
	case commandUpdate:
		return "update"
	case commandKeywords:
		return "keywords"
	case commandPersist:
		return "persist"
	case commandStatus:
		return "status"
	case commandFlushattrs:
		return "flushattrs"
	case commandSphinxql:
		return "sphinxql"
	case commandPing:
		return "ping"
	case commandUvar:
		return "uvar"
	case commandJson:
		return "json"
	case commandCallpq:
		return "callpq"
	case commandClusterpq:
		return "clusterpq"
	default:
		return fmt.Sprintf("wrong(%d)", uint16(vl))
	}
}

// known command versions
type uCommandVersion uint16

const verCommandWrong uCommandVersion = 0

var searchdcommandv = [commandTotal]uCommandVersion{

	0x121,           // search
	0x104,           // excerpt
	0x103,           // update
	0x101,           // keywords
	verCommandWrong, // persist
	0x101,           // status
	verCommandWrong, // _
	0x100,           // flushattrs
	0x100,           // sphinxql
	0x100,           // ping
	verCommandWrong, // delete
	0x100,           // uvar
	verCommandWrong, // insert
	verCommandWrong, // replace
	verCommandWrong, // commit
	verCommandWrong, // suggest
	0x100,           // json
	0x100,           // callpq
	0x100,           // clusterpq
}

func (vl uCommandVersion) String() string {
	return fmt.Sprintf("%d.%02d", byte(vl>>8), byte(vl&0xFF))
}

const (
	cphinxClientVersion uint32 = 1
	cphinxSearchdProto  uint32 = 1
	SphinxPort          uint16 = 9312 // Default IANA port for Sphinx API
)

// Document ID type
type DocID uint64

const DocidMax DocID = 0xffffffffffffffff

// eAggrFunc describes aggregate function in search query.
// Used in master-agent extensions for commandSearchMaster>=15
type eAggrFunc uint32

const (
	AggrNone eAggrFunc = iota // None
	AggrAvg                   // Avg()
	AggrMin                   // Min()
	AggrMax                   // Max()
	AggrSum                   // Sum()
	AggrCat                   // Cat()
)

// EAttrType represents known attribute types.
// See comments in constants for concrete meaning. Values of this type will be returned with resultset schema,
// you don't need to use them yourself.
type EAttrType uint32

const (
	AttrNone       EAttrType = iota       // not an attribute at all
	AttrInteger                           // unsigned 32-bit integer
	AttrTimestamp                         // this attr is a timestamp
	_                                     // there was SPH_ATTR_ORDINAL=3 once
	AttrBool                              // this attr is a boolean bit field
	AttrFloat                             // floating point number (IEEE 32-bit)
	AttrBigint                            // signed 64-bit integer
	AttrString                            // string (binary; in-memory)
	_                                     // there was SPH_ATTR_WORDCOUNT=8 once
	AttrPoly2d                            // vector of floats, 2D polygon (see POLY2D)
	AttrStringptr                         // string (binary, in-memory, stored as pointer to the zero-terminated string)
	AttrTokencount                        // field token count, 32-bit integer
	AttrJson                              // JSON subset; converted, packed, and stored as string
	AttrUint32set  EAttrType = 0x40000001 // MVA, set of unsigned 32-bit integers
	AttrInt64set   EAttrType = 0x40000002 // MVA, set of signed 64-bit integers
)

// these types are runtime only
// used as intermediate types in the expression engine
const (
	AttrMaparg      EAttrType = 1000 + iota
	AttrFactors               // packed search factors (binary, in-memory, pooled)
	AttrJsonField             // points to particular field in JSON column subset
	AttrFactorsJson           // packed search factors (binary, in-memory, pooled, provided to Client json encoded)
)

// eCollation is collation of search query. Used in master-agent extensions for commandSearchMaster>=1
type eCollation uint32

const (
	CollationLibcCi        eCollation = iota // Libc CI
	CollationLibcCs                          // Libc Cs
	CollationUtf8GeneralCi                   // Utf8 general CI
	CollationBinary                          // Binary

	CollationDefault = CollationLibcCi
)

// eFilterType describes different filters types. Internal.
type eFilterType uint32

const (
	FilterValues     eFilterType = iota // filter by integer values set
	FilterRange                         // filter by integer range
	FilterFloatrange                    // filter by float range
	FilterString                        // filter by string value
	FilterNull                          // filter by NULL
	FilterUservar                       // filter by @uservar
	FilterStringList                    // filter by string list
	FilterExpression                    // filter by expression
)

/* EGroupBy selects search query grouping mode. It is used as a param when calling `SetGroupBy()` function.

GroupbyDay

GroupbyDay extracts year, month and day in YYYYMMDD format from timestamp.

GroupbyWeek

GroupbyWeek extracts year and first day of the week number (counting from year start) in YYYYNNN format from timestamp.

GroupbyMonth

GroupbyMonth extracts month in YYYYMM format from timestamp.

GroupbyYear

GroupbyYear extracts year in YYYY format from timestamp.

GroupbyAttr

GroupbyAttr uses attribute value itself for grouping.

GroupbyMultiple

GroupbyMultiple group by on multiple attribute values. Allowed plain attributes and json fields; MVA and full JSONs are not allowed.
*/
type EGroupBy uint32

const (
	GroupbyDay      EGroupBy = iota // group by day
	GroupbyWeek                     // group by week
	GroupbyMonth                    // group by month
	GroupbyYear                     // group by year
	GroupbyAttr                     // group by attribute value
	_                               // GroupbyAttrpair, group by sequential attrs pair (rendered redundant by 64bit attrs support; removed)
	GroupbyMultiple                 // group by on multiple attribute values
)

// eQueryoption describes keyword expansion mode. Used only in master-agent mode of search query for commandSearchMaster>=16
type eQueryoption uint32

const (
	QueryOptDefault   eQueryoption = iota // Default
	QueryOptDisabled                      // Disabled
	QueryOptEnabled                       // Enabled
	QueryOptMorphNone                     // None morphology expansion
)

/*
ERankMode selects query relevance ranking mode. It is set via `SetRankingMode()` and `SetRankingExpression()` functions.

Manticore ships with a number of built-in rankers suited for different purposes. A number of them uses two factors,
phrase proximity (aka LCS) and BM25. Phrase proximity works on the keyword positions, while BM25 works on the keyword
frequencies. Basically, the better the degree of the phrase match between the document body and the query, the higher
is the phrase proximity (it maxes out when the document contains the entire query as a verbatim quote). And BM25 is
higher when the document contains more rare words. We’ll save the detailed discussion for later.

Currently implemented rankers are:

RankProximityBm25

RankProximityBm25, the default ranking mode that uses and combines both phrase proximity and BM25 ranking.

RankBm25

RankBm25, statistical ranking mode which uses BM25 ranking only (similar to most other full-text engines).
This mode is faster but may result in worse quality on queries which contain more than 1 keyword.

RankNone

RankNone, no ranking mode. This mode is obviously the fastest. A weight of 1 is assigned to all matches.
This is sometimes called boolean searching that just matches the documents but does not rank them.

RankWordcount

RankWordcount, ranking by the keyword occurrences count. This ranker computes the per-field keyword occurrence counts,
then multiplies them by field weights, and sums the resulting values.

RankProximity

RankProximity, returns raw phrase proximity value as a result. This mode is internally used to emulate MatchAll queries.

RankMatchany

RankMatchany, returns rank as it was computed in SPH_MATCH_ANY mode earlier, and is internally used to emulate MatchAny queries.

RankFieldmask

RankFieldmask, returns a 32-bit mask with N-th bit corresponding to N-th fulltext field, numbering from 0.
The bit will only be set when the respective field has any keyword occurrences satisfying the query.

RankSph04

RankSph04, is generally based on the default SPH_RANK_PROXIMITY_BM25 ranker, but additionally boosts the matches when
they occur in the very beginning or the very end of a text field. Thus, if a field equals the exact query,
SPH04 should rank it higher than a field that contains the exact query but is not equal to it. (For instance, when
the query is “Hyde Park”, a document entitled “Hyde Park” should be ranked higher than a one entitled “Hyde Park,
London” or “The Hyde Park Cafe”.)

RankExpr

RankExpr, lets you specify the ranking formula in run time. It exposes a number of internal text factors and lets you
define how the final weight should be computed from those factors.

RankExport

RankExport, rank by BM25, but compute and export all user expression factors

RankPlugin

RankPlugin, rank by user-defined ranker provided as UDF function.
*/
type ERankMode uint32

const (
	RankProximityBm25 ERankMode = iota // default mode, phrase proximity major factor and BM25 minor one (aka SPH03)
	RankBm25                           // statistical mode, BM25 ranking only (faster but worse quality)
	RankNone                           // no ranking, all matches get a weight of 1
	RankWordcount                      // simple word-count weighting, rank is a weighted sum of per-field keyword occurence counts
	RankProximity                      // phrase proximity (aka SPH01)
	RankMatchany                       // emulate old match-any weighting (aka SPH02)
	RankFieldmask                      // sets bits where there were matches
	RankSph04                          // codename SPH04, phrase proximity + bm25 + head/exact boost
	RankExpr                           // rank by user expression (eg. "sum(lcs*user_weight)*1000+bm25")
	RankExport                         // rank by BM25, but compute and export all user expression factors
	RankPlugin                         // user-defined ranker
	RankTotal
	RankDefault = RankProximityBm25
)

// ESearchdstatus describes known return codes. Also status codes for search command (but there 32bit)
type ESearchdstatus uint16

const (
	StatusOk      ESearchdstatus = iota // general success, command-specific reply follows
	StatusError                         // general failure, error message follows
	StatusRetry                         // temporary failure, error message follows, Client should retry late
	StatusWarning                       // general success, warning message and command-specific reply follow
)

// Stringer interface for ESearchdstatus type
func (vl ESearchdstatus) String() string {
	switch vl {
	case StatusOk:
		return "ok"
	case StatusError:
		return "error"
	case StatusRetry:
		return "retry"
	case StatusWarning:
		return "warning"
	default:
		return fmt.Sprintf("unknown(%d)", uint16(vl))
	}
}

/*
ESortOrder selects search query sorting orders

There are the following result sorting modes available:

SortRelevance

SortRelevance sorts by relevance in descending order (best matches first).

SortAttrDesc

SortAttrDescmode sorts by an attribute in descending order (bigger attribute values first).

SortAttrAsc

SortAttrAsc mode sorts by an attribute in ascending order (smaller attribute values first).

SortTimeSegments

SortTimeSegments sorts by time segments (last hour/day/week/month) in descending order, and then by relevance in descending order.
Attribute values are split into so-called time segments, and then sorted by time segment first, and by relevance second.

The segments are calculated according to the current timestamp at the time when the search is performed,
so the results would change over time. The segments are as follows:

last hour,

last day,

last week,

last month,

last 3 months,

everything else.

These segments are hardcoded, but it is trivial to change them if necessary.

This mode was added to support searching through blogs, news headlines, etc. When using time segments, recent records
would be ranked higher because of segment, but within the same segment, more relevant records would be ranked higher -
unlike sorting by just the timestamp attribute, which would not take relevance into account at all.

SortExtended

SortExtended sorts by SQL-like combination of columns in ASC/DESC order. You can specify an SQL-like sort expression
with up to 5 attributes (including internal attributes), eg:

@relevance DESC, price ASC, @id DESC

Both internal attributes (that are computed by the engine on the fly) and user attributes that were configured for this
index are allowed. Internal attribute names must start with magic @-symbol; user attribute names can be used as is.
In the example above, @relevance and @id are internal attributes and price is user-specified.

Known internal attributes are:

@id (match ID)

@weight (match weight)

@rank (match weight)

@relevance (match weight)

@random (return results in random order)

@rank and @relevance are just additional aliases to @weight.

SortExpr

SortExpr sorts by an arithmetic expression.

`SortRelevance` ignores any additional parameters and always sorts matches by relevance rank.
All other modes require an additional sorting clause, with the syntax depending on specific mode.
SortAttrAsc, SortAttrDesc and SortTimeSegments modes require simply an attribute name.
SortRelevance is equivalent to sorting by “@weight DESC, @id ASC” in extended sorting mode,
SortAttrAsc is equivalent to “attribute ASC, @weight DESC, @id ASC”,
and SortAttrDesc to “attribute DESC, @weight DESC, @id ASC” respectively.
*/
type ESortOrder uint32

const (
	SortRelevance    ESortOrder = iota // sort by document relevance desc, then by date
	SortAttrDesc                       // sort by document data desc, then by relevance desc
	SortAttrAsc                        // sort by document data asc, then by relevance desc
	SortTimeSegments                   // sort by time segments (hour/day/week/etc) desc, then by relevance desc
	SortExtended                       // sort by SQL-like expression (eg. "@relevance DESC, price ASC, @id DESC")
	SortExpr                           // sort by arithmetic expression in descending order (eg. "@id + max(@weight,1000)*boost + log(price)")
	SortTotal
)

/*
Qflags is bitmask with query flags which is set by calling Search.SetQueryFlags()
Different values have to be combined with '+' or '|' operation from following constants:

QflagReverseScan

Control the order in which full-scan query processes the rows.
  0 direct scan
  1 reverse scan

 QFlagSortKbuffer

Determines sort method for resultset sorting.
The result set is in both cases the same; picking one option or the other
may just improve (or worsen!) performance.
  0 priority queue
  1 k-buffer (gives faster sorting for already pre-sorted data, e.g. index data sorted by id)

QflagMaxPredictedTime

Determines if query has or not max_predicted_time option as an extra parameter
  0 no predicted time provided
  1 query contains predicted time metric

QflagSimplify

Switch on query boolean simplification to speed it up
If set to 1, daemon will simplify complex queries or queries that produced by different algos to eliminate and
optimize different parts of query.
  0 query will be calculated without transformations
  1 query will be transformed and simplified.

List of performed transformation is:

 common NOT
  ((A !N) | (B !N)) -> ((A|B) !N)

 common compound NOT
  ((A !(N C)) | (B !(N D))) -> (((A|B) !N) | (A !C) | (B !D)) // if cost(N) > cost(A) + cost(B)

 common sub-term
  ((A (X | C)) | (B (X | D))) -> (((A|B) X) | (A C) | (B D)) // if cost(X) > cost(A) + cost(B)

 common keywords
  (A | "A B"~N) -> A
  ("A B" | "A B C") -> "A B"
  ("A B"~N | "A B C"~N) -> ("A B"~N)

 common PHRASE
  ("X A B" | "Y A B") -> (("X|Y") "A B")

 common AND NOT factor
  ((A !X) | (A !Y) | (A !Z)) -> (A !(X Y Z))

 common OR NOT
  ((A !(N | N1)) | (B !(N | N2))) -> (( (A !N1) | (B !N2) ) !N)

 excess brackets
  ((A | B) | C) -> ( A | B | C )
  ((A B) C) -> ( A B C )

 excess AND NOT
  ((A !N1) !N2) -> (A !(N1 | N2))

 QflagPlainIdf

 Determines how BM25 IDF will be calculated. Below ``N'' is collection size, and ``n'' is number of matched documents
  1 plain IDF = log(N/n), as per Sparck-Jonesor
  0 normalized IDF = log((N-n+1)/n), as per Robertson et al

QflagGlobalIdf

Determines whether to use global statistics (frequencies) from the global_idf file for IDF computations,
rather than the local index statistics.
  0 use local index statistics
  1 use global_idf file (see https://docs.manticoresearch.com/latest/html/conf_options_reference/index_configuration_options.html#global-idf)

QflagNormalizedTfIdf

Determines whether to divide IDF value additionally by query word count, so that TF*IDF fits into [0..1] range
  0 don't divide IDF by query word count
  1 divide IDF by query word count

Notes for QflagPlainIdf and QflagNormalizedTfIdf flags

The historically default IDF (Inverse Document Frequency) in Manticore is equivalent to
QflagPlainIdf=0, QflagNormalizedTfIdf=1, and those normalizations may cause several undesired effects.

First, normalized idf (QflagPlainIdf=0) causes keyword penalization. For instance, if you search for [the | something]
and [the] occurs in more than 50% of the documents, then documents with both keywords [the] and [something] will get
less weight than documents with just one keyword [something]. Using QflagPlainIdf=1 avoids this. Plain IDF varies
in [0, log(N)] range, and keywords are never penalized; while the normalized IDF varies in [-log(N), log(N)] range,
and too frequent keywords are penalized.

Second, QflagNormalizedTfIdf=1 causes IDF drift over queries. Historically, we additionally divided IDF by query
keyword count, so that the entire sum(tf*idf) over all keywords would still fit into [0,1] range. However, that
means that queries [word1] and [word1 | nonmatchingword2] would assign different weights to the exactly same result
set, because the IDFs for both “word1” and “nonmatchingword2” would be divided by 2. QflagNormalizedTfIdf=0
fixes that. Note that BM25, BM25A, BM25F() ranking factors will be scaled accordingly once you
disable this normalization.

QflagLocalDf

Determines whether to automatically sum DFs over all the local parts of a distributed index,
so that the IDF is consistent (and precise) over a locally sharded index.
  0 don't sum local DFs
  1 sum local DFs

QflagLowPriority

Determines priority for executing the query
  0 run the query in usual (normal) priority
  1 run the query in idle priority

QflagFacet

Determines slave role of the query in multi-query facet
  0 query is not a facet query, or is main facet query
  1 query is depended (slave) part of facet multiquery

QflagFacetHead

Determines slave role of the query in multi-query facet
  0 query is not a facet query, or is slave of facet query
  1 query is main (head) query of facet multiquery

QflagJsonQuery

Determines if query is originated from REST api and so, must be parsed as one of JSON syntax
  0 query is API query
  1 query is JSON query
*/
type Qflags uint32

const (
	QflagReverseScan      Qflags = 1 << iota // direct or reverse full-scans
	QFlagSortKbuffer                         // pq or kbuffer for sorting
	QflagMaxPredictedTime                    // has or not max_predicted_time value
	QflagSimplify                            // apply or not boolean simplification
	QflagPlainIdf                            // plain or normalized idf
	QflagGlobalIdf                           // use or not global idf
	QflagNormalizedTfIdf                     // plain or normalized tf-idf
	QflagLocalDf                             // sum or not DFs over a locally sharderd (distributed) index
	QflagLowPriority                         // run query in idle priority
	QflagFacet                               // query is part of facet batch query
	QflagFacetHead                           // query is main facet query
	QflagJsonQuery                           // query is JSON query (otherwise - API query)
)
