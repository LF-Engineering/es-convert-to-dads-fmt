package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	urlLib "net/url"

	jsoniter "github.com/json-iterator/go"
)

const (
	cGitHubURLRoot = "https://github.com/"
)

var (
	gESURL            string
	gNotAnalyzeString = []byte(`{"dynamic_templates":[{"notanalyzed":{"match":"*","match_mapping_type":"string","mapping":{"type":"keyword"}}},{"formatdate":{"match":"*","match_mapping_type":"date","mapping":{"type":"date","format":"strict_date_optional_time||epoch_millis"}}}]}`)
	gDSTypes          = map[string]struct{}{
		"git":                 {},
		"github/issue":        {},
		"github/pull_request": {},
	}
	gMapping = map[string][]byte{
		"github/issue":        []byte(`{"dynamic":true,"properties":{"metadata__updated_on":{"type":"date","format":"strict_date_optional_time||epoch_millis"},"merge_author_geolocation":{"type":"geo_point"},"assignee_geolocation":{"type":"geo_point"},"state":{"type":"keyword"},"user_geolocation":{"type":"geo_point"},"title_analyzed":{"type":"text","index":true},"body_analyzed":{"type":"text","index":true}},"dynamic_templates":[{"notanalyzed":{"match":"*","unmatch":"body","match_mapping_type":"string","mapping":{"type":"keyword"}}},{"formatdate":{"match":"*","match_mapping_type":"date","mapping":{"format":"strict_date_optional_time||epoch_millis","type":"date"}}}]}`),
		"github/pull_request": []byte(`{"dynamic":true,"properties":{"metadata__updated_on":{"type":"date","format":"strict_date_optional_time||epoch_millis"},"merge_author_geolocation":{"type":"geo_point"},"assignee_geolocation":{"type":"geo_point"},"state":{"type":"keyword"},"user_geolocation":{"type":"geo_point"},"title_analyzed":{"type":"text","index":true},"body_analyzed":{"type":"text","index":true}},"dynamic_templates":[{"notanalyzed":{"match":"*","unmatch":"body","match_mapping_type":"string","mapping":{"type":"keyword"}}},{"formatdate":{"match":"*","match_mapping_type":"date","mapping":{"format":"strict_date_optional_time||epoch_millis","type":"date"}}}]}`),
		"git":                 []byte(`{"dynamic":true,"properties":{"file_data":{"type":"nested"},"authors_signed":{"type":"nested"},"authors_co_authored":{"type":"nested"},"authors_tested":{"type":"nested"},"authors_approved":{"type":"nested"},"authors_reviewed":{"type":"nested"},"authors_reported":{"type":"nested"},"authors_informed":{"type":"nested"},"authors_resolved":{"type":"nested"},"authors_influenced":{"type":"nested"},"author_name":{"type":"keyword"},"metadata__updated_on":{"type":"date","format":"strict_date_optional_time||epoch_millis"},"message_analyzed":{"type":"text","index":true}},"dynamic_templates":[{"notanalyzed":{"match":"*","unmatch":"message_analyzed","match_mapping_type":"string","mapping":{"type":"keyword"}}},{"formatdate":{"match":"*","match_mapping_type":"date","mapping":{"format":"strict_date_optional_time||epoch_millis","type":"date"}}}]}`),
	}
	gNoCopyFields = map[string]map[string]struct{}{
		"github/issue": {
			"user_data_gender_acc":        {},
			"user_data_gender":            {},
			"repository_labels":           {},
			"project_1":                   {},
			"metadata__gelk_version":      {},
			"metadata__gelk_backend_name": {},
			"metadata__filter_raw":        {},
			"item_type":                   {},
			"id":                          {},
			"author_gender_acc":           {},
			"author_gender":               {},
			"assignee_data_gender_acc":    {},
			"assignee_data_gender":        {},
		},
		"github/pull_request": {
			"user_data_gender_acc":        {},
			"user_data_gender":            {},
			"repository_labels":           {},
			"project_1":                   {},
			"metadata__gelk_version":      {},
			"metadata__gelk_backend_name": {},
			"metadata__filter_raw":        {},
			"item_type":                   {},
			"id":                          {},
			"author_gender_acc":           {},
			"author_gender":               {},
			"assignee_data_gender_acc":    {},
			"assignee_data_gender":        {},
			"merged_by_data_gender_acc":   {},
			"merged_by_data_gender":       {},
		},
	}
	gCopyFields = map[string]map[[2]string]struct{}{
		"github/issue": {
			[2]string{"origin", "repo_name"}: {},
			[2]string{"id", "issue_id"}:      {},
		},
		"github/pull_request": {
			[2]string{"origin", "repo_name"}:   {},
			[2]string{"id", "pull_request_id"}: {},
			[2]string{"merged", "is_approved"}: {},
		},
	}
)

func fatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		msg := fmt.Sprintf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Printf("%s", msg)
		fmt.Fprintf(os.Stderr, "%s", msg)
		panic("stacktrace")
	}
}

func fatalf(f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...))
}

func getThreadsNum() int {
	st := os.Getenv("NCPUS")
	if st == "" {
		return runtime.NumCPU()
	}
	n, err := strconv.Atoi(st)
	if err != nil || n <= 0 {
		return runtime.NumCPU()
	}
	runtime.GOMAXPROCS(n)
	return n
}

func bytesToStringTrunc(data []byte, maxLen int, addLenInfo bool) (str string) {
	lenInfo := ""
	if addLenInfo {
		lenInfo = "(" + strconv.Itoa(len(data)) + "): "
	}
	if len(data) <= maxLen {
		return lenInfo + string(data)
	}
	half := maxLen >> 1
	str = lenInfo + string(data[:half]) + "(...)" + string(data[len(data)-half:])
	return
}

func interfaceToStringTrunc(iface interface{}, maxLen int, addLenInfo bool) (str string) {
	data := fmt.Sprintf("%+v", iface)
	lenInfo := ""
	if addLenInfo {
		lenInfo = "(" + strconv.Itoa(len(data)) + "): "
	}
	if len(data) <= maxLen {
		return lenInfo + data
	}
	half := maxLen >> 1
	str = "(" + strconv.Itoa(len(data)) + "): " + data[:half] + "(...)" + data[len(data)-half:]
	return
}

func stringToCookie(s string) (c *http.Cookie) {
	ary := strings.Split(s, "===")
	if len(ary) < 2 {
		return
	}
	c = &http.Cookie{Name: ary[0], Value: ary[1]}
	return
}

func cookieToString(c *http.Cookie) (s string) {
	if c.Name == "" && c.Value == "" {
		return
	}
	s = c.Name + "===" + c.Value
	return
}

func keysOnly(i interface{}) (o map[string]interface{}) {
	if i == nil {
		return
	}
	is, ok := i.(map[string]interface{})
	if !ok {
		return
	}
	o = make(map[string]interface{})
	for k, v := range is {
		o[k] = keysOnly(v)
	}
	return
}

func dumpKeys(i interface{}) string {
	return strings.Replace(fmt.Sprintf("%v", keysOnly(i)), "map[]", "", -1)
}

