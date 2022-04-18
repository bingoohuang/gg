# Exponential Backoff

This is a Go port of the exponential backoff algorithm
from [Google's HTTP Client Library for Java][google-http-java-client].

[Exponential backoff][exponential backoff wiki]
is an algorithm that uses feedback to multiplicatively decrease the rate of some process,
in order to gradually find an acceptable rate.
The retries exponentially increase and stop increasing when a certain threshold is met.

## Usage

[original code](https://github.com/cenkalti/backoff)

Import path is `github.com/bingoohuang/gg/pkg/backoff`.

## Resources

[google-http-java-client]: https://github.com/google/google-http-java-client/blob/da1aa993e90285ec18579f1553339b00e19b3ab5/google-http-client/src/main/java/com/google/api/client/util/ExponentialBackOff.java

[exponential backoff wiki]: http://en.wikipedia.org/wiki/Exponential_backoff

[advanced example]: https://pkg.go.dev/github.com/cenkalti/backoff/v4?tab=doc#pkg-examples
