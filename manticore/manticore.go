// Copyright (c) 2001-2016, Andrew Aksyonoff
// Copyright (c) 2008-2016, Sphinx Technologies Inc
// Copyright (c) 2019, Manticore Software LTD (http://manticoresearch.com)
// All rights reserved
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU Library General Public License. You should
// have received a copy of the LGPL license along with this program; if you
// did not, you can find it at http://www.gnu.org/

// Package manticore implements Client to work with manticoresearch over it's internal binary protocol.
// Also in many cases it may be used to work with sphinxsearch daemon as well.
// It implements Client connector which may be used as
//    cl := NewClient()
//    res, err := cl.Query("hello")
//    ...
// Set of functions is mostly imitates API description of Manticoresearch for PHP, but with few
// changes which are specific to Go language as more effective and mainstream for that language
package manticore

import (
	"errors"
	"time"
)


// BuildExcerpts generates excerpts (snippets)
// of given documents for given query. returns nil on failure,
// an array of snippets on success. If necessary it will connect to the server before processing.
//
// `docs` is a plain slice of strings that carry the documents’ contents.
//
// `index` is an index name string. Different settings (such as charset, morphology, wordforms)
// from given index will be used.
//
// `words` is a string that contains the keywords to highlight.
// They will be processed with respect to index settings. For instance, if English stemming is
// enabled in the index, shoes will be highlighted even if keyword is shoe. Keywords can contain wildcards,
// that work similarly to star-syntax available in queries.
//
// `opts` is an optional struct SnippetOptions which may contain
// additional optional highlighting parameters, it may be created by calling of ``NewSnippetOptions()'' and then tuned
// for your needs. If `opts` is omitted, default will be used.
//
// Snippets extraction algorithm currently favors better passages (with closer phrase matches),
// and then passages with keywords not yet in snippet. Generally, it will try to highlight the best match
// with the query, and it will also to highlight all the query keywords, as made possible by the limits.
// In case the document does not match the query, beginning of the document trimmed down according to the limits
// will be return by default. You can also return an empty snippet instead case by setting allow_empty option to true.
//
// Returns false on failure. Returns a plain array of strings with excerpts (snippets) on success.
func (cl *Client) BuildExcerpts(docs []string, index,
	words string, opts ...SnippetOptions) ([]string, error) {

	var popts *SnippetOptions
	if len(opts) > 0 {
		popts = &opts[0]
	} else {
		popts = NewSnippetOptions()
	}

	if len(docs) == 0 {
		return nil, errors.New("invalid arguments (docs must not be empty)")
	}

	if index == "" {
		return nil, errors.New("invalid arguments (index must not be empty)")
	}

	if words == "" {
		return nil, errors.New("invalid arguments (words must not be empty)")
	}

	ndocs := len(docs)
	snippets, err := cl.netQuery(commandExcerpt,
		buildSnippetRequest(popts, docs, index, words),
		parseSnippetAnswer(ndocs))
	if snippets==nil {
		return nil, err
	}
	return snippets.([]string), err
}

// BuildKeywords extracts keywords from query using tokenizer settings
// for given index, optionally with per-keyword occurrence statistics.
// Returns an array of hashes with per-keyword information. If necessary it will connect to the server before processing.
//
// `query` is a query to extract keywords from.
//
// `index` is a name of the index to get tokenizing settings and keyword
// occurrence statistics from.
//
// `hits` is a boolean flag that indicates whether keyword occurrence statistics are required.
func (cl *Client) BuildKeywords(query, index string, hits bool) ([]Keyword, error) {

	if query == "" {
		return nil, errors.New("invalid arguments (query must not be empty)")
	}

	if index == "" {
		return nil, errors.New("invalid arguments (index must not be empty)")
	}

	keywords, err := cl.netQuery(commandKeywords,
		buildKeywordsRequest(query, index, hits),
		parseKeywordsAnswer(hits))
	if keywords==nil {
		return nil, err
	}
	return keywords.([]Keyword), err
}

