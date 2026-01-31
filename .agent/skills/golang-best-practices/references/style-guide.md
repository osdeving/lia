# Go Style Guide References for LIA

## External Resources

1. **[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)**
   - Comprehensive guide used by Uber's Go team
   - Covers error handling, performance, testing patterns

2. **[Effective Go](https://go.dev/doc/effective_go)**
   - Official Go documentation
   - Canonical source for idiomatic Go

3. **[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)**
   - Common comments made during Go code reviews
   - Maintained by the Go team

## LIA-Specific Principles

- **Determinism First**: All outputs must be reproducible
- **Canonical Representations**: Use codec.SortAll() before serialization
- **No map[string]any**: Use structs for normative IR
- **Hash Everything**: Enable content-addressable artifacts
