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
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
)

// ModNew creates a new Module with the specified name and file path.
//
// This method directly creates a new module instance using the audmod.New function,
// bypassing any registered callback functions since this implementation provides
// the default behavior for module creation. The module represents a complete source code unit
// that can contain multiple packages and is typically used when parsing Go source files
// to create top-level modules.
//
// Parameters:
//   - path: The import path of the module (e.g., "github.com/example/project")
//   - file: The file path associated with the module, usually the declaration package file
//
// Returns:
//   - A newly created Module instance that represents a complete source code unit
func (o *mdl) ModNew(path, file string) audmod.Module {
	return audmod.New(path, file)
}

// ModNewOrMerge creates a new module or merges with an existing one if it already exists.
//
// This method attempts to create a new module with the given name and path.
// If a module with the same identity already exists in the collection, it merges
// the new data into the existing module instead of creating a duplicate.
//
// Parameters:
//   - name: The name of the module
//   - path: The file path associated with the module
//
// Returns:
//   - The newly created or merged Module instance
func (o *mdl) ModNewOrMerge(name string, path string) audmod.Module {
	m := audmod.New(name, path)
	if m == nil {
		return nil
	}

	o.mod.Store(m)
	return m
}

// ModGet retrieves a module from the collection by its unique identifier.
//
// This method fetches an existing module from the internal module collection using
// its unique identifier. It provides access to previously created modules for further processing
// or analysis within the AST management system. If no module with the specified ID exists,
// this method returns nil.
//
// Parameters:
//   - id: The unique identifier of the module to retrieve
//
// Returns:
//   - The Module instance if found in the collection, otherwise nil
func (o *mdl) ModGet(id audids.ID) audmod.Module {
	return o.mod.Load(id)
}

// ModAdd stores a module in the collection for later retrieval and processing.
//
// This method persists a module in the internal module collection for use during AST parsing operations.
// If a module with the same ID already exists and is not empty, it merges the new module with the existing one
// before storing to maintain data integrity. This approach ensures that modules can be incrementally built
// or updated rather than replaced entirely.
//
// Parameters:
//   - mod: The Module instance to add to the collection
func (o *mdl) ModAdd(mod audmod.Module) {
	o.mod.Merge(mod)
}

// ModGetOrAdd retrieves an existing module or adds a new one if it doesn't exist.
//
// This method first attempts to find a module matching the provided module's identity.
// If found, it returns the existing module. If not found, it adds the provided module
// to the collection and returns it.
//
// Parameters:
//   - mod: The Module instance to retrieve or add
//
// Returns:
//   - The existing or newly added Module instance
func (o *mdl) ModGetOrAdd(mod audmod.Module) audmod.Module {
	return o.mod.GetOrStore(mod)
}

// ModDel removes a module from the collection by its unique identifier.
//
// This method deletes a module from the internal module collection using its unique identifier.
// It also removes any associated packages for the module before deleting the module itself,
// ensuring that all related data is properly cleaned up. This cleanup process prevents
// orphaned package references and maintains consistency in the AST structure.
//
// Parameters:
//   - id: The unique identifier of the module to delete
func (o *mdl) ModDel(id audids.ID) {
	o.pkgDelForMod(id)
	o.mod.Delete(id)
}

// ModLen returns the number of modules currently stored in the collection.
//
// This method provides a count of all modules managed by this manager instance,
// which can be useful for monitoring or debugging purposes. It accesses the underlying
// module collection's length method to return the total number of stored modules.
//
// Returns:
//   - The total number of modules in the internal collection
func (o *mdl) ModLen() int {
	return o.mod.Len()
}

// ModWalk iterates over all modules in the collection using the provided callback function.
//
// This method executes a callback function for each module in the internal module collection,
// allowing for processing or analysis of all stored modules. The iteration continues until
// all modules have been processed or until the callback returns false, indicating early termination.
//
// Parameters:
//   - f: A callback function that receives a module and returns a boolean indicating
//     whether to continue iteration (true) or stop (false)
func (o *mdl) ModWalk(f func(audmod.Module) bool) {
	o.mod.Walk(f)
}
