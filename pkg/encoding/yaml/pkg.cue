// Code generated by cuelang.org/go/pkg/gen. DO NOT EDIT.
package yaml

funcs: Marshal: {
	in: #A0: _
	out: string
}
funcs: MarshalStream: {
	in: #A0: _
	out: string
}
funcs: Unmarshal: {
	in: #A0: bytes | string
	out: _
}
funcs: UnmarshalStream: {
	in: #A0: bytes | string
	out: _
}
funcs: Validate: {
	in: {
		#A0: bytes | string
		#A1: _
	}
	out: bool
}
funcs: ValidatePartial: {
	in: {
		#A0: bytes | string
		#A1: _
	}
	out: bool
}
