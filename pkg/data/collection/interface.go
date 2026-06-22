//
// MIT License
//
// Copyright (c) 2026 Nicolas JUHEL
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

// Package collection provides a generic and thread-safe container for the
// model instances handled by the auditor.
//
// A Collection[M] stores the instances of type M in a map, keyed by their
// audids.ID identifier, and protects every access with a sync.RWMutex. The
// generic type parameter M allows the same implementation to manage the
// modules, the packages and the entries of the auditor.
//
// # Operations
//
// The collection exposes the following operations:
//   - Load: retrieves an instance, and lazily creates a missing one.
//   - Store: stores an instance, and merges it into an existing one.
//   - Merge: merges an instance into an existing one, and delegates to Store.
//   - GetOrStore: retrieves an instance, and stores it when it is missing.
//   - Delete: removes an instance by its identifier.
//   - Len: returns the number of the stored instances.
//   - Walk: iterates over the stored instances, with an early termination.
//   - Keys: returns the identifiers of the stored instances.
//   - Clean: empties the collection.
//
// The per-instance status mutators (SetStatus* and UnSetStatus*) address a
// single element of the collection by its identifier.
//
// # Serialization
//
// The collection supports the json, yaml, toml and cbor formats.
//
// # Thread safety
//
// Every read operation (Load, Len, Keys, Walk) and every write operation
// (Store, Merge, GetOrStore, Delete, Clean) is protected by the read-write
// mutex of the collection.
//
// The Marshal* methods are the only exception: they read the underlying map
// without acquiring the read lock. A concurrent modification of the
// collection is therefore a data race, and the caller must serialize the
// serialization with the modifications by itself.
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
	// This method attempts to retrieve an existing model instance from the
	// collection using its unique identifier. If no instance exists with the
	// specified ID, it creates and returns a new instance using the constructor
	// function provided during the creation of the collection. If the existing
	// instance is empty (as determined by IsEmpty), it also returns a new
	// instance.
	//
	// The key feature here is the **lazy initialization** of missing instances.
	// When Load is called with an ID that doesn't exist, it returns a brand new
	// instance created by calling the constructor function. This approach
	// ensures that only valid, populated instances are stored in the collection.
	//
	// Thread safety: This method acquires a read lock before accessing the
	// underlying map, ensuring that concurrent reads do not cause data races
	// while allowing multiple readers to access the collection simultaneously.
	//
	// Parameters:
	//   - id: The unique identifier of the model instance to retrieve.
	//     Type: audids.ID
	//     Purpose: Identifies the specific model instance to retrieve from the
	//     collection.
	//     How used: Passed directly to the map lookup, allowing for O(1)
	//     average-case retrieval of instances.
	//
	// Returns:
	//   - M: The model instance associated with the given ID.
	//     Construction: If an instance exists with the specified ID and is not
	//     empty, it's retrieved directly from the map. If not, a new instance is
	//     created by calling the constructor function provided during the
	//     creation of the collection.
	//
	// Notes:
	//   - If the ID is invalid (0), returns an empty instance (or nil for
	//     nilable types).
	//   - If the instance is empty (IsEmpty == true), returns a new instance
	//     created by the constructor function.
	Load(audids.ID) M

	// Store saves a model instance in the collection.
	//
	// This method updates the collection by storing a new or an existing model
	// instance under the specified ID. If the instance is empty (IsEmpty ==
	// true), the operation is ignored to avoid storing invalid data.
	//
	// The key behavior here is the **conditional update** mechanism. If an
	// instance already exists with the given ID, the method attempts to merge
	// the new data into the existing instance using the data instance's Merge
	// method. If the merge fails or the existing instance is not a
	// data.Data type, the new instance replaces the existing one. If no
	// instance exists, the new instance is stored directly.
	//
	// Thread safety: This method acquires a write lock before modifying the
	// underlying map, ensuring exclusive access during storage operations to
	// maintain consistency across concurrent modifications.
	//
	// Parameters:
	//   - val: The model instance to store in the collection.
	//     Type: M
	//     Purpose: Provides the new or updated data to be stored in the
	//     collection.
	//     How used: Passed to the map update, allowing the instance to be
	//     stored or updated directly.
	//
	// Returns:
	//   - ok: true when the instance has been stored, false when the operation
	//     has been ignored because the instance is empty, because its ID is
	//     lower than 1, or because it does not implement data.Data.
	//   - up: true when the instance has been merged into an already stored
	//     instance, false when it has been stored as a new instance or when it
	//     has replaced an existing one.
	//
	// Notes:
	//   - If the instance is empty (IsEmpty == true), the operation is ignored.
	//   - If the instance ID is less than 1, the operation is ignored.
	//   - If the instance is not a data.Data type, the operation is ignored.
	//   - If an existing instance exists and is a data.Data type, the method
	//     attempts to merge the new data into the existing instance. If the merge
	//     succeeds, the existing instance is kept; otherwise, it is replaced.
	Store(M) (ok bool, up bool)

	// Merge updates an existing model instance in the collection with new data.
	//
	// This method delegates to Store, which handles the merge logic. If the
	// instance exists, the new data is merged into the existing instance using
	// the data instance's Merge method. If the instance does not exist, it is
	// created with the provided data.
	//
	// The key behavior here is the **merge operation**. When merging data into
	// an existing instance, the method uses the data instance's Merge method to
	// combine the new data with the existing data. This approach allows for
	// complex, structured data types to be updated in a thread-safe manner.
	//
	// Thread safety: This method acquires a write lock before modifying the
	// underlying map, ensuring exclusive access during merge operations to
	// maintain consistency across concurrent modifications.
	//
	// Parameters:
	//   - val: The model instance with updated data to merge into the
	//     collection.
	//     Type: M
	//     Purpose: Provides the new or updated data to be merged into the
	//     existing instances.
	//     How used: Passed to the map update, allowing the data to be merged
	//     into the existing instances directly.
	//
	// Returns:
	//   - ok: true when the instance has been stored, false when the operation
	//     has been ignored, with the same conditions as the Store method.
	//   - up: true when the instance has been merged into an already stored
	//     instance, false otherwise.
	//
	// Notes:
	//   - If the instance is empty (IsEmpty == true), the operation is ignored.
	//   - If the instance does not exist, it is created using the constructor
	//     function.
	Merge(val M) (ok bool, up bool)

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

	// SetStatusNone clears all status bits for the model instance.
	//
	// This method resets the status to the default "None" state where no
	// processing phase is active. After this call, all other status bits
	// are cleared and the instance returns to its initial state.
	SetStatusNone(audids.ID)

	// SetStatusParsed marks the AST parse phase as completed.
	//
	// This method sets bit 0 to indicate that the AST parsing phase has
	// successfully completed for the model instance.
	SetStatusParsed(audids.ID)

	// SetStatusVendor marks the AST vendor resolution phase as completed.
	//
	// This method sets bit 1 to indicate that the AST vendor resolution
	// phase has successfully completed for the model instance.
	SetStatusVendor(audids.ID)

	// SetStatusAnalyzed marks the summary analysis phase as completed.
	//
	// This method sets bit 2 to indicate that the summary analysis phase
	// has successfully completed for the model instance.
	SetStatusAnalyzed(audids.ID)

	// SetStatusAudited1 marks the first audit report as generated.
	//
	// This method sets bit 3 to indicate that the first audit report
	// has been generated for the model instance.
	SetStatusAudited1(audids.ID)

	// SetStatusAudited2 marks the second audit report as generated.
	//
	// This method sets bit 4 to indicate that the second audit report
	// has been generated for the model instance.
	SetStatusAudited2(audids.ID)

	// SetStatusAudited3 marks the third audit report as generated.
	//
	// This method sets bit 5 to indicate that the third audit report
	// has been generated for the model instance.
	SetStatusAudited3(audids.ID)

	// SetStatusAudited4 marks the fourth audit report as generated.
	//
	// This method sets bit 6 to indicate that the fourth audit report
	// has been generated for the model instance.
	SetStatusAudited4(audids.ID)

	// SetStatusOutdated marks the item as out of date.
	//
	// This method sets bit 7 to indicate that the model instance has
	// been flagged as out of date.
	SetStatusOutdated(audids.ID)

	// UnSetStatusParsed clears the AST parse phase status bit.
	//
	// This method clears bit 0 from the model instance status, removing
	// the indication that AST parsing has completed. Other status bits
	// remain unchanged.
	UnSetStatusParsed(audids.ID)

	// UnSetStatusVendor clears the AST vendor resolution phase status bit.
	//
	// This method clears bit 1 from the model instance status, removing
	// the indication that vendor resolution has completed. Other status
	// bits remain unchanged.
	UnSetStatusVendor(audids.ID)

	// UnSetStatusAnalyzed clears the summary analysis phase status bit.
	//
	// This method clears bit 2 from the model instance status, removing
	// the indication that summary analysis has completed. Other status
	// bits remain unchanged.
	UnSetStatusAnalyzed(audids.ID)

	// UnSetStatusAudited1 clears the first audit report status bit.
	//
	// This method clears bit 3 from the model instance status, removing
	// the indication that the first audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited1(audids.ID)

	// UnSetStatusAudited2 clears the second audit report status bit.
	//
	// This method clears bit 4 from the model instance status, removing
	// the indication that the second audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited2(audids.ID)

	// UnSetStatusAudited3 clears the third audit report status bit.
	//
	// This method clears bit 5 from the model instance status, removing
	// the indication that the third audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited3(audids.ID)

	// UnSetStatusAudited4 clears the fourth audit report status bit.
	//
	// This method clears bit 6 from the model instance status, removing
	// the indication that the fourth audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited4(audids.ID)

	// UnSetStatusOutdated clears the out of date status bit.
	//
	// This method clears bit 7 from the model instance status, removing
	// the out of date flag. Other status bits remain unchanged.
	UnSetStatusOutdated(audids.ID)
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
