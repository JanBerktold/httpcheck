# httpcheck
[godoc](https://godoc.org/github.com/JanBerktold/httpcheck)

This is a fork of [httpcheck](github.com/ivpusic/httpcheck)

Contrary to the original package, it always persists cookies and does not start a "real" HTTP server but instead hooks into the handler and passes the request object directly; this allows for minimizedtest times as well as not dealing with the hassle of opening a socket on the testing machine.

## Contribution Guidelines
- Implement fix/feature
- Write tests for fix/feature
- Make sure all tests are passing
- Send Pull Request

# License
*MIT*
