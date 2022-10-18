binson-go
=========

A light-weight one-file Golang implementation of a Binson parser (decoder) and writer (encoder).

Binson is like JSON, but faster, binary and even simpler.
See [binson.org](https://binson.org/).

This library is a Go port of the Java lib: [github.com/franslundberg/binson-java-light](https://github.com/franslundberg/binson-java-light).

For a Go library that handle whole Binson object in memory, see
[github.com/hakanols/binson-go](https://github.com/hakanols/binson-go)
by Håkan Olsson.


## Install

Just add '"github.com/assaabloy-ppi/binson-go/binson"` to your import list,
your build tool will do the rest. 



## Code examples

Useful code examples. The source code is also available from 
`binson-go/blob/master/examples/examples.go`.

NOTE: fields must be sorted on alphabetical order
(see binson.org for exact sort order) to be real Binson objects. This light-weight implementation does not check this. Invalid Binson bytes can be produced with this library.

**Example 1**. The code below first creates Binson bytes with two fields: 
one integer named `a` and one string named `s`. Then the bytes are parsed to 
retrieve the original values.

```go
//
// {"a":123, "s":"Hello world!"}
//
var b bytes.Buffer
var e = binson.NewEncoder(&b)

e.Begin()
e.Name("a")
e.Integer(123)
e.Name("s")
e.String("Hello world!")
e.End()
e.Flush()

var d = binson.NewDecoder(&b)

d.Field("a")
fmt.Println(d.Value) // -> 123
d.Field("s")
fmt.Println(d.Value) // -> Hello world!
```

**Example 2**. This example demonstrates how a nested Binson object
can be handled.

```go
//
// {"a":{"b":2},"c":3}
//
var b bytes.Buffer
var e = binson.NewEncoder(&b)

e.Begin()
e.Name("a")
e.Begin()
e.Name("b")
e.Integer(2)
e.End()
e.Name("c")
e.Integer(3)
e.End()
e.Flush()

var d = binson.NewDecoder(&b)

d.Field("a")
d.GoIntoObject()
d.Field("b")
fmt.Println(d.Value) // -> 2
d.GoUpToObject()
d.Field("c")
fmt.Println(d.Value) // -> 3
```

**Example 3**. This example shows how arrays are generated and parsed.

```go
//
// {"arr":[123, "hello"]}
//
var b bytes.Buffer
var e = binson.NewEncoder(&b)

e.Begin()
e.Name("arr")
e.BeginArray()
e.Integer(123)
e.String("hello")
e.EndArray()
e.End()
e.Flush()

var d = binson.NewDecoder(&b)

d.Field("arr")
d.GoIntoArray()
gotValue := d.NextArrayValue()
fmt.Println(gotValue)                      // -> true
fmt.Println(binson.Integer == d.ValueType) // -> true
fmt.Println(d.Value)                       // -> 123

gotValue = d.NextArrayValue()
fmt.Println(gotValue)                     // -> true
fmt.Println(binson.String == d.ValueType) // -> true
fmt.Println(d.Value)                      // -> hello
```

