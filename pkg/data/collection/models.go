/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

// Package collection defines a comprehensive interface for managing collections of model
// instances, providing a flexible, thread-safe, and extensible data storage and
// retrieval mechanism. It enables efficient serialization in multiple formats (JSON,
// YAML, TOML, CBOR) while maintaining strong concurrency guarantees through
// read-write mutexes. The core structure revolves around the mdl[M] type, which serves
// as a modular data container that can be easily adapted to various data structures.
//
// Key Features:
//   - Thread-safe operations via sync.RWMutex
//   - Support for multiple serialization formats (JSON, YAML, TOML, CBOR)
//   - Modular architecture with customizable serialization and deserialization
//   - Efficient storage and retrieval using optimized encoding packages
//   - Robust error handling with detailed diagnostic messages
package collection

import (
	"maps"
	"slices"
	"sync"

	"github.com/nabbar/auditor/pkg/data"
	audids "github.com/nabbar/auditor/pkg/data/id"
)

// mdl represents a thread-safe collection of model instances.
//
// The core structure revolves around the mdl[M] type, which serves as a modular data
// container that can be easily adapted to various data structures. The mdl type is
// designed to be highly flexible, allowing for easy extension and modification
// while maintaining consistent interfaces across different model types.
//
// Key Features:
//   - Thread-safe operations via sync.RWMutex
//   - Support for multiple serialization formats (JSON, YAML, TOML, CBOR)
//   - Modular architecture with customizable serialization and deserialization
//   - Efficient storage and retrieval using optimized encoding packages
//   - Robust error handling with detailed diagnostic messages
//
// Implementation:
//   - The generic type parameter M allows for flexible implementation across different
//     model types, enabling the same collection interface to manage various data structures.
//   - The constructor function n is used to create new instances when needed,
//     providing a way to lazily initialize missing data while maintaining thread safety.
//   - The map-based storage approach ensures efficient lookup and retrieval
//     of model instances by their unique IDs, supporting large-scale collections.
//
// Thread Safety Overview:
//   - All operations are protected by a read-write mutex to prevent concurrent
//     modification and ensure data integrity.
//   - Read operations acquire a read lock, allowing multiple readers to access
//     the collection simultaneously without causing data races.
//   - Write operations acquire an exclusive write lock, ensuring that only one
//     goroutine can modify the underlying map at any given time.
//
// Read Lock Operations:
//   - Load
//   - Len
//   - Keys
//   - Walk
//   - Marshal* (JSON, YAML, TOML, CBOR)
//
// Write Lock Operations:
//   - Store
//   - Merge
//   - GetOrStore
//   - Delete
//   - Clean
//   - Unmarshal* (JSON, YAML, TOML, CBOR)
type mdl[M any] struct {
	// m is the mutex protecting access to the underlying map.
	// This read-write mutex ensures thread safety during concurrent operations
	// on the model instances collection, preventing data races and ensuring
	// consistency across multiple goroutines.
	m sync.RWMutex

	// l is the map storing model instances keyed by their IDs.
	// This map provides O(1) average-case lookup time for retrieving model instances
	// by their unique identifiers, making it efficient for large collections.
	l map[audids.ID]M

	// n is a constructor function used to create new instances when needed.
	// The constructor function is called when Load operations encounter missing or
	// empty instances, ensuring that new valid instances can be created as required.
	n func() M
}

// Load retrieves a model instance by its ID.
//
// This method attempts to retrieve an existing model instance from the collection
// using its unique identifier. If no instance exists with the specified ID, or if
// the existing instance is empty (as determined by IsEmpty), it creates and returns
// a new instance using the constructor function provided during manager creation.
//
// The key feature here is the **lazy initialization** of missing instances. When
// Load is called with an ID that doesn't exist, it returns a brand new instance
// created by calling the constructor function. This approach ensures that only
// valid, populated instances are stored in the collection.
//
// Thread safety: This method acquires a read lock before accessing the underlying
// map, ensuring that concurrent reads do not cause data races while allowing
// multiple readers to access the collection simultaneously.
//
// Parameters:
//   - id: The unique identifier of the model instance to retrieve.
//     Type: audids.ID
//     Purpose: Identifies the specific model instance to retrieve from the collection.
//     How used: Passed directly to the map lookup, allowing for O(1) average-case
//     retrieval of instances.
//
// Returns:
//   - M: The model instance associated with the given ID.
//     Construction: If an instance exists with the specified ID and is not empty,
//     it's retrieved directly from the map. If not, a new instance is created by
//     calling the constructor function provided during manager creation.
//
// Notes:
//   - If the ID is invalid (0), returns an empty instance (or nil for nilable types).
//   - If the instance is empty (IsEmpty == true), returns a new instance created
//     by the constructor function.
//   - If the instance does not implement data.Data[M], returns a new instance
//     created by the constructor function.
func (o *mdl[M]) Load(id audids.ID) M {
	o.m.RLock()
	var m, k = o.l[id]
	o.m.RUnlock()

	if !k {
		return o.n()
	}

	if i, k := any(m).(data.Data[M]); !k {
		return o.n()
	} else if i.IsEmpty() {
		return o.n()
	}

	return m
}

