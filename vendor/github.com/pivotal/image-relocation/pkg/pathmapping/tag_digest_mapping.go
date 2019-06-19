/*
 * Copyright (c) 2019-Present Pivotal Software, Inc. All rights reserved.
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
	"github.com/pivotal/image-relocation/pkg/image"
)

// FlattenRepoPathPreserveTagDigest maps the given Name to a new Name with a given repository prefix.
// It aims to avoid collisions between repositories and to include enough of the original name
// to make it recognizable by a human being. It preserves any tag and/or digest.
func FlattenRepoPathPreserveTagDigest(repoPrefix string, originalImage image.Name) image.Name {
	rn := FlattenRepoPath(repoPrefix, originalImage)

	// Preserve any tag
	if tag := originalImage.Tag(); tag != "" {
		var err error
		rn, err = rn.WithTag(tag)
		if err != nil {
			panic(err) // should never occur
		}
	}

	// Preserve any digest
	if dig := originalImage.Digest(); dig != image.EmptyDigest {
		var err error
		rn, err = rn.WithDigest(dig)
		if err != nil {
			panic(err) // should never occur
		}
	}

	return rn
}
