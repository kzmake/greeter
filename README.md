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
make up
```

## request api
```
make request
```

## generate certs for mTLS
```bash
make certs
```

## publish image
```bash
make publish
```
