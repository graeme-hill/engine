# Development

Request / Response models are generated from the `controller/pb/controller.proto` file.
Run the following command to regenate

```bash
make proto 
```

Make sure to update the Open API specs here: https://github.com/battlesnakeio/docs/blob/master/apis/engine/spec

Run the tests

```bash
make test
```
