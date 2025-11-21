# `equal` function

Declares that a possibly-unknown value should be equal to a given known value.

```hcl
provider::assume::equal(actual_value, expected_value)
```

When `actual_value` is unknown, this function immediately returns
`expected_value`.

Once `actual_value` becomes known this function checks whether it's equal to
`expected_value`. If not, it raises an error complaining about the mismatch.
If so, it returns `expected_value` again.

If the actual value and expected value are not of the same type then this
function will attempt to convert the actual value to match the type of the
expected value before comparing. If that conversion fails then this function
returns a type mismatch error because two values of differing types can never
compare as equal.

**Do not use sensitive values in the `actual_value` argument**, because if
the assumption fails then the returned error message may include the actual
value in cleartext.

Assuming that an known value is _fully equal_ to another value is somewhat
esoteric, since if you already know what value you're expecting then it might
be simpler to just use that value. Other functions in this provider might be
more appropriate since they only _partially_ constrain the expected value.

This function's narrow use-case is when a module author wants to reduce the
uncertainty of values during the planning phase by specifying what value
Terraform should expect when planning changes for downstream resources, while
still checking whether that assumption was upheld during the apply phase in
practice.

For example, if you're using the `hashicorp/aws` provider you may be using
a resource type which sets its `arn` attribute to an unknown string when
the object hasn't been created yet, but you know from the documentation of
the AWS service in question what ARN format you're expecting, in which case
you can tell Terraform to assume that the ARN will be what you expected during
planning so that a derived IAM policy (or similar) will have predictable
content in the plan.

The following demonstrates that situation using an IAM role whose ARN is then
used as part of an IAM policy:

```hcl
data "aws_caller_identity" "current" {}
data "aws_partition" "current {}

resource "aws_iam_role" "example" {
  name = "example"
  # Must not specify "path" here without also updating local.example_role_arn

  assume_role_policy = jsonencode({
    # (...trust policy rules...)
  })
}

locals {
  # We expect the role ARN to match the documented conventions for IAM roles:
  # https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsidentityandaccessmanagementiam.html#awsidentityandaccessmanagementiam-resources-for-iam-policies
  example_role_arn = provider::assume::equal(
    aws_iam_role.example.arn,
    provider::aws::arn_build(
      data.aws_partition.current.id,
      "iam",
      "", // Roles are global objects, so no region specified
      aws_caller_identity.current.account_id,
      "role/${aws_iam_role.example.name}", # NOTE: assumes that the role specifies no "path" argument
    ),
  )
}

resource "aws_s3_bucket_policy" "example" {
  bucket = "example"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:*",
        ]
        Effect   = "Allow"
        Principal = {
          # Use local.example_role_arn instead of referring directly to
          # aws_iam_role.example.arn so that this will be known before
          # the role has been created.
          AWS = local.example_role_arn
        }
      },
    ]
  })
}
```

The above could therefore allow Terraform to present the full expected S3
bucket policy in the plan output even when the role hasn't been created yet,
which allows operators to check whether other parts of the policy seem correct
before approving the plan.

Following this pattern is a risk tradeoff: it produces more information during
the planning phase that might therefore make the plan easier to review, but if
the final value _doesn't_ match the expected value then this will cause an
error during the apply phase for something that could otherwise have succeeded.

It's important then to make sure that the expected value is actually correct.
To assist future maintainers in updating the assumed value as necessary,
consider including comments in or around the call to `provider::assume::local`
which specify why the author believes the value to be correct, and what changes
could potentially cause the assumption to become incorrect.
