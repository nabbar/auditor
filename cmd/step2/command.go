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

package step2

import (
	"fmt"

	audpkg "github.com/nabbar/auditor/pkg"
	spfcbr "github.com/spf13/cobra"
)

const (
	CmdName  = "step2"
	CmdShort = "Perform an LLM-powered deep structural analysis of the scanned codebase"
	CmdLong  = `
Execute advanced semantic dependency resolution and asynchronous, multi-threaded 
LLM-powered structural auditing on the scanned codebase.

  The 'step2' command represents the central analytical intelligence of the auditor 
  toolset. While the initial scanning phase is strictly limited to extracting raw 
  syntax and generating Abstract Syntax Trees (AST), this second step is responsible 
  for transforming that rigid structural data into a deeply understood semantic 
  knowledge base. It achieves this by bridging the gap between raw code parsing 
  and artificial intelligence analysis.

  The execution of this command triggers a highly complex, multi-stage workflow. 
  First, the engine traverses the internal database to reconstruct the semantic 
  dependency graphs. This involves linking declarations to their implementations 
  across disparate packages, resolving symbol references, and building an 
  inter-package reference map. This map is absolutely critical because it allows 
  the auditor to understand the global flow of data, the architectural constraints 
  of the codebase, and the hidden relationships between seemingly isolated modules.

  Once the dependency graph is fully resolved, the command initiates the core 
  analysis sequence. It spawns an asynchronous, multi-threaded worker pool designed 
  to maximize throughput without violating API rate limits. For every single code 
  entry registered in the database, the engine dynamically constructs a prompt 
  containing the entry's raw code, its semantic dependencies, its parameter 
  signatures, and its composition hierarchy. This highly contextualized prompt is 
  then dispatched to a configured Large Language Model (LLM). 

  The LLM acts as an expert software architect, evaluating the structural integrity 
  of the entry, deducing its business logic purpose, analyzing the validity of its 
  input parameters and return types, and summarizing its internal fields or methods. 
  The responses returned by the LLM are parsed, validated, and persistently stored 
  back into the internal database, ready to be compiled into comprehensive human-
  readable reports in subsequent operational steps.

Prerequisites:
  Before invoking the 'step2' command, the user must ensure that the target source 
  code has been successfully processed by the scanning engine. The integrity of the 
  internal SQLite database is strictly enforced prior to execution. Specifically, 
  the pre-run validation checks mandate that the internal database must contain at 
  least one fully initialized module, at least one parsed package, and at least one 
  distinct code entry. If any of these fundamental requirements are not met, the 
  command will immediately abort its execution and return a fatal error, preventing 
  the generation of corrupt or incomplete dependency graphs.

Supported Programming Languages:
  The auditor toolset is designed to be inherently polyglot, leveraging a robust 
  language identification subsystem defined within the internal 'ast' package. When 
  defining configurations, you must use the strict language codes recognized by the 
  system. The currently supported programming languages include:

  - Assembler ("asm"): Represents low-level assembly code instructions. Analyzing 
    this language requires the LLM to understand hardware registers, memory 
    allocation at the lowest abstraction layer, and microprocessor-specific opcodes.
  - C ("c"): Represents standard C source code. The analysis focuses heavily on 
    memory management, pointer arithmetic, macro definitions, and procedural 
    function execution logic.
  - C++ ("cpp"): An extension of C that introduces object-oriented paradigms. The 
    engine expects the LLM to evaluate complex template metaprogramming, class 
    inheritances, polymorphism, and standard template library usage.
  - Go ("go"): The native language of the auditor itself. Analysis is highly optimized 
    for concurrent programming structures (goroutines, channels), structural typing, 
    and strict interface implementations.
  - Rust ("rust"): A systems programming language focusing on safety. The LLM prompts 
    are geared toward evaluating memory ownership rules, borrowing mechanics, 
    lifetimes, and safe concurrency paradigms.
  - Python ("python"): A high-level, dynamically typed language. Auditing focuses on 
    duck typing behaviors, class hierarchies, dynamic memory references, and standard 
    library integrations.
  - Java ("java"): A strictly object-oriented, class-based language. The engine 
    evaluates deep inheritance trees, abstract classes, garbage collection 
    implications, and JVM-specific structural patterns.
  - PHP ("php"): A popular server-side scripting language. The analysis targets 
    object-oriented structures, dynamic variable typing, trait implementations, 
    and web-centric architectural patterns.
  - JavaScript ("js"): A prototype-based, multi-paradigm language. Auditing 
    encompasses asynchronous execution flows (promises, async/await), prototype 
    chaining, closures, and event-driven architectures.

Code Abstraction Types:
  Beyond understanding the language, the internal 'types' package classifies every 
  scanned element into an enumerated "CodeType". This abstraction allows the configuration 
  engine to apply specific LLM behaviors based on the structural nature of the code. 
  The recognized code types are:

  - "function" (FUNCTION): This type encompasses standalone subroutines, procedural 
    functions, and object-oriented methods. It is the primary executable unit of logic. 
    Analysis of this type generally focuses on algorithmic complexity, input validation, 
    state mutation, and the specific nature of the data returned to the caller.
  - "struct" (STRUCT): This type represents composite data structures, classes, or 
    data transfer objects. Since structs do not natively execute logic, the LLM is 
    instructed to analyze the cohesion of the encapsulated fields, data normalization, 
    memory alignment considerations, and the overarching purpose of the data model.
  - "interface" (INTERFACE): This type represents behavioral contracts. The engine 
    evaluates the design of the interface, checking for adherence to the Interface 
    Segregation Principle, the clarity of the defined method signatures, and the 
    expected behaviors of any concrete types that claim to implement it.
  - "primitive" (PRIMITIVE): This refers to the fundamental building blocks of a 
    language's type system (e.g., integers, booleans, floating-point numbers, 
    strings). For primitives, the system often requires an additional string identifier 
    and evaluates how the primitive is aliased or utilized within a domain context.
  - "custom" (CUSTOM): This type acts as a catch-all for user-defined type aliases, 
    complex compound types, or framework-specific constructs that do not cleanly map 
    to standard functions or structs. Like primitives, this requires additional textual 
    context for accurate evaluation.
  - "depend" (DEPEND): Represents external or internal structural dependencies that 
    are referenced by the analyzed code but are not necessarily defined within the 
    current analytical scope.

Supported LLM providers:
  The auditor utilizes an abstraction layer (the 'apitype' package) to interface 
  seamlessly with various artificial intelligence vendors. You must specify one of 
  the following supported providers in the "type" field of your configuration:

  - "OpenAI": Integrates with the official OpenAI REST API. This provider is known 
    for highly accurate instruction following and robust handling of complex software 
    architecture prompts. It natively supports advanced models like GPT-4.
  - "Anthropic": Interfaces with the Claude family of models. This provider excels 
    at processing extremely large context windows, making it ideal for analyzing 
    massive, deeply interconnected codebases with extensive documentation requirements.
  - "Gemini": Connects to Google's Gemini models. It provides excellent multi-modal 
    reasoning and fast inference capabilities, particularly useful when analyzing 
    code that interacts with Google Cloud architectures.
  - "Mistral": Supports the open-weight and commercial models provided by Mistral AI. 
    This is often used for high-performance, cost-effective auditing when using 
    models like Mixtral or Mistral Large.
  - "Ollama": A critical provider for air-gapped or localized environments. Ollama 
    allows you to run open-source models (like Llama 3 or Qwen) directly on your 
    local hardware, ensuring that proprietary source code never leaves your internal 
    network.
  - "Copilot": Hooks into the GitHub Copilot infrastructure, leveraging enterprise 
    developer environments and code-specific fine-tuned models for highly relevant 
    codebase summarizations.

Configuration Architecture:
  To provide absolute control over the auditing process, the tool relies on a highly 
  sophisticated, cascading configuration file located at "<working folder>/.auditor/llmconfig.json". 
  This JSON file governs everything from API endpoints to authentication mechanisms 
  and token consumption limits.

  The configuration system relies on a strict inheritance hierarchy consisting of 
  three tiers. If a specific setting is not defined at a lower tier, the system 
  automatically falls back to the definition provided in the tier above it.

  Tier 1: Global Default ("default"). This block defines the baseline behavior for 
  the entire application. If no other settings are provided, every API call will 
  use the credentials and limits defined here.
  
  Tier 2: Language Overrides ("lang"). You can define an object using any of the 
  supported language codes (e.g., "go", "python"). Any configuration block placed 
  inside a specific language will completely override the Global Default for files 
  written in that language. This allows you, for instance, to use OpenAI for Go code 
  but Anthropic for Python code.

  Tier 3: Code Type Overrides ("types"). Nested within a language block, you can 
  further refine settings based on the CodeType (e.g., "function", "struct"). This 
  provides extreme granularity, allowing you to allocate expensive, highly capable 
  models strictly for complex "function" logic, while routing simple "struct" 
  evaluations to a faster, cheaper local Ollama instance.

Configuration Schema Details:
  Whether defining configurations at the default, language, or type level, the 
  available parameters remain consistent. The configuration block must adhere to 
  the following exhaustive schema definitions:

  1. The API Definition Object ("api"):
     This block dictates exactly where and how the network requests are routed.
     - "hostname": A string representing the Fully Qualified Domain Name (FQDN) or 
       IP address, optionally including a port number (e.g., "api.openai.com:443" 
       or "127.0.0.1:11434" for local Ollama).
     - "model": A string specifying the exact model identifier to request from the 
       provider (e.g., "gpt-4-turbo", "claude-3-opus-20240229", "llama3:8b").
     - "timeout": An integer defining the maximum number of seconds the HTTP client 
       will wait for a response before terminating the connection and attempting a 
       retry operation.
     - "options": A flexible JSON object containing key-value pairs (strings, integers, 
       or booleans) that are passed directly into the provider's specific API payload. 
       This is used to handle proprietary flags required by different SDKs.

  2. The Authentication Object ("auth"):
     This block manages security and access control. It supports three distinct 
     paradigms. You should only configure the paradigm required by your chosen provider.
     - "basic": An object containing a "username" string and a "password" string. 
       This is traditionally used for proxy authentications or legacy local services.
     - "api": An object containing a single "api-key" string. This is the standard 
       Bearer token authentication utilized by OpenAI, Anthropic, Mistral, and Gemini.
     - "oauth": A complex object for environments enforcing OAuth 2.0 security policies, 
       typically seen in enterprise network perimeters.
         - "login-uri": The exact URL endpoint to which the client must send its 
           credentials to negotiate a temporary access token.
         - "timeout-login": An integer specifying the maximum time in seconds allowed 
           for the authentication handshake to complete.
         - "client-id": A string representing the unique identifier of the application 
           requesting authorization.
         - "secret-id": A secure string acting as the cryptographic password for the 
           client application.
         - "grant-type": A string defining the specific OAuth 2.0 flow being executed 
           (most commonly "client_credentials" for machine-to-machine communication).

  3. The Limitation and Quota Object ("limit"):
     Given the asynchronous, multi-threaded nature of the auditor, it is incredibly 
     easy to accidentally exceed API rate limits or incur massive billing charges. 
     This object provides strict governors to prevent catastrophic runaway executions.
     - "maxToken": An integer defining the absolute maximum length of the response 
       generated by the LLM. Note that some providers (like older OpenAI endpoints) 
       handle output limits differently, but this sets a baseline constraint.
     - "temperature": A floating-point value (typically between 0.0 and 1.0 or 2.0) 
       that controls the stochastic nature of the LLM. A value closer to 0.0 forces 
       the model to be highly deterministic, analytical, and repetitive—ideal for 
       strict code analysis. A higher value induces creativity, which increases the 
       risk of hallucinations and inaccurate structural audits.
     - "max-request-second": An integer acting as a hard throttle on the multi-threaded 
       execution pool. It dictates the maximum number of concurrent HTTP requests that 
       can be initiated per second. If the thread pool exceeds this, it will artificially 
       sleep to respect provider limits (e.g., preventing HTTP 429 Too Many Requests).
     - "max-request-token-hour": An integer representing a rolling budget for outbound 
       context tokens per hour. The system heuristically calculates the token size of 
       the generated prompts using specialized algorithms. Once this hourly budget is 
       exhausted, the engine will pause execution.
     - "max-response-token-hour": An integer representing a rolling budget for inbound 
       generation tokens per hour. This allows for strict control over billing costs in 
       commercial cloud environments.

Custom Analysis Templates:
  While the internal code provides a robust baseline for LLM communication, 
  advanced users have the capability to entirely rewrite the instructional prompts 
  sent to the artificial intelligence. The auditor will look for custom template files 
  located within the "<working folder>/.auditor/analyze/" directory. 
  
  These templates are processed as dynamic string blocks where the engine injects 
  contextual variables (such as the raw code, the detected dependencies, and the 
  architectural metadata). By editing these templates, an administrator can force 
  the LLM to adopt a specific persona, adhere to strict internal corporate coding 
  guidelines, or output its analysis in specialized proprietary formats (like 
  custom XML or specific JSON schemas).

Troubleshooting and Execution Behavior:
  When running the 'step2' command, the standard output will display real-time 
  progress logs detailing the dependency resolution phase followed by the core deep 
  auditing phase. Due to the asynchronous design, errors from individual LLM 
  requests will not halt the entire process. Instead, failed requests are logged 
  and queued for retry mechanisms based on exponential back-off algorithms.
  
  If the command appears to stall, it is highly likely that the "max-request-second" 
  limit is set too low for your hardware capabilities, or the "max-request-token-hour" 
  budget has been consumed, forcing the engine into a waiting state. If the resulting 
  reports contain factual inaccuracies regarding your code architecture, verify that 
  the "temperature" setting in your "limit" block is configured close to zero to 
  prevent hallucinatory generations.

`
	CmdUsage   = ""
	CmdExample = ""
)

