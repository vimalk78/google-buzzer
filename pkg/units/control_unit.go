// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package units implements the business logic to make the fuzzer work
package units

import (
	"fmt"

	"buzzer/pkg/strategies/parse_verifier"
	"buzzer/pkg/strategies/playground"
	"buzzer/pkg/strategies/pointer_arithmetic"
	"buzzer/pkg/strategies/stack_corruption"
	"buzzer/pkg/strategies"
)

// RunMode are the modes of operation for the server.
type RunMode string

// StrategyInterface contains all the methods that a fuzzing strategy should
// implement.
type StrategyInterface interface {
	Fuzz(e strategies.ExecutorInterface, cm strategies.CoverageManager) error
}

// ControlUnit directs the execution of the fuzzer.
type ControlUnit struct {
	strat StrategyInterface
	ex    strategies.ExecutorInterface
	rm    RunMode
	cm    strategies.CoverageManager
	rdy   bool
}

// Init prepares the control unit to be used.
func (cu *ControlUnit) Init(executor strategies.ExecutorInterface, coverageManager strategies.CoverageManager, runMode, fuzzStrategyFlag string) error {
	cu.ex = executor

	switch fuzzStrategyFlag {
	case parseverifier.StrategyName:
		cu.strat = &parseverifier.StrategyParseVerifierLog{}
	case pointerarithmetic.StrategyName:
		cu.strat = &pointerarithmetic.Strategy{
			// 60 is an arbitrary number.
			InstructionCount: 60,
		}
	case playground.StrategyName:
		cu.strat = &playground.Strategy{}
	case stackcorruption.StrategyName:
		cu.strat = &stackcorruption.Strategy{}
	default:
		return fmt.Errorf("unknown fuzzing strategy: %s", fuzzStrategyFlag)
	}

	cu.rdy = true
	return nil
}

// IsReady indicates to the caller if the ControlUnit is initialized successully.
func (cu *ControlUnit) IsReady() bool {
	return cu.rdy
}

// RunFuzzer kickstars the fuzzer in the mode that was specified at Init time.
func (cu *ControlUnit) RunFuzzer() error {
	return cu.strat.Fuzz(cu.ex, cu.cm)
}
