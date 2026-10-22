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

// Package collection provides a generic, thread-safe, map-backed collection for
// storing and managing model instances that implement the data.Data[M] interface.
//
// # Overview
//
// A Collection[M] holds instances of type M in a Go map keyed by their
// audids.ID. Every access to the underlying map is synchronized through a
// sync.RWMutex. The generic type parameter M allows the same implementation
// to manage modules, packages, or code entries without code duplication.
//
// The package exposes two public symbols:
//   - New[M](func() M): the constructor function that creates a thread-safe
//     collection, optionally backed by a user-supplied instance factory.
//   - Collection[M]: the interface that declares every operation on the
//     collection and is also the return type of New.
//
// The non-exported type mdl[M any] is the concrete implementation of
// Collection[M]. It is not exported; callers interact exclusively through the
// interface returned by New.
//
// # Prerequisites
//
// Models stored in a collection must implement the data.Data[M] interface.
// This interface guarantees the presence of GetID(), IsEmpty(), and Merge().
// All mutating and querying operations that go beyond simple map access rely
// on these methods to validate, initialize, and merge instances.
//
// # Public API
//
// The Collection[M] interface declares the following public methods:
//
// Retrieval:
//   - Load(id audids.ID) M: returns the instance for the given ID, or a new
//     instance created by the constructor function when the ID is unknown
//     or the existing instance is empty.
//   - GetOrStore(val M) M: returns the existing instance for val's ID if it
//     exists and is not empty; otherwise stores and returns val.
//   - Keys() []audids.ID: returns a slice of all IDs currently in the
//     collection.
//   - Len() int: returns the number of instances in the collection.
//   - Walk(f func(M) bool): calls f for each instance until f returns false.
//
// Modification:
//   - Store(val M) (ok bool, up bool): stores or merges val into the
//     collection. Returns (false, false) when the operation is rejected,
//     (true, false) when a new entry was created or an existing one was
//     replaced, and (true, true) when an existing instance was merged.
//   - Merge(val M) (ok bool, up bool): delegates to Store with the same
//     return semantics.
//   - Delete(id audids.ID): removes the instance with the given ID. No-op
//     if the ID is zero.
//   - Clean(): replaces the underlying map with an empty one.
//
// Status mutators (operating on a single element identified by its ID):
//   - SetStatusNone(id), SetStatusParsed(id), SetStatusVendor(id),
//     SetStatusAnalyzed(id), SetStatusAudited1-4(id), SetStatusOutdated(id):
//     each sets one status bit on the instance addressed by id.
//   - UnSetStatusParsed(id), UnSetStatusVendor(id), UnSetStatusAnalyzed(id),
//     UnSetStatusAudited1-4(id), UnSetStatusOutdated(id):
//     each clears one status bit on the instance addressed by id.
//   These methods are no-ops when the ID is unknown, the element is empty,
//   or the element does not implement data.Data.
//
// Serialization (Marshal and Unmarshal for four formats:
// json, yaml, toml, cbor, plus the encoding.BinaryMarshaler / Unmarshaler):
//   - MarshalJSON, MarshalYAML, MarshalTOML, MarshalCBOR
//   - UnmarshalJSON, UnmarshalYAML, UnmarshalTOML, UnmarshalCBOR
//   - MarshalBinary, UnmarshalBinary (delegate to CBOR)
//
// # Data Structure
//
// The mdl[M] type holds three internal fields:
//   - m sync.RWMutex: protects all map access; a shared read lock is
//     acquired for read operations, an exclusive write lock for write
//     operations.
//   - l map[audids.ID]M: the underlying storage, indexed by the unique
//     identifier of each model instance.
//   - n func() M: the constructor function used by Load and GetOrStore
//     to create instances when needed. It is set once by New.
//
// # Thread Safety
//
// Every read operation (Load, Len, Keys, Walk) acquires a shared read lock
// before accessing the underlying map, allowing multiple concurrent readers.
// Every write operation (Store, Merge, GetOrStore, Delete, Clean, all
// status mutators, and all Unmarshal methods) acquires an exclusive write
// lock.
//
// The Marshal* methods do not acquire any lock. They read the map directly,
// which means a concurrent modification is a data race. The caller must
// serialize Marshal* calls with any concurrent write operations.
//
// # Lazy Initialization
//
// Load uses the constructor function n when the requested ID is not in the
// map or when the existing value for that ID reports IsEmpty(). This enables
// a caller to query for an instance without having to check existence first.
// When New is called with a nil constructor, the default constructor returns
// the zero value of M.
//
// # Store Semantics
//
// Store performs a conditional update. Before storing, it validates that:
//   - val implements data.Data[M],
//   - val is not empty (IsEmpty returns false),
//   - the ID returned by val.GetID() is at least 1.
// If any of these conditions fails, the operation is rejected and
// (false, false) is returned.
//
// When the key exists in the map, Store attempts to merge val into the
// existing instance by calling Merge on the existing element. If Merge
// returns true, the existing instance is kept in the map and (true, true)
// is returned. If Merge returns false or the existing element does not
// implement data.Data[M], val replaces the existing element and
// (true, false) is returned. When the key is new, val is stored directly
// and (true, false) is returned.
//
// Merge delegates entirely to Store and returns the same result.
//
// # GetOrStore Semantics
//
// GetOrStore first validates val the same way Store does. If val is
// rejected, a new instance from the constructor is returned without
// storing anything. If val passes validation, GetOrStore acquires a write
// lock and checks the map. If the key does not exist, val is stored and
// returned. If the key exists but the existing element is not a data.Data or
// is empty, val replaces it and is returned. Otherwise the existing element
// is returned unchanged.
//
// # Walk Semantics
//
// Walk calls Keys to obtain the current set of IDs, then Load to retrieve
// each instance. The callback f is called with each instance in turn. If f
// returns false, Walk stops immediately. A nil callback is silently ignored.
//
// # Serialization Details
//
// Each Marshal method serializes the underlying map directly. The map keys
// (audids.ID) become the serialized keys. Each model instance is serialized
// by its own Marshaler implementation (the data.Data interface embeds the
// marshaler contracts for all four formats).
//
// UnmarshalJSON, UnmarshalYAML, and UnmarshalCBOR replace the entire
// underlying map. UnmarshalCBOR takes extra care: it first unmarshals the
// input into a map[audids.ID]any, then for each value it marshals it to CBOR
// and feeds it to the model's own UnmarshalCBOR, ensuring the value is
// reconstructed as the correct concrete type. UnmarshalTOML accepts either
// []byte or string input.
//
// MarshalBinary and UnmarshalBinary are convenience methods that delegate
// to MarshalCBOR and UnmarshalCBOR respectively.
//
// # Status Bit Field
//
// Each model instance carries a status bit field (audsts.Status) that tracks
// the completion of processing phases. The bits are defined as follows:
//   - bit 0 (Parsed): AST parsing phase completed.
//   - bit 1 (Vendor): vendor resolution phase completed.
//   - bit 2 (Analyzed): summary analysis phase completed.
//   - bit 3 (Audited1): first audit report generated.
//   - bit 4 (Audited2): second audit report generated.
//   - bit 5 (Audited3): third audit report generated.
//   - bit 6 (Audited4): fourth audit report generated.
//   - bit 7 (Outdated): instance is out of date.
//
// The status mutator methods on the Collection interface (SetStatusP*,
// UnSetStatusP*, SetStatusNone) operate on a single element identified by
// its ID. They look up the element, verify it implements data.Data and is
// not empty, then call the corresponding method on the element and write it
// back into the map.
//
// # Data Flow
//
// An auditor pipeline populates a collection through the following stages:
//
//   1. The collection is created via New, providing a constructor that knows
//      how to build a new instance of the model type. Typically this is
//      called once per kind of model (module collection, package collection,
//      code-entry collection).
//
//   2. Model instances are created externally (by the scanner or loader) and
//      stored through Store. Store merges compatible data into existing
//      instances and rejects empty or invalid ones, so the collection always
//      holds usable data.
//
//   3. As the pipeline progresses, the status bits on each model are set
//      through the SetStatusP methods on the collection (e.g.,
//      SetStatusParsed(id) after AST parsing completes).
//
//   4. When a model is fully audited, its data can be serialized for
//      persistence using any of the Marshal methods. After serialization,
//      the collection can be discarded or cleaned for the next run.
//
//   5. To restore a collection from a serialized snapshot, construct a new
//      collection and call the corresponding Unmarshal method. The instances
//      are reconstructed from the serialized data.
//
// The flow can be summarized as:
//
//   New(constructor) --> Store/Merge --> SetStatusP --> Marshal -> persistence
//   Unmarshal -> Load --> SetStatusP --> (repeat until fully audited) -> Marshal

package collection
