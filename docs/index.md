# Terraform "assume" Provider

This is a utility provider for Terraform (and other software that supports
Terraform's provider protocol) which allows annotating potentially-unknown
values with assumptions that can give Terraform more information during the
planning phase and thus help to produce a more complete plan.

To use the functions in this provider, your module must declare that it
requires the provider:

```hcl
terraform {
  required_providers {
    assume = {
      source = "apparentlymart/assume"
    }
  }
}
```

The above declares that this module will use the local name "assume" to refer
to this provider, and so its functions would be called using the prefix
`provider::assume::`. For example the `notnull` function would be called
as `provider::assume::notnull`. The documentation for each function assumes
that local name, but you can choose a different local name if you wish.

Support for functions in providers was added in Terraform v1.8, and so this
provider effectively requires Terraform v1.8.0 or later.

## How Terraform uses these assumptions

During the planning phase there are often values that Terraform cannot yet
determine, because they will be decided by a remote system only during the
apply step.

However, Terraform can still track some approximate information about these
unknown values, and will make use of that information when evaluating
expressions that refer to unknown values.

For example, if a value `v` is marked as "not null" then Terraform can produce
a known value for `v != null` even if the final value of `v` is not decided yet.

Ideally explicit assumption annotations would not be needed because all other
Terraform and provider features would annotate their results with this
additional information automatically. However, this possibility was only
introduced to Terraform in v1.6 and so not all providers have yet been updated
to produce the necessary metadata, and so this provider can potentially help
fill those gaps until the providers are updated.
