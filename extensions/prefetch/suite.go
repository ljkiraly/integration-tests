// Copyright (c) 2021 Doc.ai and/or its affiliates.
//
// Copyright (c) 2022 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package prefetch exports suite that can do prefetch of required images once per suite.
package prefetch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"

	"github.com/ljkiraly/integration-tests/extensions/prefetch/images"
	"github.com/networkservicemesh/gotestmd/pkg/suites/shell"
)

// Config is env config to setup images prefetching.
type Config struct {
	ImagesPerDaemonset int    `default:"10" desc:"Number of images created per DaemonSet" split_words:"true"`
	Timeout            string `default:"10m" desc:"Kubectl rollout status timeout for the DaemonSet" split_words:"true"`
}

// Suite creates `prefetch` daemonset which pulls all test images for all cluster nodes.
type Suite struct {
	shell.Suite
	SourcesURLs []string
}

var once sync.Once

// SetupSuite prefetches docker images for each k8s node.
func (s *Suite) SetupSuite() {
	once.Do(s.initialize)
}

func (s *Suite) initialize() {
	var config Config
	require.NoError(s.T(), envconfig.Usage("prefetch", &config))
	require.NoError(s.T(), envconfig.Process("prefetch", &config))

	prefetchImages := images.ReteriveList(s.SourcesURLs, func(s string) bool {
		return strings.HasSuffix(s, ".yaml") && !IsExcluded(s)
	}).Images

	prefetchImages = removeDuplicates(prefetchImages)

	wd, err := os.Getwd()
	require.NoError(s.T(), err)
	tmpDir := filepath.Join(wd, uuid.NewString())
	require.NoError(s.T(), os.MkdirAll(tmpDir, 0750))
	s.T().Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	r := s.Runner(tmpDir)

	var daemonSets []string
	for d := 0; d*config.ImagesPerDaemonset < len(prefetchImages); d++ {
		var containers string
		for c := 0; c < config.ImagesPerDaemonset && d*config.ImagesPerDaemonset+c < len(prefetchImages); c++ {
			containers += container(uuid.NewString(), prefetchImages[d*config.ImagesPerDaemonset+c])
		}

		r.Run(createDaemonSet(d, containers))

		daemonSets = append(daemonSets, fmt.Sprintf("prefetch-%d", d))
	}

	r.Run("kubectl create ns prefetch")
	s.T().Cleanup(func() {
		r.Run("kubectl describe pods -n prefetch")
		r.Run("kubectl delete ns prefetch")
	})

	var wg sync.WaitGroup
	for _, daemonSet := range daemonSets {
		wg.Add(1)
		go func(daemonSet string) {
			defer wg.Done()

			dr := s.Runner(tmpDir)
			dr.Run(fmt.Sprintf("kubectl -n prefetch apply -f %s.yaml", daemonSet))
			dr.Run(fmt.Sprintf("kubectl -n prefetch rollout status daemonset/%s --timeout=%s", daemonSet, config.Timeout))
			dr.Run(fmt.Sprintf("kubectl -n prefetch delete -f %s.yaml", daemonSet))
		}(daemonSet)
	}
	wg.Wait()
}

func removeDuplicates(source []string) []string {
	var allKeys = make(map[string]bool)
	var result []string
	for _, item := range source {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			result = append(result, item)
		}
	}
	return result
}
