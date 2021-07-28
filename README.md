This project was migrated to https://github.com/tsouza/terraform-provider-elasticstack. Thanks!

# Terraform Provider for Elastic Stack

### Development

1. Install go 1.16
1. Clone the repo
1. Run `make install`

In your terraform file, use the provider by adding a `required_providers` block like

```
terraform {

  required_providers {
    elastic-stack = {
      source  = "github.com/stevewritescode/elastic-stack"
    }
  }
}
```