func request(
	url, method string,
	headers map[string]string,
	payload []byte,
	cookies []string,
	jsonStatuses, errorStatuses, okStatuses map[[2]int]struct{},
) (result interface{}, status int, isJSON bool, outCookies []string, outHeaders map[string][]string, err error) {
	var (
		payloadBody *bytes.Reader
		req         *http.Request
	)
	if len(payload) > 0 {
		payloadBody = bytes.NewReader(payload)
		req, err = http.NewRequest(method, url, payloadBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		sPayload := bytesToStringTrunc(payload, 1024, true)
		err = fmt.Errorf("new request error:%+v for method:%s url:%s payload:%s", err, method, url, sPayload)
		return
	}
	for _, cookieStr := range cookies {
		cookie := stringToCookie(cookieStr)
		req.AddCookie(cookie)
	}
	for header, value := range headers {
		req.Header.Set(header, value)
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		sPayload := bytesToStringTrunc(payload, 1024, true)
		err = fmt.Errorf("do request error:%+v for method:%s url:%s headers:%v payload:%s", err, method, url, headers, sPayload)
		if strings.Contains(err.Error(), "socket: too many open files") {
			fmt.Printf("too many open socets detected, sleeping for 3 seconds\n")
			time.Sleep(time.Duration(3) * time.Second)
		}
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sPayload := bytesToStringTrunc(payload, 1024, true)
		err = fmt.Errorf("read request body error:%+v for method:%s url:%s headers:%v payload:%s", err, method, url, headers, sPayload)
		return
	}
	_ = resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		outCookies = append(outCookies, cookieToString(cookie))
	}
	outHeaders = resp.Header
	status = resp.StatusCode
	hit := false
	for r := range jsonStatuses {
		if status >= r[0] && status <= r[1] {
			hit = true
			break
		}
	}
	if hit {
		err = jsoniter.Unmarshal(body, &result)
		if err != nil {
			sPayload := bytesToStringTrunc(payload, 1024, true)
			sBody := bytesToStringTrunc(body, 1024, true)
			err = fmt.Errorf("unmarshall request error:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s", err, method, url, headers, status, sPayload, sBody)
			return
		}
		isJSON = true
	} else {
		result = body
	}
	hit = false
	for r := range errorStatuses {
		if status >= r[0] && status <= r[1] {
			hit = true
			break
		}
	}
	if hit {
		sPayload := bytesToStringTrunc(payload, 1024, true)
		sBody := bytesToStringTrunc(body, 1024, true)
		var sResult string
		bResult, bOK := result.([]byte)
		if bOK {
			sResult = bytesToStringTrunc(bResult, 1024, true)
		} else {
			sResult = interfaceToStringTrunc(result, 1024, true)
		}
		err = fmt.Errorf("status error:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s result:%+v", err, method, url, headers, status, sPayload, sBody, sResult)
	}
	if len(okStatuses) > 0 {
		hit = false
		for r := range okStatuses {
			if status >= r[0] && status <= r[1] {
				hit = true
				break
			}
		}
		if !hit {
			sPayload := bytesToStringTrunc(payload, 1024, true)
			sBody := bytesToStringTrunc(body, 1024, true)
			var sResult string
			bResult, bOK := result.([]byte)
			if bOK {
				sResult = bytesToStringTrunc(bResult, 1024, true)
			} else {
				sResult = interfaceToStringTrunc(result, 1024, true)
			}
			err = fmt.Errorf("status not success:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s result:%+v", err, method, url, headers, status, sPayload, sBody, sResult)
		}
	}
	return
}

