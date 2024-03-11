# `notnull` function

Annotates a value as definitely not null.

```hcl
provider::assume::notnull(value)
```

When given an unknown value, this function returns the same value annotated
with a guarantee that its final value will not be null.

When given a known value, this function either returns that value verbatim
or returns an error if the value is null.

For example, if you are writing a module that returns the ID of an object,
and you know from the provider's documentation that the ID cannot possibly
be null as long as the object is created successfully, you can report that
to callers of your module by using this function in an `output` block:

```hcl
output "subnet_id" {
  value = provider::assume::notnull(aws_subnet.example.id)
}
```

This will then ensure that any comparison of this value to `null` in the
calling module will return a known boolean value.
