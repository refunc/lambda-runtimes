# Lambda Runtimes

Lambda Runtime images and [Xenvs](https://github.com/refunc/refunc/blob/7bb8d133f2af66affa02d6d616797f869b69f48a/pkg/apis/refunc/v1beta3/xenv.go#L29) for [Refunc](https://github.com/refunc/refunc)

## How it works

We use image from [lambci/docker-lambda](https://github.com/lambci/docker-lambda) as base image for each runtime, and convert it's runtime to [aws cunstom runtime](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html) in order to run old runtimes on refunc

## Supported runtimes

- [x] prvoided
- [x] python3.7
- [x] ruby2.5
- [x] python3.6
- [x] python2.7
- [x] nodejs8.10
- [x] nodejs6.10
- [ ] java8
- [ ] go1.x
- [ ] dotnetcore2.1
- [ ] dotnetcore2.0