func forEachESItem(
	dsType, idxFrom, idxTo, idField string,
	ufunct func(string, string, string, string, int, *[]interface{}, *[]interface{}, bool) error,
	uitems func(string, string, string, string, int, []interface{}, *[]interface{}) error,
) (err error) {
	packSize := 1000
	var (
		scroll *string
		res    interface{}
		status int
	)
	headers := map[string]string{"Content-Type": "application/json"}
	attemptAt := time.Now()
	total := 0
	// Defer free scroll
	defer func() {
		if scroll == nil {
			return
		}
		url := gESURL + "/_search/scroll"
		payload := []byte(`{"scroll_id":"` + *scroll + `"}`)
		_, _, _, _, _, err := request(
			url,
			"DELETE",
			headers,
			payload,
			[]string{},
			nil,
			nil,                                 // Error statuses
			map[[2]int]struct{}{{200, 200}: {}}, // OK statuses
		)
		if err != nil {
			fmt.Printf("Error releasing scroll %s: %+v, ignored\n", *scroll, err)
			err = nil
		}
	}()
	thrN := getThreadsNum()
	fmt.Printf("Using %d threads\n", thrN)
	nThreads := 0
	var (
		mtx *sync.Mutex
		ch  chan error
	)
	docs := []interface{}{}
	outDocs := []interface{}{}
	if thrN > 1 {
		mtx = &sync.Mutex{}
		ch = make(chan error)
	}
	funct := func(c chan error, last bool) (e error) {
		defer func() {
			if thrN > 1 {
				mtx.Unlock()
			}
			if c != nil {
				c <- e
			}
		}()
		if thrN > 1 {
			mtx.Lock()
		}
		e = ufunct(dsType, idxFrom, idxTo, idField, thrN, &docs, &outDocs, last)
		return
	}
	for {
		var (
			url     string
			payload []byte
		)
		if scroll == nil {
			url = gESURL + "/" + idxFrom + "/_search?scroll=15m&size=1000"
		} else {
			url = gESURL + "/_search/scroll"
			payload = []byte(`{"scroll":"15m","scroll_id":"` + *scroll + `"}`)
		}
		res, status, _, _, _, err = request(
			url,
			"POST",
			headers,
			payload,
			[]string{},
			map[[2]int]struct{}{{200, 200}: {}}, // JSON statuses
			nil,                                 // Error statuses
			map[[2]int]struct{}{{200, 200}: {}, {404, 404}: {}, {500, 500}: {}}, // OK statuses
		)
		fatalOnError(err)
		if status == 404 {
			if scroll != nil && strings.Contains(string(res.([]byte)), "No search context found for id") {
				fmt.Printf("scroll %s probably expired, retrying\n", *scroll)
				scroll = nil
				err = nil
				continue
			}
			fatalf("got status 404 but not because of scroll context expiration:\n%s\n", string(res.([]byte)))
		}
		if status == 500 {
			if scroll == nil && status == 500 && strings.Contains(string(res.([]byte)), "Trying to create too many scroll contexts") {
				time.Sleep(5)
				now := time.Now()
				elapsed := now.Sub(attemptAt)
				fmt.Printf("%d retrying scroll, first attempt at %+v, elapsed %+v/%ds\n", len(res.(map[string]interface{})), attemptAt, elapsed, 600)
				if elapsed.Seconds() > 600 {
					fatalf("Tried to acquire scroll too many times, first attempt at %v, elapsed %v/%ds", attemptAt, elapsed, 600)
				}
				continue
			}
			fatalf("got status 500 but not because of too many scrolls:\n%s\n", string(res.([]byte)))
		}
		sScroll, ok := res.(map[string]interface{})["_scroll_id"].(string)
		if !ok {
			err = fmt.Errorf("Missing _scroll_id in the response %+v", dumpKeys(res))
			return
		}
		scroll = &sScroll
		items, ok := res.(map[string]interface{})["hits"].(map[string]interface{})["hits"].([]interface{})
		if !ok {
			err = fmt.Errorf("Missing hits.hits in the response %+v", dumpKeys(res))
			return
		}
		nItems := len(items)
		if nItems == 0 {
			break
		}
		if thrN > 1 {
			mtx.Lock()
		}
		err = uitems(dsType, idxFrom, idxTo, idField, thrN, items, &docs)
		if err != nil {
			return
		}
		nDocs := len(docs)
		if nDocs >= packSize {
			if thrN > 1 {
				go func() {
					_ = funct(ch, false)
				}()
				nThreads++
				if nThreads == thrN {
					err = <-ch
					if err != nil {
						return
					}
					nThreads--
				}
			} else {
				err = funct(nil, false)
				if err != nil {
					return
				}
			}
		}
		if thrN > 1 {
			mtx.Unlock()
		}
		total += nItems
	}
	if thrN > 1 {
		mtx.Lock()
	}
	if thrN > 1 {
		go func() {
			_ = funct(ch, true)
		}()
		nThreads++
		if nThreads == thrN {
			err = <-ch
			if err != nil {
				return
			}
			nThreads--
		}
	} else {
		err = funct(nil, true)
		if err != nil {
			return
		}
	}
	if thrN > 1 {
		mtx.Unlock()
	}
	for thrN > 1 && nThreads > 0 {
		err = <-ch
		nThreads--
		if err != nil {
			return
		}
	}
	fmt.Printf("Total number of items processed: %d\n", total)
	return
}

