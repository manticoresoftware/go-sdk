package manticore

import (
	"errors"
	"fmt"
	"time"
)

// version of master-agent SEARCH command extension
const commandSearchMaster uint32 = 16

/* EMatchMode selects search query matching mode.
So-called matching modes are a legacy feature that used to provide (very) limited query syntax and ranking support.
Currently, they are deprecated in favor of full-text query language and so-called Available built-in rankers.
It is thus strongly recommended to use `MatchExtended` and proper query syntax rather than any other legacy mode.
All those other modes are actually internally converted to extended syntax anyway. SphinxAPI still defaults to
`MatchAll` but that is for compatibility reasons only.

There are the following matching modes available:

MatchAll

MatchAll matches all query words.

MatchAny

MatchAny matches any of the query words.

MatchPhrase

MatchPhrase, matches query as a phrase, requiring perfect match.

MatchBoolean

MatchBoolean, matches query as a boolean expression (see Boolean query syntax).

MatchExtended

MatchExtended2

MatchExtended, MatchExtended2 (alias) matches query as an expression in Manticore internal query language
(see Extended query syntax). This is default matching mode if nothing else specified.

MatchFullscan

MatchFullscan, matches query, forcibly using the “full scan” mode as below. NB, any query terms will be ignored,
such that filters, filter-ranges and grouping will still be applied, but no text-matching. MatchFullscan mode will be
automatically activated in place of the specified matching mode when the query string is empty (ie. its length is zero).

In full scan mode, all the indexed documents will be considered as matching.
Such queries will still apply filters, sorting, and group by, but will not perform any full-text searching.
This can be useful to unify full-text and non-full-text searching code, or to offload SQL server
(there are cases when Manticore scans will perform better than analogous MySQL queries).
An example of using the full scan mode might be to find posts in a forum. By selecting the forum’s user ID via
SetFilter() but not actually providing any search text, Manticore will match every document (i.e. every post)
where SetFilter() would match - in this case providing every post from that user. By default this will be ordered by
relevancy, followed by Manticore document ID in ascending order (earliest first).
*/
type EMatchMode uint32

const (
	MatchAll       EMatchMode = iota // match all query words
	MatchAny                         // match any query word
	MatchPhrase                      // match this exact phrase
	MatchBoolean                     // match this boolean query
	MatchExtended                    // match this extended query
	MatchFullscan                    // match all document IDs w/o fulltext query, apply filters
	MatchExtended2                   // extended engine V2 (TEMPORARY, WILL BE REMOVED IN 0.9.8-RELEASE)

	MatchTotal
)

type searchFilter struct {
	Attribute  string
	FilterType eFilterType
	Exclude    bool
	FilterData interface{}
}

// Search represents one search query.
// Exported fields may be set directly. Unexported which bind by internal dependencies and constrains
// intended to be set wia special methods.
type Search struct {
	Offset        int32 // offset into resultset (0)
	Limit         int32 // count of resultset (20)
	MaxMatches    int32
	CutOff        int32
	RetryCount    int32
	MaxQueryTime  time.Duration
	RetryDelay    time.Duration
	predictedTime time.Duration
	MatchMode     EMatchMode // Matching mode
	ranker        ERankMode
	sort          ESortOrder
	rankexpr      string
	sortby        string
	FieldWeights  map[string]int32 // bind per-field weights by name
	IndexWeights  map[string]int32 // bind per-index weights by name
	IDMin         DocID            // set IDs range to match (from)
	IDMax         DocID            // set IDs range to match (to)
	filters       []searchFilter
	geoLatAttr    string
	geoLonAttr    string
	geoLatitude   float32
	geoLongitude  float32
	Groupfunc     EGroupBy
	GroupBy       string
	GroupSort     string
	GroupDistinct string // count-distinct attribute for group-by queries
	SelectClause  string // select-list (attributes or expressions), SQL-like syntax
	queryflags    Qflags
	outerorderby  string
	outeroffset   int32
	outerlimit    int32
	hasouter      bool
	tokenFlibrary string
	tokenFname    string
	tokenFopts    string
	Indexes       string
	Comment       string
	Query         string
}

