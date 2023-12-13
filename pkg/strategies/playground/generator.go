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

package playground

import (
	. "buzzer/pkg/ebpf"
)

// Generator is responsible for constructing the ast for this strategy.
type Generator struct{}

// Generate is the main function that builds the ast for this strategy.
func (g *Generator) Generate() []Instruction {
	root, _ := InstructionSequence(
		Mov64(RegR0, 0),
		Exit(),
	)
	return root
}
