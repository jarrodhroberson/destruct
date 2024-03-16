# destruct
Destructuring package for structs in Go.

This is a work in progress, I accidentally published it to pkg.go.dev prematurely.

This package attempts to break any struct or pointer to struct down into its most primary parts to so a hash can be
calculated on the entire graph to generate a sort of identity ETag string that should be able to used to compare to see
if two things are the same field by field.
