/*
 * Copyright (c) 2018-Present Pivotal Software, Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pathmapping

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"

	"github.com/pivotal/image-relocation/pkg/image"
)

// PathMapping is a type of function which maps a given Name to a new Name by apply a repository prefix.
type PathMapping func(repoPrefix string, originalImage image.Name) image.Name

// FlattenRepoPath maps the given Name to a new Name with a given repository prefix.
// It aims to avoid collisions between repositories and to include enough of the original name
// to make it recognizable by a human being.
func FlattenRepoPath(repoPrefix string, originalImage image.Name) image.Name {
	hasher := md5.New()
	hasher.Write([]byte(originalImage.Name()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	available := reference.NameTotalLengthMax - len(mappedPath(repoPrefix, "", hash))
	fp := flatPath(originalImage.Path(), available)
	var mp string
	if fp == "" {
		mp = fmt.Sprintf("%s/%s", repoPrefix, hash)
	} else {
		mp = mappedPath(repoPrefix, fp, hash)
	}
	mn, err := image.NewName(mp)
	if err != nil {
		panic(err) // should not happen
	}
	return mn
}

func mappedPath(repoPrefix string, repoPath string, hash string) string {
	return fmt.Sprintf("%s/%s-%s", repoPrefix, repoPath, hash)
}

func flatPath(repoPath string, size int) string {
	return strings.Join(crunch(strings.Split(repoPath, "/"), size), "-")
}

func crunch(components []string, size int) []string {
	for n := len(components); n > 0; n-- {
		comp := reduce(components, n)
		if len(strings.Join(comp, "-")) <= size {
			return comp
		}

	}
	if len(components) > 0 && len(components[0]) <= size {
		return []string{components[0]}
	}
	return []string{}
}

func reduce(components []string, n int) []string {
	if len(components) < 2 || len(components) <= n {
		return components
	}

	tmp := make([]string, len(components))
	copy(tmp, components)

	last := components[len(tmp)-1]
	if n < 2 {
		return []string{last}
	}

	front := tmp[0 : n-1]
	return append(front, "-", last)
}
