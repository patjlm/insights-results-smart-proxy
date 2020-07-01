// Copyright 2020 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package content_test

import (
	"net/http"
	"testing"
	"time"

	cs_content "github.com/RedHatInsights/insights-content-service/content"
	ics_server "github.com/RedHatInsights/insights-content-service/server"
	"github.com/RedHatInsights/insights-results-aggregator-data/testdata"
	"github.com/stretchr/testify/assert"

	"github.com/RedHatInsights/insights-results-smart-proxy/content"
	"github.com/RedHatInsights/insights-results-smart-proxy/tests/helpers"
)

const (
	testTimeout = 10 * time.Second
)

func TestGetRuleContent(t *testing.T) {
	helpers.RunTestWithTimeout(t, func(t *testing.T) {
		defer helpers.CleanAfterGock(t)
		helpers.GockExpectAPIRequest(t, helpers.DefaultServicesConfig.ContentBaseEndpoint, &helpers.APIRequest{
			Method:   http.MethodGet,
			Endpoint: ics_server.AllContentEndpoint,
		}, &helpers.APIResponse{
			StatusCode: http.StatusOK,
			Body:       helpers.MustGobSerialize(t, testdata.RuleContentDirectory3Rules),
		})

		content.UpdateContent(helpers.DefaultServicesConfig)

		ruleContent, err := content.GetRuleContent(testdata.Rule1ID)
		helpers.FailOnError(t, err)
		assert.NotNil(t, ruleContent)

		assert.Equal(t, testdata.RuleContent1, *ruleContent)
	}, testTimeout)
}

func TestGetRuleContent_CallMultipleTimes(t *testing.T) {
	const N = 10

	helpers.RunTestWithTimeout(t, func(t *testing.T) {
		defer helpers.CleanAfterGock(t)
		helpers.GockExpectAPIRequest(t, helpers.DefaultServicesConfig.ContentBaseEndpoint, &helpers.APIRequest{
			Method:   http.MethodGet,
			Endpoint: ics_server.AllContentEndpoint,
		}, &helpers.APIResponse{
			StatusCode: http.StatusOK,
			Body:       helpers.MustGobSerialize(t, testdata.RuleContentDirectory3Rules),
		})

		content.UpdateContent(helpers.DefaultServicesConfig)

		for i := 0; i < N; i++ {
			ruleContent, err := content.GetRuleContent(testdata.Rule1ID)
			helpers.FailOnError(t, err)
			assert.NotNil(t, ruleContent)

			assert.Equal(t, testdata.RuleContent1, *ruleContent)
		}
	}, testTimeout)
}

func TestUpdateContent_CallMultipleTimes(t *testing.T) {
	const N = 10

	helpers.RunTestWithTimeout(t, func(t *testing.T) {
		defer helpers.CleanAfterGock(t)

		for i := 0; i < N; i++ {
			helpers.GockExpectAPIRequest(t, helpers.DefaultServicesConfig.ContentBaseEndpoint, &helpers.APIRequest{
				Method:   http.MethodGet,
				Endpoint: ics_server.AllContentEndpoint,
			}, &helpers.APIResponse{
				StatusCode: http.StatusOK,
				Body:       helpers.MustGobSerialize(t, testdata.RuleContentDirectory3Rules),
			})
		}

		for i := 0; i < N; i++ {
			content.UpdateContent(helpers.DefaultServicesConfig)
		}

		for i := 0; i < N; i++ {
			ruleContent, err := content.GetRuleContent(testdata.Rule1ID)
			helpers.FailOnError(t, err)
			assert.NotNil(t, ruleContent)

			assert.Equal(t, testdata.RuleContent1, *ruleContent)
		}
	}, testTimeout)
}

func TestUpdateContentBadTime(t *testing.T) {
	// using testdata.RuleContent4 because contains datetime in a different format
	ruleContentDirectory := cs_content.RuleContentDirectory{
		Config: cs_content.GlobalRuleConfig{
			Impact: testdata.ImpactStrToInt,
		},
		Rules: map[string]cs_content.RuleContent{
			"rc4": testdata.RuleContent4,
		},
	}

	content.LoadRuleContent(&ruleContentDirectory)
	content.RuleContentDirectoryReady.L.Lock()
	content.RuleContentDirectoryReady.Broadcast()
	content.RuleContentDirectoryReady.L.Unlock()

	_, err := content.GetRuleWithErrorKeyContent(testdata.Rule4ID, testdata.ErrorKey4)
	helpers.FailOnError(t, err)
}