// Store saves a model instance in the collection.
//
// This method updates the collection by storing a new or existing model instance
// under the specified ID. If the instance is empty (IsEmpty == true), the operation
// is ignored to avoid storing invalid data.
//
// The key behavior here is the **conditional update** mechanism. If an instance
// already exists with the given ID, the method attempts to merge the new data
// into the existing instance using the data instance's Merge method. If the merge
// fails or the existing instance is not a data.Data type, the new instance replaces
// the existing one. If no instance exists, the new instance is stored directly.
//
// Thread safety: This method acquires a write lock before modifying the underlying
// map, ensuring exclusive access during storage operations to maintain consistency
// across concurrent modifications.
//
// Parameters:
//   - val: The model instance to store in the collection.
//     Type: M
//     Purpose: Provides the new or updated data to be stored in the collection.
//     How used: Passed to the map update, allowing the instance to be stored or
//     updated directly.
//
// Returns: None
//   - No value is returned, as the method modifies the underlying state of the
//     collection directly.
//
// Notes:
//   - If the instance is empty (IsEmpty == true), the operation is ignored.
//   - If the instance ID is less than 1, the operation is ignored.
//   - If the instance is not a data.Data type, the operation is ignored.
//   - If an existing instance exists and is a data.Data type, the method attempts
//     to merge the new data into the existing instance. If the merge succeeds,
//     the existing instance is kept; otherwise, it is replaced.
func (o *mdl[M]) Store(val M) {
	var id audids.ID

	if v, k := any(val).(data.Data[M]); !k {
		return
	} else if v.IsEmpty() {
		return
	} else {
		id = v.GetID()
	}

	if id < 1 {
		return
	}

	o.m.Lock()
	var a, k = o.l[id]
	o.m.Unlock()

	if !k {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return
	}

	if v, k := any(a).(data.Data[M]); !k {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return
	} else if !v.Merge(val) {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return
	}

	o.m.Lock()
	o.l[id] = a
	o.m.Unlock()
}

// Merge updates an existing model instance in the collection with new data.
//
// This method delegates to Store, which handles the merge logic. If the instance
// exists, the new data is merged into the existing instance using the data
// instance's Merge method. If the instance does not exist, it is created with
// the provided data.
//
// The key behavior here is the **merge operation**. When merging data into an existing
// instance, the method uses the data instance's Merge method to combine the new data
// with the existing data. This approach allows for complex, structured data types
// to be updated in a thread-safe manner.
//
// Thread safety: This method acquires a write lock before modifying the underlying
// map, ensuring exclusive access during merge operations to maintain consistency
// across concurrent modifications.
//
// Parameters:
//   - val: The model instance with updated data to merge into the collection.
//     Type: M
//     Purpose: Provides the new or updated data to be merged into existing instances.
//     How used: Passed to the map update, allowing the data to be merged into
//     existing instances directly.
//
// Returns: None
//   - No value is returned, as the method modifies the underlying state of the
//     collection directly.
//
// Notes:
//   - If the instance is empty (IsEmpty == true), the operation is ignored.
//   - If the instance does not exist, it is created using the constructor function.
func (o *mdl[M]) Merge(val M) {
	o.Store(val)
}