func sendToElastic(idxName, key string, items []interface{}) (err error) {
	fmt.Printf("%s(key=%s) ES bulk uploading %d items\n", idxName, key, len(items))
	url := gESURL + "/" + idxName + "/_bulk?refresh=true"
	// {"index":{"_id":"uuid"}}
	payloads := []byte{}
	newLine := []byte("\n")
	var (
		doc    []byte
		hdr    []byte
		status int
	)
	for _, item := range items {
		doc, err = jsoniter.Marshal(item)
		if err != nil {
			return
		}
		iID, ok := item.(map[string]interface{})[key]
		if !ok {
			err = fmt.Errorf("missing %s property in %+v", key, dumpKeys(item))
			return
		}
		id, ok := iID.(string)
		if !ok {
			err = fmt.Errorf("%s property is %T not string %+v", key, iID, iID)
			return
		}
		hdr = []byte(`{"index":{"_id":"` + id + "\"}}\n")
		payloads = append(payloads, hdr...)
		payloads = append(payloads, doc...)
		payloads = append(payloads, newLine...)
	}
	var result interface{}
	result, status, _, _, _, err = request(
		url,
		"POST",
		map[string]string{"Content-Type": "application/x-ndjson"},
		payloads,
		[]string{},
		map[[2]int]struct{}{{200, 200}: {}}, // JSON statuses
		map[[2]int]struct{}{{400, 599}: {}}, // error statuses: 400-599
		nil,                                 // OK statuses
	)
	resp, ok := result.(map[string]interface{})
	if ok {
		ers, ok := resp["errors"].(bool)
		if ok && ers {
			msg := interfaceToStringTrunc(result, 1000, true)
			fmt.Printf("%s(key=%s): bulk upload failed: status=%d, %s\n", idxName, key, status, msg)
			err = fmt.Errorf("%s", msg)
		}
	}
	if err == nil {
		fmt.Printf("%s(key=%s) ES bulk upload saved %d items\n", idxName, key, len(items))
		return
	}
	var sResp string
	bResp, ok := result.([]byte)
	if ok {
		sResp = bytesToStringTrunc(bResp, 1024, true)
	}
	fmt.Printf("%s(key=%s) ES bulk upload of %d items failed, falling back to one-by-one mode, response: %s\n", idxName, key, len(items), sResp)
	fmt.Printf("%s(key=%s) ES bulk upload error: %+v\n", idxName, key, err)
	err = nil
	// Fallback to one-by-one inserts
	indexName := idxName
	url = gESURL + "/" + indexName + "/_doc/"
	headers := map[string]string{"Content-Type": "application/json"}
	var itemStatus int
	for _, item := range items {
		doc, _ = jsoniter.Marshal(item)
		id, _ := item.(map[string]interface{})[key].(string)
		id = urlLib.PathEscape(id)
		_, itemStatus, _, _, _, err = request(
			url+id,
			"PUT",
			headers,
			doc,
			[]string{},
			nil,                                 // JSON statuses
			map[[2]int]struct{}{{400, 599}: {}}, // error statuses: 400-599
			map[[2]int]struct{}{{200, 201}: {}}, // OK statuses: 200-201
		)
		if err != nil {
			fmt.Printf("sendToElastic: error: %+v, status=%d for %+v\n", err, itemStatus, item)
			err = nil
		}
	}
	fmt.Printf("%s(key=%s) ES bulk upload saved %d items (in non-bulk mode)\n", idxName, key, len(items))
	return
}

