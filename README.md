# stdx

Generic building blocks that Go's standard library does not provide yet: optional values, result types, tuples, typed mutexes, and custom base conversion.

Licensed under Apache-2.0. You may use this in proprietary projects without open-sourcing your code.

## Requirements

Go 1.21 or later.

## Install

```bash
go get github.com/Jordanzuo/stdx
```

## Packages

| Package | Import | Description |
| --- | --- | --- |
| `option` | `github.com/Jordanzuo/stdx/option` | Optional value (`Option[T]`) without nil pointers |
| `result` | `github.com/Jordanzuo/stdx/result` | Success/error (`Result[T]`) and optional+error (`OptionResult[T]`) |
| `tuple` | `github.com/Jordanzuo/stdx/tuple` | Heterogeneous tuples (`Tuple2` … `Tuple16`) |
| `mutex` | `github.com/Jordanzuo/stdx/mutex` | Typed mutex wrappers (`Mutex[T]`, `RWMutex[T]`) |
| `base` | `github.com/Jordanzuo/stdx/base` | Convert `uint64` values to and from custom base strings |

## Examples

### Option

```go
import "github.com/Jordanzuo/stdx/option"

id := option.NewDataOption(int64(100))
if id.HasData() {
    _ = id.Data()
}

name := option.NewNoneOption[string]()
display := name.DataOrDefault("guest")
```

An `Option` must be created with `NewNoneOption` or `NewDataOption`. The zero value is uninitialized and panics on use. Use `NewDataOption(0)` to represent a present zero value.

### Result

```go
import "github.com/Jordanzuo/stdx/result"

func parseID(s string) result.Result[int64] {
    // return result.NewDataResult(id) or result.NewErrorResult[int64](err)
    return result.NewDataResult(int64(1))
}

r := parseID("1")
if r.HasData() {
    _ = r.Data()
}
```

A `Result` must be created with `NewDataResult` or `NewErrorResult`. The zero value is uninitialized and panics on use. Use `NewDataResult(0)` to represent a successful zero value.

### OptionResult

Use when an operation can succeed with a value, succeed with no value, or fail with an error.

```go
import "github.com/Jordanzuo/stdx/result"

r := result.NewNoneOptionResult[string]()
if r.HasNone() {
    // not found
}
```

An `OptionResult` must be created with `NewDataOptionResult`, `NewErrorOptionResult`, or `NewNoneOptionResult`. The zero value is uninitialized and panics on use. Use `NewDataOptionResult((*T)(nil))` to represent a present nil pointer, distinct from `NewNoneOptionResult`.

### Tuple

```go
import "github.com/Jordanzuo/stdx/tuple"

pair := tuple.NewTuple2(int64(1), "Jordan")
id, name := pair.Values()
```

### Base

```go
import "github.com/Jordanzuo/stdx/base"

hex := base.NewBase16(false)
hex.Transform(16) // "10"

hexUpper := base.NewBase16(true)
hexUpper.Transform(255) // "FF"

binary := base.NewBase2()
binary.Transform(16) // "10000"

custom := base.New("0123456789ABCDEF")
if custom.HasError() {
    // handle error
}
_ = custom.Data().Transform(255) // "FF"
```

Use `Transform` to convert a decimal `uint64` to a base string, and `Parse` to convert back. Predefined helpers cover binary, octal, hexadecimal, base-26, base-36, and base-62.

Parsing is **case-sensitive**. `NewBase16(false)` accepts `"ff"` but not `"FF"`; `NewBase16(true)` is the opposite. Pick the `uppercase` argument to match your input, or use `NewBase62` when both letter cases are valid digits.

`Transform` returns the shortest encoding without leading zeros. `Parse` accepts leading zeros (`"00ff"` and `"ff"` both parse to 255), but re-encoding with `Transform` yields `"ff"`. Compare parsed numeric values, not raw strings.

`NewBase62` includes visually similar digits that encode different values, such as `0`/`O`/`o` and `1`/`l`/`I`. Prefer another base when people must type or read codes by hand.

For custom alphabets from `New`, visually similar Unicode characters (such as Latin and Cyrillic letters) are distinct digits. Invalid UTF-8 bytes are read as U+FFFD (`�`). Create `Base` values only through `New` or the `NewBase*` helpers; a nil `*Base` panics when used.

### Mutex

```go
import "github.com/Jordanzuo/stdx/mutex"

m := mutex.NewMutex(&Config{})
m.WriteData(func(cfg *Config) {
    cfg.Enabled = true
})
```

`WriteData` and `ReadData` pass a copy of `T` to the callback. In-place mutation works when `T` is a pointer, map, or other reference type. For value types, use `Lock`, `SetDataAfterLock`, and `Unlock`:

```go
v := this.Lock()
defer this.Unlock()
v.Count++
this.SetDataAfterLock(v)
```

## Development

```bash
go test ./...
```

## License

See [LICENSE](./LICENSE) for the full Apache-2.0 text.
