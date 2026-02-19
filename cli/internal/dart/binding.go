package dart

// #cgo CFLAGS: -std=c11 -fPIC -I.
// #include "tree_sitter/parser.h"
// TSLanguage *tree_sitter_dart();
import "C"

import (
	"unsafe"

	sitter "github.com/smacker/go-tree-sitter"
)

func GetLanguage() *sitter.Language {
	return sitter.NewLanguage(unsafe.Pointer(C.tree_sitter_dart()))
}
