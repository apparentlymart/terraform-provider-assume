# `setlength` functions

Annotates the upper bound, lower bound, or both bounds of a set's length.

```hcl
provider::assume::setlength(set, min_length, max_length)
provider::assume::setlengthmin(set, min_length)
provider::assume::setlengthmax(set, max_length)
```

When given an unknown value, this function returns the same value annotated
with a guarantee that its final value will have a length within the given
range.

When given a known value, this function either returns that value verbatim
or returns an error if the value does not have a length in the promised range.

For example, if you are writing a module that returns a set of ids where
it's guaranteed that there will be one element for each element of a collection
given as an input variable, you can report that to callers of your module by
using this function in an `output` block:

```hcl
output "subnet_ids" {
  value = provider::assume::setlength(
    toset(values(aws_subnet.example)[*].id),
    length(var.availability_zones), length(var.availability_zones),
  )
}
```

If you also know that the value will never be `null`, consider also using
[`notnull`](./notnull.md) to report that.
