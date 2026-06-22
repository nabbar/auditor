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
	"encoding/json"
	"sync"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Collection represents a generic interface for managing model instance collections,
// providing a standardized set of operations that can be implemented for various
// data structures while maintaining thread safety and serialization capabilities.
// The generic type parameter M allows for flexible implementation across different
// model types while maintaining consistent interfaces.
//
// The primary operational methods include Load, Store, Merge, GetOrStore, Delete, Len, and Walk.
// Load allows retrieval of instances by ID, while Store updates existing instances or creates new ones.
// Merge enables updating existing instances with new data, and GetOrStore provides a single operation
// for retrieving or creating instances. Delete removes instances by ID, and Len provides the collection size.
// Walk enables batch processing of all instances, and Clean resets the collection to an empty state.
// All operations are designed to be thread-safe, ensuring consistent access patterns across concurrent calls.
//
// Key Implementation Details:
//   - Load uses a lazy initialization mechanism, creating new instances when needed
//   - Store performs conditional updates, ignoring empty instances and attempting to merge with existing data
//   - Merge delegates to Store, combining new data with existing instances using data instance-specific methods
//   - GetOrStore merges retrieval and creation into a single operation, returning the stored or existing instance
//   - Delete performs atomic removal using a single map delete operation
//   - Len provides a snapshot of the collection size
//   - Walk enables early termination, allowing for efficient partial collection processing
//
// Interface Methods:
//
//	Load(id audids.ID) M - Retrieves an instance by ID, creating a new instance if missing
//	Store(val M) - Updates an instance or creates a new one if it doesn't exist
//	Merge(val M) - Updates an existing instance with new data (delegates to Store)
//	GetOrStore(val M) M - Retrieves an instance or creates a new one if it doesn't exist
//	Delete(id audids.ID) - Removes an instance by ID
//	Len() int - Returns the collection size
//	Walk(f func(M) bool) - Iterates over all instances, allowing for early termination
//	Keys() []audids.ID - Returns a slice of all IDs in the collection
//
// Type Definition:
//
//	type Collection[M any] interface {
//	    // ... implementation details
//	}
//
// Implementation:
//
//	type mdl[M any] struct {
//	    m sync.RWMutex
//	    l map[audids.ID]M
//	    n func() M
//	}
//
//	func New[M any](n func() M) Collection[M] - Creates a new collection instance
//
// Serialization Capabilities:
//   - JSON.Marshaler - Serializes to JSON format
//   - json.Unmarshaler - Deserializes from JSON format
//   - yaml.Marshaler - Serializes to YAML format
//   - yaml.Unmarshaler - Deserializes from YAML format
//   - toml.Marshaler - Serializes to TOML format
//   - toml.Unmarshaler - Deserializes from TOML format
//   - cbor.Marshaler - Serializes to CBOR format
//   - cbor.Unmarshaler - Deserializes from CBOR format
//
// Thread Safety:
//   - All operations are protected by a read-write mutex
//   - Read operations acquire a read lock, allowing multiple concurrent readers
//   - Write operations acquire exclusive write access
//
// Error Handling:
//   - Detailed diagnostic messages for deserialization failures
//   - Early detection of invalid data formats
//   - Specific error messages for each failure case
type Collection[M any] interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// Load retrieves a model instance by its ID.
	//
	// This method attempts to retrieve an existing model instance from the collection
	// using its unique identifier. If no instance exists with the specified ID, it
	// creates and returns a new instance using the constructor function provided
	// during manager creation. If the existing instance is empty (as determined by
	// IsEmpty), it also returns a new instance.
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
	Load(audids.ID) M

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
	Store(M)

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
	Merge(val M)

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
	GetOrStore(val M) M

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
	Delete(audids.ID)

	// Len returns the number of model instances in the collection.
	//
	// This method provides a snapshot of the collection size at the moment it is called.
	// It is useful for monitoring collection size or determining whether any instances
	// exist before performing further operations.
	//
	// Thread safety: This method acquires a read lock before accessing the underlying
	// map, ensuring that concurrent reads do not cause data races while allowing multiple
	// readers to access the collection simultaneously.
	Len() int

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
	Walk(func(M) bool)

	// Keys returns a slice of all IDs currently stored in the collection.
	//
	// This method provides a snapshot of all unique identifiers present in the collection.
	// It is used internally by Walk and other methods to iterate over all instances in
	// the collection.
	//
	// Thread safety: This method acquires a read lock before accessing the underlying
	// map, ensuring that concurrent reads do not cause data races while allowing multiple
	// readers to access the collection simultaneously.
	Keys() []audids.ID

	// Clean removes all model instances from the collection, resetting it to an empty state.
	//
	// This method acquires a write lock before modifying the underlying map, ensuring
	// exclusive access during cleanup operations to prevent concurrent modification.
	// It replaces the underlying map with a new empty map.
	Clean()
}

// New creates and returns a new Collection instance for managing model instances of type M.
//
// This function initializes and returns a new Collection implementation that handles
// collections of model instances. It accepts an optional constructor function that
// is used to create new instances when needed (e.g., during Load operations for missing
// instances). If no constructor function is provided, a default one is used that creates
// an uninitialized instance.
//
// The returned Collection is thread-safe and suitable for concurrent access.
//
// Parameters:
//   - n: Optional constructor function to create new instances. If nil, a default
//     constructor is provided that returns an uninitialized instance of type M.
//
// Returns:
//   - Collection[M]: A thread-safe collection implementation capable of managing
//     model instances of type M.
func New[M any](n func() M) Collection[M] {
	if n == nil {
		n = func() M {
			var v M
			return v
		}
	}

	return &mdl[M]{
		m: sync.RWMutex{},
		l: make(map[audids.ID]M),
		n: n,
	}
}
