## How to compile

- Unit Test: `make test`
- Run locally: `make run`

## Using Technology
Using [operator-framework](https://github.com/operator-framework/operator-sdk) to buildOperator：v0.8.1

## Added Unit Test

- Reference the kubebuilder, add dependency for BDD。Add 2 lines to Gopkg.toml's required sections.
```
  "github.com/onsi/ginkgo", # for test framework
  "github.com/onsi/gomega", # for test matchers
```

-  Run`dep ensure`