func InitCmd() *spfcbr.Command {
	cmd := audpkg.GetCobra().NewCommand(CmdName, CmdShort, CmdLong, CmdUsage, CmdExample)
	cmd.Args = spfcbr.NoArgs
	cmd.PreRunE = preRun
	cmd.RunE = run
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceErrors = true

	cmd.Aliases = []string{"analyze"}

	return cmd
}

func preRun(_ *spfcbr.Command, _ []string) error {
	if audpkg.GetDBManager().ModLen() < 1 {
		return fmt.Errorf("scanning files result no modules")
	} else if audpkg.GetDBManager().PkgLen() < 1 {
		return fmt.Errorf("scanning files result no packages")
	} else if audpkg.GetDBManager().EntLen() < 1 {
		return fmt.Errorf("scanning files result no entries")
	}

	return nil
}

func run(_ *spfcbr.Command, _ []string) error {
	audpkg.GetUIManager().Info("⛓️ Resolving Semantic Dependency Graphs & Inter-Package Reference Maps...")
	if e := audpkg.GetEngine().ReOrder(); e != nil {
		return fmt.Errorf("error trigger on Resolving Semantic Dependency Graphs: %w", e)
	}

	audpkg.GetUIManager().Info("🛡️ Asynchronous MultiThreaded Structural Core Deep Auditing...")
	if e := audpkg.GetEngine().Analyze(); e != nil {
		return fmt.Errorf("error trigger on Core Deep Auditing: %w", e)
	}

	return nil
}