// NewSearch construct default search which then may be customized. You may just customize 'Query' and m.b. 'Indexes'
// from default one, and it will work like a simple 'Query()' call.
func NewSearch(query, index, comment string) Search {
	return Search{
		0, 20, 1000, 0, 0,
		0, 0, 0,
		MatchAll,
		RankDefault,
		SortRelevance,
		"", "",
		nil, nil,
		0, 0,
		nil,
		"", "",
		0, 0,
		GroupbyDay,
		"", "@group desc", "",
		"",
		QflagNormalizedTfIdf,
		"",
		0, 0,
		false,
		"", "", "",
		index, comment, query,
	}
}

/*
AddFilter adds new integer values set filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name

`values` must be a plain slice containing integer values.

`exclude` controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index matches any of the values from `values` slice
will be matched (or rejected, if `exclude` is true).
 */
func (q *Search) AddFilter(attribute string, values []int64, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterValues, exclude, values})
}

/*
AddFilterExpression adds new filter by expression.

On this call, additional new filter is added to the existing list of filters.

The only value `expression` must contain filtering expression which returns bool.

Expression has SQL-like syntax and may refer to columns (usually json fields) by name, and may look like: 'j.price - 1 > 3 OR j.tag IS NOT null'
Documents either filtered by 'true' expression, either (if `exclude` is set to true) by 'false'.
*/
func (q *Search) AddFilterExpression(expression string, exclude bool) {
	q.filters = append(q.filters, searchFilter{expression, FilterExpression, exclude, nil})
}

/*
AddFilterFloatRange adds new float range filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name.

`fmin` and `fmax` must be floats that define the acceptable attribute values range (including the boundaries).

`exclude` controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index is between `fmin` and `fmax`
(including values that are exactly equal to `fmin` or `fmax`) will be matched (or rejected, if `exclude` is true).
 */
func (q *Search) AddFilterFloatRange(attribute string, fmin, fmax float32, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterFloatrange, exclude, []float32{fmin, fmax}})
}

// AddFilterNull adds new IsNull filter.
//
//On this call, additional new filter is added to the existing list of filters. Documents where `attribute` is null will match,
//(if `isnull` is true) or not match (if `isnull` is false).
func (q *Search) AddFilterNull(attribute string, isnull bool) {

	q.filters = append(q.filters, searchFilter{attribute, FilterNull, false, isnull})
}

/*
AddFilterRange adds new integer range filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name.

`imin` and `imax` must be integers that define the acceptable attribute values range (including the boundaries).

`exclude` controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index is between `imin` and `imax`
(including values that are exactly equal to `imin` or `imax`) will be matched (or rejected, if `exclude` is true).
*/
func (q *Search) AddFilterRange(attribute string, imin, imax int64, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterRange, exclude, []int64{imin, imax}})
}

/*
AddFilterString adds new string value filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name.

`value` must be a string.

`exclude` must be a boolean value; it controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index equal to string value from `value` will be matched (or rejected, if `exclude` is true).
 */
func (q *Search) AddFilterString(attribute string, value string, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterString, exclude, value})
}

/*
AddFilterStringList adds new string list filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name.

`values` must be slice of strings

`exclude` must be a boolean value; it controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index equal to one of string values from `values` will be matched (or rejected, if `exclude` is true).
*/
func (q *Search) AddFilterStringList(attribute string, values []string, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterStringList, exclude, values})
}