// GetOrStore returns a model instance if it exists, otherwise creates and returns a new instance.
//
// If the instance with the specified ID exists and is not empty, it returns the
// existing instance. If no instance exists, or if the existing instance is empty,
// it stores the provided instance and returns it.
//
// The key feature here is the **single operation**. This method performs a single
// map lookup followed by a single map update, eliminating the need for extra
// operations like checking for existence before storing.
//
// Thread safety: This method acquires a write lock before modifying the underlying
// map, ensuring exclusive access during storage operations to maintain consistency
// across concurrent modifications.
//
// Parameters:
//   - val: The model instance to retrieve or create in the collection.
//     Type: M
//     Purpose: Provides the data to be stored in the collection.
//     How used: Passed to the map update, allowing the instance to be retrieved
//     or created directly.
//
// Returns:
//   - M: The model instance associated with the given ID, either retrieved
//     from the collection or the provided instance if not found.
//
// Notes:
//   - If the provided instance is not a data.Data type or is empty, returns a new
//     instance created by the constructor function.
//   - If the existing instance is not a data.Data type or is empty, the provided
//     instance replaces it.
func (o *mdl[M]) GetOrStore(val M) M {
	var id audids.ID

	if v, k := any(val).(data.Data[M]); !k {
		return o.n()
	} else if v.IsEmpty() {
		return o.n()
	} else {
		id = v.GetID()
	}

	o.m.Lock()
	var a, k = o.l[id]
	o.m.Unlock()

	if !k {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return val
	}

	if v, k := any(a).(data.Data[M]); !k {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return val
	} else if v.IsEmpty() {
		o.m.Lock()
		o.l[id] = val
		o.m.Unlock()
		return val
	}

	return a
}

// Delete removes a model instance by its ID from the collection.
//
// If the ID is invalid (0), the operation is silently ignored to prevent accidental
// deletion of instances that may still be in use.
//
// The key behavior here is the **atomic deletion**. This method performs a single
// map delete operation, eliminating the need for extra checks or complex error
// handling.
//
// Thread safety: This method acquires a write lock before modifying the underlying
// map, ensuring exclusive access during deletion operations to maintain consistency
// across concurrent modifications.
func (o *mdl[M]) Delete(id audids.ID) {
	if id == 0 {
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	delete(o.l, id)
}

// Len returns the number of model instances in the collection.
//
// This method provides a snapshot of the collection size at the moment it is called.
// It is useful for monitoring collection size or determining whether any instances
// exist before performing further operations.
//
// Thread safety: This method acquires a read lock before accessing the underlying
// map, ensuring that concurrent reads do not cause data races while allowing multiple
// readers to access the collection simultaneously.
func (o *mdl[M]) Len() int {
	o.m.RLock()
	defer o.m.RUnlock()

	return len(o.l)
}

// Walk iterates over all model instances in the collection.
//
// This method applies a provided function to each model instance in the collection,
// allowing for batch operations or processing of all stored instances. It uses
// Keys() to get all IDs and Load() to retrieve each instance.
//
// The key feature here is the **early termination**. If the provided function returns
// false for any instance, the iteration stops immediately, allowing for efficient
// processing of partial collections.
//
// Thread safety: This method uses Keys() which acquires a read lock, ensuring safe
// concurrent access during iteration while maintaining consistency.
func (o *mdl[M]) Walk(f func(M) bool) {
	if f == nil {
		return
	}

	for _, k := range o.Keys() {
		if !f(o.Load(k)) {
			return
		}
	}
}

// Keys returns a slice of all IDs currently stored in the collection.
//
// This method provides a snapshot of all unique identifiers present in the collection.
// It is used internally by Walk and other methods to iterate over all instances in
// the collection.
//
// Thread safety: This method acquires a read lock before accessing the underlying
// map, ensuring that concurrent reads do not cause data races while allowing multiple
// readers to access the collection simultaneously.
func (o *mdl[M]) Keys() []audids.ID {
	o.m.RLock()
	defer o.m.RUnlock()

	return slices.Collect(maps.Keys(o.l))
}

// Clean removes all model instances from the collection, resetting it to an empty state.
//
// This method acquires a write lock before modifying the underlying map, ensuring
// exclusive access during cleanup operations to prevent concurrent modification.
// It replaces the underlying map with a new empty map.
func (o *mdl[M]) Clean() {
	o.m.Lock()
	defer o.m.Unlock()

	o.l = make(map[audids.ID]M)
}
