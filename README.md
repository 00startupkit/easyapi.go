<p align="center">
  <img src="docs/easyapi-go-logo-3.png" alt="quickapi logo" height="35" />
</p>
<p align="center">
  <img src="https://github.com/00startupkit/easyapi.go/actions/workflows/go_tests.yml/badge.svg" alt="ci status" />
</p>

Setup a REST API to serve your data easily.

### Examle Usage

```go
config := &Config {
  // The defined routes will be prefixed with this root.
  Root: "/api/v1",
  Schema: []*Schema{
    {
      Name: "Users",
      // Plug in the data provider for the "Users" dataset
      // here so that the REST API knows how to access the data
      // to be served.
      // Read below for data providers for your persistent storage solution.
      Provider: users_data_provider,
    },
    // ... define more shemas to be served by the REST API
  },
}

err := InitEasyApi(config)
// TODO: Put example of plugging in the api router.
```

### Data Providers
- [MySQL Data Provider](https://github.com/00startupkit/easyapi-mysql-provider.go): Configure to serve data from your MySQL database.