/*
AddFilterUservar adds new uservar filter.

On this call, additional new filter is added to the existing list of filters.

`attribute` must be a string with attribute name.

`uservar` must be name of user variable, containing list of filtering values, as "@var"

`exclude` must be a boolean value; it controls whether to accept the matching documents (default mode, when `exclude` is false) or reject them.

Only those documents where `attribute` column value stored in the index equal to one of the values stored in `uservar` variable on daemon side (or rejected, if `exclude` is true).
Such filter intended to save huge list of variables once on the server, and then refer to it by name. Saving the list might be done by separate call of 'SetUservar()'
*/
func (q *Search) AddFilterUservar(attribute string, uservar string, exclude bool) {
	q.filters = append(q.filters, searchFilter{attribute, FilterUservar, exclude, uservar})
}

// ChangeQueryFlags changes (set or reset) query flags by mask `flags`.
func (q *Search) ChangeQueryFlags(flags Qflags, set bool) {
	if set {
		q.queryflags |= flags
	} else {
		q.queryflags &^= flags
		if !q.hasSetQueryFlag(QflagMaxPredictedTime) {
			q.predictedTime = 0
		}
	}
}

/*
ResetFilters clears all currently set search filters.

This call is only normally required when using multi-queries. You might want to set different filters for different
queries in the batch. To do that, you may either create another Search request and fill it from the scratch, either
copy existing (last one) and modify. To change all the filters in the copy you can call ResetFilters() and add new
filters using the respective calls.
 */
func (q *Search) ResetFilters() {
	q.geoLatAttr, q.geoLonAttr = "", ""
	q.filters = nil
}

/*
ResetGroupBy clears all currently group-by settings, and disables group-by.

This call is only normally required when using multi-queries. You might want to set different
group-by settings in the batch. To do that, you may either create another Search request and fill ot from the scratch, either
copy existing (last one) and modify. In last case you can change individual group-by settings using SetGroupBy() and SetGroupDistinct() calls,
but you can not disable group-by using those calls. ResetGroupBy() fully resets previous group-by settings and
disables group-by mode in the current Search query.
 */
func (q *Search) ResetGroupBy() {
	q.GroupDistinct, q.GroupBy = "", ""
	q.GroupSort = "@group desc"
	q.Groupfunc = GroupbyDay
}

/*
ResetOuterSelect clears all outer select settings

This call is only normally required when using multi-queries. You might want to set different
outer select settings in the batch. To do that, you may either create another Search request and fill ot from the scratch, either
copy existing (last one) and modify. In last case you can change individual group-by settings using SetOuterSelect() calls,
but you can not disable outer statement by this calls. ResetOuterSelect() fully resets previous outer select settings.
 */
func (q *Search) ResetOuterSelect() {
	q.outerorderby, q.outeroffset, q.outerlimit, q.hasouter = "", 0, 0, false
}

/*
ResetQueryFlags resets query flags of Select query to default value, and also reset value set by SetMaxPredictedTime() call.

This call is only normally required when using multi-queries. You might want to set different
flags of Select queries in the batch. To do that, you may either create another Search request and fill ot from the scratch, either
copy existing (last one) and modify. In last case you can change individual or many flags using SetQueryFlags() and ChangeQueryFlags() calls.
This call just one-shot set all the flags to default value `QflagNormalizedTfIdf`, and also set predicted time to 0.
 */
func (q *Search) ResetQueryFlags() {
	q.queryflags = QflagNormalizedTfIdf
	q.predictedTime = 0
}

/*
SetGeoAnchor sets anchor point for and geosphere distance (geodistance) calculations, and enable them.

`attrlat` and `attrlong` contain the names of latitude and longitude attributes, respectively.

`lat` and `long` specify anchor point latitude and longitude, in radians.

Once an anchor point is set, you can use magic @geodist attribute name in your filters and/or sorting expressions.
Manticore will compute geosphere distance between the given anchor point and a point specified by latitude and
longitude attributes from each full-text match, and attach this value to the resulting match. The latitude and l
ongitude values both in SetGeoAnchor and the index attribute data are expected to be in radians. The result will
be returned in meters, so geodistance value of 1000.0 means 1 km. 1 mile is approximately 1609.344 meters.
 */
