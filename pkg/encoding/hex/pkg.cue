// Code generated by cuelang.org/go/pkg/gen. DO NOT EDIT.
package hex

funcs: EncodedLen: {
	in: #A0: int64
	out: int64
}
funcs: DecodedLen: {
	in: #A0: int64
	out: int64
}
funcs: Decode: {
	in: #A0: string
	out: bytes | string
}
funcs: Dump: {
	in: #A0: bytes | string
	out: string
}
funcs: Encode: {
	in: #A0: bytes | string
	out: string
}