// Close closes previously opened persistent connection. If no connection active, it fire error 'not connected' which
// is just informational and safe to ignore.
func (cl *Client) Close() (bool, error) {
	if !cl.connected {
		return false, errors.New("not connected")
	}
	err := cl.conn.Close()
	cl.conn = nil
	cl.connected = false
	return err == nil, err
}

// FlushAttributes forces searchd to flush pending attribute updates to disk, and blocks until completion.
// Returns a non-negative internal flush tag on success, or -1 and error.
//
// Attribute values updated using UpdateAttributes() API call are kept in a memory mapped file.
// Which means the OS decides when the updates are actually written to disk. FlushAttributes() call lets you enforce
// a flush, which writes all the changes to disk. The call will block until searchd finishes writing the data to disk,
// which might take seconds or even minutes depending on the total data size (.spa file size).
// All the currently updated indexes will be flushed.
//
// Flush tag should be treated as an ever growing magic number that does not mean anything.
// It’s guaranteed to be non-negative. It is guaranteed to grow over time, though not necessarily in a sequential
// fashion; for instance, two calls that return 10 and then 1000 respectively are a valid situation.
// If two calls to FlushAttrs() return the same tag, it means that there were no actual attribute updates in between
// them, and therefore current flushed state remained the same (for all indexes).
//
// Usage example:
//
//  status, err := cl.FlushAttributes ()
//  if err!=nil {
//    fmt.Println(err.Error())
//  }
func (cl *Client) FlushAttributes() (int, error) {
	tag, err := cl.netQuery(commandFlushattrs, nil, parseDwordAnswer())
	if tag==nil{
		return -1, err
	}
	return tag.(int), err
}

// GetLastWarning returns last warning message, as a string, in human readable format.
// If there were no warnings during the previous API call, empty string is returned.
//
// You should call it to verify whether your request (such as Query()) was completed but with warnings.
// For instance, search query against a distributed index might complete successfully even if several remote agents
// timed out. In that case, a warning message would be produced.
//
// The warning message is not reset by this call; so you can safely call it several times if needed.
// If you issued multi-query by running RunQueries(), individual warnings will not be written in client; instead
// check the Warning field in each returned result of the slice.
func (cl *Client) GetLastWarning() string {
	return cl.lastWarning
}

// IsConnectError checks whether the last error was a network error on API side, or a remote error reported by searchd.
// Returns true if the last connection attempt to searchd failed on API side, false otherwise
// (if the error was remote, or there were no connection attempts at all).
func (cl *Client) IsConnectError() bool {
	return cl.connError
}

// Open opens persistent connection to the server.
func (cl *Client) Open() (bool, error) {

	if cl.connected {
		return false, errors.New("already connected")
	}
	_, err := cl.netQuery(commandPersist, buildBoolRequest(true), nil)
	return err == nil, err
}

// Query connects to searchd server, run given simple search query string through given indexes,
// and return the search result.
//
// This is simplified function which accepts only 1 query string parameter and no options
// Internally it will run with ranker 'RankProximityBm25', mode 'MatchAll' with 'max_matches=1000' and 'limit=20'
// It is good to be used in kind of a demo run. If you want more fine-tuned options, consider to use `RunQuery()`
// and `RunQueries()` functions which provide you full spectre of possible tuning options.
//
// `query` is a query string.
//
// `indexes` is an index name (or names) string. Default value for `indexes` is "*" that means to query all local indexes.
// Characters allowed in index names include Latin letters (a-z), numbers (0-9) and underscore (_);
// everything else is considered a separator. Note that index name should not start with underscore character.
// Internally 'Query' is just invokes 'RunQuery' with default Search, where only `query` and `index` fields are customized.
//
// Therefore, all of the following samples calls are valid and will search the same two indexes:
//  cl.Query ( "test query", "main delta" )
//  cl.Query ( "test query", "main;delta" )
//  cl.Query ( "test query", "main, delta" )
func (cl *Client) Query(query string, indexes ...string) (*QueryResult, error) {
	index := "*"

	if len(indexes) > 0 {
		index = indexes[0]
	}

	res, err := cl.RunQuery(NewSearch(query, index, ""))

	if res==nil {
		return nil, err
	}

	if err == nil && res.Status != StatusError {
		return res, nil
	}
	return nil, err
}

