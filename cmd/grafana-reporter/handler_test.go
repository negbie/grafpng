/*
   Copyright 2018 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/IzakMarais/reporter/grafana"
	"github.com/IzakMarais/reporter/report"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

type mockReport struct {
}

func (m mockReport) Generate() (pdf io.ReadCloser, err error) {
	return ioutil.NopCloser(bytes.NewReader(nil)), nil
}

func (m mockReport) Clean() {}

func TestReportServeHandler(t *testing.T) {
	Convey("When the report server handler is called", t, func() {
		//mock new grafana client function to capture and validate its input parameters
		var clAPIToken string
		var clVars url.Values
		newGrafanaClient := func(url string, apiToken string, variables url.Values) grafana.Client {
			clAPIToken = apiToken
			clVars = variables
			return grafana.NewClient(url, apiToken, variables)
		}
		//mock new report function to capture and validate its input parameters
		var repDashName string
		newReport := func(g grafana.Client, dashName string, _ grafana.TimeRange, _ string) report.Report {
			repDashName = dashName
			return &mockReport{}
		}

		router := mux.NewRouter()
		RegisterHandlers(router, ServeReportHandler{newGrafanaClient, newReport})
		rec := httptest.NewRecorder()

		Convey("It should extract dashboard ID from the URL and forward it to the new reporter ", func() {
			req, _ := http.NewRequest("GET", "/api/report/testDash", nil)
			router.ServeHTTP(rec, req)
			So(repDashName, ShouldEqual, "testDash")
		})

		Convey("It should extract the apiToken from the URL and forward it to the new Grafana Client ", func() {
			req, _ := http.NewRequest("GET", "/api/report/testDash?apitoken=1234", nil)
			router.ServeHTTP(rec, req)
			So(clAPIToken, ShouldEqual, "1234")
		})

		Convey("It should extract the grafana variables and forward them to the new Grafana Client ", func() {
			req, _ := http.NewRequest("GET", "/api/report/testDash?var-test=testValue", nil)
			router.ServeHTTP(rec, req)
			expected := url.Values{}
			expected.Add("var-test", "testValue")
			So(clVars, ShouldResemble, expected)

			Convey("Variables should not contain other query parameters ", func() {
				req, _ := http.NewRequest("GET", "/api/report/testDash?var-test=testValue&apitoken=1234", nil)
				router.ServeHTTP(rec, req)
				expected := url.Values{}
				expected.Add("var-test", "testValue") //apitoken not expected here
				So(clVars, ShouldResemble, expected)
			})
		})
	})
}