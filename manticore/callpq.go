package manticore

import (
	"errors"
	"fmt"
	"time"
)

/*
Pqflags determines boolean parameter flags for CallQP options
This flags are unified into one bitfield used instead of bunch of separate flags.

There are the following flags for CallPQ modes available:

NeedDocs

NeedDocs require to provide numbers of matched documents. It is either order numbers from the set of provided documents,
or DocIDs, if documents are JSON and you pointed necessary field which contains DocID. (NOTE: json PQ calls are not yet
implemented via API, it will be done later).

NeedQuery

NeedQuery require to return not only QueryID of the matched queries, but also another information about them.
It may include query itself, tags and filters.

Verbose

Verbose, require to return additional meta-information about matching and queries. It causes daemon to fill fields TmSetup, TmTotal,
QueriesFailed, EarlyOutQueries and QueryDT of SearchPqResponse structure.

SkipBadJson

SkipBadJson, require to not fail on bad (ill-formed) jsons, but warn and continue processing. This flag works only for
bson queries and useless for plain text (may even cause warning if provided there).
*/
type Pqflags uint32

const (
	NeedDocs Pqflags = (1 << iota)
	NeedQuery
	jsonDocs
	Verbose
	SkipBadJson
)

/*
SearchPqOptions incapsulates params to be passed to CallPq function.

Flags

Flags is instance of Pqflags, different bites described there.

IdAlias

IdAlias determines name of the field in supplied json documents, which contain DocumentID. If NeedDocs flag is set,
this value will be used in resultset to identify documents instead of just plain numbers of them.

Shift

Shift is used if daemon returns order number of the documents (i.e. when NeedDoc flag is set, but no IdAlias provided,
or if documents are just plain texts and can't contain such field at all). Shift then is just added to every number of
the doc, helping move the whole range. Say, if you provide 2 documents, they may be returned as numbers 1 and 2.
Buf if you also give Shift=100, they will became 101 and 102. It may help if you distribute bit docset over several
instances and want to keep the numbers. Daemon itself uses this value for the same purpose.
*/
type SearchPqOptions = struct {
	Flags   Pqflags
	IdAlias string
	Shift   int32
}

/*
NewSearchPqOptions creates empty instance of search options. Prefer to use this function when you need options,
since it may set necessary defaults
*/
func NewSearchPqOptions() SearchPqOptions {
	return SearchPqOptions{jsonDocs, "", 0}
}

func buildCallpqRequest(index string, values []string, opts SearchPqOptions) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putDword(uint32(opts.Flags))
		buf.putString(opts.IdAlias)
		buf.putString(index)
		buf.putInt(opts.Shift)
		buf.putLen(len(values))
		for i := 0; i < len(values); i++ {
			buf.putString(values[i])
		}
	}
}

/*
PqResponseFlags determines boolean flags came in SearchPqResponse result
These flags are unified into one bitfield used instead of bunch of separate flags.

There are following bits available:

HasDocs

HasDocs indicates that each QueryDesc of Queries result array have array of documents in Docs field. Otherwise this field
there is nil.

DumpQueries

DumpQueries indicates that each query contains additional info, like query itself, tags and filters. Otherwise it have only
the number - QueryID and nothing more.

HasDocids

HasDocids, came in pair with HasDocs, indicates that array of documents in Queries[]Docs field is array of
uint64 with document ids, provided in documents of original query. Otherwise it is array of int32 with order
numbers, may be shifted by Shift param.
*/
type PqResponseFlags uint32

const (
	HasDocs PqResponseFlags = (1 << iota)
	DumpQueries
	HasDocids
)

/*
SearchPqResponse represents whole response to CallPQ and CallPQBson calls
*/
type SearchPqResponse = struct {
	Flags           PqResponseFlags
	TmTotal         time.Duration // total time spent for matching the document(s)
	TmSetup         time.Duration // time spent to initial setup of matching process - parsing docs, setting options, etc.
	QueriesMatched  int           // how many stored queries match the document(s)
	QueriesFailed   int           // number of failed queries
	DocsMatched     int           // how many times the documents match the queries stored in the index
	TotalQueries    int           // how many queries are stored in the index at all
	OnlyTerms       int           // how many queries in the index have terms. The rest of the queries have extended query syntax
	EarlyOutQueries int           // num of queries which wasn’t fall into full routine, but quickly matched and rejected with filters or other conditions
	QueryDT         []int         // detailed times per each query
	Warnings        string
	Queries         []QueryDesc // queries themselve. See QueryDesc structure for details
}

//func (r *SearchPqResponse) String() string {
//	return ""
//}

/*
QueryDescFlags is bitfield describing internals of PqQuery struct
This flags are unified into one bitfield used instead of bunch of separate flags.

There are following bits available:

QueryPresent

QueryPresent indicates that field Query is valid. Otherwise it is not touched ("" by default)

TagsPresent

TagsPresent indicates that field Tags is valid. Otherwise it is not touched ("" by default)

FiltersPresent

FiltersPresent indicates that field Filters is valid. Otherwise it is not touched ("" by default)

QueryIsQl

QueryIsQl indicates that field Query (if present) is query in sphinxql syntax. Otherwise it is query in json syntax.
PQ index can store indexes in both format, and this flag in resultset helps you to distinguish them (both are
text, but syntax m.b. different)
*/
type QueryDescFlags uint32

