# `stringprefix` function

Annotates a string as definitely having a specific prefix.

```hcl
provider::assume::stringprefix(string, prefix)
```

When given an unknown value, this function returns the same value annotated
with a guarantee that its final value will start with the given prefix.

When given a known value, this function either returns that value verbatim
or returns an error if the value does not have the promised prefix.

For example, if you are writing a module that returns the ID of an object,
and you know from the vendor's documentation that the ID will definitely
have a specific prefix, you can report that to callers of your module by
using this function in an `output` block:

```hcl
output "subnet_id" {
  value = provider::assume::stringprefix(aws_subnet.example.id, "subnet-")
}
```

If you also know that the value will never be `null`, consider also using
[`notnull`](./notnull.md) to report that.

Terraform can use this additional information, for example, to return a known
boolean value if the string is compared to `""`, because a string with the
prefix `subnet-` could not possibly be the empty string.
