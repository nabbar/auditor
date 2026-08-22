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

// Package ast provides language identification and AST (Abstract Syntax Tree) parsing capabilities
// for different programming languages supported by the auditor tool.
//
// This package implements a comprehensive system for identifying various programming languages
// and parsing their source code into Abstract Syntax Trees (ASTs). It supports multiple languages
// including Assembly, C, C++, Go, Rust, Python, Java, PHP, and JavaScript. The package leverages
// specialized AST parsers for each supported language to provide accurate syntax analysis.
package ast

import (
	"errors"
	"fmt"
	"strings"

	astgol "github.com/nabbar/auditor/pkg/ast/astgol"
	astgen "github.com/nabbar/auditor/pkg/ast/generic"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audpkg "github.com/nabbar/auditor/pkg/generic"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
)

// Lang represents a programming language type used for identification and parsing.
//
// This enumeration defines the supported programming languages within the auditor tool.
// Each language constant represents a specific programming language that can be parsed
// and analyzed by the AST processing system. The Lang type serves as a unified interface
// for handling different language-specific parsers and identification logic.
//
// Supported languages include:
//   - Assembly (LangAsm)
//   - C (LangC)
//   - Go (LangGo)
//   - C++ (LangCpp)
//   - Rust (LangRust)
//   - Python (LangPython)
//   - Java (LangJava)
//   - PHP (LangPhp)
//   - JavaScript (LangJS)
//
// The Lang type enables consistent handling of language-specific operations across
// different programming languages through a common interface.
type Lang uint8

const (
	// LangUnknown represents an unknown or unsupported language
	LangUnknown Lang = iota

	// LangAsm represents Assembly language
	LangAsm

	// LangC represents C programming language
	LangC

	// LangGo represents Go programming language
	LangGo

	// LangCpp represents C++ programming language
	LangCpp

	// LangRust represents Rust programming language
	LangRust

	// LangPython represents Python programming language
	LangPython

	// LangJava represents Java programming language
	LangJava

	// LangPhp represents PHP programming language
	LangPhp

	// LangJS represents JavaScript programming language
	LangJS
)

// String returns the human-readable name of the language.
//
// This method provides a descriptive string representation of the language enumeration
// value. It is particularly useful for logging, debugging, and user-facing displays where
// a human-readable format is required instead of the numeric representation.
//
// The returned string corresponds to the standard name of the programming language
// associated with the Lang value. For unsupported languages, it returns "unknown".
//
// Returns:
//   - string: The human-readable name of the language, or "unknown" for invalid values
func (o Lang) String() string {
	switch o {
	case LangAsm:
		return "Assembler" // placeholder waiting package AST for Assembler
	case LangC:
		return "C" // placeholder waiting package AST for C
	case LangGo:
		return astgol.Lang
	case LangCpp:
		return "C++" // placeholder waiting package AST for C++
	case LangRust:
		return "Rust" // placeholder waiting package AST for Rust
	case LangPython:
		return "Python" // placeholder waiting package AST for Python
	case LangJava:
		return "Java" // placeholder waiting package AST for Java
	case LangPhp:
		return "PHP" // placeholder waiting package AST for PHP
	case LangJS:
		return "JavaScript" // placeholder waiting package AST for JavaScript
	default:
		return audpkg.Unknown
	}
}

func Parse(s string) Lang {
	switch strings.ToLower(s) {
	case strings.ToLower(LangAsm.String()):
		return LangAsm
	case strings.ToLower(LangC.String()):
		return LangC
	case strings.ToLower(LangGo.String()):
		return LangGo
	case strings.ToLower(LangCpp.String()):
		return LangCpp
	case strings.ToLower(LangRust.String()):
		return LangRust
	case strings.ToLower(LangPython.String()):
		return LangPython
	case strings.ToLower(LangJava.String()):
		return LangJava
	case strings.ToLower(LangPhp.String()):
		return LangPhp
	case strings.ToLower(LangJS.String()):
		return LangJS
	}

	return LangUnknown
}

// Identify determines whether a given file matches the language type.
//
// This method implements language-specific identification logic to detect if a file
// belongs to this programming language category. It uses language-specific heuristics
// and patterns to determine if the file content matches the expected format for the
// specified language type. The method is essential for automatic language detection
// in source code analysis.
//
// Parameters:
//   - file: The full path to the file being checked for language matching
//
// Returns:
//   - bool: True if the file matches this language type, false otherwise. This indicates
//     whether the specified file conforms to the syntax and structure expected by this language parser
func (o Lang) Identify(file string) bool {
	switch o {
	case LangAsm:
		return false // Placeholder for assembly language identification
	case LangC:
		return false // Placeholder for C language identification
	case LangGo:
		return astgol.Identify(file) // Uses Go-specific identification logic
	case LangCpp:
		return false // Placeholder for C++ language identification
	case LangRust:
		return false // Placeholder for Rust language identification
	case LangPython:
		return false // Placeholder for Python language identification
	case LangJava:
		return false // Placeholder for Java language identification
	case LangPhp:
		return false // Placeholder for PHP language identification
	case LangJS:
		return false // Placeholder for JavaScript language identification
	default:
		return false // Default fallback for unknown languages
	}
}

