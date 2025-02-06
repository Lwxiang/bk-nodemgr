/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package conv ...
package conv

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// ToInt64 conv interface{} to int64.
func ToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, errors.New("value is nil")
	}

	i, done, err := convNormalTypeToInt64(value)
	if done {
		return i, err
	}

	number, err := convCustomTypeToInt64(value)
	if err != nil {
		return 0, err
	}

	return number, nil
}

// convCustomTypeToInt64 convert custom type to int64.
func convCustomTypeToInt64(value interface{}) (int64, error) {
	val := reflect.ValueOf(value)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, errors.New("value is nil pointer")
		}
		val = val.Elem()
	}

	// get the underlying value.
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number := val.Uint()
		if number > math.MaxInt64 {
			return 0, errors.New("value is overflow")
		}

		return int64(number), nil
	case reflect.Float32, reflect.Float64:
		number := val.Float()
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, errors.New("value is not a number")
		}

		err := floatToInt64BoundaryCheck(number)
		if err != nil {
			return 0, err
		}

		return int64(number), nil
	default:
		return 0, fmt.Errorf("cannot convert interface to int64, kind(%v)", val.Kind())
	}
}

// convNormalTypeToInt64 convert base type to int64.
func convNormalTypeToInt64(value interface{}) (int64, bool, error) {
	// this is the most common case, but it can't handle custom types.
	switch v := value.(type) {
	case int64:
		return v, true, nil
	case int:
		return int64(v), true, nil
	case int32:
		return int64(v), true, nil
	case int16:
		return int64(v), true, nil
	case int8:
		return int64(v), true, nil
	case uint:
		if v > math.MaxInt64 {
			return 0, true, errors.New("value is overflow")
		}

		return int64(v), true, nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, true, errors.New("value is overflow")
		}

		return int64(v), true, nil
	case uint32:
		return int64(v), true, nil
	case uint16:
		return int64(v), true, nil
	case uint8:
		return int64(v), true, nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, true, errors.New("value is not a number")
		}

		err := floatToInt64BoundaryCheck(v)
		if err != nil {
			return 0, true, err
		}

		return int64(v), true, nil
	case float32:
		err := floatToInt64BoundaryCheck(float64(v))
		if err != nil {
			return 0, true, err
		}

		return int64(v), true, nil
	case string:
		if len(v) == 0 {
			return 0, true, errors.New("empty string")
		}

		number, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, true, err
		}

		return number, true, nil
	case json.Number:
		number, err := v.Int64()
		if err != nil {
			return 0, true, err
		}

		return number, true, nil
	}

	return 0, false, nil
}

func floatToInt64BoundaryCheck(v float64) error {
	// notice: float can't be lossless converted to int64. so we shrank accuracy.
	if v >= math.MaxInt64 {
		return errors.New("value is overflow")
	}

	if v <= math.MinInt64 {
		return errors.New("value is underflow")
	}
	return nil
}

// ToInt64Default conv interface{} to int64, return defaultVal if error.
func ToInt64Default(value interface{}, defaultVal int64) int64 {
	val, err := ToInt64(value)
	if err != nil {
		return defaultVal
	}

	return val
}