func (q *Search) SetGeoAnchor(attrlat, attrlong string, lat, long float32) {
	q.geoLatAttr, q.geoLonAttr = attrlat, attrlong
	q.geoLatitude, q.geoLongitude = lat, long
}

func (q *Search) hasGeoAnchor() bool {
	return q.geoLatAttr != "" && q.geoLonAttr != ""
}

/*
SetGroupBy sets grouping attribute, function, and groups sorting mode; and enables grouping.

`attribute` is a string that contains group-by attribute name.

`func` is a constant that chooses a function applied to the attribute value in order to compute group-by key.

`groupsort` is optional clause that controls how the groups will be sorted.

Grouping feature is very similar in nature to GROUP BY clause from SQL.
Results produces by this function call are going to be the same as produced by the following pseudo code:
 SELECT ... GROUP BY func(attribute) ORDER BY groupsort

Note that it’s `groupsort` that affects the order of matches in the final result set.
Sorting mode (see `SetSortMode()`) affect the ordering of matches within group, ie. what match will be selected
as the best one from the group. So you can for instance order the groups by matches count and select the most relevant
match within each group at the same time.

Grouping on string attributes is supported, with respect to current collation.
 */
func (q *Search) SetGroupBy(attribute string, gfunc EGroupBy, groupsort ...string) {
	if len(groupsort) > 0 {
		q.GroupSort = groupsort[0]
	}

	q.GroupBy = attribute
	q.Groupfunc = gfunc
}

// SetQueryFlags set query flags. New flags are |-red to existing value, previously set flags are not affected.
// Note that default flags has set QflagNormalizedTfIdf bit, so if you need to reset it, you need to explicitly invoke
// ChangeQueryFlags(QflagNormalizedTfIdf,false) for it.
func (q *Search) SetQueryFlags(flags Qflags) {
	q.queryflags |= flags
}

func (q *Search) hasSetQueryFlag(flag Qflags) bool {
	return (q.queryflags & flag) != 0
}

// SetMaxPredictedTime set max predicted time and according query flag
func (q *Search) SetMaxPredictedTime(predtime time.Duration) {
	q.predictedTime = predtime
	q.SetQueryFlags(QflagMaxPredictedTime)
}

// SetOuterSelect determines outer select conditions for Search query.
//
// `orderby` specify clause with SQL-like syntax as "foo ASC, bar DESC, baz" where name of the items (`foo`, `bar`, `baz` in example) are the names of columns originating from internal query.
//
// `offset` and `limit` has the same meaning as fields Offset and Limit in the clause, but applied to outer select.
//
// Outer select currently have 2 usage cases:
//
// 1. We have a query with 2 ranking UDFs, one very fast and the other one slow and we perform a full-text search will a big match result set. Without outer the query would look like
//
//  q := NewSearch("some common query terms", "index", "")
//  q.SelectClause = "id, slow_rank() as slow, fast_rank as fast"
//  q.SetSortMode( SortExtended, "fast DESC, slow DESC" )
//  // q.Limit=20, q.MaxMatches=1000 - are default, so we don't set them explicitly
//
// With subselects the query can be rewritten as :
//  q := NewSearch("some common query terms", "index", "")
//  q.SelectClause = "id, slow_rank() as slow, fast_rank as fast"
//  q.SetSortMode( SortExtended, "fast DESC" )
//  q.Limit=100
//  q.SetOuterSelect("slow desc", 0, 20)
//
// In the initial query the slow_rank() UDF is computed for the entire match result set.
// With subselects, only fast_rank() is computed for the entire match result set, while slow_rank() is only computed for a limited set.
//
// 2. The second case comes handy for large result set coming from a distributed index.
//
// For this query:
//
//  q := NewSearch("some conditions", "my_dist_index", "")
//  q.Limit = 50000
// If we have 20 nodes, each node can send back to master a number of 50K records, resulting in 20 x 50K = 1M records,
// however as the master sends back only 50K (out of 1M), it might be good enough for us for the nodes to send only the
// top 10K records. With outer select we can rewrite the query as:
//
//  q := NewSearch("some conditions", "my_dist_index", "")
//  q.Limit = 10000
//  q.SetOuterSelect("some_attr", 0, 50000)
//In this case, the nodes receive only the inner query and execute. This means the master will receive only 20x10K=200K
// records. The master will take all the records received, reorder them by the OUTER clause and return the best 50K
// records. The outer select helps reducing the traffic between the master and the nodes and also reduce the master’s
// computation time (as it process only 200K instead of 1M).
func (q *Search) SetOuterSelect(orderby string, offset, limit int32) {
	q.outerorderby = orderby
	q.outeroffset, q.outerlimit, q.hasouter = offset, limit, true
}

