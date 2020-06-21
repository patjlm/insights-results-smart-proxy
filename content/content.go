// Package content provides API to get rule's content by its `rule id` and `error key`.
// It takes all the work of caching rules taken from content service
package content

import (
	"fmt"
	"sync"
	"time"

	"github.com/RedHatInsights/insights-content-service/content"
	ics_content "github.com/RedHatInsights/insights-content-service/content"
	"github.com/RedHatInsights/insights-results-aggregator-utils/types"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-results-smart-proxy/services"
)

var (
	ruleContentDirectory      *content.RuleContentDirectory
	ruleContentDirectoryReady = make(chan struct{})
)

type ruleIDAndErrorKey struct {
	RuleID   types.RuleID
	ErrorKey types.ErrorKey
}

// RulesWithContentStorage is a key:value structure to store processed rules.
// It's thread safe
type RulesWithContentStorage struct {
	sync.RWMutex
	rulesWithContent map[ruleIDAndErrorKey]*types.RuleWithContent
	rules            map[types.RuleID]*ics_content.RuleContent
}

func (s *RulesWithContentStorage) GetRuleWithErrorKeyContent(
	ruleID types.RuleID, errorKey types.ErrorKey,
) (*types.RuleWithContent, bool) {
	s.RLock()
	defer s.RUnlock()

	res, found := s.rulesWithContent[ruleIDAndErrorKey{
		RuleID:   ruleID,
		ErrorKey: errorKey,
	}]
	return res, found
}

func (s *RulesWithContentStorage) GetRuleContent(ruleID types.RuleID) (*ics_content.RuleContent, bool) {
	s.RLock()
	defer s.RUnlock()

	res, found := s.rules[ruleID]
	return res, found
}

func (s *RulesWithContentStorage) SetRuleWithContent(
	ruleID types.RuleID, errorKey types.ErrorKey, ruleWithContent *types.RuleWithContent,
) {
	s.Lock()
	defer s.Unlock()

	s.rulesWithContent[ruleIDAndErrorKey{
		RuleID:   ruleID,
		ErrorKey: errorKey,
	}] = ruleWithContent
}

func (s *RulesWithContentStorage) SetRule(
	ruleID types.RuleID, ruleContent *ics_content.RuleContent,
) {
	s.Lock()
	defer s.Unlock()

	s.rules[ruleID] = ruleContent
}

var rulesWithContentStorage = RulesWithContentStorage{
	rulesWithContent: map[ruleIDAndErrorKey]*types.RuleWithContent{},
	rules:            map[types.RuleID]*ics_content.RuleContent{},
}

// GetRuleWithErrorKeyContent returns content for rule with provided `rule id` and `error key`.
// Caching is done under the hood, don't worry about it.
func GetRuleWithErrorKeyContent(
	ruleID types.RuleID, errorKey types.ErrorKey,
) (*types.RuleWithContent, error) {
	// to be sure data is there
	// it will wait only on opened channel
	// something like condition variable
	_, _ = <-ruleContentDirectoryReady

	res, found := rulesWithContentStorage.GetRuleWithErrorKeyContent(ruleID, errorKey)
	if !found {
		return nil, &types.ItemNotFoundError{ItemID: fmt.Sprintf("%v/%v", ruleID, errorKey)}
	}

	return res, nil
}

// GetRuleContent returns content for rule with provided `rule id`
// Caching is done under the hood, don't worry about it.
func GetRuleContent(ruleID types.RuleID) (*ics_content.RuleContent, error) {
	// to be sure data is there
	// it will wait only on opened channel
	// something like condition variable
	_, _ = <-ruleContentDirectoryReady

	res, found := rulesWithContentStorage.GetRuleContent(ruleID)
	if !found {
		return nil, &types.ItemNotFoundError{ItemID: ruleID}
	}

	return res, nil
}

// RunUpdateContentLoop runs loop which updates rules content by ticker
func RunUpdateContentLoop(servicesConf services.Configuration) {
	ticker := time.NewTicker(servicesConf.GroupsPollingTime)

	for {
		updateContent(servicesConf)
		<-ticker.C
	}
}

func updateContent(servicesConf services.Configuration) {
	var err error

	ruleContentDirectory, err = services.GetContent(servicesConf)
	if err != nil {
		log.Error().Err(err).Msg("Error retrieving static content")
		return
	}

	loadRuleContent(ruleContentDirectory)

	close(ruleContentDirectoryReady)
}
