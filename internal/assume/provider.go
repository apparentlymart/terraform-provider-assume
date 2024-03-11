package assume

import (
	"github.com/apparentlymart/go-tf-func-provider/tffunc"
)

func NewProvider() *tffunc.Provider {
	p := tffunc.NewProvider()
	p.AddFunction("notnull", notnullFunc)
	p.AddFunction("stringprefix", stringprefixFunc)
	p.AddFunction("listlength", listlengthFunc)
	p.AddFunction("listlengthmin", listlengthminFunc)
	p.AddFunction("listlengthmax", listlengthmaxFunc)
	p.AddFunction("setlength", setlengthFunc)
	p.AddFunction("setlengthmin", setlengthminFunc)
	p.AddFunction("setlengthmax", setlengthmaxFunc)
	p.AddFunction("maplength", maplengthFunc)
	p.AddFunction("maplengthmin", maplengthminFunc)
	p.AddFunction("maplengthmax", maplengthmaxFunc)
	return p
}