// AST retrieves the Abstract Syntax Tree for a given file of this language type.
//
// This method returns an AST implementation specific to the language and handles errors appropriately.
// It provides the core functionality for parsing source code files of the specified language type.
// The returned AST parser instance is configured with language-specific settings to ensure accurate
// syntax analysis and semantic interpretation.
//
// Parameters:
//   - mgr: The AST Manager instance for handling collections of AST elements
//   - file: The full path to the source file to parse for AST generation
//
// Returns:
//   - astgen.AST: An AST parser instance configured for this language type, or nil if not implemented
//   - error: Any error encountered during AST creation, including invalid file names or initialization errors.
//     This includes cases where the AST parser is not yet implemented for the specific language
func (o Lang) AST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error) {
	var e = fmt.Errorf("ast parser not implemented for file %s", file)

	if mgr == nil || mgr.IsEmpty() {
		return nil, errors.New("invalid DB manager")
	}

	switch o {
	case LangAsm:
		return nil, e // Assembly language AST parser not implemented
	case LangC:
		return nil, e // C language AST parser not implemented
	case LangGo:
		return astgol.GetAST(uim, mgr, file) // Uses Go-specific AST parsing logic
	case LangCpp:
		return nil, e // C++ language AST parser not implemented
	case LangRust:
		return nil, e // Rust language AST parser not implemented
	case LangPython:
		return nil, e // Python language AST parser not implemented
	case LangJava:
		return nil, e // Java language AST parser not implemented
	case LangPhp:
		return nil, e // PHP language AST parser not implemented
	case LangJS:
		return nil, e // JavaScript language AST parser not implemented
	default:
		return nil, e // Default fallback for unknown languages
	}
}

// Pattern returns the file pattern or extension associated with this language type.
//
// This method provides a list of file patterns or extensions that are commonly used
// for files belonging to this programming language. It is primarily used for
// identifying files that match specific language types based on their naming conventions.
//
// Returns:
//   - []string: A slice of file pattern strings that represent the file extensions or naming conventions
//     associated with this language type, or nil if no patterns are defined for this language
func (o Lang) Pattern() []string {
	switch o {
	case LangAsm:
		return nil
	case LangC:
		return nil
	case LangGo:
		return astgol.Pattern()
	case LangCpp:
		return nil
	case LangRust:
		return nil
	case LangJava:
		return nil
	case LangPhp:
		return nil
	case LangJS:
		return nil
	default:
		return nil
	}
}

// Identify determines the programming language of a given file by checking all supported languages.
//
// This function iterates through all available language types and uses their Identify method to find
// the appropriate language for the specified file. It serves as the main language detection mechanism
// for the auditor tool's AST processing system. The function systematically checks each language type
// in order of precedence until a match is found, returning LangUnknown if no language matches.
//
// Parameters:
//   - file: The full path to the file whose language needs to be identified
//
// Returns:
//   - Lang: The detected programming language type, or LangUnknown if no match is found. This indicates
//     which programming language the specified file belongs to based on its content and structure
func Identify(file string) Lang {
	switch {
	case LangAsm.Identify(file):
		return LangAsm // Detected as Assembly language
	case LangC.Identify(file):
		return LangC // Detected as C language
	case LangGo.Identify(file):
		return LangGo // Detected as Go language
	case LangCpp.Identify(file):
		return LangCpp // Detected as C++ language
	case LangRust.Identify(file):
		return LangRust // Detected as Rust language
	case LangPython.Identify(file):
		return LangPython // Detected as Python language
	case LangJava.Identify(file):
		return LangJava // Detected as Java language
	case LangPhp.Identify(file):
		return LangPhp // Detected as PHP language
	case LangJS.Identify(file):
		return LangJS // Detected as JavaScript language
	default:
		return LangUnknown // Default case when no language is identified
	}
}

// Pattern returns a map of file patterns for all supported languages.
//
// This function aggregates the file pattern information for all supported programming languages
// and constructs a mapping from language types to their corresponding file patterns. It is useful
// for applications that need to identify files based on their extensions or naming conventions.
//
// Returns:
//   - map[Lang][]string: A map where keys are language types and values are slices of file pattern strings
//     representing the extensions or naming conventions associated with each language type
func Pattern() map[Lang][]string {
	var r = make(map[Lang][]string)

	for _, k := range []Lang{
		LangAsm,
		LangC,
		LangGo,
		LangCpp,
		LangRust,
		LangJava,
		LangPhp,
		LangJS,
	} {
		if l := k.Pattern(); len(l) > 0 {
			r[k] = l
		}
	}

	return r
}

func List() []Lang {
	return []Lang{
		LangAsm,
		LangC,
		LangGo,
		LangCpp,
		LangRust,
		LangJava,
		LangPhp,
		LangJS,
	}
}