// SetRankingExpression assigns ranking expression, and also set ranking mode to RankExpr
//
// `rankexpr` provides ranking formula, for example, "sum(lcs*user_weight)*1000+bm25" - this is the same
// as RankProximityBm25, but written explicitly.
// Since using ranking expression assumes RankExpr ranker, it is also set by this function.
func (q *Search) SetRankingExpression(rankexpr string) {
	q.rankexpr = rankexpr
	if q.ranker != RankExpr && q.ranker != RankExport {
		q.SetRankingMode(RankExpr)
	}
}

// SetRankingMode assigns ranking mode and also adjust MatchMode to MatchExtended2 (since otherwise rankers are useless)
func (q *Search) SetRankingMode(ranker ERankMode) {
	q.ranker = ranker
	if q.MatchMode != MatchExtended && q.MatchMode != MatchExtended2 {
		q.MatchMode = MatchExtended2
	}
}

// SetSortMode sets matches sorting mode
//
// `sort` determines sorting mode.
//
// `sortby` determines attribute or expression used for sorting.
//
// If `sortby` set in Search query is empty (it is not necessary set in this very call, it might be set earlier!), then `sort`
// is explicitly set as SortRelevance
func (q *Search) SetSortMode(sort ESortOrder, sortby ...string) {
	q.sort = sort
	if len(sortby) > 0 {
		q.sortby = sortby[0]
	}
	if q.sortby == "" {
		q.sort = SortRelevance
	}
}

/*
SetTokenFilter setups UDF token filter

`library` is the name of plugin library, as "mylib.so"

`name` is the name of token filtering function in the library, as "email_process"

`opts` is string parameters which passed to udf filter, like "field=email;split=.io". Format of the options determined by UDF plugin.
 */
func (q *Search) SetTokenFilter(library, name string, opts string) {
	q.tokenFlibrary = library
	q.tokenFname = name
	q.tokenFopts = opts
}









// iOStats is internal structure, used only in master-agent communication
type iOStats struct {
	ReadTime, ReadBytes, WriteTime, WriteBytes int64
	ReadOps, WriteOps                          uint32
}

// ColumnInfo represents one attribute column in resultset schema
type ColumnInfo struct {
	Name string    // name of the attribute
	Type EAttrType // type of the attribute
}


// Match represents one match (document) in result schema
type Match struct {
	DocID  DocID			// key Document ID
	Weight int 				// weight of the match
	Attrs  []interface{} 	// optional array of attributes, quantity and types depends from schema
}

// Stringer interface for Match type
func (vl Match) String() (line string) {
	line = fmt.Sprintf("Doc: %v, Weight: %v, attrs: %v", vl.DocID, vl.Weight, vl.Attrs)
	return
}

// JsonOrStr is typed string with explicit flag whether it is 'just a string', or json document. It may be used, say,
// to either escape plain strings when appending to JSON structure, either add it 'as is' assuming it is alreayd json.
// Such values came from daemon as attribute values for PQ indexes.
type JsonOrStr struct {
	IsJson bool   // true, if Val is JSON document; false if it is just a plain string
	Val    string // value (string or JSON document)
}

// Stringer interface for JsonOrStr type. Just append ' (json)' suffix, if IsJson is true.
func (vl JsonOrStr) String() string {
	if vl.IsJson {
		return fmt.Sprintf("%s (json)", vl.Val)
	} else {
		return vl.Val
	}
}

