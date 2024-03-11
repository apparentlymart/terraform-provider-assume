# `maplength` functions

Annotates the upper bound, lower bound, or both bounds of a map's length.

```hcl
provider::assume::maplength(map, min_length, max_length)
provider::assume::maplengthmin(map, min_length)
provider::assume::maplengthmax(map, max_length)
```

When given an unknown value, this function returns the same value annotated
with a guarantee that its final value will have a length within the given
range.

When given a known value, this function either returns that value verbatim
or returns an error if the value does not have a length in the promised range.

For example, if you are writing a module that returns a map of ids where
it's guaranteed that there will be one element for each element of a collection
given as an input variable, you can report that to callers of your module by
using this function in an `output` block:

```hcl
output "subnet_ids" {
  value = provider::assume::maplength(
    { for az, subnet in aws_subnet.example : az => subnet.id },
    length(var.availability_zones), length(var.availability_zones),
  )
}
```

If you also know that the value will never be `null` consider also using
[`notnull`](./notnull.md) to report that.
