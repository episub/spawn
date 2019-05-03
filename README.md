Documentation in docs.

# Benchmarking

Example:

```
ab -p gql.txt -kc 200 -T application/json -t 4 'http://localhost:8080/query'
```

With `gql.txt`:

```
{"operationName":null,"variables":{},"query":"{\n  todos {\n    id\n    content\n    user {\n      id\n      username\n    }\n  }\n}\n"}
```