// WordStat describes statistic for one word in QueryResult. That is, word, num of docs and num of hits.
type WordStat struct {
	Word       string
	Docs, Hits int
}

// Stringer interface for WordStat type
func (vl WordStat) String() string {
	return fmt.Sprintf("'%s' (Docs:%d, Hits:%d)", vl.Word, vl.Docs, vl.Hits)
}

// QueryResult represents resultset from successful Query/RunQuery, or one of resultsets from RunQueries call.
type QueryResult struct {
	Error, Warning    string			// messages (if any)
	Status            ESearchdstatus	// status code for current resultset
	Fields            []string			// fields of the schema
	Attrs             []ColumnInfo		// attributes of the schema
	Id64              bool				// if DocumentID is 64-bit (always true)
	Matches           []Match			// set of matches according to schema
	Total, TotalFound int				// num of matches and total num of matches found
	QueryTime         time.Duration		// query duration
	WordStats         []WordStat		// words statistic
}

// Stringer interface for EAttrType type
func (vl EAttrType) String() string {
	switch vl {
	case AttrNone:
		return "none"
	case AttrInteger:
		return "int"
	case AttrTimestamp:
		return "timestamp"
	case AttrBool:
		return "bool"
	case AttrFloat:
		return "float"
	case AttrBigint:
		return "bigint"
	case AttrString:
		return "string"
	case AttrPoly2d:
		return "poly2d"
	case AttrStringptr:
		return "stringptr"
	case AttrTokencount:
		return "tokencount"
	case AttrJson:
		return "json"
	case AttrUint32set:
		return "uint32Set"
	case AttrInt64set:
		return "int64Set"
	case AttrMaparg:
		return "maparg"
	case AttrFactors:
		return "factors"
	case AttrJsonField:
		return "jsonField"
	case AttrFactorsJson:
		return "factorJson"
	default:
		return fmt.Sprintf("unknown(%d)", uint32(vl))
	}
}

// Stringer interface for ColumnInfo type
func (res ColumnInfo) String() string {
	return fmt.Sprintf("%s: %v", res.Name, res.Type)
}

// Stringer interface for QueryResult type
func (res QueryResult) String() string {
	line := fmt.Sprintf("Status: %v\n", res.Status)
	if res.Status==StatusError {
		line += fmt.Sprintf("Error: %v\n", res.Error)
	}
	line += fmt.Sprintf("Query time: %v\n", res.QueryTime)
	line += fmt.Sprintf("Total: %v\n", res.Total)
	line += fmt.Sprintf("Total found: %v\n", res.TotalFound)
	line += fmt.Sprintln("Schema:")
	if len(res.Fields) != 0 {
		line += fmt.Sprintln("\tFields:")
		for i := 0; i < len(res.Fields); i++ {
			line += fmt.Sprintf("\t\t%v\n", res.Fields[i])
		}
	}
	if len(res.Attrs) != 0 {
		line += fmt.Sprintln("\tAttributes:")
		for i := 0; i < len(res.Attrs); i++ {
			line += fmt.Sprintf("\t\t%v\n", res.Attrs[i])
		}
	}

	if len(res.Matches) != 0 {
		line += fmt.Sprintln("Matches:")
		for i := 0; i < len(res.Matches); i++ {
			line += fmt.Sprintf("\t%v\n", &res.Matches[i])
		}
	}
	if len(res.WordStats) != 0 {
		line += fmt.Sprintln("Word stats:")
		for i := 0; i < len(res.WordStats); i++ {
			line += fmt.Sprintf("\t%v\n", res.WordStats[i])
		}
	}
	return line
}