// RunQueries connects to searchd, runs a batch of queries, obtains and returns the result sets.
// Returns nil and error message on general error (such as network I/O failure).
// Returns a slice of result sets on success.
//
// `queries` is slice of Search structures, each represent one query. You need to prepare this slice yourself before call.
//
// Each result set in the returned array is exactly the same as the result set returned from RunQuery.
//
// Note that the batch query request itself almost always succeeds - unless there’s a network error,
// blocking index rotation in progress, or another general failure which prevents the whole request
// from being processed.
//
// However individual queries within the batch might very well fail. In this case their respective
// result sets will contain non-empty `error` message, but no matches or query statistics.
// In the extreme case all queries within the batch could fail. There still will be no general error reported,
// because API was able to successfully connect to searchd, submit the batch, and receive the results -
// but every result set will have a specific error message.
func (cl *Client) RunQueries(queries []Search) ([]QueryResult, error) {
	nreqs := len(queries)
	if nreqs == 0 {
		return nil, errors.New("no queries defined, issue AddQuery() first")
	}

	res, err := cl.netQuery(commandSearch,
		buildSearchRequest(queries),
		parseSearchAnswer(nreqs))
	if res == nil {
		return nil, err
	}
	return res.([]QueryResult), err
}

// RunQuery connects to searchd, runs a query, obtains and returns the result set.
// Returns nil and error message on general error (such as network I/O failure).
// Returns a result set on success.
//
// `query` is a single Search structure, representing the query. You need to prepare it yourself before call.
//
// Each result set in the returned array is exactly the same as the result set returned from RunQuery.
//
func (cl *Client) RunQuery(query Search) (*QueryResult, error) {
	res, err := cl.netQuery(commandSearch,
		buildSearchRequest([]Search{query}),
		parseSearchAnswer(1))
	if res==nil {
		return nil, err
	}
	result := res.([]QueryResult)[0]
	cl.lastWarning = result.Warning
	return &result, err
}

// SetConnectTimeout sets the time allowed to spend connecting to the server before giving up.
//
// Under some circumstances, the server can be delayed in responding, either due to network delays, or a query backlog.
// In either instance, this allows the client application programmer some degree of control over how their program
// interacts with searchd when not available, and can ensure that the client application does not fail due to exceeding
// the execution limits.
//
// In the event of a failure to connect, an appropriate error code should be returned back to the application
// in order for application-level error handling to advise the user.
func (cl *Client) SetConnectTimeout(timeout time.Duration) {
	cl.timeout = timeout
}

// SetMaxAlloc limits size of client's network buffer. For sending queries and receiving results client reuses byte array,
// which can grow up to required size. If the limit reached, array will be released and new one will be created. Usually
// API needs just few kilobytes of the memory, but sometimes the value may grow significantly high. For example, if you fetch a
// big resultset with many attributes. Such resultset will be properly received and processed, however at the next query
// backend array which used for it will be released, and occupied memory will be returned to runtime.
//
// `alloc` is size, in bytes. Reasonable default value is 8M.
func (cl *Client) SetMaxAlloc(alloc int) {
	cl.maxAlloc = alloc
}

// SetServer sets searchd host name and TCP port. All subsequent requests will use the new host and port settings.
// Default host and port are ‘localhost’ and 9312, respectively.
//
// `host` is either url (hostname or ip address), either unix socket path (starting with '/')
//
// `port` is optional, it has sense only for tcp connections and not used for unix socket. Default is 9312
func (cl *Client) SetServer(host string, port ...uint16) {

	if host[0] == '/' {
		cl.dialmethod = "unix"
		cl.host = host
		cl.port = 0
		return
	}

	if host[:7] == "unix://" {
		cl.dialmethod = "unix"
		cl.host = host[7:]
		cl.port = 0
		return
	}

	cl.host = host
	cl.dialmethod = "tcp"
	if len(port) > 0 {
		cl.port = port[0]
	}
}

func (cl *Client) Sphinxql(cmd string) ([]sqlresult, error) {
	blob, err := cl.netQuery(commandSphinxql,
		buildSphinxqlRequest(cmd),
		parseSphinxqlAnswer())
	if blob==nil{
		return nil, err
	}
	return blob.([]sqlresult), err
}