func esBulkUploadFunc(dsType, idxFrom, idxTo, itemID string, thrN int, docs, outDocs *[]interface{}, last bool) (e error) {
	fmt.Printf("%s(%s): %s -> %s: ES bulk uploading %d/%d func\n", dsType, itemID, idxFrom, idxTo, len(*docs), len(*outDocs))
	bulkSize := 1000
	run := func() (err error) {
		nItems := len(*outDocs)
		fmt.Printf("ES bulk uploading %d items to ES\n", nItems)
		nPacks := nItems / bulkSize
		if nItems%bulkSize != 0 {
			nPacks++
		}
		for i := 0; i < nPacks; i++ {
			from := i * bulkSize
			to := from + bulkSize
			if to > nItems {
				to = nItems
			}
			fmt.Printf("ES bulk upload: bulk uploading pack #%d %d-%d (%d/%d) to ES\n", i+1, from, to, to-from, nPacks)
			err = sendToElastic(idxTo, itemID, (*outDocs)[from:to])
			if err != nil {
				return
			}
		}
		return
	}
	nDocs := len(*docs)
	nOutDocs := len(*outDocs)
	fmt.Printf("ES bulk upload pack size %d/%d last %v\n", nDocs, nOutDocs, last)
	for _, doc := range *docs {
		*outDocs = append(*outDocs, doc)
		nOutDocs = len(*outDocs)
		if nOutDocs >= bulkSize {
			fmt.Printf("ES bulk pack size %d/%d reached, flushing\n", nOutDocs, bulkSize)
			e = run()
			if e != nil {
				return
			}
			*outDocs = []interface{}{}
		}
	}
	if last {
		nOutDocs := len(*outDocs)
		if nOutDocs > 0 {
			e = run()
			if e != nil {
				return
			}
			*outDocs = []interface{}{}
		}
	}
	*docs = []interface{}{}
	nOutDocs = len(*outDocs)
	if nOutDocs > 0 {
		fmt.Printf("ES bulk upload %d items left (last %v)\n", nOutDocs, last)
	}
	return
}

func handleMapping(idx string, mapping []byte, useDefault bool) (err error) {
	// Create index, ignore if exists (see status 400 is not in error statuses)
	url := gESURL + "/" + idx
	fmt.Printf("index: %s\n", url)
	var (
		result interface{}
		status int
	)
	stringResult := func(r interface{}) string {
		bR, ok := r.([]byte)
		if ok {
			return string(bR)
		}
		iR, ok := r.(map[string]interface{})
		if ok {
			return fmt.Sprintf("%+v", iR)
		}
		return fmt.Sprintf("%+v", r)
	}
	result, status, _, _, _, err = request(
		url+"?wait_for_active_shards=all",
		"PUT",
		nil,                                 // headers
		[]byte{},                            // payload
		[]string{},                          // cookies
		nil,                                 // JSON statuses
		map[[2]int]struct{}{{401, 599}: {}}, // error statuses: 401-599
		nil,                                 // OK statuses
	)
	fmt.Printf("index %s created: status=%d, result: %+v\n", url, status, stringResult(result))
	fatalOnError(err)
	// DS specific raw index mapping
	url += "/_mapping"
	result, status, _, _, _, err = request(
		url,
		"PUT",
		map[string]string{"Content-Type": "application/json"},
		mapping,
		[]string{},
		nil,
		nil,
		map[[2]int]struct{}{{200, 200}: {}},
	)
	fmt.Printf("index mapping %s -> status=%d, result: %+v\n", url, status, stringResult(result))
	//fmt.Printf("mapping: %+v\n", string(mapping))
	fatalOnError(err)
	if useDefault {
		// Global not analyze string mapping
		result, status, _, _, _, err = request(
			url,
			"PUT",
			map[string]string{"Content-Type": "application/json"},
			gNotAnalyzeString,
			[]string{},
			nil,
			nil,
			map[[2]int]struct{}{{200, 200}: {}},
		)
		fmt.Printf("index not analyze string mapping %s -> status=%d, result: %+v\n", url, status, stringResult(result))
		fatalOnError(err)
	}
	return
}
func translate(in map[string]interface{}, ds string) (map[string]interface{}, error) {
	switch ds {
	case "github/issue":
		return translateGithubIssue(in)
	case "github/pull_request":
		return translateGithubPullRequest(in)
	default:
		return nil, fmt.Errorf("translate for %s ds type not implemented", ds)
	}
}