func (buf *apibuf) buildSearchRequest(q *Search) {

	buf.putDword(uint32(q.queryflags))
	buf.putInt(q.Offset)
	buf.putInt(q.Limit)

	buf.putDword(uint32(q.MatchMode))
	buf.putDword(uint32(q.ranker))
	if q.ranker == RankExport || q.ranker == RankExpr {
		buf.putString(q.rankexpr)
	}

	buf.putInt(int32(q.sort))
	buf.putString(q.sortby)

	buf.putString(q.Query)
	buf.putInt(0)
	buf.putString(q.Indexes)
	buf.putInt(1)

	// default full id range (any Client range must be in filters at this stage)
	buf.putDocid(0)
	buf.putDocid(DocidMax)

	docidfilter := 0
	if (q.IDMin != 0) || (q.IDMax != DocidMax && q.IDMax != 0) {
		docidfilter = 1
	}

	buf.putLen(len(q.filters) + docidfilter) // N of filters
	// filters goes here
	for _, filter := range q.filters {
		buf.putString(filter.Attribute)
		buf.putDword(uint32(filter.FilterType))
		switch filter.FilterType {
		case FilterString:
			buf.putString(filter.FilterData.(string))
		case FilterUservar:
			buf.putString(filter.FilterData.(string))
		case FilterNull:
			buf.putBoolByte(filter.FilterData.(bool))
		case FilterRange:
			foo := filter.FilterData.([]int64)
			buf.putInt64(foo[0])
			buf.putInt64(foo[1])
		case FilterFloatrange:
			foo := filter.FilterData.([]float32)
			buf.putFloat(foo[0])
			buf.putFloat(foo[1])
		case FilterValues:
			foo := filter.FilterData.([]int64)
			buf.putLen(len(foo))
			for _, value := range foo {
				buf.putInt64(value)
			}
		case FilterStringList:
			foo := filter.FilterData.([]string)
			buf.putLen(len(foo))
			for _, value := range foo {
				buf.putString(value)
			}
		}
		buf.putBoolDword(filter.Exclude)
	}

	// docid filter, if any, we put as the last one
	if docidfilter == 1 {
		buf.putString("@id")
		buf.putDword(uint32(FilterRange))
		buf.putDocid(q.IDMin)
		buf.putDocid(q.IDMax)
		buf.putBoolDword(false)
	}

	buf.putDword(uint32(q.Groupfunc))
	buf.putString(q.GroupBy)
	buf.putInt(q.MaxMatches)

	buf.putString(q.GroupSort)
	buf.putInt(q.CutOff)

	buf.putInt(q.RetryCount)
	buf.putDuration(q.RetryDelay)

	buf.putString(q.GroupDistinct)

	buf.putBoolDword(q.hasGeoAnchor())
	if q.hasGeoAnchor() {
		buf.putString(q.geoLatAttr)
		buf.putString(q.geoLonAttr)
		buf.putFloat(q.geoLatitude)
		buf.putFloat(q.geoLongitude)
	}

	buf.putLen(len(q.IndexWeights))
	for idx, iw := range q.IndexWeights {
		buf.putString(idx)
		buf.putInt(iw)
	}

	buf.putDuration(q.MaxQueryTime)
	buf.putLen(len(q.FieldWeights))
	for idx, iw := range q.FieldWeights {
		buf.putString(idx)
		buf.putInt(iw)
	}

	buf.putString(q.Comment)
	buf.putInt(0) // N of overrides

	buf.putString(q.SelectClause)

	if q.hasSetQueryFlag(QflagMaxPredictedTime) {
		buf.putDuration(q.predictedTime)
	}

	buf.putString(q.outerorderby)
	buf.putInt(q.outeroffset)
	buf.putInt(q.outerlimit)
	buf.putBoolDword(q.hasouter)
	buf.putString(q.tokenFlibrary)
	buf.putString(q.tokenFname)
	buf.putString(q.tokenFopts)

	buf.putInt(0) // N of filter tree elems
}

func (result *QueryResult) makeError(erstr string) error {
	result.Error = erstr
	if erstr == "" {
		return nil
	}
	return errors.New(erstr)
}

