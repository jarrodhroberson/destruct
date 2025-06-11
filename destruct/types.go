package destruct

import (
	"github.com/joomcode/errorx"
)

var destructNamespace = errorx.NewNamespace("destruct")
var UnmatchedStrategyError = errorx.NewType(destructNamespace, "no strategy matched")