const (
	QueryPresent QueryDescFlags = (1 << iota)
	TagsPresent
	FiltersPresent
	QueryIsQl
)

/*
QueryDesc represents an elem of Queries array from SearchPqResponse and describe one returned stored query.

QueryID

QueryID is namely, Query ID. In most minimal query it is the only returned field.

Docs

Docs is filled only if flag HasDocs is set, and contains either array of DocID (which are uint64) - if flag HasDocids is set,
either array of doc ordinals (which are int32), if flag HasDocids is NOT set.

Query

Query is query meta, in addition to QueryID. It is filled only if in the query options they were requested via
bit NeedQuery, and may contain query string, tags and filters.
*/
type QueryDesc = struct {
	QueryID uint64
	Docs    interface{}
	Query   PqQuery
}

/*
PqQuery describes one separate query info from resultset of CallPQ/CallPQBson

Flags determines type of the Query, and also whether other fields of the struct are filled or not.

Query, Tags, Filters - attributes saved with query, all are optional
*/
type PqQuery = struct {
	Flags   QueryDescFlags
	Query   string
	Tags    string
	Filters string
}

func parseCallpqAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		var rs SearchPqResponse
		rs.Flags = PqResponseFlags(answer.getDword())
		hasDocs := rs.Flags&HasDocs != 0
		hasDocids := rs.Flags&HasDocids != 0
		dumpQueries := rs.Flags&DumpQueries != 0
		nqueries := int(answer.getDword())
		rs.Queries = make([]QueryDesc, nqueries)
		for i := 0; i < nqueries; i++ {
			rs.Queries[i].QueryID = answer.getUint64()
			if hasDocs {
				ndocs := int(answer.getDword())
				if hasDocids {
					docids := make([]uint64, ndocs)
					for j := 0; j < ndocs; j++ {
						docids[j] = answer.getUint64()
					}
					rs.Queries[i].Docs = docids
				} else {
					docids := make([]int32, ndocs)
					for j := 0; j < ndocs; j++ {
						docids[j] = int32(answer.getInt())
					}
					rs.Queries[i].Docs = docids
				}
			}
			if dumpQueries {
				flags := QueryDescFlags(answer.getDword())
				rs.Queries[i].Query.Flags = flags
				if flags&QueryPresent != 0 {
					rs.Queries[i].Query.Query = answer.getString()
				}
				if flags&TagsPresent != 0 {
					rs.Queries[i].Query.Tags = answer.getString()
				}
				if flags&FiltersPresent != 0 {
					rs.Queries[i].Query.Filters = answer.getString()
				}
			}
		}
		rs.TmTotal = time.Microsecond * time.Duration(answer.getUint64())
		rs.TmSetup = time.Microsecond * time.Duration(answer.getUint64())
		rs.QueriesMatched = answer.getInt()
		rs.QueriesFailed = answer.getInt()
		rs.DocsMatched = answer.getInt()
		rs.TotalQueries = answer.getInt()
		rs.OnlyTerms = answer.getInt()
		rs.EarlyOutQueries = answer.getInt()
		dts := answer.getInt()
		if dts != 0 {
			rs.QueryDT = make([]int, dts)
			for i := 0; i < dts; i++ {
				rs.QueryDT[i] = answer.getInt()
			}
		}
		rs.Warnings = answer.getString()
		return &rs
	}
}

/*
CallQP perform check if a document matches any of the predefined criterias (queries)
It returns list of matched queries and may be additional info as matching clause, filters, and tags.

`index` determines name of PQ index you want to call into. It can be either local, either distributed
built from several PQ agents

`values` is the list of the index. Each value regarded as separate index. Order num of matched indexes then may
be returned in resultset

`opts` packed options. See description of SearchPqOptions for details.
In general you need to make instance of options by calling NewSearchPqOptions(), set
desired flags and options, and then invoke CallPQ, providing desired index, set of documents and the options.

Since this function expects plain text documents, it will remove all flags about json from the options, and also will
not use IdAlias, if any provided.

For example:
  ..
  po := NewSearchPqOptions()
  po.Flags = NeedDocs | Verbose | NeedQuery
  resp, err := cl.CallPQ("pq",[]string{"angry test","filter test doc2",},po)
  ...
*/
func (cl *Client) CallPQ(index string, values []string, opts SearchPqOptions) (*SearchPqResponse, error) {
	opts.Flags &^= jsonDocs
	opts.Flags &^= SkipBadJson
	opts.IdAlias = ""
	ans, err := cl.netQuery(commandCallpq,
		buildCallpqRequest(index, values, opts),
		parseCallpqAnswer())

	if ans == nil {
		return nil, err
	}
	return ans.(*SearchPqResponse), err
}

/*
CallPQBson perform check if a document matches any of the predefined criterias (queries)
It returns list of matched queries and may be additional info as matching clause, filters, and tags.

It works very like CallPQ, but expects documents in BSON form. With this function it is have sense to use
flags as SkipBadJson, and param IdAlias which are not used for plain queries.

This function is not yet implemented in SDK, it is stub.
*/
func (cl *Client) CallPQBson(index string, values []byte, opts SearchPqOptions) (*SearchPqResponse, error) {
	return nil, errors.New(fmt.Sprintln("Not yet implemented"))

}
