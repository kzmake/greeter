# greeter
```proto
service Greeter {
  rpc Hello(HelloRequest) returns (HelloResponse) {
    option (google.api.http) = {
      post: "/hello"
      body: "*"
    };
  }
}

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string msg = 1;
}
```

## run app
```bash
make kind dev
```

## request api
```
make request
```

## regenerate certs for mTLS
```bash
make certs
```
