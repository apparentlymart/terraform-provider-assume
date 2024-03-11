# `listlength` functions

Annotates the upper bound, lower bound, or both bounds of a list's length.

```hcl
provider::assume::listlength(list, min_length, max_length)
provider::assume::listlengthmin(list, min_length)
provider::assume::listlengthmax(list, max_length)
```

When given an unknown value, this function returns the same value annotated
with a guarantee that its final value will have a length within the given
range.

When given a known value, this function either returns that value verbatim
or returns an error if the value does not have a length in the promised range.

For example, if you are writing a module that returns a list of ids where
it's guaranteed that there will be one element for each element of a list
given as an input variable, you can report that to callers of your module by
using this function in an `output` block:

```hcl
output "subnet_ids" {
  value = provider::assume::listlength(
    aws_subnet.example[*].id,
    length(var.cidr_blocks), length(var.cidr_blocks),
  )
}
```

If you also know that the value will never be `null`, consider also using
[`notnull`](./notnull.md) to report that.

If you specify the same number as both the lower and upper bound of the
length of a list, Terraform can automatically transform an unknown list into
the equivalent known list whose values are all unknown, thereby allowing
`length(list)` to produce a known result even if the elements are unknown.
