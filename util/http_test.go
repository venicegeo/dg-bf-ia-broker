// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSubmitSinglePart(t *testing.T) {
	SetMockClient(nil, 250)
	method := "TRACE"
	url := "http://testURL.net"
	bodyStr := "testBody"
	authKey := "testAuthKey"

	resp, err := SubmitSinglePart(method, bodyStr, url, authKey)
	if err != nil {
		t.Error(`received error on basic test of SubmitSinglePart.  Error message: ` + err.Error())
	} else {
		req := (*http.Request)(resp.Request)
		if req.Header.Get("Content-Type") != "application/json" {
			t.Error(`SubmitSinglePart: Content-Type not application/json.`)
		}
		if req.Header.Get("Authorization") != authKey {
			t.Error(`SubmitSinglePart: Authorization not sustained properly.`)
		}
		if req.URL.String() != url {
			t.Error(`SubmitSinglePart: URL not sustained properly.`)
		}
		if req.Method != method {
			t.Error(`SubmitSinglePart: method not sustained properly.`)
		}
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		if string(bodyBytes) != bodyStr {
			t.Error(`SubmitSinglePart: body string not sustained properly.`)
		}
	}

	SetMockClient(nil, 500)
	resp, err = SubmitSinglePart(method, bodyStr, url, authKey)
	if err == nil {
		t.Error(`SubmitSinglePart: did not respond to http status error properly.`)
	}

	SetMockClient(nil, 100)
	resp, err = SubmitSinglePart(method, bodyStr, url, authKey)
	if err == nil {
		t.Error(`SubmitSinglePart: did not respond to http status error properly.`)
	}
}

func TestSubmitMultipart(t *testing.T) {
	SetMockClient(nil, 250)
	bodyStr := "testBody"
	url := "http://testURL.net"
	fileName := "name"
	authKey := "testAuthKey"
	testData := []byte("testtesttest")

	_, err := SubmitMultipart(bodyStr, url, fileName, authKey, testData)
	if err != nil {
		t.Errorf("TestSubmitMultipart: failed on what should have been good run. %v", err.Error())
	}
	SetMockClient(nil, 550)
	_, err = SubmitMultipart(bodyStr, url, fileName, authKey, testData)
	if err == nil {
		t.Error(`TestSubmitMultipart: passed on what should have been bad status code.`)
	}

}

func TestRequestKnownJSON(t *testing.T) {
	outStrs := []string{
		`{"PercentComplete":0, "TimeRemaining":"blah", "TimeSpent":"blah"}`,
		`XXXXX`,
	}
	SetMockClient(outStrs, 250)
	method := "TRACE"
	url := "http://testURL.net"
	bodyStr := "testBody"
	authKey := "testAuthKey"
	jp := make(map[string]interface{})
	_, err := RequestKnownJSON(method, bodyStr, url, authKey, &jp)
	if err != nil {
		t.Errorf("TestSubmitMultipart: failed on what should have been good run. %T", err)
		t.Errorf("TestSubmitMultipart: failed on what should have been good run. %v", err.Error())
	}
	_, err = RequestKnownJSON(method, bodyStr, url, authKey, &jp)
	if err == nil {
		t.Error(`TestRequestKnownJSON: passed on what should have been bad JSON.`)
	}
	_, err = RequestKnownJSON("", "", "flack", "flank", &jp)
	if err == nil {
		t.Error(`TestRequestKnownJSON: passed on what should have been bad call.`)
	}
}

func TestReqByObjJSON(t *testing.T) {
	SetMockClient(nil, 250)
	method := "TEST"
	url := "http://testURL.net"
	authKey := "testAuthKey"
	var emptyHolder interface{}
	_, err := ReqByObjJSON(method, url, authKey, emptyHolder, emptyHolder)
	if err == nil {
		t.Error(`TestReqByObjJSON: passed on what should have been a bad run.`)
	}

}
func TestHttpResponseWriter(t *testing.T) {

	var emptyHolder interface{}
	rr, _, _ := GetMockResponseWriter()
	HTTPOut(rr, "Test", 200)
	PrintJSON(rr, emptyHolder, 200)
	/*
		method := "TRACE"
		url := "http://testURL.net"
		testData := []byte("testtesttest")
		xx := bytes.NewBuffer(testData)
		req := httptest.NewRequest(method, url, xx)
		if Preflight(rr, req) {
			t.Log("Options")
		} else {
			t.Log(req.Header.Get("Origin"))
			t.Log(req.Header.Get("Access-Control-Allow-Origin"))
			t.Log(req.Header.Get("Access-Control-Allow-Methods"))
			t.Log(req.Header.Get("Access-Control-Allow-Headers"))
		}
		req.Header.Add("Origin", "set")
		req.Header.Add("Access-Control-Allow-Origin", "Hit")
		req.Header.Add("Access-Control-Allow-Methods", "Hat")
		req.Header.Add("Access-Control-Allow-Headers", "Hot")

		if Preflight(rr, req) {
			t.Log("Options")
		} else {
			t.Log(req.Header.Get("Origin"))
			t.Log(req.Header.Get("Access-Control-Allow-Origin"))
			t.Log(req.Header.Get("Access-Control-Allow-Methods"))
			t.Log(req.Header.Get("Access-Control-Allow-Headers"))
		}*/

}

func TestHTTPError(t *testing.T) {
	writer, _, _ := GetMockResponseWriter()
	lc := &BasicLogContext{}
	LogSimpleErr(lc, "Test Error.", errors.New("Test Error for TestHTTPError."))
	HTTPError(writer, lc)
	LogInfo(lc, writer.OutputString)
}

func TestReadBodyJSON(t *testing.T) {

	bStrings := []string{``, `b`, `{}`, `{"PercentComplete":50}`}

	for i, bstr := range bStrings {
		jp := make(map[string]interface{})
		body := GetMockReadCloser(bstr)
		_, err := ReadBodyJSON(&jp, body)
		if i < 2 && err == nil {
			t.Error("ReadBodyJson did not throw error on test ", i)
		}
		if i >= 2 && err != nil {
			t.Error("ReadBodyJson threw error on test ", i, ".  Error: ", err.Error())
		}
	}
}

func TestTestUtils(t *testing.T) {
	testData := []byte("testtesttest")
	mockRespWrite, _, _ := GetMockResponseWriter()
	mockRespWrite.Header()
	returnInt, err := mockRespWrite.Write(testData)
	if err != nil {
		t.Error("MockRespHeader Write Failed")
	} else {
		t.Logf("MockRespHeader Write returns %v", returnInt)
	}
	mockRespWrite.WriteHeader(10)
}

func TestUtils(t *testing.T) {
	uuidStrings := [3]string{}

	uuidStrings[0], _ = PsuUUID()
	uuidStrings[1], _ = PsuUUID()
	uuidStrings[2], _ = PsuUUID()

	uuidSlice := uuidStrings[0:2]
	t.Log(SliceToCommaSep(uuidSlice))

}
