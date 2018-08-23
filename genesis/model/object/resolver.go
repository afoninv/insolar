/*
 *    Copyright 2018 INS Ecosystem
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package object

import (
	"fmt"
)

// Resolver marks that instance have ability to get proxy objects by its reference.
type Resolver interface {
	GetObject(reference interface{}, cls interface{}) (interface{}, error)
}

func checkClass(class Proxy, expected interface{}) error {
	if class == expected {
		return nil
	}
	if expected == nil {
		_, okF := class.(Factory)
		_, okCF := class.(CompositeFactory)
		if okF || okCF {
			return nil
		}

	}

	return fmt.Errorf("instance class is not equal received")
}