func (cl *Client) Ping(cookie uint32) (uint32, error) {
	answer, err := cl.netQuery(commandPing,
		buildDwordRequest(cookie),
		parseDwordAnswer())
	if answer==nil{
		return 0, err
	}
	return answer.(uint32), err
}

// Status queries searchd status, and returns an array of status variable name and value pairs.
//
// `global` determines whether you take global status, or meta of the last query.
//  true: receive global daemon status
//  false: receive meta of the last executed query
//
// Usage example:
//  status, err := cl.Status(false)
//	if err != nil {
//		fmt.Println(err.Error())
//	} else {
//		for key, line := range (status) {
//			fmt.Printf("%v:\t%v\n", key, line)
//		}
//	}
// example output:
//  time:	0.000
//  keyword[0]:	query
//  docs[0]:	1235
//  hits[0]:	1474
//  total:	3
//  total_found:	3
func (cl *Client) Status(global bool) (map[string]string, error) {
	status, err := cl.netQuery(commandStatus,
		buildBoolRequest(global),
		parseStatusAnswer())
	if status==nil {
		return nil, err
	}
	return status.(map[string]string), err
}

// UpdateAttributes instantly updates given attribute values in given documents. Returns number of actually updated
// documents (0 or more) on success, or -1 on failure with error.
//
// `index` is a name of the index (or indexes) to be updated. It can be either a single index name or a list,
// like in Query(). Unlike Query(), wildcard is not allowed and all the indexes to update must be specified explicitly.
// The list of indexes can include distributed index names. Updates on distributed indexes will be pushed to all agents.
//
// `attrs` is a slice with string attribute names, listing attributes that are updated.
//
// `values` is a map with documents IDs as keys and new attribute values, see below.
//
// `vtype` type parameter, see EUpdateType description for values.
//
// `ignorenonexistent` points that the update will silently ignore any warnings about trying to update a column which
// is not exists in current index schema.
//
// Usage example:
//
// 	upd, err := cl.UpdateAttributes("test1", []string{"group_id"}, map[DocID][]interface{}{1:{456}}, UpdateInt, false)
//
// Here we update document 1 in index test1, setting group_id to 456.
//
//  upd, err := cl.UpdateAttributes("products", []string{"price", "amount_in_stock"}, map[DocID][]interface{}{1001:{123,5}, 1002:{37,11}, 1003:{25,129}}, UpdateInt, false)
//
// Here we update documents 1001, 1002 and 1003 in index products.
// For document 1001, the new price will be set to 123 and the new amount in stock to 5;
// for document 1002, the new price will be 37 and the new amount will be 11; etc.
func (cl *Client) UpdateAttributes(index string, attrs []string, values map[DocID][]interface{},
	vtype EUpdateType, ignorenonexistent bool) (int, error) {

	if attrs == nil || len(attrs) == 0 {
		return -1, errors.New("invalid arguments (attrs must not empty)")
	}

	if index == "" {
		return -1, errors.New("invalid arguments (index must not be empty)")
	}

	if values == nil || len(values) == 0 {
		return -1, errors.New("invalid arguments (values must not be empty)")
	}

	updated, err := cl.netQuery(commandUpdate,
		buildUpdateRequest(index, attrs, values, vtype, ignorenonexistent),
		parseDwordAnswer())
	if updated==nil {
		return -1, err
	}
	return int(updated.(uint32)), err
}

// EscapeString escapes characters that are treated as special operators by the query language parser.
//
// `from` is a string to escape.
// This function might seem redundant because it’s trivial to implement in any calling application.
// However, as the set of special characters might change over time, it makes sense to have an API call that is
// guaranteed to escape all such characters at all times.
// Returns escaped string.
func EscapeString(from string) string {
	dest := make([]byte, 0, 2*len(from))
	for i := 0; i < len(from); i++ {
		c := from[i]
		switch c {
		case '\\', '(', ')', '|', '-', '!', '@', '~', '"', '&', '/', '^', '$', '=', '<':
			dest = append(dest, '\\')
		}
		dest = append(dest, c)
	}
	return string(dest)
}
