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

// Package manager provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
package manager

import (
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// EntNew creates a new Entry within a package using the specified parameters.
//
// This method directly creates a new entry (code element) using the audent.New function,
// bypassing any registered callback functions since this implementation provides
// the default behavior for entry creation. Entries represent individual code constructs
// such as functions, variables, or types that are part of a package.
//
// Parameters:
//   - pkg: The package ID to associate with the new entry (provides context for the entry)
//   - name: The name of the source element (e.g., function name, variable name)
//   - typ: The type of the source element (e.g., "func", "var", "type")
//   - codeType: The code type identifier (defines the category of code element)
//   - lang: The language identifier for the source content
//   - src: The raw source content (byte slice containing the actual source code)
//
// Returns:
//   - A newly created Entry instance that represents an individual code element
func (o *mdl) EntNew(pkg audids.ID, name, typ string, codeType audtps.CodeType, lang string, src []byte) audent.Entry {
	return audent.New(pkg, name, typ, codeType, lang, src)
}

// EntNewOrMerge creates a new entry or merges with an existing one if it already exists.
//
// This method attempts to create a new entry with the given parameters.
// If an entry with the same identity already exists in the collection, it merges
// the new data into the existing entry instead of creating a duplicate.
//
// Parameters:
//   - pkg: The package ID to associate with the new entry
//   - name: The name of the source element
//   - typ: The type of the source element
//   - code: The code type identifier
//   - lang: The language identifier for the source content
//   - src: The raw source content
//
// Returns:
//   - The newly created or merged Entry instance
func (o *mdl) EntNewOrMerge(pkg audids.ID, name string, typ string, code audtps.CodeType, lang string, src []byte) audent.Entry {
	e := audent.New(pkg, name, typ, code, lang, src)
	if e == nil {
		return nil
	}

	o.ent.Store(e)
	return o.ent.Load(e.GetID())
}

// EntGet retrieves an Entry from the collection by its unique identifier.
//
// This method fetches an existing entry from the internal entry collection using
// its unique identifier. It provides access to previously created entries for further processing
// or analysis within the AST management system. If no entry with the specified ID exists,
// this method returns nil.
//
// Parameters:
//   - id: The unique identifier of the entry to retrieve
//
// Returns:
//   - The Entry instance if found in the collection, otherwise nil
func (o *mdl) EntGet(id audids.ID) audent.Entry {
	return o.ent.Load(id)
}

// EntAdd stores an Entry in the collection for later retrieval and processing.
//
// This method persists an entry in the internal entry collection for use during AST parsing operations.
// If an entry with the same ID already exists and is not empty, it merges the new entry with the existing one
// before storing to maintain data integrity. This approach ensures that entries can be incrementally built
// or updated rather than replaced entirely.
//
// Parameters:
//   - ent: The Entry instance to add to the collection
func (o *mdl) EntAdd(ent audent.Entry) {
	o.ent.Merge(ent)
}

// EntGetOrAdd retrieves an existing entry or adds a new one if it doesn't exist.
//
// This method first attempts to find an entry matching the provided entry's identity.
// If found, it returns the existing entry. If not found, it adds the provided entry
// to the collection and returns it.
//
// Parameters:
//   - ent: The Entry instance to retrieve or add
//
// Returns:
//   - The existing or newly added Entry instance
func (o *mdl) EntGetOrAdd(ent audent.Entry) audent.Entry {
	return o.ent.GetOrStore(ent)
}

// EntDel removes an Entry from the collection by its unique identifier.
//
// This method deletes an entry from the internal entry collection using its unique identifier.
// It is typically called when cleaning up or removing entries that are no longer needed,
// such as during resource cleanup or when processing changes in source code structure.
//
// Parameters:
//   - id: The unique identifier of the entry to delete
func (o *mdl) EntDel(id audids.ID) {
	o.ent.Delete(id)
}

// EntLen returns the number of entries currently stored in the collection.
//
// This method provides a count of all entries managed by this manager instance,
// which can be useful for monitoring or debugging purposes. It accesses the underlying
// entry collection's length method to return the total number of stored entries.
//
// Returns:
//   - The total number of entries in the internal collection
func (o *mdl) EntLen() int {
	return o.ent.Len()
}

// EntWalk iterates over all entries in the collection using the provided callback function.
//
// This method executes a callback function for each entry in the internal entry collection,
// allowing for processing or analysis of all stored entries. The iteration continues until
// all entries have been processed or until the callback returns false, indicating early termination.
//
// Parameters:
//   - f: A callback function that receives an entry and returns a boolean indicating
//     whether to continue iteration (true) or stop (false)
func (o *mdl) EntWalk(f func(audent.Entry) bool) {
	o.ent.Walk(f)
}

// entDelForPkg deletes all entries associated with a specific package.
//
// This internal helper method removes all entries from the collection that belong
// to the specified package ID. It is called during package deletion to ensure that
// all related entry data is properly cleaned up before removing the package itself.
// This prevents orphaned entry references and maintains data consistency.
//
// Parameters:
//   - id: The unique identifier of the package for which entries should be deleted
func (o *mdl) entDelForPkg(id audids.ID) {
	var ids = make([]audids.ID, 0, o.ent.Len())

	o.EntWalk(func(ent audent.Entry) bool {
		if ent.GetPkg().GetID() == id {
			ids = append(ids, ent.GetID())
		}
		return true
	})

	for _, k := range ids {
		o.EntDel(k)
	}
}
