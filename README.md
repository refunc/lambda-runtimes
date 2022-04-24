# Lambda Runtimes

Lambda Runtime images and [Xenvs](https://github.com/refunc/refunc/blob/7bb8d133f2af66affa02d6d616797f869b69f48a/pkg/apis/refunc/v1beta3/xenv.go#L29) for [Refunc](https://github.com/refunc/refunc)

## How it works

We use aws lambda runtime interface client for each runtime, eg [aws-lambda-python-runtime-interface-client](https://github.com/aws/aws-lambda-python-runtime-interface-client), and convert it's runtime to [aws cunstom runtime](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html) in order to run lambda on refunc.

## Supported runtimes

- [x] python3.7
- [x] python3.8
- [x] python3.9
- [x] nodejs12.x
- [x] golang1.17

## Dev in local docker without a k8s cluster

[main.go](./main.go) is a code runner that implements a stdin/stdout [sidecar](https://github.com/refunc/refunc/blob/59a7964b60a8914e08b4016a77ce64a8af97e937/pkg/sidecar/sidecar.go#L20) to hosts a [AWS lambda custom runtime](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html)

The runner will setup and exec bootstrap for a given runtime, it reads json string from stdin and forward to function and print output to stdout

When creating a new runtime, we can start a container and using runner to debug, for details plz check [Makefile](./Makefile)

## Artifacts

We leverage [docker's auto builds](https://docs.docker.com/docker-hub/builds/) to create images for our ported runtimes, new image will name under [refunc/lambda:${runtime_name}](https://hub.docker.com/r/refunc/lambda)

## License

Copyright (c) 2018 [refunc.io](http://refunc.io)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.