func (result *QueryResult) parseResult(req *apibuf) error {

	// extract status
	result.Status = ESearchdstatus(req.getDword())
	switch result.Status {
	case StatusError:
		return result.makeError(req.getString())
	case StatusRetry:
		return result.makeError(req.getString())
	case StatusWarning:
		result.Warning = req.getString()
	}

	result.parseSchema(req)
	nmatches := req.getInt()
	result.Id64 = req.getIntBool()

	// parse matches
	result.Matches = make([]Match, nmatches)
	for j := 0; j < nmatches; j++ {
		result.parseMatch(&result.Matches[j], req)
	}

	// read totals (retrieved count, total count, query time, word count)
	result.Total = req.getInt()
	result.TotalFound = req.getInt()
	result.QueryTime = time.Millisecond * time.Duration(req.getInt())

	nwords := req.getInt()
	result.WordStats = make([]WordStat, nwords)

	// read per-word stats
	for i := 0; i < nwords; i++ {
		result.WordStats[i].Word = req.getString()
		result.WordStats[i].Docs = req.getInt()
		result.WordStats[i].Hits = req.getInt()
	}
	return result.makeError("")
}

func (result *QueryResult) parseSchema(req *apibuf) {

	// read Fields
	nfields := req.getInt()
	result.Fields = make([]string, nfields)
	for j := 0; j < nfields; j++ {
		result.Fields[j] = req.getString()
	}

	// read attributes
	nattrs := req.getInt()
	result.Attrs = make([]ColumnInfo, nattrs)
	for j := 0; j < nattrs; j++ {
		result.Attrs[j].Name = req.getString()
		result.Attrs[j].Type = EAttrType(req.getDword())
	}
}

func (result *QueryResult) parseMatch(match *Match, req *apibuf) {

	// docid
	if result.Id64 {
		match.DocID = req.getDocid()
	} else {
		match.DocID = DocID(req.getDword())
	}

	// weight
	match.Weight = req.getInt()
	match.Attrs = make([]interface{}, len(result.Attrs))

	// attributes
	for i, item := range result.Attrs {
		switch item.Type {

		case AttrUint32set:
			iValues := int(req.getDword())
			values := make([]uint32, iValues)
			for j := 0; j < iValues; j++ {
				values[j] = req.getDword()
			}
			match.Attrs[i] = values

		case AttrInt64set:
			iValues := int(req.getDword())
			values := make([]uint64, iValues)
			for j := 0; j < iValues; j++ {
				values[j] = req.getUint64()
			}
			match.Attrs[i] = values

		case AttrFloat:
			match.Attrs[i] = req.getFloat()

		case AttrBigint:
			match.Attrs[i] = req.getUint64()

		case AttrStringptr:
		case AttrString:
			foo := req.getRefBytes()
			ln := len(foo)
			var res JsonOrStr
			if foo[ln-2] == 0 { // this is typed
				res.IsJson = foo[ln-1] == 0
				res.Val = string(foo[:ln-2])
			} else {
				res.Val = string(foo)
			}
			match.Attrs[i] = res

		case AttrJson:
		case AttrFactors:
		case AttrFactorsJson:
			match.Attrs[i] = req.getBytes()

		case AttrJsonField:
			var bs bsonField
			bs.etype = eBsonType(req.getByte())
			if bs.etype != bsonEof {
				bs.blob = req.getBytes()
			}
			match.Attrs[i] = bs

		case AttrTimestamp:
			foo := req.getDword()
			match.Attrs[i] = time.Unix(int64(foo), 0)

		default:
			match.Attrs[i] = req.getDword()
		}
	}
}

func buildSearchRequest(queries []Search) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putUint(0) // that is cl!
		buf.putLen(len(queries))
		for j := 0; j < len(queries); j++ {
			buf.buildSearchRequest(&queries[j])
		}
	}
}

func parseSearchAnswer(nreqs int) func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		resp := make([]QueryResult, nreqs)
		for j := 0; j < nreqs; j++ {
			_ = resp[j].parseResult(answer)
		}
		return resp
	}
}