func translateGithubIssue(in map[string]interface{}) (out map[string]interface{}, err error) {
	out = make(map[string]interface{})
	noCopyFields := gNoCopyFields["github/issue"]
	for k, v := range in {
		_, noCopy := noCopyFields[k]
		if noCopy {
			continue
		}
		out[k] = v
	}
	copyFields := gCopyFields["github/issue"]
	for data := range copyFields {
		from := data[0]
		to := data[1]
		out[to], _ = in[from]
	}
	_, ok := in["project"]
	if ok {
		out["project_ts"] = time.Now().Unix()
	}
	githubRepo, _ := in["origin"].(string)
	if strings.HasSuffix(githubRepo, ".git") {
		githubRepo = githubRepo[:len(githubRepo)-4]
	}
	if strings.Contains(githubRepo, cGitHubURLRoot) {
		githubRepo = strings.Replace(githubRepo, cGitHubURLRoot, "", -1)
	}
	out["github_repo"] = githubRepo
	var repoShortName string
	arr := strings.Split(githubRepo, "/")
	if len(arr) > 1 {
		repoShortName = arr[1]
	}
	out["repo_short_name"] = repoShortName
	out["n_total_comments"] = 0
	out["n_reactions"] = 0
	out["n_comments"] = 0
	out["n_commenters"] = 0
	out["n_assignees"] = 0
	out["type"] = "issue"
	pr, _ := in["pull_request"]
	if pr.(bool) {
		out["item_type"] = "issue pull request"
	} else {
		out["item_type"] = "issue"
	}
	number, _ := in["id_in_repo"]
	out["id"] = githubRepo + "/issue/" + number.(string)
	out["commenters"] = []interface{}{}
	out["assignees_data"] = []interface{}{}
	out["category"] = "issue"
	// p2o doesn't have it
	out["project_slug"] = nil
	return
}

func translateGithubPullRequest(in map[string]interface{}) (out map[string]interface{}, err error) {
	out = make(map[string]interface{})
	noCopyFields := gNoCopyFields["github/pull_request"]
	for k, v := range in {
		_, noCopy := noCopyFields[k]
		if noCopy {
			continue
		}
		out[k] = v
	}
	copyFields := gCopyFields["github/pull_request"]
	for data := range copyFields {
		from := data[0]
		to := data[1]
		out[to], _ = in[from]
	}
	_, ok := in["project"]
	if ok {
		out["project_ts"] = time.Now().Unix()
	}
	githubRepo, _ := in["origin"].(string)
	if strings.HasSuffix(githubRepo, ".git") {
		githubRepo = githubRepo[:len(githubRepo)-4]
	}
	if strings.Contains(githubRepo, cGitHubURLRoot) {
		githubRepo = strings.Replace(githubRepo, cGitHubURLRoot, "", -1)
	}
	out["github_repo"] = githubRepo
	var repoShortName string
	arr := strings.Split(githubRepo, "/")
	if len(arr) > 1 {
		repoShortName = arr[1]
	}
	out["repo_short_name"] = repoShortName
	out["item_type"] = "pull request"
	out["type"] = "pull_request"
	number, _ := in["id_in_repo"]
	out["id"] = githubRepo + "/pull_request/" + number.(string)
	out["review_commenters"] = []interface{}{}
	out["requested_reviewers_data"] = []interface{}{}
	out["assignees_data"] = []interface{}{}
	out["n_total_comments"] = 0
	out["n_review_comments"] = 0
	out["n_review_commenters"] = 0
	out["n_requested_reviewers"] = 0
	out["n_reactions"] = 0
	out["n_comments"] = 0
	out["n_commenters"] = 0
	out["n_assignees"] = 0
	out["category"] = "pull_request"
	// p2o doesn't have it
	out["project_slug"] = nil
	return
}

