// Package vm embeds the project-owned Core 2.0 interpreter and native backend. Engines compile
// concurrently; each Store serializes guest execution and mutable object access.
// Host callbacks use Caller for controlled reentry. Values carry exact numeric
// bits or opaque reference roots which callers explicitly Release. Close is a
// nonblocking cancellation request, and WaitClosed awaits final reclamation.
package vm
