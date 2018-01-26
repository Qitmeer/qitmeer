// Copyright 2017 The nox-project Authors
// This file is part of the nox-project library.
//
// The nox-project library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The nox-project library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the nox-project library. If not, see <http://www.gnu.org/licenses/>.

package graph

import (
	"testing"
)

func TestIfIncluded(t *testing.T) {
	earlier_noxes := []string{"a", "b", "c"}
	current_nox := "c"

	ok := IfIncluded(earlier_noxes, current_nox)
	if !ok {
		t.Errorf("Expected %v included %s got %v", earlier_noxes, current_nox, ok)
	}
}
