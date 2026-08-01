// Copyright 2026 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package textinput

import "testing"

func TestSessionQueuesInputBetweenStarts(t *testing.T) {
	var s session

	first, _ := s.start()
	s.send(textInputState{Text: "\uD55C", Committed: true})
	s.end()

	if got := <-first; got.Text != "\uD55C" || !got.Committed {
		t.Fatalf("first state = %+v, want committed Hangul", got)
	}
	if _, ok := <-first; ok {
		t.Fatal("first session channel remained open")
	}

	// Windows can deliver the ASCII character that ended an IME composition
	// after the composition's session has already closed. Preserve it for the
	// session that Field starts on the next tick.
	s.send(textInputState{Text: ".", Committed: true})
	second, end := s.start()
	defer end()

	if got := <-second; got.Text != "." || !got.Committed {
		t.Fatalf("queued state = %+v, want committed period", got)
	}
}

func TestSessionPreservesQueuedInputOrder(t *testing.T) {
	var s session
	for _, text := range []string{"a", "b", "c"} {
		s.send(textInputState{Text: text, Committed: true})
	}

	ch, end := s.start()
	defer end()
	for _, want := range []string{"a", "b", "c"} {
		if got := <-ch; got.Text != want {
			t.Fatalf("queued state = %q, want %q", got.Text, want)
		}
	}
}

func TestSessionDropsQueuedInputWhenCleared(t *testing.T) {
	var s session
	s.send(textInputState{Text: "stale", Committed: true})
	s.clearQueue()

	ch, end := s.start()
	defer end()
	select {
	case got := <-ch:
		t.Fatalf("unexpected queued state after clear: %+v", got)
	default:
	}
}