func itemsFunc(dsType, idxFrom, idxTo, idField string, thrN int, items []interface{}, docs *[]interface{}) (err error) {
	fmt.Printf("%s(%s): %s -> %s: %d items, %d threads\n", dsType, idField, idxFrom, idxTo, len(items), thrN)
	var (
		mtx *sync.Mutex
		ch  chan error
	)
	if thrN > 1 {
		mtx = &sync.Mutex{}
		ch = make(chan error)
	}
	procItem := func(c chan error, item interface{}) (e error) {
		defer func() {
			if c != nil {
				c <- e
			}
		}()
		doc, ok := item.(map[string]interface{})["_source"]
		if !ok {
			e = fmt.Errorf("Missing _source in item %+v", dumpKeys(item))
			return
		}
		in, _ := doc.(map[string]interface{})
		out, e := translate(in, dsType)
		fatalOnError(e)
		if thrN > 1 {
			mtx.Lock()
		}
		*docs = append(*docs, out)
		if thrN > 1 {
			mtx.Unlock()
		}
		return
	}
	if thrN > 1 {
		nThreads := 0
		for _, item := range items {
			go func(it interface{}) {
				_ = procItem(ch, it)
			}(item)
			nThreads++
			if nThreads == thrN {
				err = <-ch
				if err != nil {
					return
				}
				nThreads--
			}
		}
		for nThreads > 0 {
			err = <-ch
			nThreads--
			if err != nil {
				return
			}
		}
		return
	}
	for _, item := range items {
		err = procItem(nil, item)
		if err != nil {
			return
		}
	}
	return
}

func convertGitHubIssue(idxFrom, idxTo string) (err error) {
	fatalOnError(handleMapping(idxTo, gMapping["github/issue"], false))
	err = forEachESItem("github/issue", idxFrom, idxTo, "url_id", esBulkUploadFunc, itemsFunc)
	return
}

func convertGitHubPullRequest(idxFrom, idxTo string) (err error) {
	fatalOnError(handleMapping(idxTo, gMapping["github/pull_request"], false))
	err = forEachESItem("github/pull_request", idxFrom, idxTo, "url_id", esBulkUploadFunc, itemsFunc)
	return
}

func convertGit(idxFrom, idxTo string) (err error) {
	fatalOnError(handleMapping(idxTo, gMapping["git"], false))
	err = forEachESItem("git", idxFrom, idxTo, "uuid", esBulkUploadFunc, itemsFunc)
	return
}

func convert(dsType, idxFrom, idxTo string) (err error) {
	switch dsType {
	case "git":
		err = convertGit(idxFrom, idxTo)
	case "github/issue":
		err = convertGitHubIssue(idxFrom, idxTo)
	case "github/pull_request":
		err = convertGitHubPullRequest(idxFrom, idxTo)
	default:
		err = fmt.Errorf("%s support not implemented", dsType)
	}
	return
}

func main() {
	if len(os.Args) < 3 {
		fatalf("ES_URL=... %s: ds-type from-index to-index\n", os.Args[0])
		return
	}
	gESURL = os.Getenv("ES_URL")
	if gESURL == "" {
		fatalf("%s: you need to set ES_URL environment variable\n", os.Args[0])
		return
	}
	dsType := os.Args[1]
	_, ok := gDSTypes[dsType]
	if !ok {
		fatalf("%s: %s is not a know ds-type, allowed are: %+v\n", os.Args[0], dsType, gDSTypes)
		return
	}
	fatalOnError(convert(dsType, os.Args[2], os.Args[3]))
}
